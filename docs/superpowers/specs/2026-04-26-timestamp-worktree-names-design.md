# 2026-04-26 — タイムスタンプベースの worktree / セッション名（`ccw-<owner>-<repo>-<yymmdd>-<hhmmss>`）

## 背景

直前 PR (#48) で `internal/namegen.Generate()` は `ccw-<owner>-<repo>-<shorthash6>` 形式の決定論的な名前を返すようになった（`quick-falcon-7bd2` のランダム slug から置換）。`<shorthash6>` はローカル default branch tip の short SHA。

ただし運用してみると、Claude Code 上で resume 候補や作業中セッション名を見たときに **「いつから作業しているセッションか」** が直感的に分からない。short SHA は default branch HEAD と緩く相関するだけで、人間が時間軸として読み取るのは難しい。

`namegen.Generate` は **新規作成時のみ** 呼ばれる（`cmd/ccw/main.go:75, 102`）。RESUME は picker が既存 worktree 一覧から選ばせる方式で、生成名の決定性には依存しない。したがって命名規則を時間ベースに置き換えても resume 機構は影響を受けない。

## ゴール

worktree 名 / セッション名から **「いつ始めた作業か」が一目で分かる** ようにする。

```text
ccw-<owner>-<repo>-<yymmdd>-<hhmmss>
```

例:

| 状況 | 生成名 |
|---|---|
| `tqer39/ccw-cli`, 2026-04-26 14:30:55 (local) に作成 | `ccw-tqer39-ccw-cli-260426-143055` |
| `Anthropic/claude-code`, 同上 | `ccw-anthropic-claude-code-260426-143055` |
| `origin` 未設定（ローカル専用 repo）, 同上 | `ccw-local-<basename>-260426-143055` |

## 非ゴール

- 既存の hash 形式 worktree（例: `ccw-tqer39-ccw-cli-9d3dc6`）のマイグレーション。picker は文字列としてディレクトリ名を扱うだけなので新旧混在で問題ない
- ブランチ名（`worktree-<name>`）の短縮 — claude 側の自動生成挙動は変えない
- タイムゾーンの設定可能化（ローカル固定）

## 要件

### 名前構成

```text
ccw-<owner>-<repo>-<yymmdd>-<hhmmss>
```

- 接頭辞 `ccw-` は固定（変更なし）
- `<owner>` / `<repo>`: 既存ロジックを完全に流用（`origin` URL parse → 正規化）
- `<yymmdd>-<hhmmss>`: **作成時刻のローカルタイム**を `time.Now().Format("060102-150405")` で文字列化
  - 例: 2026-04-26 14:30:55 (local) → `260426-143055`
  - ms（ミリ秒）は付けない。同一秒の連続作成は既存の `-N` suffix で吸収する

### タイムゾーン

ローカルタイム固定。理由:

- ccw はローカルターミナルから対話的に起動される CLI なので、ユーザーの体感時刻 = ローカル時刻
- UTC にすると「いつ始めた作業か」を逆算する手間が増える

### 衝突回避

既存の `-N` suffix ロジックをそのまま流用:

- 1 個目: `ccw-tqer39-ccw-cli-260426-143055`
- 2 個目: `ccw-tqer39-ccw-cli-260426-143055-2`
- 3 個目: `ccw-tqer39-ccw-cli-260426-143055-3`

検知条件・上限（99）も PR #48 から変更なし。同一秒に手動で複数回 `ccw -n` を叩く / 自動化テストで連続作成する場合の保険。

### `origin` 未設定時の fallback

既存ロジックを流用:

```text
ccw-local-<basename>-<yymmdd>-<hhmmss>
```

### 失敗モード

時刻取得は `time.Now()` で失敗しないため、新たなエラー経路は発生しない。

旧版で発生していた以下のエラーは消滅する:

- default branch 解決失敗
- `rev-parse --short=6` 失敗（空 repo 等）

これにより `cmd/ccw/main.go` のエラーメッセージから「`git remote set-head origin -a` を実行するか、`main` ブランチを作成してください」という hint も不要になる（メッセージ更新が必要）。

## 設計

### パッケージ構成

| パッケージ | 役割 | 変更内容 |
|---|---|---|
| `internal/namegen` | 名前生成 | `shortHashFn` / `defaultBranchFn` を削除、`nowFn = time.Now` を導入 |
| `internal/gitx` | git コマンド薄ラッパ | `ShortHash` を削除（namegen 以外で未使用） |
| `cmd/ccw` | エントリポイント | `Generate` のエラー hint メッセージを更新 |

### `internal/namegen` の変更

```go
// 追加
var nowFn = time.Now

// 削除
// var shortHashFn = gitx.ShortHash
// var defaultBranchFn = gitx.DefaultBranch
```

`Generate(mainRepo)` のシグネチャは変更なし。エラー戻り値も維持（owner/repo 解決と `takenNames` の I/O エラーは残る）。

```go
func Generate(mainRepo string) (string, error) {
    owner, repo, err := resolveOwnerRepo(mainRepo)
    if err != nil {
        return "", err
    }
    ts := nowFn().Format("060102-150405")
    taken, err := takenNames(mainRepo)
    if err != nil {
        return "", err
    }
    return buildName(owner, repo, ts, taken)
}
```

`buildName` のシグネチャは変えない（第 3 引数の意味だけが「shorthash」→「timestamp」に変わる）。引数名のみ `shorthash` → `tail` などに変更してドキュメンテーションを反映。エラーメッセージ中の `shorthash is empty` は文言を一般化（例: `tail is empty`）。

パッケージ doc コメント（`namegen.go:1-3`）を新フォーマットに合わせて更新。

### `internal/gitx` の変更

`branch.go` から `ShortHash` 関数を削除。`branch_test.go` の `TestShortHash_Length` / `TestShortHash_MissingRef` も削除。

`DefaultBranch` は他で使われている可能性があるため **触らない**（実装フェーズで grep 確認 → 他で未使用なら別 PR で削除を検討）。

### `cmd/ccw/main.go` の変更

`Generate` 失敗時の hint メッセージから「`git remote set-head origin -a` / main ブランチ作成」の文言を除去。残るエラー要因（origin URL parse 失敗、worktrees ディレクトリ読み込み失敗、衝突 99 件超過）に合った文言に差し替え。2 箇所（`flags.NewWorktree` パスと picker の `ActionNew` パス）両方を更新。

### 既存 worktree との互換

picker / `worktree.List` はディレクトリ名を文字列として扱うだけなので、旧形式 (`ccw-tqer39-ccw-cli-9d3dc6`) と新形式が同居しても問題ない。`takenNames` は両方を見るので衝突検出も問題なし。マイグレーション不要。

## テスト

### `internal/namegen` (unit)

`namegen_test.go` の差し替え:

- `shortHashFn` の差し替えを使っているテストを `nowFn` の差し替えに置換
  - 固定時刻 `time.Date(2026, 4, 26, 14, 30, 55, 0, time.Local)` → `260426-143055`
  - `nowFn` の差し替えヘルパ（save / restore パターン）を `t.Cleanup` で実装
- `TestGenerate_ShortHashError` を **削除**（時刻取得はエラーが起きない）
- `TestGenerate_DefaultBranchError` 相当があれば削除（同上）
- collision テスト（`-2`, `-3` 付与）はフォーマット変更に追従。固定 `nowFn` を返す前提で期待値を更新
- `buildName` の table-driven test は引数の意味変更に合わせて文字列だけ調整（`shorthash` → `timestamp` にネーミング）

### `internal/gitx`

`TestShortHash_*` を削除。他の影響なし。

### `cmd/ccw` (smoke)

既存テストハーネスのうち、エラー文言を assert している箇所があれば文言更新に追随。

## ドキュメント変更

- `README.md` / `docs/README.ja.md` の命名規則記述（"Naming convention" 周辺）を新フォーマット `ccw-<owner>-<repo>-<yymmdd>-<hhmmss>` に更新
  - 例: `ccw-tqer39-ccw-cli-260426-143055`
  - タイムゾーンはローカルである旨を明記
- 両ファイル同期は `readme-sync` skill 経由で確認

## ロールバック

`internal/namegen` の差分が中心。`gitx.ShortHash` 削除を含むので revert 時は import が戻る点だけ注意。既存 worktree は無影響。

## PR スコープ

この spec は **単独 PR** 用。実装プランは `docs/superpowers/plans/2026-04-26-timestamp-worktree-names.md` として別途作成。
