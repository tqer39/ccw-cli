# 2026-04-25 — 決定論的な worktree / セッション名（`ccw-<owner>-<repo>-<shorthash6>`）

## 背景

現状 `internal/namegen.Generate()` は `quick-falcon-7bd2` のようなランダム slug を返し、それが「worktree ディレクトリ名 = `claude -n` セッション名」に使われている（`cmd/ccw/main.go:75,98`）。Claude Code 側の picker や履歴一覧でセッション名だけを見たとき、**どのリポジトリのセッションか** が判別できない。複数のリポジトリで ccw を併用するユーザーにとって認知負荷が高い。

## ゴール

worktree 名 / セッション名から「どのリポジトリの、どの base commit から切ったセッションか」が一目で分かるようにする。

具体的には:

```text
ccw-<owner>-<repo>-<shorthash6>
```

例:

| 状況 | 生成名 |
|---|---|
| `tqer39/ccw-cli`, ローカル `main` tip = `a3f2b1c…` | `ccw-tqer39-ccw-cli-a3f2b1` |
| `Anthropic/claude-code`, tip = `9f8e7d6…` | `ccw-anthropic-claude-code-9f8e7d` |
| `origin` 未設定（ローカル専用 repo） | `ccw-local-<basename>-<shorthash6>` |

## 非ゴール

- 既存の `quick-falcon-7bd2` 形式 worktree のマイグレーション（後方互換のため新規生成のみ切り替え）
- ブランチ名（`worktree-<name>`）の短縮 — claude 側の自動生成挙動なので別 issue
- `claude --worktree` の名前検証ロジックへの介入

## 要件

### 名前構成

```text
ccw-<owner>-<repo>-<shorthash6>
```

- 接頭辞 `ccw-` は固定。「ccw が作ったセッション」と一目で分かるため
- `<owner>` / `<repo>`: `git remote get-url origin` の URL を parse して抽出
  - SSH 形式 (`git@github.com:owner/repo.git`) と HTTPS 形式 (`https://github.com/owner/repo.git`) の両方に対応
  - `.git` サフィックスは除去
  - GitLab の nested group (`group/subgroup/repo`) は `subgroup` を採用（最後から 2 番目）。owner 階層を全部入れると長すぎるため
- `<shorthash6>`: ローカルの default branch tip の short SHA（6 文字）
  - default branch の解決優先順:
    1. `git symbolic-ref refs/remotes/origin/HEAD`（例: `refs/remotes/origin/main`）から末尾を抽出
    2. fallback: `main` → `master` → 取得できなければエラー
  - `git rev-parse --short=6 <default_branch>` で取得

### 正規化

`<owner>` と `<repo>` 単位で以下を適用 (順番):

1. `.git` サフィックス除去
2. 全文字を ASCII lowercase 化
3. `[a-z0-9-]` 以外の文字をすべて `-` に置換
4. 連続 `-` を 1 個に圧縮
5. 先頭・末尾の `-` を trim

これにより `Anthropic/claude-code.git` → `anthropic` / `claude-code`、`my org/my repo!` → `my-org` / `my-repo` のように安定化。

### 衝突回避

同じ default branch tip から複数 worktree を切ると同名になる。検知して末尾に `-N` を付ける:

- 1 個目: `ccw-tqer39-ccw-cli-a3f2b1`
- 2 個目: `ccw-tqer39-ccw-cli-a3f2b1-2`
- 3 個目: `ccw-tqer39-ccw-cli-a3f2b1-3`

検知条件: 候補名 `<name>` について、以下のどちらかに該当すれば衝突とみなす:

- `<mainRepo>/.claude/worktrees/<name>` がディレクトリとして存在する
- `git worktree list --porcelain` の `worktree` 行のうち、basename が `<name>` と一致するエントリが存在する

`N` は 2 から開始、上限 99（実用上ほぼ届かない）。99 を超えたらエラー。

ランダム fallback は採用しない（決定論性を優先）。

### `origin` 未設定時の fallback

`git remote get-url origin` が失敗 or 空の場合:

```text
ccw-local-<basename>-<shorthash6>
```

- `<basename>` は `filepath.Base(mainRepo)` を上記正規化ルールで処理した値
- shorthash の取得方法は通常時と同じ（local の default branch を参照）

`local` という固定 owner を入れることで、origin あり版とのフォーマット一貫性を保つ。

### shorthash 取得失敗時の挙動

以下のいずれかに該当する場合は **エラーで止める**（`ccw -n` または picker の `[+ new]` の結果として表示）:

- default branch が `main` / `master` のどれにも該当しない、かつ `origin/HEAD` も未設定
- `rev-parse --short=6` が失敗（commit が一つも無い空 repo 等）

ユーザーには「`git remote set-head origin -a` を実行するか、`main` ブランチを作成してください」と案内する。

## 設計

### パッケージ構成

| パッケージ | 役割 | 変更内容 |
|---|---|---|
| `internal/namegen` | 名前生成 | `Generate()` のシグネチャ変更、ロジック全置換 |
| `internal/gitx` | git コマンド薄ラッパ | `OriginURL`, `DefaultBranch`, `ShortHash` を追加 |
| `cmd/ccw` | エントリポイント | `namegen.Generate(mainRepo)` 呼び出し 2 箇所をエラー対応に |

### `internal/namegen` の新 API

```go
// Generate returns "ccw-<owner>-<repo>-<shorthash6>" with collision suffixing.
// Returns an error when default branch / origin cannot be resolved.
func Generate(mainRepo string) (string, error)
```

内部関数を 2 つに分解してテスタビリティを確保:

```go
// pure function — no git/file I/O, easy to table-test
func buildName(owner, repo, shorthash string, takenNames map[string]bool) (string, error)

// normalize applies the canonical lowercase / dash-collapse rules.
func normalize(s string) string
```

### `internal/gitx` の追加 API

```go
// OriginURL returns the origin remote URL or "" when not set.
func OriginURL(mainRepo string) (string, error)

// DefaultBranch returns the default branch name (e.g. "main").
// Resolution order: refs/remotes/origin/HEAD → "main" → "master".
// Returns an error when none exist.
func DefaultBranch(mainRepo string) (string, error)

// ShortHash returns `git rev-parse --short=<n> <ref>` output (trimmed).
func ShortHash(mainRepo, ref string, length int) (string, error)
```

URL parse 用に `internal/gitx/url.go` 相当のヘルパ（owner/repo 抽出）をパッケージ内に置く（gitx は git CLI ラッパなので、URL parse はここに置くのが自然）。

```go
// ParseOriginURL extracts (owner, repo) from an origin URL.
// Supports SSH and HTTPS forms; nested groups collapse to the last two segments.
func ParseOriginURL(url string) (owner, repo string, err error)
```

### `cmd/ccw/main.go` の差分

```go
// before
name := namegen.Generate()

// after
name, err := namegen.Generate(mainRepo)
if err != nil {
    ui.Error("generate worktree name: %v", err)
    return 1
}
```

`flags.NewWorktree` パスと picker の `ActionNew` パスの 2 箇所。

### 既存 worktree との互換

picker (`internal/picker`, `internal/worktree.List`) はディレクトリ名を文字列として扱うだけなので、旧形式 (`quick-falcon-7bd2`) と新形式が同居しても何もしなくて良い。`worktree.HasSession` も path 基準。マイグレーションは行わない。

### ブランチ名

`claude --worktree <name>` が `worktree-<name>` を自動生成する claude 側の挙動はそのまま。結果として:

```text
worktree-ccw-tqer39-ccw-cli-a3f2b1
```

長いが、識別性のメリットを優先。短縮要望が出たら別 issue。

## テスト

### `internal/namegen` (unit)

`buildName` と `normalize` は純粋関数なので table-driven test を充実させる:

- `normalize` (segment 単位、slash を含まない入力前提):
  - `Anthropic` → `anthropic`
  - `My Org` → `my-org`
  - `_underscore_` → `underscore`
  - `--double--dash--` → `double-dash`
  - `repo.git` → `repo`（`.git` は `ParseOriginURL` で除去済み前提だが、normalize 単独でも `.` は `-` 置換 → 連続圧縮 → trim で消える）
  - 空文字 → 空文字
- `buildName`:
  - `("tqer39", "ccw-cli", "a3f2b1", {})` → `"ccw-tqer39-ccw-cli-a3f2b1"`
  - 衝突 1 個 → `-2` 付与
  - 衝突 2 個 → `-3` 付与
  - 衝突 99 個まで → `-99` 付与、100 でエラー
- `Generate` (integration): `gitx` 関数を関数値で差し替え可能にして hermetic に
  - happy path
  - `origin` 無し → `local` fallback
  - default branch 解決失敗 → エラー

### `internal/gitx` (integration)

実 git に対する temp repo テストで:

- `OriginURL`: SSH / HTTPS / 未設定 / `.git` 付き / `.git` 無し
- `DefaultBranch`: `origin/HEAD` あり / `main` のみ / `master` のみ / どれもない (error)
- `ShortHash`: 6 文字、commit 無し時のエラー
- `ParseOriginURL`:
  - `git@github.com:tqer39/ccw-cli.git` → `("tqer39", "ccw-cli", nil)`
  - `https://github.com/tqer39/ccw-cli` → `("tqer39", "ccw-cli", nil)`
  - `https://gitlab.com/group/sub/repo.git` → `("sub", "repo", nil)`
  - 不正な URL → error

### `cmd/ccw` (smoke)

main.go レベルでの統合テストはこれまで通り（既存のテストハーネスを使う）。`Generate` のエラー伝播が正しく ui.Error → exit 1 に乗ることを確認。

## ドキュメント変更

`README.md` / `docs/README.ja.md` の "Naming convention" セクション:

- 旧: `<name>` is generated like `quick-falcon-7bd2`.
- 新: `<name>` is generated as `ccw-<owner>-<repo>-<shorthash6>` (e.g. `ccw-tqer39-ccw-cli-a3f2b1`). owner/repo come from the `origin` remote; shorthash is the 6-char short SHA of the local default branch tip at creation time.

両ファイル同期は `readme-sync` skill 経由で確認。

## ロールバック

`internal/namegen` 内の変更だけで完結する。リバートは PR revert で十分。既存 worktree は無影響。

## PR スコープ

この spec は **単独 PR** 用。実装プランは `docs/superpowers/plans/` 配下に別途作成。
