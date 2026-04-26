# 2026-04-26 — `-s` superpowers flag の廃止と `.claude/settings.json` への移行

## 背景

`ccw -s` / `--superpowers` は次の 3 つを 1 つの flag に詰め込んでいる:

1. `superpowers` プラグインが未導入なら `claude plugin install` を呼んで導入する（`internal/superpowers/detect.go::EnsureInstalled`）
2. 固定文字列のプリアンブル（`internal/superpowers/preamble.txt`、日本語 3 行）を `claude -- "<text>"` で初回プロンプトとして渡す
3. `--new` を含意する

検討の結果、これらは Claude Code の標準機能とほぼ重複していることが判明した:

| `-s` の機能 | Claude Code 標準機能 |
|---|---|
| 初期プロンプトを渡す | `claude -- "..."`（`ccw` も `--` 以降を passthrough 済み） |
| プリアンブルでフローを案内 | `superpowers:using-superpowers` skill がセッション開始時に自動発火し同等内容を主張する |
| プラグイン導入 | `.claude/settings.json` の `enabledPlugins` 宣言 → Claude Code が初回起動時に導入を促す |
| `--new` 含意 | `ccw -n` |

`ccw` は README で「a thin launcher for Claude Code's `--worktree`」と謳っており、プロンプト本文の組み立てはスコープ外である。

## ゴール

- `-s` / `--superpowers` flag を廃止し、関連コードを削除する。
- 開発者が ccw-cli を clone した時点で superpowers プラグインの導入を促されるよう、リポジトリに `.claude/settings.json` をコミットする。

## 非ゴール

- 任意プリセットや `--message` 系の汎用初期プロンプト flag の導入。
- `-s` の deprecation 期間（warning だけ出して機能継続）の提供。今回はクリーンカットで削除する。
- `superpowers` プラグイン本体や `using-superpowers` skill の改変。

## 決定事項

### D1. `-s` は即時削除（クリーンカット）

`-s` / `--superpowers` を渡すと spf13/pflag の標準 "unknown flag" エラーで弾かれる。追加のガイダンスメッセージは入れない（YAGNI）。

ccw は v0.x、Homebrew tap で配布中だが、deprecation 期間を設けるほどの利用規模ではない。CHANGELOG / リリースノートで breaking change として明記する。

### D2. `internal/superpowers/` パッケージは丸ごと削除

`preamble.go` / `preamble.txt` / `detect.go` と各 `_test.go` を削除。`cmd/ccw/main.go` の import と `maybeSuperpowers` 関数、呼び出しも削除。

`EnsureInstalled` の自動インストールロジックは Claude Code 標準の `enabledPlugins` 経由の prompt に委譲する。

### D3. `internal/claude/claude.go` の `preamble` 引数を撤去

`BuildNewArgs(name, preamble, extra)` / `BuildInWorktreeArgs(name, preamble, extra)` から `preamble` を削除。preamble が空でない場合に `--` を挿入していたコードパスも削除。passthrough（ユーザー指定の `--` 以降）の取り扱いはそのまま残す。

呼び出し元 `LaunchNew` / `LaunchInWorktree` および `cmd/ccw/main.go` から実引数を取り除く。テストも追従。

### D4. `.claude/settings.json` をコミット

内容:

```json
{
  "enabledPlugins": {
    "superpowers@claude-plugins-official": true
  }
}
```

`claude-plugins-official` は Claude Code が起動時に自動登録するため、`extraKnownMarketplaces` の追加は不要。開発者は ccw-cli フォルダを trust した時点で「このリポジトリは superpowers を要求しています」と促される。

### D5. `.gitignore` に `.claude/settings.local.json` を追加

ローカル設定（個人の hooks / permissions など）を誤コミットしないよう、既存の `.claude/worktrees/` の隣に `.claude/settings.local.json` を明示的に除外する。

### D6. README を更新

`README.md`（英語）と `docs/README.ja.md` の両方で、現状 `-s` を参照している以下 3 箇所をそれぞれ更新する（行番号は 2026-04-26 時点）:

- `:36` Features の "Design first startup" 箇条書き — 削除する（superpowers の自動発火は plugin 側の責務であり ccw 固有機能ではないため）。
- `:48` Quick Start のコマンド例 `ccw -s` — 削除する。
- `:118` Optional dependency の `-s` 言及 — 「`.claude/settings.json` で `enabledPlugins` 宣言済み。`claude` 起動時に未導入なら導入を促される」旨に書き換える。

`readme-sync` skill を使って両 README を同期更新する。

## アーキテクチャへの影響

`internal/superpowers/` の削除と `internal/claude/` の signature 変更により、`cmd/ccw/main.go` の起動フローは preamble に関する分岐が消え単純化される。他の internal パッケージ（`gitx`, `gh`, `picker`, `worktree`, `i18n` など）には影響しない。

## 検証

- `go build ./...` および `go test ./...` が pass する。
- `internal/cli/parse_test.go` から `-s` 関連 4 ケース（"short superpowers implies new", "long superpowers implies new", "new and superpowers combined", "superpowers with passthrough"）が削除され、残りのテストが通る。
- `ccw -s` を実行すると pflag の unknown flag エラーで終了する。
- `ccw` を ccw-cli 内で起動し、Claude Code が superpowers 未導入時にインストール prompt を出すことを目視確認する（このリポジトリで開発する開発者に対しての主要シナリオ）。

## 移行ノート

破壊的変更（breaking change）。リリース時に CHANGELOG とリリースノートで以下を明記:

- `-s` / `--superpowers` は削除された。
- 開発者は `claude` 起動時に促される `enabledPlugins` の prompt に従って superpowers をインストールするか、`/plugin install superpowers@claude-plugins-official` を手動実行する。
- 初期プロンプトを渡したい場合は `ccw -n -- "<text>"` を使う（既存機能、変更なし）。
