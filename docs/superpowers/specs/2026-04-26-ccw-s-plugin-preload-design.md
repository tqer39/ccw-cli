# 2026-04-26 — `ccw -s` 起動時 superpowers skill 未ロード問題の解消

## 背景

`ccw -s`（PR #64 で再追加された「preamble 注入専用」フラグ）は次のように `claude` を起動する:

```text
claude --permission-mode auto --worktree <name> -n <name> [passthrough...] -- "<preamble>"
```

`<preamble>` は「`superpowers:brainstorming` → `writing-plans` → `executing-plans` の順で進めて」という指示文。`.claude/settings.json` には `enabledPlugins.superpowers@claude-plugins-official: true` が宣言済みのため、`claude` は起動時に当該プラグインを読み込むはずである。

しかし実測で次の事象が再現する:

- セッションの最初のターンでは superpowers 系 skill (`superpowers:brainstorming` 等) が **skill レジストリ未登録** の状態
- preamble は届いているので Claude は brainstorming を呼びにいくが、`Unknown skill: superpowers:brainstorming` で失敗する
- ユーザーが `/reload-plugins` を手動実行すると以降は使える

upstream 調査（claude-code-guide エージェント）の結論:

- `--` 経由の preamble はテキスト扱い。`/reload-plugins` を含めても slash command としては実行されない
- SessionStart hook は shell コマンドのみ実行可能で、Claude Code に slash command を送る公式 API は存在しない
- 関連 issue: [anthropics/claude-code#52967](https://github.com/anthropics/claude-code/issues/52967)（`/reload-plugins` のインプロセス化 FR）、[#53438](https://github.com/anthropics/claude-code/issues/53438)（plugin slash commands が起動時に budget trim で落ちる既知バグ）、[#9716](https://github.com/anthropics/claude-code/issues/9716)（skills が起動直後に認識されない）
- `claude` CLI には `--plugin-dir <path>`（セッション限定でプラグインディレクトリを明示注入、repeatable）が存在する

## ゴール

- `ccw -s` 起動直後の最初のターンから `superpowers:*` skill を使える状態にする
- ユーザーが `/reload-plugins` を手で打つ必要をなくす
- 既存挙動（preamble 注入、`-n` 含意、passthrough）は維持する

## 非ゴール

- superpowers 以外のプラグインへの汎用化（`-s` は superpowers 専用フラグなので、まず superpowers に閉じて解決する。汎用化は別タスク）
- upstream Claude Code への issue 提起や仕様変更を待つ姿勢（先回りで動く対策を入れる）
- `--plugin-dir` で完全に解決しなかった場合の追加リトライ機構

## 検討した選択肢

- **A1. `--plugin-dir` で superpowers cache path を明示注入** — `claude --help` に存在する公式フラグ。セッション限定でプラグインを load できる
- **A2. `claude --print "/reload-plugins"` 事前ウォームアップ** — 実測で `--print` モードでは slash command が反応せず、また別プロセスのプラグイン状態が次セッションに継承される根拠もない。**棄却**
- **A3. SessionStart hook で `/reload-plugins` を発火** — hook は shell のみ実行可能で slash command を送る手段がない。**棄却**
- **A4. preamble だけで済ませる（半自動）** — preamble に「skill が無ければ `/reload-plugins` を打って」と書く。ユーザー操作が残るため A（完全自動）の要件を満たさない。ただし A1 のフォールバックとして併用する価値あり

## 決定: A1（`--plugin-dir` 注入）+ preamble 自己修復文（ベルト＆サスペンダー）

### D1. `flags.Superpowers` 時に `--plugin-dir <path>` を `claude` に渡す

`cmd/ccw/main.go::run()` の `claude.LaunchNew(mainRepo, name, preamble, flags.Passthrough)` 呼び出し直前で、`flags.Superpowers == true` のときだけ superpowers プラグインの cache パスを解決し、解決できれば `--plugin-dir <path>` 2 要素を `flags.Passthrough` の **先頭** に prepend する。

`flags.Passthrough` は `claude.BuildNewArgs` 内で `extra...` として展開され、`-n <name>` の後・`--` の前に挿入される。`--plugin-dir` は claude のフラグ位置として正しい場所に並ぶ。

### D2. プラグインパス解決のカスケード

新規ファイル `internal/superpowers/plugindir.go` に `ResolvePluginDir() (string, bool)` を追加し、以下を **(a) → (b) → (c) の順** に試して最初に成功した時点で返す:

- **(a)** `~/.claude/plugins/installed_plugins.json` を読み、`superpowers` を含むキー（`superpowers@<marketplace>` 形式）の cache パスを取得。スキーマ未確定なため、JSON が存在しパースできて、それらしいパスフィールドが取れた場合のみヒット扱い
- **(b)** well-known パス `~/.claude/plugins/cache/claude-plugins-official/superpowers/latest` の `.claude-plugin/plugin.json` 存在確認
- **(c)** glob `~/.claude/plugins/cache/*/superpowers/latest/.claude-plugin/plugin.json` の最初のヒット。複数 marketplace が存在する場合は alphabetical で先頭

ヒットしたパスは `.../latest`（`.claude-plugin/` の親）を返す。`--plugin-dir` の引数として正しい形。

### D3. パス解決失敗時のフォールバック (i)

`ResolvePluginDir()` が `(_, false)` を返したら:

1. `ui.Warn` で警告表示（i18n 対応）:
   - JA: `superpowers プラグインのパスを解決できませんでした。skill が読み込まれていない場合は Claude Code 内で /reload-plugins を実行してください。`
   - EN: `Could not resolve the superpowers plugin path. If skills are not yet loaded, run /reload-plugins inside Claude Code.`
2. `--plugin-dir` は付けない
3. preamble は通常通り送信し、`claude` を起動する（`-s` の最低保証＝ preamble は届く、を維持）

### D4. preamble に自己修復行を追記

`internal/superpowers/preamble_ja.txt` と `preamble_en.txt` に以下の最終行を追記する:

- JA:

  ```text
  もし superpowers の skill がまだ読み込まれていない場合は、`/reload-plugins` を実行してから brainstorming を始めてください。
  ```

- EN:

  ```text
  If the superpowers skills are not yet loaded, run `/reload-plugins` first, then begin brainstorming.
  ```

`--plugin-dir` で解決できなかったエッジケース（パス解決失敗 + preamble だけ届くケース、または `--plugin-dir` を渡したのに何らかの理由で skill が認識されない upstream バグ #53438 系）でも、Claude が自己修復をユーザーに促せる。

### D5. パッケージ境界

`internal/superpowers/` パッケージに閉じる新規 API は次の 1 関数のみ:

```go
// ResolvePluginDir は superpowers プラグインの cache ディレクトリを返す。
// 見つからない場合は ("", false) を返し、呼び出し側はフォールバック処理を行う。
func ResolvePluginDir() (path string, ok bool)
```

`cmd/ccw/main.go` 側は `flags.Superpowers` 分岐内で `ResolvePluginDir()` を呼び、結果に応じて `flags.Passthrough` を組み立て直す。`internal/claude/` パッケージのシグネチャは変更しない。

## アーキテクチャへの影響

- `internal/superpowers/plugindir.go` 新規追加（パス解決ロジック）
- `internal/superpowers/preamble_*.txt` の文言追記
- `cmd/ccw/main.go::run()` で `flags.Superpowers` 時の passthrough 組み立てを変更
- `internal/claude/` には触らない
- `i18n` ロケールに警告メッセージキーを 1 つ追加

## 検証

- `go build ./...` および `go test ./...` が pass する
- `internal/superpowers/plugindir_test.go` の新規テスト:
  - (a) `installed_plugins.json` が読める場合のヒット
  - (b) well-known パスのみ存在する場合のヒット
  - (c) 別 marketplace 配下にしか存在しない場合の glob ヒット
  - 全て失敗する場合に `("", false)` を返す
  - テストは `t.TempDir()` + `HOME` 環境変数の差し替えでファイルシステムを隔離する
- `internal/superpowers/preamble_test.go` の既存テストを新文言に追従
- 手動検証:
  1. `ccw -s` を実行し、最初のターンで `superpowers:brainstorming` skill が利用可能であること（このバグの直接再現）
  2. `~/.claude/plugins/cache` を一時的にリネームし、警告が表示され、preamble は届くこと（フォールバック動作）
  3. `claude` 起動コマンドラインに `--plugin-dir <path>` が含まれることを `ps -ef` 等で目視確認

## 移行ノート

- 後方互換: 既存の `ccw -s` 利用者にとって挙動は改善のみ（破壊的変更なし）
- preamble 文言が変わるため、preamble の固定文字列に依存している外部ツール（あれば）には影響あり。ccw リポジトリ内では `internal/superpowers/preamble_test.go` のみが文字列を参照
- `internal/superpowers/plugindir.go` は `os.UserHomeDir()` に依存。CI 等で `HOME` 未設定の場合は (a)(b)(c) すべて失敗 → フォールバックパス。動作上の問題なし

## 関連

- 起点コミット: `bf75469` (PR #64) — `-s` の preamble 注入再追加
- 過去の関連設計: `2026-04-26-remove-s-flag-design.md`（`-s` 一旦廃止 → #64 で再導入されたが、本 spec はそれを前提とする）
- upstream issue: [#52967](https://github.com/anthropics/claude-code/issues/52967), [#53438](https://github.com/anthropics/claude-code/issues/53438), [#9716](https://github.com/anthropics/claude-code/issues/9716)
