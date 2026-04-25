# Worktree ↔ Claude Code Session Resume Integration — Design

- **Date**: 2026-04-25
- **Status**: Draft (awaiting user review)
- **Owner**: tqer39

## 背景

ccw-cli は `claude --worktree` を薄くラップして worktree 選択 UI を提供しているが、worktree を選んで `[r] run` した時の claude 起動は **常に fresh セッション**であり、過去会話を resume できない。README.ja.md:79-80 には「`-- --resume ID` のパススルーは非推奨」と明記されており、当時は worktree 作成と resume の同時指定が path ズレを起こすという理由で見送られていた。

しかし Claude Code v2.1.82 で「resume したセッションは元 worktree に自動で戻る」、v2.1.97 で「同 repo 内の別 worktree のセッションを直接 resume できる」と仕様が改善された。さらに `-n / --name <name>` で起動時にセッションへ表示名を付与でき、plan 承認による上書きも防がれる（`/rename` も同様）。

本仕様は **worktree 名 = session 名 の 1:1 マッピング** を ccw が裏で確立し、worktree picker からの選択でそのまま過去会話を resume できる体験を提供する。

## ゴール / 非ゴール

### ゴール

- ccw 経由で作成した worktree は **session 名が worktree 名と一致**する
- picker から既存 worktree を選択 → 自動で過去会話を resume（無ければ fresh 起動）
- picker 行に **resume 可否を明示**（`💬 RESUME` / `⚡ NEW`）
- README の旧警告（README.ja.md:79-80）を更新し、新動作と命名規約・関連 TIPS を記載
- picker footer にランダム TIPS を表示

### 非ゴール

- ユーザーが手動 `/rename` でセッション名を乖離させた場合の自動補正（何もしない）
- worktree のリネーム機能
- `claude --from-pr` 連携（独立した paths として、TIPS で紹介するに留める）
- 過去 session 検索 UI（Claude 標準 `/resume` に委ねる）

## 前提検証項目（実装初期に確認）

設計が成立するために確認が必要な事項:

1. `claude --worktree <name> -n <name>` のフラグ併用が正しく動作する
2. `claude --continue` の no-session 時の終了挙動（クリーンに非ゼロ exit するか、picker を開くか）
3. 最低 Claude Code バージョン: `2.1.118`（README 依存欄 `>= 2.1.49` から引き上げ）

不成立の場合は **E1 / E2** のフォールバックパスに従う。

## アーキテクチャ

### スコープ

| 領域 | 変更 |
|---|---|
| `internal/claude/` | フラグ生成（`-n <name>`、`--continue`）、リネーム |
| `internal/worktree/` | `Info.HasSession` 追加、マーカー管理パッケージ追加 |
| `internal/picker/` | L2 4 行レイアウト、`💬 RESUME` / `⚡ NEW` バッジ、footer TIPS |
| `internal/tips/`（新規）| ランダム TIPS 文字列の管理 |
| `cmd/ccw/main.go` | マーカー作成呼び出し、`--continue` フォールバック |
| `docs/README.md` / `docs/README.ja.md` | 旧警告撤去、新動作・命名規約・TIPS 追加 |

### マーカーファイル

ccw が「この worktree でセッションを起動済み」と知るためのマーカーを配置:

```text
<main-repo>/.git/worktrees/<worktree-name>/ccw-session-active
```

- 0 byte の空ファイル
- `git worktree remove` で worktree が削除される時、git 自身が `.git/worktrees/<name>/` ごと消すので **マーカーも自動削除**
- `.git/` 配下なので git 追跡対象外、`.gitignore` 更新不要
- ccw の領域内に閉じ、`~/.claude/` を読みに行かない（権限プロンプトを避ける）

### コンポーネント一覧

| パッケージ / ファイル | 変更内容 |
|---|---|
| `internal/claude/claude.go` | `BuildNewArgs(name, preamble, extra)` シグネチャ変更（name 引数追加）。`--worktree <name>` と `-n <name>` を生成。`BuildResumeArgs` を `BuildContinueArgs` にリネーム、`--continue` 付与。`Resume` を `Continue` にリネーム |
| `internal/worktree/marker.go`（新規）| `MarkSessionActive(repoRoot, name) error` / `HasSession(repoRoot, name) bool` |
| `internal/worktree/info.go`（既存）| `Info` 構造体に `HasSession bool` 追加 |
| `internal/picker/delegate.go` | L2 レイアウト実装。`Height()` を 4 に拡張。新ヘルパ `resumeBadge(hasSession bool) string` |
| `internal/picker/style.go` | `RESUME` / `NEW` バッジのスタイル定義（Lipgloss + NO_COLOR フォールバック） |
| `internal/picker/view.go` | footer のランダム TIPS 表示 |
| `internal/tips/tips.go`（新規）| TIPS 文字列スライス + `PickRandom(seed int64) string` |
| `cmd/ccw/main.go` | 新規 worktree 起動時 / 既存 run 時に `MarkSessionActive` を呼ぶ。`Continue` 失敗時は `LaunchNew(... -n <name>)` フォールバック |
| `docs/README.md` / `docs/README.ja.md` | 旧 `--resume` 警告（README.ja.md:79-80）を撤去。新動作・命名規約・`claude --from-pr` の TIPS を追加 |

### 依存関係

- `picker` → `worktree`（HasSession）→ ファイル存在確認のみ
- `claude` パッケージは worktree 名を引数で受け取る純粋関数として保つ
- `tips` は他に依存しない単独パッケージ

## データフロー

### A. 新規 worktree 作成

注: `claude` 起動は blocking。マーカー作成は claude exit 後に行う（worktree 作成は claude 自身が `--worktree` で行うため、exec 前には `.git/worktrees/<name>/` が存在しない）。

```text
1. ccw: 名前を決定（user 指定 or 自動生成）
2. ccw: exec claude --worktree <name> -n <name> [extra]   ← blocking
   (この間 claude が .claude/worktrees/<name>/ 作成、branch worktree-<name>
    を生成、対話セッションを開始)
3. claude が exit、ccw に制御が戻る
4. ccw: worktree.MarkSessionActive(repoRoot, name)
   → touch <main-repo>/.git/worktrees/<name>/ccw-session-active
5. ccw: claude の exit code を返して終了
```

トレードオフ: 長時間セッション中に別ターミナルで ccw picker を開くと、当該 worktree は `⚡ NEW` 表示になる（マーカー未作成のため）。セッション終了後の次回起動からは正しく `💬 RESUME` 表示される。許容する。

### B. 既存 worktree を選択 → `[r] run`

注: `.git/worktrees/<name>/` は既に存在するので、マーカー作成は exec 前に行える。

```text
1. ccw picker: worktree 一覧表示（HasSession で RESUME/NEW バッジ判定）
2. user が worktree 選択 → [r] run
3. ccw: cd <worktree-path>
4. if HasSession(repoRoot, name):
     ccw: exec claude --continue
       → 即時非ゼロ exit なら ccw: MarkSessionActive + exec claude -n <name> で再試行
   else:
     ccw: MarkSessionActive(repoRoot, name)        ← exec 前に作成
     ccw: exec claude -n <name>
5. claude が exit、ccw 終了
```

### C. picker 表示時の HasSession 判定

```text
1. ccw: git worktree list で worktree 一覧取得
2. 各 worktree について:
   marker_path = <main-repo>/.git/worktrees/<name>/ccw-session-active
   info.HasSession = file_exists(marker_path)
3. delegate.Render: HasSession に応じて 💬 RESUME / ⚡ NEW を描画
```

### D. TIPS 表示

```text
1. ccw 起動時: tips.PickRandom() で 1 件選択（seed: time.Now().UnixNano()）
2. picker.View() の footer に gh ヒントと並列表示
   - gh 不在時:  "💡 Install gh to see PR titles here"
   - gh 利用可:  "💡 Tip: <ランダム TIPS>"
3. session 中は同じ TIPS を表示（毎レンダで再選択しない）
```

TIPS 候補（初期セット）:

- `worktree 名 = session 名。手で /rename しても ccw は何もしません`
- `claude --from-pr <番号> で PR 連携セッションを直接 resume できます`
- `--clean-all で push 済 worktree を一括削除`
- `ccw -- --model <id> で claude にフラグを素通し`

### E. worktree 削除

```text
1. user: [d] delete or --clean-all
2. ccw: git worktree remove <path>
3. git: <main-repo>/.git/worktrees/<name>/ ディレクトリごと削除
   → ccw-session-active マーカーも自動削除
4. picker 次回起動時: HasSession=false（worktree 自体が消えてるので影響なし）
```

## TUI レイアウト（L2: 4 行ラベル付き + resume 強調）

### 表示例

```text
> 💬 RESUME · foo                            [PUSHED]  ↑0 ↓0
    branch:  feature/auth
    pr:      [OPEN] #123 "feat: add auth"
    path:    ~/.claude/worktrees/foo

  ⚡ NEW · bugfix-x                           [LOCAL]   ↑2 ↓0
    branch:  bugfix-x
    pr:      (no PR)
    path:    ~/.claude/worktrees/bugfix-x

  💬 RESUME · experiment                      [DIRTY]   ↑0 ↓0 ✎3
    branch:  worktree-experiment
    pr:      [DRAFT] #99 "wip: explore"
    path:    ~/.claude/worktrees/experiment
```

### スタイル指針

- `💬 RESUME` → 背景緑/シアン塗りの強調バッジ。RESUME 可能行は全体が「目に入る」
- `⚡ NEW` → 控えめグレー
- worktree 名（= session 名）が RESUME バッジ直後に並ぶ（1:1 マッピングを視覚化）
- status badge / indicators は右寄せ
- NO_COLOR モードでは `[RESUME]` / `[NEW]` の括弧バッジに退化
- 4 行 × 件数で長くなる。極狭端末（< 60 cols）でラベル崩れの可能性は許容（YAGNI）

## エラーハンドリング / エッジケース

### E1. `--worktree` と `-n` のフラグ併用が動かない

- 起動時 `claude --version` 確認、最低 `2.1.118` を要求
- 併用不可と判明したら `-n` を諦め、claude 起動後にユーザーへ「`/rename <name>` してください」とヒント表示

### E2. `claude --continue` がセッション無しで挙動不明

- `HasSession()` 事前判定で原則ヒットせず
- マーカー有り＋実セッション無し時、claude が即時非ゼロ exit すれば `claude -n <name>` で再試行
- claude が picker を開いた場合は user に委ねる（自動 fallback しない）

### E3. マーカーファイル作成失敗

- ログに WARN 出して続行（claude 起動は妨げない）
- 次回 picker で `⚡ NEW` 表示になるだけ（致命的でない）
- `CCW_DEBUG=1` で詳細ログ

### E4. main repo の `.git/worktrees/` 不在（bare repo / submodule 等）

- `gitx.MainRepoRoot()` 失敗時は既存ロジックでエラー終了
- マーカー機能は no-op、HasSession は常に false。機能劣化するが破綻しない

### E5. ユーザーが手動 `/rename foo` でセッション名乖離

- 何もしない（Q3 (a) 方針）
- 次回 ccw run: `--continue` は cwd 基準で最新セッションを resume → リネーム後の `foo` セッションが見つかり、復元可能
- `claude --resume <worktree-name>` は当然マッチしないが、ccw のフローでは使わない

### E6. 同 worktree 内に複数セッション

- `--continue` は最新を resume
- 古い側を resume したい場合: ユーザーは Claude 内 `/resume` で picker
- ccw は介入しない

### E7. worktree 名に特殊文字

- 既存の worktree 命名バリデーション再利用
- マーカーファイル名は固定文字列 `ccw-session-active` を使用

### E8. picker レンダリング width 不足

- 既存 `truncateToWidth` で各行を幅切り
- < 80 cols 端末でのレイアウト最適化は将来課題

### E9. plan mode 承認による名前上書き

- 公式仕様で `-n`/`/rename` 済みなら上書きされない
- 対応不要

## テスト

### 単体テスト

**`internal/claude/claude_test.go`**

- `BuildNewArgs("foo", "", nil)` → `["--permission-mode", "auto", "--worktree", "foo", "-n", "foo"]`
- `BuildNewArgs("foo", "preamble", []string{"--model", "x"})` → 期待引数列
- `BuildContinueArgs(nil)` → `["--permission-mode", "auto", "--continue"]`
- `BuildContinueArgs([]string{"--debug"})` → `--continue` の前後でフラグ位置確認
- 名前に空白を含む場合の引数化

**`internal/worktree/marker_test.go`（新規）**

- `MarkSessionActive` で `<repoRoot>/.git/worktrees/<name>/ccw-session-active` が作られる
- 存在しない `.git/worktrees/<name>/` への書き込みは error 返却
- `HasSession` が file_exists を反映
- bare repo / `.git` がファイルの worktree からの呼び出し対応

**`internal/picker/delegate_test.go`**

- L2 4 行レイアウト: meta / branch / pr / path 各行が期待文字列
- `HasSession=true` → `💬 RESUME` バッジ
- `HasSession=false` → `⚡ NEW` バッジ
- NO_COLOR モード: `[RESUME]` / `[NEW]` 括弧バッジ
- selected 時の prefix `>` 付与
- width 不足時の truncate

**`internal/picker/style_test.go`**

- RESUME / NEW スタイルの NO_COLOR 切替

**`internal/tips/tips_test.go`（新規）**

- `PickRandom()` が候補配列から 1 件返す
- 空配列でも panic しない
- seed 指定で決定的な結果

**`internal/picker/view_test.go`**

- footer に gh ヒント / TIPS が排他表示
- gh 不在時 → gh ヒント、gh 利用可 → TIPS

### 統合テスト

`tests/resume_flow_test.go`（新規）— `claude` を fake バイナリでスタブ:

- 新規 worktree 作成 → fake claude が `-n <name>` を受け取り、ccw がマーカー作成
- 既存 worktree で run → fake claude が `--continue` を受け取る
- マーカー無し worktree で run → fake claude が `-n <name>` を受け取る
- `--continue` が exit 1 → ccw が `-n <name>` で再試行
- worktree 削除 → マーカーが自動削除

### 手動検証チェックリスト（PR に記載）

- [ ] `claude --worktree foo -n foo` が実機で動作
- [ ] 同 worktree で 2 回目に `claude --continue` が前回会話を復元
- [ ] picker で RESUME / NEW バッジが正しく分かれる
- [ ] `/rename` 後も `--continue` で復元できる（E5）
- [ ] `git worktree remove` でマーカーも消える
- [ ] NO_COLOR=1 で表示崩れなし
- [ ] 80 cols 端末で L2 が読める
- [ ] CCW_DEBUG=1 でマーカー作成ログが出る
- [ ] TIPS が起動ごとに変わる

### CI

既存の `go test ./...` / `go vet ./...` / lefthook pre-commit に乗る。新規依存なし。

## ドキュメント変更

### `docs/README.ja.md` / `README.md`

**削除**: 旧 `-- --resume ID` パススルー非推奨警告（README.ja.md:79-80 相当）

**追加**:

- 「ccw は worktree 作成時に session 名を worktree 名と同期する」旨の説明
- picker の RESUME / NEW バッジ説明（既存の status / PR バッジ表と並列）
- 命名規約: `<フォルダ名> = <worktree 名> = <session 名>`、ブランチ名は `worktree-<name>`、PR タイトルは独立軸
- TIPS 例: `claude --from-pr <番号>` で PR 起点 resume も可能
- 依存欄の Claude Code 最低バージョンを `>= 2.1.118` に引き上げ

## 公開仕様への影響（破壊的変更）

- `--worktree` の挙動: 表面上同じ。内部で `-n <name>` が追加されるだけ
- `[r] run` の挙動: **fresh 起動から resume へデフォルト変更**。ユーザーが意図せず過去会話に戻る可能性 → README で明示
- 環境変数: 変更なし
- 終了コード: 変更なし

## 参考

- Claude Code 公式: [Resume previous conversations](https://code.claude.com/docs/en/common-workflows#resume-previous-conversations)
- CHANGELOG 該当エントリ: v2.1.82 / v2.1.97 / v2.1.101 / v2.1.115 / v2.1.118
- README.ja.md:79-80 の旧警告（本仕様で撤去対象）
