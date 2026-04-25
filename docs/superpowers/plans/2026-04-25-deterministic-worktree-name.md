# 決定論的な worktree 名 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** worktree / セッション名のランダム slug (`quick-falcon-7bd2`) を、決定論的な `ccw-<owner>-<repo>-<shorthash6>` 形式（例: `ccw-tqer39-ccw-cli-a3f2b1`）に置き換える。

**Architecture:** `internal/gitx` に origin URL 解析と default branch / short SHA 取得関数を追加。`internal/namegen` の `Generate()` を `Generate(mainRepo string) (string, error)` に変更し、純粋関数 `normalize` / `buildName`（衝突回避）と `gitx` 呼び出し（関数値で差し替え可能）を組み合わせる。`cmd/ccw/main.go` の 2 箇所の呼び出しをエラー対応に。既存 worktree との互換は無条件で保つ（マイグレーション無し）。

**Tech Stack:** Go 1.25, 標準ライブラリ (`net/url`, `regexp`), 既存の `internal/gitx` ヘルパ。

**Spec:** `docs/superpowers/specs/2026-04-25-deterministic-worktree-name-design.md`

---

## File Structure

- Create: `internal/gitx/origin.go` — `OriginURL(mainRepo)`, `ParseOriginURL(url) (owner, repo, error)`
- Create: `internal/gitx/origin_test.go`
- Create: `internal/gitx/branch.go` — `DefaultBranch(mainRepo)`, `ShortHash(mainRepo, ref, length)`
- Create: `internal/gitx/branch_test.go`
- Modify: `internal/namegen/namegen.go` — `Generate` シグネチャ変更、`normalize` / `buildName` 追加、adj/noun テーブル削除
- Modify: `internal/namegen/namegen_test.go` — テスト全置換
- Modify: `cmd/ccw/main.go` — 呼び出し 2 箇所 (`L75`, `L98`) を `(name, err) := namegen.Generate(mainRepo)` 形に
- Modify: `README.md` — "Naming convention" セクション (L86-94) 書き換え
- Modify: `docs/README.ja.md` — 対応セクション書き換え

---

## Task 1: `gitx.ParseOriginURL` を追加（純粋関数）

**Files:**

- Create: `internal/gitx/origin.go`
- Create: `internal/gitx/origin_test.go`

- [ ] **Step 1: 失敗テストを書く**

`internal/gitx/origin_test.go` を新規作成:

```go
package gitx

import "testing"

func TestParseOriginURL(t *testing.T) {
 cases := []struct {
  name      string
  url       string
  owner     string
  repo      string
  wantError bool
 }{
  {"ssh github", "git@github.com:tqer39/ccw-cli.git", "tqer39", "ccw-cli", false},
  {"ssh github no .git", "git@github.com:tqer39/ccw-cli", "tqer39", "ccw-cli", false},
  {"https github", "https://github.com/tqer39/ccw-cli.git", "tqer39", "ccw-cli", false},
  {"https github no .git", "https://github.com/tqer39/ccw-cli", "tqer39", "ccw-cli", false},
  {"https with trailing slash", "https://github.com/tqer39/ccw-cli/", "tqer39", "ccw-cli", false},
  {"gitlab nested", "https://gitlab.com/group/sub/repo.git", "sub", "repo", false},
  {"ssh gitlab nested", "git@gitlab.com:group/sub/repo.git", "sub", "repo", false},
  {"empty", "", "", "", true},
  {"only host", "git@github.com:", "", "", true},
  {"single segment", "https://example.com/repo.git", "", "", true},
 }
 for _, tc := range cases {
  t.Run(tc.name, func(t *testing.T) {
   owner, repo, err := ParseOriginURL(tc.url)
   if tc.wantError {
    if err == nil {
     t.Fatalf("ParseOriginURL(%q) want error, got owner=%q repo=%q", tc.url, owner, repo)
    }
    return
   }
   if err != nil {
    t.Fatalf("ParseOriginURL(%q) unexpected error: %v", tc.url, err)
   }
   if owner != tc.owner || repo != tc.repo {
    t.Errorf("ParseOriginURL(%q) = (%q, %q), want (%q, %q)", tc.url, owner, repo, tc.owner, tc.repo)
   }
  })
 }
}
```

- [ ] **Step 2: 失敗を確認**

Run: `go test ./internal/gitx/ -run TestParseOriginURL -v`
Expected: コンパイルエラー `undefined: ParseOriginURL`

- [ ] **Step 3: 最小実装を書く**

`internal/gitx/origin.go` を新規作成:

```go
package gitx

import (
 "fmt"
 "strings"
)

// ParseOriginURL extracts (owner, repo) from a git remote URL.
// Supports SSH (git@host:owner/repo[.git]) and HTTPS (https://host/owner/repo[.git]) forms.
// Nested path segments (e.g. GitLab subgroups) collapse to the last two segments.
// Returns an error for empty, malformed, or single-segment paths.
func ParseOriginURL(rawURL string) (string, string, error) {
 url := strings.TrimSpace(rawURL)
 if url == "" {
  return "", "", fmt.Errorf("empty origin url")
 }
 var path string
 switch {
 case strings.HasPrefix(url, "git@"):
  idx := strings.Index(url, ":")
  if idx < 0 {
   return "", "", fmt.Errorf("malformed ssh url: %q", rawURL)
  }
  path = url[idx+1:]
 case strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "ssh://") || strings.HasPrefix(url, "git://"):
  idx := strings.Index(url, "://")
  rest := url[idx+3:]
  slash := strings.Index(rest, "/")
  if slash < 0 {
   return "", "", fmt.Errorf("malformed url: %q", rawURL)
  }
  path = rest[slash+1:]
 default:
  return "", "", fmt.Errorf("unsupported url scheme: %q", rawURL)
 }
 path = strings.TrimSuffix(strings.TrimSuffix(strings.Trim(path, "/"), ".git"), "/")
 parts := strings.Split(path, "/")
 if len(parts) < 2 || parts[0] == "" || parts[len(parts)-1] == "" {
  return "", "", fmt.Errorf("origin url has fewer than 2 path segments: %q", rawURL)
 }
 owner := parts[len(parts)-2]
 repo := parts[len(parts)-1]
 return owner, repo, nil
}
```

- [ ] **Step 4: テスト pass を確認**

Run: `go test ./internal/gitx/ -run TestParseOriginURL -v`
Expected: 全 case PASS

- [ ] **Step 5: コミット**

```bash
git add internal/gitx/origin.go internal/gitx/origin_test.go
git commit -m "feat(gitx): add ParseOriginURL for SSH/HTTPS git remotes"
```

---

## Task 2: `gitx.OriginURL` を追加（git wrapper）

**Files:**

- Modify: `internal/gitx/origin.go`
- Modify: `internal/gitx/origin_test.go`

- [ ] **Step 1: 失敗テストを追加**

`internal/gitx/origin_test.go` の末尾に追記:

```go
func TestOriginURL_Configured(t *testing.T) {
 dir := initRepo(t)
 mustRun(t, dir, "git", "remote", "add", "origin", "git@github.com:tqer39/ccw-cli.git")
 got, err := OriginURL(dir)
 if err != nil {
  t.Fatalf("OriginURL: %v", err)
 }
 if got != "git@github.com:tqer39/ccw-cli.git" {
  t.Errorf("OriginURL = %q, want %q", got, "git@github.com:tqer39/ccw-cli.git")
 }
}

func TestOriginURL_NotConfigured(t *testing.T) {
 dir := initRepo(t)
 got, err := OriginURL(dir)
 if err != nil {
  t.Fatalf("OriginURL on no-origin repo: want nil error, got %v", err)
 }
 if got != "" {
  t.Errorf("OriginURL = %q, want \"\"", got)
 }
}
```

(`initRepo` / `mustRun` は同パッケージ既存ヘルパを再利用)

- [ ] **Step 2: 失敗を確認**

Run: `go test ./internal/gitx/ -run TestOriginURL -v`
Expected: コンパイルエラー `undefined: OriginURL`

- [ ] **Step 3: 実装を追加**

`internal/gitx/origin.go` の末尾に追記:

```go
// OriginURL returns the URL of the `origin` remote, or "" when not configured.
// The empty/no-origin case is treated as a normal branch (no error).
func OriginURL(mainRepo string) (string, error) {
 out, err := OutputSilent(mainRepo, "remote", "get-url", "origin")
 if err != nil {
  return "", nil
 }
 return strings.TrimSpace(out), nil
}
```

- [ ] **Step 4: テスト pass を確認**

Run: `go test ./internal/gitx/ -run TestOriginURL -v`
Expected: 2 case PASS

- [ ] **Step 5: コミット**

```bash
git add internal/gitx/origin.go internal/gitx/origin_test.go
git commit -m "feat(gitx): add OriginURL wrapper that treats no-origin as empty"
```

---

## Task 3: `gitx.DefaultBranch` を追加

**Files:**

- Create: `internal/gitx/branch.go`
- Create: `internal/gitx/branch_test.go`

- [ ] **Step 1: 失敗テストを書く**

`internal/gitx/branch_test.go` を新規作成:

```go
package gitx

import "testing"

func TestDefaultBranch_FromOriginHEAD(t *testing.T) {
 upstream := initRepo(t)
 mustRun(t, upstream, "git", "commit", "--allow-empty", "-m", "init")
 dir := initRepo(t)
 mustRun(t, dir, "git", "remote", "add", "origin", upstream)
 mustRun(t, dir, "git", "fetch", "origin")
 mustRun(t, dir, "git", "remote", "set-head", "origin", "-a")
 got, err := DefaultBranch(dir)
 if err != nil {
  t.Fatalf("DefaultBranch: %v", err)
 }
 if got != "main" {
  t.Errorf("DefaultBranch = %q, want %q", got, "main")
 }
}

func TestDefaultBranch_FallbackMain(t *testing.T) {
 dir := initRepo(t)
 mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "init")
 got, err := DefaultBranch(dir)
 if err != nil {
  t.Fatalf("DefaultBranch fallback main: %v", err)
 }
 if got != "main" {
  t.Errorf("DefaultBranch = %q, want %q", got, "main")
 }
}

func TestDefaultBranch_FallbackMaster(t *testing.T) {
 dir := initRepo(t)
 mustRun(t, dir, "git", "checkout", "-q", "-b", "master")
 mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "init")
 mustRun(t, dir, "git", "branch", "-q", "-D", "main")
 got, err := DefaultBranch(dir)
 if err != nil {
  t.Fatalf("DefaultBranch fallback master: %v", err)
 }
 if got != "master" {
  t.Errorf("DefaultBranch = %q, want %q", got, "master")
 }
}

func TestDefaultBranch_NoBranches(t *testing.T) {
 dir := initRepo(t)
 mustRun(t, dir, "git", "checkout", "-q", "-b", "feature")
 mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "init")
 mustRun(t, dir, "git", "branch", "-q", "-D", "main")
 _, err := DefaultBranch(dir)
 if err == nil {
  t.Fatal("DefaultBranch with no main/master/origin: want error, got nil")
 }
}
```

- [ ] **Step 2: 失敗を確認**

Run: `go test ./internal/gitx/ -run TestDefaultBranch -v`
Expected: コンパイルエラー `undefined: DefaultBranch`

- [ ] **Step 3: 実装を書く**

`internal/gitx/branch.go` を新規作成:

```go
package gitx

import (
 "fmt"
 "strings"
)

// DefaultBranch returns the canonical default branch name for the repo at mainRepo.
// Resolution order:
//  1. refs/remotes/origin/HEAD (e.g. "refs/remotes/origin/main") — strip prefix
//  2. local branch "main"
//  3. local branch "master"
// Returns an error when none of the above exist.
func DefaultBranch(mainRepo string) (string, error) {
 if out, err := OutputSilent(mainRepo, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
  s := strings.TrimSpace(out)
  if idx := strings.LastIndex(s, "/"); idx >= 0 && idx < len(s)-1 {
   return s[idx+1:], nil
  }
 }
 for _, name := range []string{"main", "master"} {
  if _, err := OutputSilent(mainRepo, "rev-parse", "--verify", "--quiet", "refs/heads/"+name); err == nil {
   return name, nil
  }
 }
 return "", fmt.Errorf("no default branch found (origin/HEAD, main, master all unset)")
}
```

- [ ] **Step 4: テスト pass を確認**

Run: `go test ./internal/gitx/ -run TestDefaultBranch -v`
Expected: 4 case PASS

- [ ] **Step 5: コミット**

```bash
git add internal/gitx/branch.go internal/gitx/branch_test.go
git commit -m "feat(gitx): add DefaultBranch with origin/HEAD → main → master fallback"
```

---

## Task 4: `gitx.ShortHash` を追加

**Files:**

- Modify: `internal/gitx/branch.go`
- Modify: `internal/gitx/branch_test.go`

- [ ] **Step 1: 失敗テストを追加**

`internal/gitx/branch_test.go` の末尾に追記:

```go
func TestShortHash_Length(t *testing.T) {
 dir := initRepo(t)
 mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "init")
 got, err := ShortHash(dir, "main", 6)
 if err != nil {
  t.Fatalf("ShortHash: %v", err)
 }
 if len(got) != 6 {
  t.Errorf("ShortHash length = %d, want 6 (got %q)", len(got), got)
 }
}

func TestShortHash_MissingRef(t *testing.T) {
 dir := initRepo(t)
 _, err := ShortHash(dir, "nonexistent", 6)
 if err == nil {
  t.Fatal("ShortHash missing ref: want error, got nil")
 }
}
```

- [ ] **Step 2: 失敗を確認**

Run: `go test ./internal/gitx/ -run TestShortHash -v`
Expected: コンパイルエラー `undefined: ShortHash`

- [ ] **Step 3: 実装を追加**

`internal/gitx/branch.go` の末尾に追記:

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

- [ ] **Step 4: テスト pass を確認**

Run: `go test ./internal/gitx/ -v`
Expected: 全テスト（既存も含む）PASS

- [ ] **Step 5: コミット**

```bash
git add internal/gitx/branch.go internal/gitx/branch_test.go
git commit -m "feat(gitx): add ShortHash wrapping git rev-parse --short"
```

---

## Task 5: `namegen.normalize` を実装

**Files:**

- Modify: `internal/namegen/namegen.go`（書き換え）
- Modify: `internal/namegen/namegen_test.go`（書き換え）

- [ ] **Step 1: 既存ファイルを完全置換する準備として、新テストを書く**

`internal/namegen/namegen_test.go` を以下で完全置換:

```go
package namegen

import "testing"

func TestNormalize(t *testing.T) {
 cases := []struct {
  in, want string
 }{
  {"Anthropic", "anthropic"},
  {"My Org", "my-org"},
  {"_underscore_", "underscore"},
  {"--double--dash--", "double-dash"},
  {"repo.git", "repo"},
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
```

- [ ] **Step 2: 失敗を確認**

Run: `go test ./internal/namegen/ -run TestNormalize -v`
Expected: コンパイルエラー（旧 `Generate` を参照する main も壊れるが、このタスクでは namegen 単体だけ確認）

- [ ] **Step 3: `namegen.go` を書き換え（normalize のみ追加、旧 Generate も一旦残す形にせず一気に置換）**

`internal/namegen/namegen.go` を以下で完全置換:

```go
// Package namegen generates deterministic worktree / Claude Code session names
// of the form "ccw-<owner>-<repo>-<shorthash6>".
package namegen

import (
 "fmt"
 "regexp"
 "strings"

 "github.com/tqer39/ccw-cli/internal/gitx"
)

// nonSlugRE matches anything outside [a-z0-9-]. Used by normalize.
var nonSlugRE = regexp.MustCompile(`[^a-z0-9-]+`)

// dashRunRE matches runs of two or more dashes. Used by normalize.
var dashRunRE = regexp.MustCompile(`-{2,}`)

// normalize returns a slug-safe lowercase form of s: ASCII-only, [a-z0-9-]+,
// with consecutive dashes collapsed and leading/trailing dashes trimmed.
// `.git` suffix is not stripped here — callers (e.g. ParseOriginURL) handle it.
func normalize(s string) string {
 s = strings.ToLower(s)
 s = nonSlugRE.ReplaceAllString(s, "-")
 s = dashRunRE.ReplaceAllString(s, "-")
 s = strings.Trim(s, "-")
 return s
}

// origin / branch / shorthash hooks are package-level vars so tests can
// substitute fakes without spinning up a real repo.
var (
 originURLFn     = gitx.OriginURL
 defaultBranchFn = gitx.DefaultBranch
 shortHashFn     = gitx.ShortHash
)

// Generate placeholder — fully implemented in Task 7. Returns an error so
// callers fail fast if reached before the wiring is in place.
func Generate(mainRepo string) (string, error) {
 _ = mainRepo
 return "", fmt.Errorf("namegen.Generate not yet implemented")
}
```

- [ ] **Step 4: namegen のテスト pass を確認**

Run: `go test ./internal/namegen/ -run TestNormalize -v`
Expected: `TestNormalize/*` PASS

ビルド全体は cmd/ccw が旧 API を呼んでいて壊れるが、Task 8 で直すまでは想定内。確認:

Run: `go build ./internal/namegen/`
Expected: success

- [ ] **Step 5: コミット**

```bash
git add internal/namegen/namegen.go internal/namegen/namegen_test.go
git commit -m "refactor(namegen): replace random slug API with normalize + Generate stub"
```

---

## Task 6: `namegen.buildName` を実装（衝突回避）

**Files:**

- Modify: `internal/namegen/namegen.go`
- Modify: `internal/namegen/namegen_test.go`

- [ ] **Step 1: 失敗テストを追加**

`internal/namegen/namegen_test.go` の末尾に追記:

```go
func TestBuildName(t *testing.T) {
 cases := []struct {
  name      string
  owner     string
  repo      string
  shorthash string
  taken     map[string]bool
  want      string
  wantError bool
 }{
  {
   name: "no collision",
   owner: "tqer39", repo: "ccw-cli", shorthash: "a3f2b1",
   taken: map[string]bool{},
   want:  "ccw-tqer39-ccw-cli-a3f2b1",
  },
  {
   name: "one collision",
   owner: "tqer39", repo: "ccw-cli", shorthash: "a3f2b1",
   taken: map[string]bool{"ccw-tqer39-ccw-cli-a3f2b1": true},
   want:  "ccw-tqer39-ccw-cli-a3f2b1-2",
  },
  {
   name: "two collisions",
   owner: "tqer39", repo: "ccw-cli", shorthash: "a3f2b1",
   taken: map[string]bool{
    "ccw-tqer39-ccw-cli-a3f2b1":   true,
    "ccw-tqer39-ccw-cli-a3f2b1-2": true,
   },
   want: "ccw-tqer39-ccw-cli-a3f2b1-3",
  },
  {
   name: "normalization applied",
   owner: "Anthropic", repo: "Claude.Code", shorthash: "9F8E7D",
   taken: map[string]bool{},
   want:  "ccw-anthropic-claude-code-9f8e7d",
  },
  {
   name: "empty owner errors",
   owner: "", repo: "ccw-cli", shorthash: "a3f2b1",
   taken:     map[string]bool{},
   wantError: true,
  },
  {
   name: "empty repo errors",
   owner: "tqer39", repo: "", shorthash: "a3f2b1",
   taken:     map[string]bool{},
   wantError: true,
  },
  {
   name: "empty shorthash errors",
   owner: "tqer39", repo: "ccw-cli", shorthash: "",
   taken:     map[string]bool{},
   wantError: true,
  },
 }
 for _, tc := range cases {
  t.Run(tc.name, func(t *testing.T) {
   got, err := buildName(tc.owner, tc.repo, tc.shorthash, tc.taken)
   if tc.wantError {
    if err == nil {
     t.Fatalf("buildName(%q,%q,%q) want error, got %q", tc.owner, tc.repo, tc.shorthash, got)
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
 base := "ccw-x-y-aaaaaa"
 taken[base] = true
 for i := 2; i <= 99; i++ {
  taken[base+"-"+strconv.Itoa(i)] = true
 }
 if _, err := buildName("x", "y", "aaaaaa", taken); err == nil {
  t.Fatal("buildName at 99-collision cap: want error, got nil")
 }
}
```

`namegen_test.go` の import を以下に拡張:

```go
import (
 "strconv"
 "testing"
)
```

- [ ] **Step 2: 失敗を確認**

Run: `go test ./internal/namegen/ -run TestBuildName -v`
Expected: コンパイルエラー `undefined: buildName`

- [ ] **Step 3: 実装**

`internal/namegen/namegen.go` の `Generate` 関数の上に追加:

```go
// maxCollisionSuffix is the upper bound on numeric suffixes attempted before
// buildName gives up. 99 is comfortably above any plausible real-world need.
const maxCollisionSuffix = 99

// buildName composes "ccw-<owner>-<repo>-<shorthash>" with normalization,
// suffixing "-2", "-3", ... when the candidate is in `taken`. Returns an
// error if any input segment is empty after normalization or no slot is
// available within maxCollisionSuffix.
func buildName(owner, repo, shorthash string, taken map[string]bool) (string, error) {
 o := normalize(owner)
 r := normalize(repo)
 h := normalize(shorthash)
 if o == "" {
  return "", fmt.Errorf("buildName: owner is empty after normalization (input %q)", owner)
 }
 if r == "" {
  return "", fmt.Errorf("buildName: repo is empty after normalization (input %q)", repo)
 }
 if h == "" {
  return "", fmt.Errorf("buildName: shorthash is empty after normalization (input %q)", shorthash)
 }
 base := fmt.Sprintf("ccw-%s-%s-%s", o, r, h)
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
```

- [ ] **Step 4: テスト pass を確認**

Run: `go test ./internal/namegen/ -v`
Expected: 全テスト PASS

- [ ] **Step 5: コミット**

```bash
git add internal/namegen/namegen.go internal/namegen/namegen_test.go
git commit -m "feat(namegen): add buildName with -N collision suffixing (cap 99)"
```

---

## Task 7: `namegen.Generate` を実装（gitx 連携 + taken 検出）

**Files:**

- Modify: `internal/namegen/namegen.go`
- Modify: `internal/namegen/namegen_test.go`

- [ ] **Step 1: 失敗テストを追加**

`internal/namegen/namegen_test.go` の末尾に追記:

```go
func TestGenerate_HappyPath(t *testing.T) {
 withFakes(t, fakes{
  origin:    "git@github.com:tqer39/ccw-cli.git",
  branch:    "main",
  shorthash: "a3f2b1",
 })
 got, err := Generate("/fake/repo")
 if err != nil {
  t.Fatalf("Generate: %v", err)
 }
 if got != "ccw-tqer39-ccw-cli-a3f2b1" {
  t.Errorf("Generate = %q, want %q", got, "ccw-tqer39-ccw-cli-a3f2b1")
 }
}

func TestGenerate_NoOriginFallback(t *testing.T) {
 withFakes(t, fakes{
  origin:    "",
  branch:    "main",
  shorthash: "a3f2b1",
 })
 got, err := Generate("/tmp/projects/myrepo")
 if err != nil {
  t.Fatalf("Generate: %v", err)
 }
 if got != "ccw-local-myrepo-a3f2b1" {
  t.Errorf("Generate = %q, want %q", got, "ccw-local-myrepo-a3f2b1")
 }
}

func TestGenerate_DefaultBranchError(t *testing.T) {
 withFakes(t, fakes{
  origin:      "git@github.com:tqer39/ccw-cli.git",
  branchError: true,
 })
 if _, err := Generate("/fake/repo"); err == nil {
  t.Fatal("Generate with default-branch error: want error, got nil")
 }
}

func TestGenerate_CollisionWithExistingDir(t *testing.T) {
 repo := t.TempDir()
 mustMkdir(t, repo, ".claude/worktrees/ccw-tqer39-ccw-cli-a3f2b1")
 withFakes(t, fakes{
  origin:    "git@github.com:tqer39/ccw-cli.git",
  branch:    "main",
  shorthash: "a3f2b1",
 })
 got, err := Generate(repo)
 if err != nil {
  t.Fatalf("Generate: %v", err)
 }
 if got != "ccw-tqer39-ccw-cli-a3f2b1-2" {
  t.Errorf("Generate = %q, want %q", got, "ccw-tqer39-ccw-cli-a3f2b1-2")
 }
}

// withFakes swaps namegen's gitx hooks for the duration of the test.
type fakes struct {
 origin       string
 branch       string
 branchError  bool
 shorthash    string
 shorthashErr bool
}

func withFakes(t *testing.T, f fakes) {
 t.Helper()
 origOrigin := originURLFn
 origBranch := defaultBranchFn
 origHash := shortHashFn
 t.Cleanup(func() {
  originURLFn = origOrigin
  defaultBranchFn = origBranch
  shortHashFn = origHash
 })
 originURLFn = func(string) (string, error) { return f.origin, nil }
 defaultBranchFn = func(string) (string, error) {
  if f.branchError {
   return "", fmt.Errorf("fake: no default branch")
  }
  return f.branch, nil
 }
 shortHashFn = func(string, string, int) (string, error) {
  if f.shorthashErr {
   return "", fmt.Errorf("fake: no commits")
  }
  return f.shorthash, nil
 }
}

func mustMkdir(t *testing.T, root, rel string) {
 t.Helper()
 p := filepath.Join(root, rel)
 if err := os.MkdirAll(p, 0o755); err != nil {
  t.Fatalf("mkdir %s: %v", p, err)
 }
}
```

`namegen_test.go` の import を以下に拡張:

```go
import (
 "fmt"
 "os"
 "path/filepath"
 "strconv"
 "testing"
)
```

- [ ] **Step 2: 失敗を確認**

Run: `go test ./internal/namegen/ -run TestGenerate -v`
Expected: 既存の Generate stub が `not yet implemented` を返すので 4 case とも FAIL

- [ ] **Step 3: 実装**

`internal/namegen/namegen.go` の `Generate` 関数を以下で置換:

```go
// Generate returns a deterministic worktree name of the form
// "ccw-<owner>-<repo>-<shorthash6>" for the repository at mainRepo.
// When `origin` is unset, owner becomes "local" and repo is the basename
// of mainRepo. Numeric "-N" suffixes are appended on collision (cap: 99).
func Generate(mainRepo string) (string, error) {
 owner, repo, err := resolveOwnerRepo(mainRepo)
 if err != nil {
  return "", err
 }
 branch, err := defaultBranchFn(mainRepo)
 if err != nil {
  return "", fmt.Errorf("default branch: %w", err)
 }
 shorthash, err := shortHashFn(mainRepo, branch, 6)
 if err != nil {
  return "", fmt.Errorf("short hash: %w", err)
 }
 taken, err := takenNames(mainRepo)
 if err != nil {
  return "", err
 }
 return buildName(owner, repo, shorthash, taken)
}

func resolveOwnerRepo(mainRepo string) (string, string, error) {
 url, err := originURLFn(mainRepo)
 if err != nil {
  return "", "", fmt.Errorf("origin url: %w", err)
 }
 if url == "" {
  return "local", filepath.Base(mainRepo), nil
 }
 return gitx.ParseOriginURL(url)
}

// takenNames returns the set of worktree directory names already present
// under <mainRepo>/.claude/worktrees/. Missing dir is treated as empty set.
func takenNames(mainRepo string) (map[string]bool, error) {
 dir := filepath.Join(mainRepo, ".claude", "worktrees")
 entries, err := os.ReadDir(dir)
 if err != nil {
  if os.IsNotExist(err) {
   return map[string]bool{}, nil
  }
  return nil, fmt.Errorf("read worktrees dir: %w", err)
 }
 out := make(map[string]bool, len(entries))
 for _, e := range entries {
  if e.IsDir() {
   out[e.Name()] = true
  }
 }
 return out, nil
}
```

`namegen.go` の import を `os` 込みに更新:

```go
import (
 "fmt"
 "os"
 "path/filepath"
 "regexp"
 "strings"

 "github.com/tqer39/ccw-cli/internal/gitx"
)
```

(stub の `_ = filepath.Base` 行は削除)

- [ ] **Step 4: テスト pass を確認**

Run: `go test ./internal/namegen/ -v`
Expected: 全テスト PASS

- [ ] **Step 5: コミット**

```bash
git add internal/namegen/namegen.go internal/namegen/namegen_test.go
git commit -m "feat(namegen): wire Generate to gitx + filesystem collision check"
```

---

## Task 8: `cmd/ccw/main.go` を新 API に追従

**Files:**

- Modify: `cmd/ccw/main.go`

- [ ] **Step 1: ビルドが現状壊れていることを確認**

Run: `go build ./cmd/ccw`
Expected: `not enough arguments in call to namegen.Generate` のような 2 件のコンパイルエラー（L75, L98）

- [ ] **Step 2: `flags.NewWorktree` 経路を修正**

`cmd/ccw/main.go` の以下の箇所を編集:

old:

```go
 if flags.NewWorktree {
  name := namegen.Generate()
  code, err := claude.LaunchNew(mainRepo, name, preamble, flags.Passthrough)
```

new:

```go
 if flags.NewWorktree {
  name, err := namegen.Generate(mainRepo)
  if err != nil {
   ui.Error("generate worktree name: %v", err)
   return 1
  }
  code, err := claude.LaunchNew(mainRepo, name, preamble, flags.Passthrough)
```

- [ ] **Step 3: picker の `ActionNew` 経路を修正**

old:

```go
  case picker.ActionNew:
   name := namegen.Generate()
   code, err := claude.LaunchNew(mainRepo, name, "", passthrough)
```

new:

```go
  case picker.ActionNew:
   name, err := namegen.Generate(mainRepo)
   if err != nil {
    ui.Error("generate worktree name: %v", err)
    return 1
   }
   code, err := claude.LaunchNew(mainRepo, name, "", passthrough)
```

- [ ] **Step 4: ビルドとテストが通ることを確認**

Run: `go build ./...`
Expected: success

Run: `go test ./...`
Expected: 既存テストすべて PASS（main.go は直接の unit test を持たないため、ビルド成功 + 周辺パッケージテストで担保）

- [ ] **Step 5: コミット**

```bash
git add cmd/ccw/main.go
git commit -m "feat(ccw): adopt namegen.Generate(mainRepo) signature with error handling"
```

---

## Task 9: README の Naming convention セクションを更新（EN / JA 同期）

**Files:**

- Modify: `README.md`
- Modify: `docs/README.ja.md`

- [ ] **Step 1: EN README の現状確認**

Run: `sed -n '86,95p' README.md`
Expected:

```text
### Naming convention

When ccw creates a new worktree, the worktree directory and the Claude Code session name are kept 1:1:

- Directory: `<repo>/.claude/worktrees/<name>/`
- Branch: `worktree-<name>`
- Session name: `<name>` (set via `claude -n <name>`)

`<name>` is generated like `quick-falcon-7bd2`. Renaming the session manually with `/rename` is fine — ccw does not track it, and `--continue` keys off the working directory so conversation restore is unaffected.
```

- [ ] **Step 2: EN README を編集**

`README.md` の上記ブロック末尾段落を以下で置換:

old:

```md
`<name>` is generated like `quick-falcon-7bd2`. Renaming the session manually with `/rename` is fine — ccw does not track it, and `--continue` keys off the working directory so conversation restore is unaffected.
```

new:

```md
`<name>` is generated as `ccw-<owner>-<repo>-<shorthash6>` (e.g. `ccw-tqer39-ccw-cli-a3f2b1`). `<owner>` / `<repo>` come from the `origin` remote URL; `<shorthash6>` is the 6-char short SHA of the local default branch tip at creation time. When `origin` is unset, `<owner>` becomes `local` and `<repo>` is the directory basename. Duplicate names are disambiguated with `-2`, `-3`, … Renaming the session manually with `/rename` is fine — ccw does not track it, and `--continue` keys off the working directory so conversation restore is unaffected.
```

- [ ] **Step 3: JA README の対応箇所を確認**

Run: `grep -n "quick-falcon" docs/README.ja.md`
Expected: 1 行マッチ（バージョンによって行番号は異なる）

該当ブロックを `sed` で前後 5 行ほど確認:

```bash
LINE=$(grep -n "quick-falcon" docs/README.ja.md | head -1 | cut -d: -f1)
sed -n "$((LINE-5)),$((LINE+2))p" docs/README.ja.md
```

- [ ] **Step 4: JA README を編集**

`docs/README.ja.md` の `quick-falcon-7bd2` を含む段落を以下で置換:

new（既存の周辺文を残す形で、`<name>` の生成ルール部分のみ書き換え）:

```md
`<name>` は `ccw-<owner>-<repo>-<shorthash6>`（例: `ccw-tqer39-ccw-cli-a3f2b1`）形式で生成されます。`<owner>` / `<repo>` は `origin` remote の URL から抽出、`<shorthash6>` は作成時点のローカル default branch tip の short SHA です。`origin` が未設定の場合は `<owner>` が `local`、`<repo>` がディレクトリ basename になります。同名衝突は `-2`, `-3`, … で回避します。`/rename` で手動改名しても ccw 側は追跡しないため問題ありません（`--continue` は作業ディレクトリ基準で会話を復元します）。
```

実際の置換は Edit ツールで `quick-falcon-7bd2` を含む 1 段落全体を `old_string` に取り、上記 `new_string` に置き換える。

- [ ] **Step 5: lint と読み合わせ**

Run: `lefthook run pre-commit --all-files`
Expected: markdownlint / cspell / textlint いずれも error なし（warning は許容）

もし `cspell` が `shorthash` などを未知語として落とす場合は `.cspell.json` 等プロジェクト辞書に追記（既存のパターンに従う）。

- [ ] **Step 6: コミット**

```bash
git add README.md docs/README.ja.md
git commit -m "docs: update Naming convention section for ccw-<owner>-<repo>-<shorthash6>"
```

---

## Task 10: 全体スモーク

**Files:**

- なし（実行確認のみ）

- [ ] **Step 1: 全テスト実行**

Run: `go test ./...`
Expected: PASS

Run: `go vet ./...`
Expected: clean

- [ ] **Step 2: 実バイナリでの動作確認（ローカル）**

```bash
go build -o /tmp/ccw ./cmd/ccw
cd $(mktemp -d) && git init -q -b main && git commit --allow-empty -q -m init
git remote add origin git@github.com:tqer39/ccw-cli.git
/tmp/ccw -h | head
```

Expected: ヘルプが表示される（`namegen.Generate` 自体は worktree 作成時に呼ばれるので、ここではエラーがなければ十分）。

worktree を実際に切るには `claude` CLI が必要。`claude` がインストール済みであれば:

```bash
/tmp/ccw -n -- --print "echo hi"
```

を一時 repo 上で試し、`.claude/worktrees/ccw-tqer39-ccw-cli-<shorthash6>/` が作られることを確認する（claude 未インストール環境ではこの step はスキップ）。

- [ ] **Step 3: 最終コミット（必要なら）**

スモークで何も変更が出なければ追加コミット不要。

---

## Self-Review

**Spec coverage:**

| Spec 要件 | 担当タスク |
|---|---|
| `ccw-<owner>-<repo>-<shorthash6>` 形式 | Task 6, 7 |
| owner/repo の URL parse (SSH/HTTPS, `.git` 除去) | Task 1 |
| GitLab nested → 末尾 2 segment | Task 1 |
| 正規化（lowercase, `[a-z0-9-]`, dash 圧縮, trim） | Task 5 |
| default branch 解決 (`origin/HEAD` → `main` → `master`) | Task 3 |
| short SHA 6 文字 | Task 4 |
| 衝突回避 `-2..-99`、超過でエラー | Task 6 |
| `origin` 未設定時の `local` fallback | Task 7 |
| shorthash 取得失敗時はエラー | Task 7 (`shortHashFn` の error 伝播) |
| 既存 worktree との互換（マイグレーション無し） | 設計上 picker は文字列基準 → コード変更不要 |
| `cmd/ccw/main.go` のエラー対応 | Task 8 |
| README 更新 | Task 9 |
| TDD / hermetic test | Task 1〜7 で順守 |

すべて担当タスクあり。

**Placeholder scan:** "TBD", "TODO", "implement later", 抽象的 "handle edge cases" 等 — 含まれていないことを確認済み。

**Type consistency:** `originURLFn`, `defaultBranchFn`, `shortHashFn` のシグネチャが Task 5 で導入されたものと Task 7 でのフェイク差し替え・本実装呼び出しの双方で一致。`buildName(owner, repo, shorthash, taken)` のシグネチャは Task 6 と Task 7 で一致。`Generate(mainRepo string) (string, error)` は Task 5 / 7 / 8 で一致。
