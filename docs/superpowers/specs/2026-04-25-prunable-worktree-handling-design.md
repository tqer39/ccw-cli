# prunable worktree 対応 設計書

- 作成日: 2026-04-25
- 対象: ccw-cli
- 関連 issue: なし (ユーザー報告)

## 背景

`git worktree list --porcelain` の出力には、ディレクトリ実体が消失した worktree が `prunable` 行付きで残ることがある。

例:

```text
worktree /path/to/.claude/worktrees/synchronous-finding-wave
HEAD 39f1982156e73fd92b36eb4226229c074b95db5e
branch refs/heads/worktree-synchronous-finding-wave
prunable gitdir file points to non-existent location
```

現在の `internal/gitx/worktree.go` の `ParsePorcelain` は `prunable` 行を読まないため、このエントリが通常 worktree として後段に渡る。`internal/worktree/worktree.go` の `List` は ccw 管理パス (`/.claude/worktrees/`) フィルタを通った後 `Classify` を呼び、`Classify` は存在しないパスで `git -C <missing> status --porcelain` を実行する。git は exit 128 を返し、ccw は次のメッセージで起動失敗する:

```text
✖ list worktrees: classify status: git status --porcelain: exit status 128
```

## ゴール

1. prunable な worktree が混在しても ccw が起動し、worktree 一覧を表示できる。
2. prunable な worktree を picker 上で識別でき、ユーザーが ccw 内から掃除できる。
3. 既存の通常 worktree の挙動 (status / remove / picker 表示) を変えない。

## 非ゴール

- ccw 起動時の自動 prune (破壊的操作の暗黙実行は避ける)。
- prunable な worktree の admin ファイルを git 経由でなく直接削除する。
- `git worktree repair` への対応 (現時点では掃除のみで十分)。

## 設計

### ステータス分類

`internal/worktree.Status` に `StatusPrunable` を追加する。

```go
const (
    StatusPushed Status = iota
    StatusLocalOnly
    StatusDirty
    StatusPrunable
)
```

`String()` は `"prunable"` を返す (git の用語をそのまま採用)。picker のラベルも `prunable`。

### gitx 層

`internal/gitx/worktree.go`:

- `WorktreeEntry` に `Prunable bool` を追加。
- `ParsePorcelain` で `prunable` で始まる行 (理由文字列を含む) を読み、`cur.Prunable = true` を立てる。
- 新規関数 `Prune(mainRepo string) error` を追加。`git -C <mainRepo> worktree prune` を `Run` 経由で呼ぶ。

### worktree 層

`internal/worktree/worktree.go`:

- `List` の各エントリ処理で、`e.Prunable == true` の場合は `Classify` を呼ばずに以下を直接生成して append する:

  ```go
  Info{
      Path:   e.Path,
      Branch: e.Branch,
      Status: StatusPrunable,
      // AheadCount / BehindCount / DirtyCount は 0
      // HasSession は false (ディスク上にパスが無いので意味が無い)
  }
  ```

- 新規関数 `Prune(mainRepo string) error` を追加。`gitx.Prune(mainRepo)` をラップ。

### picker 層

`internal/picker/style.go`:

- `prunable` ラベル用のスタイルを追加 (色は実装時に既存配色との調和を確認しつつ決める。灰色系を想定)。

`internal/picker/delegate.go`:

- ステータス表示分岐に `worktree.StatusPrunable` の枝を追加し、上記スタイルでラベルをレンダリング。

`internal/picker/model.go` / `update.go`:

- **単独削除時**:
  - 選択行が `StatusPrunable` の場合、既存 `stateDeleteConfirm` ではなく新規 `statePruneConfirm` に遷移する (もしくは既存 confirm の中で prunable 分岐を入れる。実装時に既存コードの形に合わせて判断)。
  - prune confirm の表示は以下のように切り替える:
    - 全 worktree 中の prunable 件数が **1 件** のとき: `Run 'git worktree prune'? [y/N]`
    - **複数件** のとき: 「以下 N 件の prunable エントリがまとめて削除されます:」 + 該当パス一覧 + `[y/N]`
  - `y` 確定で `worktree.Prune(mainRepo)` を実行し、リストを再読込。
- **bulk 削除時**:
  - 既存 `stateBulkConfirm` の選択集合に prunable を含めることを許可する。
  - 確認画面の対象一覧には prunable 行も含めて表示する (件数表示にも反映)。
  - 確定後の処理:
    1. 通常 worktree (StatusPushed / StatusLocalOnly / StatusDirty) は既存どおり `worktree.Remove(mainRepo, path, force)` を順次実行。
    2. 選択集合に prunable が 1 件以上含まれていれば、最後に `worktree.Prune(mainRepo)` を 1 回実行。
  - `git worktree prune` は全 prunable をまとめて掃除するため、選択していない prunable も巻き込まれる可能性があるが、bulk 削除確認画面には全選択 prunable が並んでいる前提で許容する (将来必要なら警告文を追加)。

### 既存挙動の保護

- 通常 worktree の `Classify` / `Remove` / picker 表示ロジックは変更しない。
- `git worktree prune` は git 公式の安全コマンド (ロック中の worktree や非 prunable には影響しない)。
- 削除系アクションは必ず確認プロンプトを通すため、ユーザーの意図しない破壊的操作は発生しない。

## テスト計画

### `internal/gitx/worktree_test.go`

- `prunable gitdir file points to non-existent location` 行を含む porcelain サンプルを追加し、対応エントリの `Prunable == true`、他のエントリの `Prunable == false` を確認。
- 行の prefix のみ検査するので、`prunable` 単独行 (理由文字列なし) でも認識されることを確認。

### `internal/worktree/worktree_test.go`

- 既存の porcelain ベースのテストフィクスチャに prunable エントリを追加。
- `List` の戻り値で対応 `Info.Status == StatusPrunable` で、ディスクアクセス無しに (= 存在しないパスでも) 返ることを確認。
- ahead/behind/dirty/session が 0 / false であることを確認。

### `internal/picker/model_test.go`

- prunable 行のみを選択して delete キーを押下 → `statePruneConfirm` (またはそれに相当する状態) へ遷移。
- prunable が 1 件のときの確認文言、複数件のときの一覧表示。
- bulk 選択で normal + prunable の混在を許容し、確認画面の件数 / 一覧表示に prunable が含まれること。
- bulk 確定の処理経路で、normal は `Remove`、prunable があれば末尾で `Prune` が 1 回呼ばれることを (テスト用 fake / モックで) 確認。

### 手動受け入れテスト

- `~/workspace/tqer39/terraform-github` (再現環境) で `ccw` を起動し、起動成功 → picker に `prunable` 行が表示されることを確認。
- 単独削除 (1 件のみ prunable) → 確認 → 一覧から消えることを確認。
- 複数 prunable を意図的に作って単独削除 → 列挙された一覧で全件消えることを確認。

## 影響範囲

- 変更ファイル: `internal/gitx/worktree.go`, `internal/worktree/worktree.go`, `internal/picker/{style,delegate,update,model}.go` (および対応するテスト)。
- 公開 API への破壊的変更なし (`WorktreeEntry` / `Status` への追加のみ)。
- ドキュメント追記 (README の picker 説明) は別タスクで対応 (本 spec には含めない)。
