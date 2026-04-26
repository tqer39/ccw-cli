# 2026-04-26 — `ccw -L` 非対話 list モード（worktree × git × PR × session の集約）

## 背景

ccw は対話 picker に「worktree × git 状態 × PR 状態 × session 有無」の情報を一画面に集約している（`internal/picker`, `internal/worktree.List`）。一方この集約は対話 UI の中に閉じており、外部から再利用できない。

特に Claude Code 自身が ccw を呼んで「他にどの worktree があるか」「どれが掃除して良いか」「自分が今いる worktree の素性は何か」を知りたい場合、現状は `git worktree list` / `git status` / `gh pr list` / session log 探索を個別に呼ぶ必要がある。ccw 内に同等の集約があるのに再実装するのは無駄、かつ判定ロジックがズレる。

Claude Code 側の resume 機能はすでに cwd / worktree でフィルタされているので（`claude --resume` の picker、`Ctrl+W` で同リポジトリの全 worktree 拡張）、**ccw が補うべきは resume 候補ではなく「git/PR/session 情報の集約」** である。

## ゴール

ccw に **非対話 list モード** を追加し、ccw が管理する worktree 群について以下を機械可読に出力する:

- worktree の path / branch / status（pushed / local-only / dirty / prunable）
- ahead / behind / dirty フラグ
- PR 情報（gh が使える場合）
- session log の有無と log path
- 直近コミット情報、作成時刻、default branch

人間が直接叩いた場合の体験も table 形式で読みやすくする。

## 非ゴール

- 対話 picker の UI 変更（picker は無関係に温存）
- worktree の作成・削除・切替・cleanup（既存の `-n` / `--clean-all` で十分）
- main repo 自身（`.claude/worktrees/` 配下でない作業ツリー）の状態出力
- JSON スキーマ v2 への拡張余地の事前設計（YAGNI、`version: 1` で固定）
- セッションのフィルタリング（Claude Code 本体で完結）

## CLI

```text
ccw -L | --list  [-d <path>]  [--json]  [--no-pr]  [--no-session]
```

### フラグ

| フラグ | 役割 | 既定値 |
|---|---|---|
| `-L` / `--list` | 非対話 list モード起動。対話 picker を開かず stdout に出力して即終了 | (off) |
| `-d <path>` | 対象ディレクトリを明示指定。cwd の代わりに使う。`-L` 必須 | `.`（cwd） |
| `--json` | 出力形式を JSON に切替 | text |
| `--no-pr` | gh 呼び出しを skip。`pr` は全エントリで `null` | (off) |
| `--no-session` | session 探索を skip。`session.exists` は false 固定 | (off) |

### 既存フラグとの組合せ規則

- `-L` と `-n`（new）は相互排他。両方指定でエラー (exit 1)。
- `-L` と `-s`（superpowers preamble）は相互排他。両方指定でエラー。
- `-L` と `--clean-all` は相互排他。
- `-L` と `--` 以降のパススルーは相互排他（claude を起動しないため意味がない）。
- `-d` は `-L` 専用フラグとし、`-L` 不在時は警告ではなくエラー (exit 1) にする。将来 `-d` を picker / new でも使えるようにする際の混乱を避けるため、いまは「list 用」と狭く定義する。

### `-d <path>` の解決

- `<path>` は任意のディレクトリでよい。worktree の中、subdirectory、main repo root、`.claude/worktrees/` 直下のどれでも受け付ける。
- 内部では既存の cwd → main repo resolve ロジック（`internal/cli` 等で `gitx` を呼ぶ）を流用し、`<path>` を起点に main repo を逆引きする。
- `<path>` が git の作業ツリー外ならエラー (exit 1, stderr に "not a git repository: <path>")。
- 解決された main repo の `.claude/worktrees/` 配下が走査対象。

## 出力スキーマ（JSON）

```jsonc
{
  "version": 1,
  "repo": {
    "owner": "tqer39",
    "name": "ccw-cli",
    "default_branch": "main",
    "main_path": "/Users/foo/workspace/tqer39/ccw-cli"
  },
  "worktrees": [
    {
      "name": "ccw-tqer39-ccw-cli-9d3dc6-4",
      "path": "/Users/foo/.../.claude/worktrees/ccw-tqer39-ccw-cli-9d3dc6-4",
      "branch": "worktree-ccw-tqer39-ccw-cli-9d3dc6-4",
      "status": "pushed",
      "ahead": 0,
      "behind": 0,
      "dirty": false,
      "default_branch": "main",
      "created_at": "2026-04-26T04:28:00+09:00",
      "last_commit": {
        "sha": "9d3dc6e",
        "subject": "feat(namegen): deterministic worktree names",
        "time": "2026-04-25T23:14:00+09:00"
      },
      "pr": {
        "state": "OPEN",
        "number": 42,
        "url": "https://github.com/tqer39/ccw-cli/pull/42",
        "title": "feat: ..."
      },
      "session": {
        "exists": true,
        "log_path": "/Users/foo/.claude/projects/-Users-foo-.../<sessionid>.jsonl"
      }
    }
  ]
}
```

### フィールド仕様

トップレベル:

| キー | 型 | 説明 |
|---|---|---|
| `version` | `int` | スキーマバージョン。常に `1` |
| `repo.owner` | `string` | `origin` URL から抽出。未設定時は `"local"` |
| `repo.name` | `string` | 同上。未設定時はディレクトリ basename を正規化したもの |
| `repo.default_branch` | `string` | `origin/HEAD` → `main` → `master` の優先順で解決 |
| `repo.main_path` | `string` | main repo の絶対パス |
| `worktrees` | `array` | 0 件でも空配列。`null` ではなく `[]` を返す |

worktree 配列要素:

| キー | 型 | 説明 |
|---|---|---|
| `name` | `string` | worktree ディレクトリ basename |
| `path` | `string` | 絶対パス |
| `branch` | `string` | `git worktree list --porcelain` の `branch` 行（`refs/heads/` プレフィクス除去後） |
| `status` | `string` | enum: `"pushed"` \| `"local-only"` \| `"dirty"` \| `"prunable"` |
| `ahead` | `int` | upstream に対する先行コミット数。`status` が `prunable` のときは `0` |
| `behind` | `int` | upstream に対する遅延コミット数。同上 |
| `dirty` | `bool` | `status == "dirty"` と同義。冗長だが Claude が status enum を解釈しなくても判定できるよう残す |
| `default_branch` | `string` | repo.default_branch と同じ値（行レベルで欲しいケースのために重複させる） |
| `created_at` | `string \| null` | worktree ディレクトリの mtime（ISO 8601 / RFC 3339）。取得失敗時 `null`。`prunable` 時は `null` |
| `last_commit` | `object \| null` | `prunable` 時 `null`、取得失敗時 `null` |
| `last_commit.sha` | `string` | 7 文字 short SHA |
| `last_commit.subject` | `string` | コミット件名 1 行 |
| `last_commit.time` | `string` | author date, RFC 3339 |
| `pr` | `object \| null` | gh が使えない / branch にマッチする PR が無い / `--no-pr` のとき `null` |
| `pr.state` | `string` | `"OPEN" \| "DRAFT" \| "MERGED" \| "CLOSED"` |
| `pr.number` | `int` | PR 番号 |
| `pr.url` | `string` | `https://github.com/<owner>/<repo>/pull/<number>` を組み立てて返す（`gh` 出力に URL は含まれないため）。`origin` 未設定なら `null` |
| `pr.title` | `string` | PR タイトル |
| `session` | `object` | 必ず存在 |
| `session.exists` | `bool` | session log ファイルが見つかれば true。`--no-session` 時は false |
| `session.log_path` | `string \| null` | exists=true 時のみパスを返す |

### 注意

- `prunable` な worktree は git 側でディレクトリが消えているため、`ahead/behind/dirty/last_commit/created_at/session` はすべて `0` / `false` / `null` 固定。`name` / `path` / `branch` / `status: "prunable"` のみ有意。
- 全フィールドはエラーがあっても可能な限り埋め、欠落部のみ `null` にする（fail-soft）。

## 出力フォーマット（text 既定）

カラム: `NAME` / `STATUS` / `AHEAD/BEHIND` / `PR` / `SESSION` / `BRANCH`

```text
NAME                            STATUS      AHEAD/BEHIND  PR        SESSION  BRANCH
ccw-tqer39-ccw-cli-9d3dc6-4     pushed      0/0           #42 OPEN  RESUME   worktree-ccw-tqer39-ccw-cli-9d3dc6-4
ccw-tqer39-ccw-cli-3e66b9-1     dirty       2/0           -         NEW      worktree-ccw-tqer39-ccw-cli-3e66b9-1
ccw-tqer39-ccw-cli-deadbeef     prunable    -             -         -        worktree-ccw-tqer39-ccw-cli-deadbeef
```

ルール:

- ANSI 装飾なし（pipe 前提、TTY 判定もしない）。`NO_COLOR` 環境変数の影響を受けない。
- カラム幅は内容に応じた自動計算。
- `PR` セルは `#<num> <state>` 形式。PR 無 / gh 不在は `-`。
- `SESSION` セルは `RESUME` / `NEW` の 2 値。`--no-session` 指定時も `NEW`。
- `prunable` 行の AHEAD/BEHIND は `-`（数値 0 を出すと誤解を招くため）。
- ヘッダー行は常に出す。worktree 0 件の場合はヘッダーのみ。

text モードで欠落するフィールド（path / dirty / created_at / last_commit / pr.url / pr.title / session.log_path / repo.* / default_branch）は `--json` で取得する設計とし、その旨を `--help` のフラグ説明に短く書く。

## エラー処理

| ケース | 挙動 |
|---|---|
| `-d` の path が存在しない / git 配下でない | exit 1, stderr に "ccw -L: not a git repository: <path>" |
| `-d` 単独指定（`-L` なし） | exit 1, stderr に "ccw: -d requires -L" |
| `-L` と排他フラグの併用 | exit 1, stderr に "ccw: -L cannot be combined with -n / -s / --clean-all / --" |
| `gh` 不在 / 認証なし | 全 entry の `pr: null`、stderr WARN 1 行 (`ccw: gh not available, PR info disabled`) |
| `gh` 取得失敗（rate limit / network / タイムアウト 5s） | 全 entry の `pr: null`、stderr WARN 1 行（理由付き） |
| session log 探索が permission 等で失敗 | 該当 entry の `session.exists: false`、stderr WARN（最大 1 行に抑制、entry 数だけ出さない） |
| `git worktree list` が失敗 | exit 2, stderr にエラー詳細 |
| `created_at` / `last_commit` の取得失敗 | 該当エントリのみ `null`、WARN は出さない |

`gh` 呼び出しは `context.WithTimeout` で 5s に制限する（picker は現状 timeout を持たないが、非対話モードでは Claude が無限待機しないよう必須）。timeout 超過は WARN を出して `pr: null`。

`--no-pr` / `--no-session` でこの fail-soft 自体を skip できる。これは Claude Code が高速化目的で渡す想定。

## exit code

| code | 意味 |
|---|---|
| 0 | 成功（worktree 0 件でも 0） |
| 1 | ユーザーエラー（無効 `-d`、フラグ組合せ不正、git でない） |
| 2 | システムエラー（git コマンド失敗等） |

`gh` 失敗 / session 取得失敗は exit code に影響しない（fail-soft）。

## アーキテクチャ

### 既存資産の再利用

`internal/worktree.List` が既に以下を返す（`Info` 構造体）:

```go
type Info struct {
    Path        string
    Branch      string
    Status      Status   // Pushed / LocalOnly / Dirty / Prunable
    AheadCount  int
    BehindCount int
    DirtyCount  int
    HasSession  bool
}
```

list モードの大半はこれをそのまま使える。**`internal/worktreeinfo` のような新パッケージは作らない**（YAGNI）。

### 拡張点

`internal/worktree.Info` に追加するフィールド:

```go
type Info struct {
    // existing fields ...
    CreatedAt   *time.Time   // worktree dir mtime
    LastCommit  *CommitInfo  // nil for prunable / failure
    SessionPath string       // empty when HasSession == false
}

type CommitInfo struct {
    SHA     string
    Subject string
    Time    time.Time
}
```

`worktree.List` の中で各エントリ確定後に上記を埋める。`prunable` のときはスキップ。失敗時は `nil` / 空文字。

`internal/gitx` に追加 API:

```go
// LastCommit returns short SHA, subject, and author time of HEAD.
func LastCommit(wt string) (sha, subject string, t time.Time, err error)
```

`internal/worktree` に追加 API:

```go
// SessionLogPath returns the absolute path to the session log
// when HasSession(absPath) is true, or "" otherwise.
// EncodeProjectPath はすでに存在するのでそれを再利用。
func SessionLogPath(absPath string) string
```

### 新規パッケージ

`internal/listmode` (新規):

```go
// Output is the top-level JSON shape (also used as the source for text rendering).
type Output struct {
    Version   int            `json:"version"`
    Repo      RepoInfo       `json:"repo"`
    Worktrees []WorktreeInfo `json:"worktrees"`
}

// Build assembles the Output from the given main repo.
// opts controls --no-pr / --no-session.
func Build(mainRepo string, opts Options) (*Output, []Warning, error)

// RenderJSON marshals Output to indented JSON.
func RenderJSON(out *Output, w io.Writer) error

// RenderText writes the table format described in the design.
func RenderText(out *Output, w io.Writer) error
```

`Options` は `NoPR bool` / `NoSession bool` の 2 フィールド。

### `cmd/ccw` への変更

- `flag.Parse` 相当のフラグ解析に `-L` `--list` `-d` `--json` `--no-pr` `--no-session` を追加。
- `-L` を見たら listmode 経路に分岐し、`Build` → `RenderText` or `RenderJSON` → `os.Stdout` → exit。
- 既存の picker / new / clean-all 経路には触らない。

### picker 側のリファクタ範囲

picker は `internal/worktree.List` を呼んでいる。新しい `Info` フィールドが増えるが、picker は Path/Branch/Status などしか参照していないため挙動は不変。コンパイルさえ通れば picker は動く。

`worktree.List` 内で `LastCommit` / `created_at` / `SessionPath` の取得が増えるぶん picker 起動が若干遅くなる可能性がある。気になるレベルか実測する。許容できなければ `Info` 取得をオプショナル化（`ListOptions` に `IncludeCommit` / `IncludeMTime` 等を持たせる）するが、初版では入れない。

## テスト

### `internal/listmode` (unit)

- `RenderText`: 各 status / PR 有無 / session 有無の組合せで snapshot test
- `RenderJSON`: 同上、JSON でフィールド型と null 処理を検証
- `Build`: `gitx` / `gh` / `worktree` をフェイクで差し替えて hermetic に
  - happy path（PR 有 / RESUME 有）
  - PR 取得失敗 → `pr: null`、warning 1 件
  - `--no-pr`, `--no-session` 各オプション
  - prunable のみ
  - worktree 0 件 → empty array

### `internal/worktree`

- 新フィールド `CreatedAt` / `LastCommit` / `SessionPath` の埋まり方
- prunable 時はすべて空のまま
- 既存テストのリグレッションなし

### `internal/gitx`

- `LastCommit`: 通常 / 空 repo（エラー）/ shallow 等

### `cmd/ccw` (smoke)

- `-L` で picker が立ち上がらないこと
- `-L --json` で valid JSON が出ること
- `-L -d <tmpdir>` で対象が切り替わること
- 排他フラグの組合せが exit 1 になること

## ドキュメント変更

`README.md` / `docs/README.ja.md`:

- "Usage" セクションに `ccw -L` / `ccw -L --json` の例を追加
- "Features" に「📋 Machine-readable worktree list (`-L --json`)」を 1 行追加
- フラグリファレンス（`ccw --help`）の出力に合わせて `--help` の文言も更新

両ファイル同期は `readme-sync` skill 経由で確認。

## ロールバック

- フラグ解析と `internal/listmode` 追加が中心、既存経路への変更は `internal/worktree.Info` のフィールド追加と `worktree.List` 内の追加取得のみ。
- 万一 picker のパフォーマンスに影響が出たら、フィールド取得をオプショナル化する patch で対応。
- リバート不要レベルにスコープを絞っているが、必要なら PR revert で完結。

## PR スコープ

この spec は **単独 PR** 用。実装プランは `docs/superpowers/plans/` 配下に別途作成する。
