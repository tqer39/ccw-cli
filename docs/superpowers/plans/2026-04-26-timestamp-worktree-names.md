# Timestamp-based worktree naming Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** worktree / Claude Code セッション名を `ccw-<owner>-<repo>-<shorthash6>` から `ccw-<owner>-<repo>-<yymmdd>-<hhmmss>`（ローカルタイム）に変更し、「いつ始めた作業か」を一目で把握できるようにする。

**Architecture:** `internal/namegen` の hook を `shortHashFn` / `defaultBranchFn` から `nowFn = time.Now` に置き換える。`gitx.ShortHash` は本変更で唯一の利用が消えるため削除。`internal/namegen.buildName` のロジックと衝突 suffix（`-2`, `-3`…）はそのまま流用。

**Tech Stack:** Go 1.x、標準 `time` パッケージ、既存の `internal/gitx` / `internal/namegen` パッケージ。

---

## File Map

| ファイル | 変更種別 | 責務 |
|---|---|---|
| `internal/namegen/namegen.go` | 修正 | `nowFn` 導入、`Generate` を timestamp 化 |
| `internal/namegen/namegen_test.go` | 修正 | `fakes` 構造体・`withFakes` を nowFn に対応、期待値更新、不要テスト削除 |
| `internal/gitx/branch.go` | 修正 | `ShortHash` 関数の削除 |
| `internal/gitx/branch_test.go` | 修正 | `TestShortHash_Length` / `TestShortHash_MissingRef` の削除 |
| `cmd/ccw/main.go` | 修正 | `Generate` 失敗時の hint メッセージ更新（2 箇所） |
| `README.md` | 修正 | 命名規約セクション更新 |
| `docs/README.ja.md` | 修正 | 命名規約セクション更新 |

---

## Task 1: namegen のテストを timestamp 形式に書き換え

**Files:**

- Modify: `internal/namegen/namegen_test.go`

このタスクではテストだけ先に書き換える。実装は Task 2 で更新するため、この時点ではビルドが通らない／テストが失敗する。Task 2 と一緒にコミットする。

- [ ] **Step 1.1: `namegen_test.go` を新しい fakes / 期待値に書き換える**

`internal/namegen/namegen_test.go` の内容を以下の完全版に置換する:

```go
package namegen

import (
 "fmt"
 "os"
 "path/filepath"
 "strconv"
 "testing"
 "time"

 "github.com/tqer39/ccw-cli/internal/gitx"
)

func TestNormalize(t *testing.T) {
 cases := []struct {
  in, want string
 }{
  {"Anthropic", "anthropic"},
  {"My Org", "my-org"},
  {"_underscore_", "underscore"},
  {"--double--dash--", "double-dash"},
  {"a..b..c", "a-b-c"},
  {"", ""},
  {"a", "a"},
  {"123", "123"},
  {"日本語repo", "repo"},
 }
 for _, tc := range cases {
  t.Run(tc.in, func(t *testing.T) {
   got := normalize(tc.in)
   if got != tc.want {
    t.Errorf("normalize(%q) = %q, want %q", tc.in, got, tc.want)
   }
  })
 }
}

func TestBuildName(t *testing.T) {
 cases := []struct {
  name      string
  owner     string
  repo      string
  tail      string
  taken     map[string]bool
  want      string
  wantError bool
 }{
  {
   name:  "no collision",
   owner: "tqer39", repo: "ccw-cli", tail: "260426-143055",
   taken: map[string]bool{},
   want:  "ccw-tqer39-ccw-cli-260426-143055",
  },
  {
   name:  "one collision",
   owner: "tqer39", repo: "ccw-cli", tail: "260426-143055",
   taken: map[string]bool{"ccw-tqer39-ccw-cli-260426-143055": true},
   want:  "ccw-tqer39-ccw-cli-260426-143055-2",
  },
  {
   name:  "two collisions",
   owner: "tqer39", repo: "ccw-cli", tail: "260426-143055",
   taken: map[string]bool{
    "ccw-tqer39-ccw-cli-260426-143055":   true,
    "ccw-tqer39-ccw-cli-260426-143055-2": true,
   },
   want: "ccw-tqer39-ccw-cli-260426-143055-3",
  },
  {
   name:  "normalization applied",
   owner: "Anthropic", repo: "Claude.Code", tail: "260426-143055",
   taken: map[string]bool{},
   want:  "ccw-anthropic-claude-code-260426-143055",
  },
  {
   name:  "empty owner errors",
   owner: "", repo: "ccw-cli", tail: "260426-143055",
   taken:     map[string]bool{},
   wantError: true,
  },
  {
   name:  "empty repo errors",
   owner: "tqer39", repo: "", tail: "260426-143055",
   taken:     map[string]bool{},
   wantError: true,
  },
  {
   name:  "empty tail errors",
   owner: "tqer39", repo: "ccw-cli", tail: "",
   taken:     map[string]bool{},
   wantError: true,
  },
 }
 for _, tc := range cases {
  t.Run(tc.name, func(t *testing.T) {
   got, err := buildName(tc.owner, tc.repo, tc.tail, tc.taken)
   if tc.wantError {
    if err == nil {
     t.Fatalf("buildName(%q,%q,%q) want error, got %q", tc.owner, tc.repo, tc.tail, got)
    }
    return
   }
   if err != nil {
    t.Fatalf("buildName: %v", err)
   }
   if got != tc.want {
    t.Errorf("buildName = %q, want %q", got, tc.want)
   }
  })
 }
}

func TestBuildName_ManyCollisions(t *testing.T) {
 taken := map[string]bool{}
 base := "ccw-x-y-260426-143055"
 taken[base] = true
 for i := 2; i <= 99; i++ {
  taken[base+"-"+strconv.Itoa(i)] = true
 }
 if _, err := buildName("x", "y", "260426-143055", taken); err == nil {
  t.Fatal("buildName at 99-collision cap: want error, got nil")
 }
}

type fakes struct {
 origin    string
 originErr bool
 now       time.Time
 worktrees []gitx.WorktreeEntry
}

// testNow / testTimestamp are the canonical fixed clock used across Generate tests.
var testNow = time.Date(2026, 4, 26, 14, 30, 55, 0, time.Local)

const testTimestamp = "260426-143055"

func withFakes(t *testing.T, f fakes) {
 t.Helper()
 origOrigin := originURLFn
 origList := worktreeListFn
 origNow := nowFn
 t.Cleanup(func() {
  originURLFn = origOrigin
  worktreeListFn = origList
  nowFn = origNow
 })
 originURLFn = func(string) (string, error) {
  if f.originErr {
   return "", fmt.Errorf("fake: origin url error")
  }
  return f.origin, nil
 }
 worktreeListFn = func(string) ([]gitx.WorktreeEntry, error) { return f.worktrees, nil }
 if !f.now.IsZero() {
  nowFn = func() time.Time { return f.now }
 } else {
  nowFn = func() time.Time { return testNow }
 }
}

func mustMkdir(t *testing.T, root, rel string) {
 t.Helper()
 p := filepath.Join(root, rel)
 if err := os.MkdirAll(p, 0o755); err != nil {
  t.Fatalf("mkdir %s: %v", p, err)
 }
}

func TestGenerate_HappyPath(t *testing.T) {
 withFakes(t, fakes{
  origin: "git@github.com:tqer39/ccw-cli.git",
  now:    testNow,
 })
 got, err := Generate(t.TempDir())
 if err != nil {
  t.Fatalf("Generate: %v", err)
 }
 want := "ccw-tqer39-ccw-cli-" + testTimestamp
 if got != want {
  t.Errorf("Generate = %q, want %q", got, want)
 }
}

func TestGenerate_NoOriginFallback(t *testing.T) {
 withFakes(t, fakes{
  origin: "",
  now:    testNow,
 })
 tmp := t.TempDir()
 repoPath := filepath.Join(tmp, "myrepo")
 if err := os.MkdirAll(repoPath, 0o755); err != nil {
  t.Fatalf("mkdir: %v", err)
 }
 got, err := Generate(repoPath)
 if err != nil {
  t.Fatalf("Generate: %v", err)
 }
 want := "ccw-local-myrepo-" + testTimestamp
 if got != want {
  t.Errorf("Generate = %q, want %q", got, want)
 }
}

func TestGenerate_OriginURLError(t *testing.T) {
 withFakes(t, fakes{
  originErr: true,
  now:       testNow,
 })
 if _, err := Generate(t.TempDir()); err == nil {
  t.Fatal("Generate with origin-url error: want error, got nil")
 }
}

func TestGenerate_CollisionWithExistingDir(t *testing.T) {
 repo := t.TempDir()
 mustMkdir(t, repo, ".claude/worktrees/ccw-tqer39-ccw-cli-"+testTimestamp)
 withFakes(t, fakes{
  origin: "git@github.com:tqer39/ccw-cli.git",
  now:    testNow,
 })
 got, err := Generate(repo)
 if err != nil {
  t.Fatalf("Generate: %v", err)
 }
 want := "ccw-tqer39-ccw-cli-" + testTimestamp + "-2"
 if got != want {
  t.Errorf("Generate = %q, want %q", got, want)
 }
}

// TestGenerate_CollisionWithGitWorktree exercises the spec rule that names
// registered with `git worktree list` count as taken even when no matching
// .claude/worktrees directory exists.
func TestGenerate_CollisionWithGitWorktree(t *testing.T) {
 repo := t.TempDir()
 withFakes(t, fakes{
  origin: "git@github.com:tqer39/ccw-cli.git",
  now:    testNow,
  worktrees: []gitx.WorktreeEntry{
   {Path: "/tmp/elsewhere/ccw-tqer39-ccw-cli-" + testTimestamp},
  },
 })
 got, err := Generate(repo)
 if err != nil {
  t.Fatalf("Generate: %v", err)
 }
 want := "ccw-tqer39-ccw-cli-" + testTimestamp + "-2"
 if got != want {
  t.Errorf("Generate = %q, want %q", got, want)
 }
}
```

変更点:

- `fakes` から `branch`, `branchError`, `shorthash`, `shorthashErr` を削除し、`now time.Time` と `originErr bool` を追加
- `withFakes` で `defaultBranchFn` / `shortHashFn` の差し替えを削除し、`nowFn` の差し替えに置換
- 固定時刻 `testNow` / 期待 timestamp `testTimestamp` をパッケージレベル定数に
- `TestBuildName` のフィールド名 `shorthash` → `tail`、期待値を新フォーマットに更新
- `TestBuildName_ManyCollisions` を新フォーマットに更新
- `TestGenerate_DefaultBranchError` / `TestGenerate_ShortHashError` を削除（時刻取得は失敗しない）
- 代わりに `TestGenerate_OriginURLError` を追加（origin URL 取得失敗の経路は依然存在する）

- [ ] **Step 1.2: テストを実行して期待どおり失敗することを確認**

```bash
go test ./internal/namegen/... -count=1
```

期待: ビルドエラー（`nowFn` 未定義、`shortHashFn` 削除済み参照など）。Task 2 で実装を更新すると pass する。

- [ ] **Step 1.3: コミットせず Task 2 へ進む**

このタスク単体ではコンパイルが通らないため、Task 2 完了時に 1 コミットでまとめる。

---

## Task 2: namegen 本体を timestamp 化

**Files:**

- Modify: `internal/namegen/namegen.go`

- [ ] **Step 2.1: `namegen.go` の冒頭 doc コメントとパッケージレベル変数を書き換える**

`internal/namegen/namegen.go` の内容を以下の完全版に置換する:

```go
// Package namegen generates timestamp-based worktree / Claude Code session names
// of the form "ccw-<owner>-<repo>-<yymmdd>-<hhmmss>" using local time.
package namegen

import (
 "fmt"
 "os"
 "path/filepath"
 "regexp"
 "strings"
 "time"

 "github.com/tqer39/ccw-cli/internal/gitx"
)

// nonSlugRE matches anything outside [a-z0-9-]. Used by normalize.
var nonSlugRE = regexp.MustCompile(`[^a-z0-9-]+`)

// dashRunRE matches runs of two or more dashes. Used by normalize.
var dashRunRE = regexp.MustCompile(`-{2,}`)

// normalize returns a slug-safe lowercase form of s: ASCII-only, [a-z0-9-]+,
// with consecutive dashes collapsed and leading/trailing dashes trimmed.
// Callers (e.g. ParseOriginURL) strip `.git` before calling.
func normalize(s string) string {
 s = strings.ToLower(s)
 s = nonSlugRE.ReplaceAllString(s, "-")
 s = dashRunRE.ReplaceAllString(s, "-")
 s = strings.Trim(s, "-")
 return s
}

// origin / worktree-list / clock hooks are package-level vars so tests can
// substitute fakes without spinning up a real repo or a real clock.
var (
 originURLFn    = gitx.OriginURL
 worktreeListFn = gitx.ListRaw
 nowFn          = time.Now
)

// maxCollisionSuffix bounds numeric suffixes attempted before giving up.
const maxCollisionSuffix = 99

// timestampLayout is Go's reference time formatted as yymmdd-hhmmss.
const timestampLayout = "060102-150405"

// buildName composes "ccw-<owner>-<repo>-<tail>" with normalization,
// suffixing "-2", "-3", ... when the candidate is in `taken`.
func buildName(owner, repo, tail string, taken map[string]bool) (string, error) {
 o := normalize(owner)
 r := normalize(repo)
 t := normalize(tail)
 if o == "" {
  return "", fmt.Errorf("buildName: owner is empty after normalization (input %q)", owner)
 }
 if r == "" {
  return "", fmt.Errorf("buildName: repo is empty after normalization (input %q)", repo)
 }
 if t == "" {
  return "", fmt.Errorf("buildName: tail is empty after normalization (input %q)", tail)
 }
 base := fmt.Sprintf("ccw-%s-%s-%s", o, r, t)
 if !taken[base] {
  return base, nil
 }
 for i := 2; i <= maxCollisionSuffix; i++ {
  candidate := fmt.Sprintf("%s-%d", base, i)
  if !taken[candidate] {
   return candidate, nil
  }
 }
 return "", fmt.Errorf("buildName: %d collisions for %q, giving up", maxCollisionSuffix, base)
}

// Generate returns a worktree name of the form
// "ccw-<owner>-<repo>-<yymmdd>-<hhmmss>" for the repository at mainRepo.
// The timestamp is formatted from time.Now() in local time.
// When `origin` is unset, owner becomes "local" and repo is the basename
// of mainRepo. Numeric "-N" suffixes are appended on collision (cap: 99).
func Generate(mainRepo string) (string, error) {
 owner, repo, err := resolveOwnerRepo(mainRepo)
 if err != nil {
  return "", err
 }
 ts := nowFn().Format(timestampLayout)
 taken, err := takenNames(mainRepo)
 if err != nil {
  return "", err
 }
 return buildName(owner, repo, ts, taken)
}

func resolveOwnerRepo(mainRepo string) (string, string, error) {
 url, err := originURLFn(mainRepo)
 if err != nil {
  return "", "", fmt.Errorf("origin url: %w", err)
 }
 if url == "" {
  return "local", filepath.Base(mainRepo), nil
 }
 owner, repo, err := gitx.ParseOriginURL(url)
 if err != nil {
  return "", "", fmt.Errorf("parse origin url: %w", err)
 }
 return owner, repo, nil
}

// takenNames returns names already in use, by union of:
//   - directory entries under <mainRepo>/.claude/worktrees/
//   - basenames of paths returned by `git worktree list --porcelain`
//
// A registered git worktree without a matching .claude/worktrees entry (e.g.
// added outside ccw, or after manual cleanup) still collides with a fresh name.
func takenNames(mainRepo string) (map[string]bool, error) {
 out := map[string]bool{}
 dir := filepath.Join(mainRepo, ".claude", "worktrees")
 entries, err := os.ReadDir(dir)
 switch {
 case err == nil:
  for _, e := range entries {
   if e.IsDir() {
    out[e.Name()] = true
   }
  }
 case !os.IsNotExist(err):
  return nil, fmt.Errorf("read worktrees dir: %w", err)
 }
 wts, err := worktreeListFn(mainRepo)
 if err != nil {
  return nil, fmt.Errorf("list git worktrees: %w", err)
 }
 for _, wt := range wts {
  out[filepath.Base(wt.Path)] = true
 }
 return out, nil
}
```

主な差分:

- import に `"time"` を追加
- パッケージ doc コメントを timestamp フォーマットに更新
- `defaultBranchFn` / `shortHashFn` を削除し、`nowFn = time.Now` を追加
- `timestampLayout = "060102-150405"` 定数を追加
- `buildName` の第 3 引数を `shorthash` → `tail` に rename。空チェックのエラー文言も `"tail is empty"` に
- `Generate` のロジックを `branch / shorthash` 取得から `nowFn().Format(timestampLayout)` に置換
- `Generate` の doc コメントを更新

- [ ] **Step 2.2: namegen テストが通ることを確認**

```bash
go test ./internal/namegen/... -count=1 -race -v
```

期待: 全テスト pass。`TestGenerate_HappyPath` が `ccw-tqer39-ccw-cli-260426-143055` を返す等。

- [ ] **Step 2.3: パッケージ全体の lint**

```bash
golangci-lint run ./internal/namegen/...
```

期待: エラーなし。

- [ ] **Step 2.4: コミット**

```bash
git add internal/namegen/namegen.go internal/namegen/namegen_test.go
git commit -m "$(cat <<'EOF'
feat(namegen): switch worktree names to local timestamp (yymmdd-hhmmss)

ccw-<owner>-<repo>-<yymmdd>-<hhmmss> 形式で worktree / セッション名を生成。
default branch HEAD short SHA への依存を削除し time.Now() に置換。
collision suffix（-2, -3, ...）と origin 未設定時の local fallback はそのまま。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: 未使用の `gitx.ShortHash` を削除

**Files:**

- Modify: `internal/gitx/branch.go`
- Modify: `internal/gitx/branch_test.go`

- [ ] **Step 3.1: 削除前に他からの参照がないことを再確認**

```bash
grep -rn "gitx.ShortHash\|gitx\.ShortHash\b" --include="*.go" cmd internal | grep -v _test.go | grep -v worktrees
```

期待: 出力なし。Task 2 で namegen 側からの参照は消えているはず。出力があれば停止して人に確認。

- [ ] **Step 3.2: `internal/gitx/branch.go` から `ShortHash` 関数を削除**

`internal/gitx/branch.go` の以下の部分を削除する:

```go
// ShortHash returns the trimmed output of `git rev-parse --short=<length> <ref>`.
func ShortHash(mainRepo, ref string, length int) (string, error) {
 out, err := Output(mainRepo, "rev-parse", fmt.Sprintf("--short=%d", length), ref)
 if err != nil {
  return "", fmt.Errorf("short hash %s: %w", ref, err)
 }
 return strings.TrimSpace(out), nil
}
```

削除後、`fmt` import は `DefaultBranch` のエラーフォーマットでまだ必要。`strings` も `DefaultBranch` で使用されているので残す。import 整理は `goimports` / `gofmt` が自動でやるので手で消さない。

- [ ] **Step 3.3: `internal/gitx/branch_test.go` から `TestShortHash_*` を削除**

`internal/gitx/branch_test.go` の以下の 2 関数を削除する:

```go
func TestShortHash_Length(t *testing.T) { ... }
func TestShortHash_MissingRef(t *testing.T) { ... }
```

ファイル末尾に余分な空行が残らないように注意。

- [ ] **Step 3.4: gitx テストを実行**

```bash
go test ./internal/gitx/... -count=1 -race
```

期待: 全テスト pass。

- [ ] **Step 3.5: 全体ビルドを確認**

```bash
go build ./...
```

期待: エラーなし。

- [ ] **Step 3.6: コミット**

```bash
git add internal/gitx/branch.go internal/gitx/branch_test.go
git commit -m "$(cat <<'EOF'
chore(gitx): drop unused ShortHash helper

namegen が timestamp ベースの命名に切り替わり、ShortHash の呼び出し元が無くなったため削除。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: `cmd/ccw/main.go` のエラー hint メッセージを更新

**Files:**

- Modify: `cmd/ccw/main.go:77` および `cmd/ccw/main.go:104`

- [ ] **Step 4.1: 1 箇所目（`flags.NewWorktree` パス）を更新**

`cmd/ccw/main.go` の以下を:

```go
  name, err := namegen.Generate(mainRepo)
  if err != nil {
   ui.Error("generate worktree name: %v\nhint: ensure a 'main' or 'master' branch with at least one commit, or run `git remote set-head origin -a`", err)
   return 1
  }
```

次のように変更する:

```go
  name, err := namegen.Generate(mainRepo)
  if err != nil {
   ui.Error("generate worktree name: %v", err)
   return 1
  }
```

(理由: timestamp 取得は失敗しないため、`Generate` のエラーは「origin URL 取得失敗」「`.claude/worktrees` 読み込み失敗」「衝突 99 件超過」のいずれか。これらは hint で誘導しても解決しない問題なので素のエラーだけ表示する。)

- [ ] **Step 4.2: 2 箇所目（picker の `ActionNew` パス）を同様に更新**

`cmd/ccw/main.go` の picker 内 `case picker.ActionNew:` ブロック内の同じエラーメッセージも同様に書き換える。

- [ ] **Step 4.3: ビルドとテスト**

```bash
go build ./...
go test ./... -race -count=1
```

期待: 全テスト pass。

- [ ] **Step 4.4: コミット**

```bash
git add cmd/ccw/main.go
git commit -m "$(cat <<'EOF'
fix(ccw): drop obsolete default-branch hint from name-generation error

namegen が default branch / short SHA に依存しなくなったため、main/master ブランチ作成や
git remote set-head を案内する hint は無効になった。素の error をそのまま表示する。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: README.md の命名規約を更新

**Files:**

- Modify: `README.md:94`

- [ ] **Step 5.1: 該当パラグラフを書き換え**

`README.md` の該当行（94 行目相当）:

```markdown
`<name>` is generated as `ccw-<owner>-<repo>-<shorthash6>` (e.g. `ccw-tqer39-ccw-cli-a3f2b1`). `<owner>` / `<repo>` come from the `origin` remote URL; `<shorthash6>` is the 6-char short SHA of the local default branch tip at creation time. When `origin` is unset, `<owner>` becomes `local` and `<repo>` is the directory basename. Duplicate names are disambiguated with `-2`, `-3`, … Renaming the session manually with `/rename` is fine — ccw does not track it, and `--continue` keys off the working directory so conversation restore is unaffected.
```

を以下に置換:

```markdown
`<name>` is generated as `ccw-<owner>-<repo>-<yymmdd>-<hhmmss>` (e.g. `ccw-tqer39-ccw-cli-260426-143055`). `<owner>` / `<repo>` come from the `origin` remote URL; the timestamp is the worktree creation time in your local timezone. When `origin` is unset, `<owner>` becomes `local` and `<repo>` is the directory basename. Duplicate names (e.g. two worktrees created within the same second) are disambiguated with `-2`, `-3`, … Renaming the session manually with `/rename` is fine — ccw does not track it, and `--continue` keys off the working directory so conversation restore is unaffected.
```

- [ ] **Step 5.2: markdown lint をローカルで確認**

```bash
markdownlint-cli2 README.md
```

期待: エラーなし（lefthook の pre-commit hook が同じものを実行する）。

---

## Task 6: docs/README.ja.md の命名規約を更新

**Files:**

- Modify: `docs/README.ja.md:94`

- [ ] **Step 6.1: 該当パラグラフを書き換え**

`docs/README.ja.md` の該当行（94 行目相当）:

```markdown
`<name>` は `ccw-<owner>-<repo>-<shorthash6>`（例: `ccw-tqer39-ccw-cli-a3f2b1`）形式で生成されます。`<owner>` / `<repo>` は `origin` remote の URL から抽出、`<shorthash6>` は作成時点のローカル default branch tip の short SHA です。`origin` が未設定の場合は `<owner>` が `local`、`<repo>` がディレクトリ basename になります。同名衝突は `-2`, `-3`, … で回避します。`/rename` で手動改名しても ccw 側は追跡しないため問題ありません（`--continue` は作業ディレクトリ基準で会話を復元します）。
```

を以下に置換:

```markdown
`<name>` は `ccw-<owner>-<repo>-<yymmdd>-<hhmmss>`（例: `ccw-tqer39-ccw-cli-260426-143055`）形式で生成されます。`<owner>` / `<repo>` は `origin` remote の URL から抽出、タイムスタンプは worktree 作成時刻（ローカルタイム）です。`origin` が未設定の場合は `<owner>` が `local`、`<repo>` がディレクトリ basename になります。同一秒に複数作成した場合などの同名衝突は `-2`, `-3`, … で回避します。`/rename` で手動改名しても ccw 側は追跡しないため問題ありません（`--continue` は作業ディレクトリ基準で会話を復元します）。
```

- [ ] **Step 6.2: markdown lint をローカルで確認**

```bash
markdownlint-cli2 docs/README.ja.md
```

期待: エラーなし。

- [ ] **Step 6.3: 両 README をコミット**

```bash
git add README.md docs/README.ja.md
git commit -m "$(cat <<'EOF'
docs: update worktree naming convention to timestamp form (yymmdd-hhmmss)

README (英) と docs/README.ja.md (日) の命名規約を新フォーマットに同期。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: 最終検証

- [ ] **Step 7.1: フルビルド + テスト + lint**

```bash
make test
make lint
make build
```

期待: 全 pass、`./ccw` バイナリが生成される。

- [ ] **Step 7.2: 実際に名前生成を確認（読み取り専用 smoke）**

```go
// scratch test (no need to commit) — もし不安なら手で書いて捨てる
// あるいは:
go run ./cmd/ccw -h | head -5
```

`-h` で起動できれば cmd/ccw 側のビルド・基本動作 OK。

- [ ] **Step 7.3: コミット履歴を確認**

```bash
git log --oneline -10
```

期待: Task 2 / 3 / 4 / 6 のコミットが順に並んでいる（Task 5 は Task 6 と同じコミットに含まれる）。

- [ ] **Step 7.4: PR 作成は別途人間が判断**

実装プランの完了。PR を作成する場合は別途 `git:create-pr` skill を使用する。

---

## Acceptance Checklist

- [ ] `go test ./... -race -count=1` が全 pass
- [ ] `make lint` がエラーなし
- [ ] `make build` でバイナリ生成成功
- [ ] `internal/namegen.Generate` が `ccw-<owner>-<repo>-<yymmdd>-<hhmmss>` 形式の文字列を返す（テストで確認済み）
- [ ] `gitx.ShortHash` 関数とそのテストが完全に削除されている
- [ ] `cmd/ccw/main.go` の Generate 失敗時 hint が古い文言（main/master 案内）を含まない
- [ ] README.md / docs/README.ja.md の命名規約が新フォーマットに更新されている
- [ ] 既存の hash 形式 worktree（例: `ccw-tqer39-ccw-cli-9d3dc6`）が picker に出る（手動確認は任意、コードパス上は無影響）
