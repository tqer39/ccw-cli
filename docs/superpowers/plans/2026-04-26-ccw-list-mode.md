# ccw `-L` 非対話 list モード Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** ccw に `-L` / `--list` フラグを追加し、`.claude/worktrees/` 配下を「worktree × git 状態 × PR × session」の集約として stdout に table または JSON で出力する非対話モードを実装する。Claude Code が ccw を呼んで状態把握する用途を主眼にする。

**Architecture:** 既存 `internal/worktree.List` の `Info` 構造体を 3 フィールド拡張し、新規 `internal/listmode` パッケージで Build / RenderJSON / RenderText を持つ。`internal/gitx` には `LastCommit` を、`internal/gh` には timeout 付き呼び出しを足す。`cmd/ccw` のフラグ解析と main 経路に list 分岐を追加する。

**Tech Stack:** Go 1.25, `pflag`, 標準 `testing`, `os/exec`. spec: `docs/superpowers/specs/2026-04-26-ccw-list-mode-design.md`

---

## File Structure

| 新規 / 変更 | パス | 役割 |
|---|---|---|
| 新規 | `internal/gitx/commit.go` | `LastCommit(wt) (sha, subject, time, error)` |
| 新規 | `internal/gitx/commit_test.go` | LastCommit のテスト |
| 変更 | `internal/worktree/worktree.go` | `Info` に `CreatedAt`, `LastCommit`, `SessionPath` フィールド追加。`CommitInfo` 型新設 |
| 変更 | `internal/worktree/worktree_test.go` | 新フィールドのテスト |
| 新規 | `internal/worktree/session_path.go` | `SessionLogPath(absPath) string` |
| 新規 | `internal/worktree/session_path_test.go` | 同上のテスト |
| 変更 | `internal/gh/gh.go` | `PRStatusWithTimeout(timeout, branches)` 追加 |
| 変更 | `internal/gh/gh_test.go` | timeout 版のテスト |
| 新規 | `internal/listmode/types.go` | `Output`, `RepoInfo`, `WorktreeEntry`, `Options`, `PRInfo`, `SessionInfo`, `CommitInfo` (or 再export) |
| 新規 | `internal/listmode/build.go` | `Build(mainRepo, opts) (*Output, []Warning, error)` |
| 新規 | `internal/listmode/build_test.go` | hermetic な fake 注入 + 統合テスト |
| 新規 | `internal/listmode/render_json.go` | `RenderJSON(out, w) error` |
| 新規 | `internal/listmode/render_json_test.go` | snapshot tests |
| 新規 | `internal/listmode/render_text.go` | `RenderText(out, w) error` |
| 新規 | `internal/listmode/render_text_test.go` | snapshot tests |
| 変更 | `internal/cli/parse.go` | `Flags` に `List`, `TargetDir`, `JSON`, `NoPR`, `NoSession` 追加。排他チェック実装 |
| 変更 | `internal/cli/parse_test.go` | 新フラグの parse テスト + 排他組合せエラーテスト |
| 変更 | `internal/cli/help.go` | usage 文字列に `-L` セクション追加 |
| 変更 | `cmd/ccw/main.go` | `flags.List` 分岐を追加。`-d` での起点切替に対応 |
| 新規 | `tests/list_mode_test.go` | binary レベルの統合テスト（fake gh / 0 worktree / 1 worktree） |
| 変更 | `README.md` | Usage / Features に `-L` を追記 |
| 変更 | `docs/README.ja.md` | 同上を日本語で同期 |

---

## Task 1: `internal/gitx.LastCommit`

worktree の HEAD コミットの short SHA / 件名 / author 時刻を返す薄い git ラッパ。

**Files:**

- Create: `internal/gitx/commit.go`
- Test: `internal/gitx/commit_test.go`

- [ ] **Step 1.1: Write the failing test**

`internal/gitx/commit_test.go`:

```go
package gitx

import (
 "testing"
 "time"
)

func TestLastCommit_HappyPath(t *testing.T) {
 dir := initRepo(t)
 mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "first commit")

 sha, subject, ts, err := LastCommit(dir)
 if err != nil {
  t.Fatalf("LastCommit: %v", err)
 }
 if len(sha) != 7 {
  t.Errorf("sha length = %d, want 7 (got %q)", len(sha), sha)
 }
 if subject != "first commit" {
  t.Errorf("subject = %q, want %q", subject, "first commit")
 }
 if ts.IsZero() {
  t.Errorf("time is zero")
 }
 if time.Since(ts) > time.Minute {
  t.Errorf("time too old: %v", ts)
 }
}

func TestLastCommit_EmptyRepo(t *testing.T) {
 dir := initRepo(t)
 if _, _, _, err := LastCommit(dir); err == nil {
  t.Fatal("LastCommit on empty repo: want error, got nil")
 }
}

func TestLastCommit_SubjectWithSpaces(t *testing.T) {
 dir := initRepo(t)
 mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "feat(x): add multi word subject")

 _, subject, _, err := LastCommit(dir)
 if err != nil {
  t.Fatalf("LastCommit: %v", err)
 }
 if subject != "feat(x): add multi word subject" {
  t.Errorf("subject = %q, want full multi-word", subject)
 }
}
```

- [ ] **Step 1.2: Run test to verify failure**

```bash
go test ./internal/gitx/ -run TestLastCommit -v
```

Expected: FAIL — `undefined: LastCommit`.

- [ ] **Step 1.3: Implement**

`internal/gitx/commit.go`:

```go
package gitx

import (
 "fmt"
 "strings"
 "time"
)

// LastCommit returns short SHA (7 chars), subject line, and author time of HEAD
// at the working tree wt. Errors when the repo has no commits.
func LastCommit(wt string) (string, string, time.Time, error) {
 // %h = abbreviated commit hash; %s = subject; %aI = author date ISO 8601 strict
 out, err := Output(wt, "log", "-1", "--no-color", "--format=%h%x1f%s%x1f%aI", "HEAD")
 if err != nil {
  return "", "", time.Time{}, fmt.Errorf("last commit: %w", err)
 }
 parts := strings.SplitN(strings.TrimSpace(out), "\x1f", 3)
 if len(parts) != 3 {
  return "", "", time.Time{}, fmt.Errorf("last commit: malformed output %q", out)
 }
 ts, err := time.Parse(time.RFC3339, parts[2])
 if err != nil {
  return "", "", time.Time{}, fmt.Errorf("last commit: parse time %q: %w", parts[2], err)
 }
 return parts[0], parts[1], ts, nil
}
```

- [ ] **Step 1.4: Run test to verify pass**

```bash
go test ./internal/gitx/ -run TestLastCommit -v
```

Expected: PASS (3 tests).

- [ ] **Step 1.5: Commit**

```bash
git add internal/gitx/commit.go internal/gitx/commit_test.go
git commit -m "feat(gitx): add LastCommit helper for HEAD short sha/subject/time"
```

---

## Task 2: `internal/worktree.SessionLogPath`

session log の絶対パスを返す（`HasSession` の path 版）。

**Files:**

- Create: `internal/worktree/session_path.go`
- Test: `internal/worktree/session_path_test.go`

- [ ] **Step 2.1: Write the failing test**

`internal/worktree/session_path_test.go`:

```go
package worktree

import (
 "os"
 "path/filepath"
 "testing"
)

func TestSessionLogPath_FoundReturnsFirst(t *testing.T) {
 home := t.TempDir()
 t.Setenv("HOME", home)

 wt := "/Users/foo/repo/.claude/worktrees/bar"
 dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(wt))
 if err := os.MkdirAll(dir, 0o755); err != nil {
  t.Fatal(err)
 }
 logPath := filepath.Join(dir, "abc123.jsonl")
 if err := os.WriteFile(logPath, []byte("{}"), 0o644); err != nil {
  t.Fatal(err)
 }

 if got := SessionLogPath(wt); got != logPath {
  t.Errorf("SessionLogPath = %q, want %q", got, logPath)
 }
}

func TestSessionLogPath_NotFoundReturnsEmpty(t *testing.T) {
 home := t.TempDir()
 t.Setenv("HOME", home)
 if got := SessionLogPath("/nonexistent"); got != "" {
  t.Errorf("SessionLogPath = %q, want empty", got)
 }
}

func TestSessionLogPath_HomeUnsetReturnsEmpty(t *testing.T) {
 t.Setenv("HOME", "")
 if got := SessionLogPath("/x"); got != "" {
  t.Errorf("SessionLogPath = %q, want empty", got)
 }
}
```

- [ ] **Step 2.2: Run test to verify failure**

```bash
go test ./internal/worktree/ -run TestSessionLogPath -v
```

Expected: FAIL — `undefined: SessionLogPath`.

- [ ] **Step 2.3: Implement**

`internal/worktree/session_path.go`:

```go
package worktree

import (
 "os"
 "path/filepath"
 "strings"
)

// SessionLogPath returns the absolute path to the first *.jsonl session log
// for absPath under ~/.claude/projects/<encoded>/, or "" if none.
// Mirrors HasSession's lookup so callers can use both consistently.
func SessionLogPath(absPath string) string {
 home := os.Getenv("HOME")
 if home == "" {
  return ""
 }
 dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(absPath))
 entries, err := os.ReadDir(dir)
 if err != nil {
  return ""
 }
 for _, e := range entries {
  if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
   return filepath.Join(dir, e.Name())
  }
 }
 return ""
}
```

- [ ] **Step 2.4: Run test to verify pass**

```bash
go test ./internal/worktree/ -run TestSessionLogPath -v
```

Expected: PASS (3 tests).

- [ ] **Step 2.5: Commit**

```bash
git add internal/worktree/session_path.go internal/worktree/session_path_test.go
git commit -m "feat(worktree): add SessionLogPath returning absolute jsonl path"
```

---

## Task 3: Extend `worktree.Info` with new fields

`Info` 構造体に `CreatedAt`, `LastCommit`, `SessionPath` を追加。`worktree.List` で値を埋める。

**Files:**

- Modify: `internal/worktree/worktree.go`
- Modify: `internal/worktree/worktree_test.go`

- [ ] **Step 3.1: Write the failing test (extend existing test)**

既存の `internal/worktree/worktree_test.go` には `initMainRepo(t) string` と `addWorktree(t, main, name) string` のヘルパがある。これを再利用する。

`internal/worktree/worktree_test.go` の最後に追加:

```go
func TestList_PopulatesCommitAndCreatedAt(t *testing.T) {
 main := initMainRepo(t)
 addWorktree(t, main, "newfields")

 infos, err := List(main)
 if err != nil {
  t.Fatalf("List: %v", err)
 }
 if len(infos) == 0 {
  t.Fatal("expected at least one worktree")
 }
 w := infos[0]
 if w.Status == StatusPrunable {
  t.Fatalf("setup invariant: worktree should not be prunable")
 }
 if w.CreatedAt == nil {
  t.Errorf("CreatedAt is nil for non-prunable entry")
 }
 if w.LastCommit == nil {
  t.Errorf("LastCommit is nil for non-prunable entry")
 } else if len(w.LastCommit.SHA) != 7 {
  t.Errorf("LastCommit.SHA len = %d, want 7", len(w.LastCommit.SHA))
 }
}

func TestList_SessionPathMatchesHasSession(t *testing.T) {
 home := t.TempDir()
 t.Setenv("HOME", home)

 main := initMainRepo(t)
 wt := addWorktree(t, main, "withsession")
 wtResolved, err := filepath.EvalSymlinks(wt)
 if err != nil {
  t.Fatal(err)
 }
 dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(wtResolved))
 if err := os.MkdirAll(dir, 0o755); err != nil {
  t.Fatal(err)
 }
 logPath := filepath.Join(dir, "abc.jsonl")
 if err := os.WriteFile(logPath, []byte("{}"), 0o644); err != nil {
  t.Fatal(err)
 }

 infos, err := List(main)
 if err != nil {
  t.Fatal(err)
 }
 var found bool
 for _, w := range infos {
  if w.Path != wtResolved {
   continue
  }
  found = true
  if !w.HasSession {
   t.Error("HasSession = false, want true")
  }
  if w.SessionPath != logPath {
   t.Errorf("SessionPath = %q, want %q", w.SessionPath, logPath)
  }
 }
 if !found {
  t.Fatalf("worktree %s not in List output", wtResolved)
 }
}

func TestList_PrunableLeavesNewFieldsZero(t *testing.T) {
 main := initMainRepo(t)
 wt := addWorktree(t, main, "prunablenew")
 if err := os.RemoveAll(wt); err != nil {
  t.Fatal(err)
 }

 infos, err := List(main)
 if err != nil {
  t.Fatal(err)
 }
 var found bool
 for _, w := range infos {
  if w.Status != StatusPrunable {
   continue
  }
  found = true
  if w.CreatedAt != nil {
   t.Errorf("CreatedAt = %v on prunable, want nil", w.CreatedAt)
  }
  if w.LastCommit != nil {
   t.Errorf("LastCommit = %v on prunable, want nil", w.LastCommit)
  }
  if w.SessionPath != "" {
   t.Errorf("SessionPath = %q, want empty", w.SessionPath)
  }
 }
 if !found {
  t.Fatal("no prunable worktree in fixture")
 }
}
```

- [ ] **Step 3.2: Run test to verify failure**

```bash
go test ./internal/worktree/ -run TestList -v
```

Expected: FAIL — `w.CreatedAt undefined` 等のコンパイルエラー。

- [ ] **Step 3.3: Modify `internal/worktree/worktree.go`**

`Info` 構造体と `List` 内の組立部を以下に置換:

```go
import (
 "fmt"
 "os"
 "strings"
 "time"

 "github.com/tqer39/ccw-cli/internal/gitx"
)

// CommitInfo summarizes the HEAD commit of a worktree.
type CommitInfo struct {
 SHA     string
 Subject string
 Time    time.Time
}

// Info is a ccw-managed worktree entry with its classified status and
// quantitative indicators (ahead/behind commits, dirty file count).
// AheadCount/BehindCount are meaningful for StatusPushed and StatusLocalOnly.
// DirtyCount is meaningful only when Status == StatusDirty.
// HasSession indicates whether a Claude Code session exists for this worktree.
// CreatedAt / LastCommit / SessionPath are populated for non-prunable entries
// when retrieval succeeds; otherwise nil / empty.
type Info struct {
 Path        string
 Branch      string
 Status      Status
 AheadCount  int
 BehindCount int
 DirtyCount  int
 HasSession  bool
 CreatedAt   *time.Time
 LastCommit  *CommitInfo
 SessionPath string
}
```

そして `List` 内、prunable でないエントリ確定後に以下を追加（`HasSession` を埋めている直後）:

```go
  info.HasSession = HasSession(e.Path)
  if info.HasSession {
   info.SessionPath = SessionLogPath(e.Path)
  }
  if st, err := os.Stat(e.Path); err == nil {
   t := st.ModTime()
   info.CreatedAt = &t
  }
  if sha, subject, ts, err := gitx.LastCommit(e.Path); err == nil {
   info.LastCommit = &CommitInfo{SHA: sha, Subject: subject, Time: ts}
  }
  result = append(result, info)
```

(既存の `result = append(result, info)` 行を削除して上記で置き換える)

- [ ] **Step 3.4: Run test to verify pass**

```bash
go test ./internal/worktree/ -v
```

Expected: 既存テスト + 新規 2 テスト全 PASS。

- [ ] **Step 3.5: Run full unit suite to confirm picker still compiles**

```bash
go vet ./...
go build ./...
```

Expected: no errors.

- [ ] **Step 3.6: Commit**

```bash
git add internal/worktree/worktree.go internal/worktree/worktree_test.go
git commit -m "feat(worktree): expose CreatedAt/LastCommit/SessionPath on Info"
```

---

## Task 4: `gh.PRStatusWithTimeout`

list モードでは Claude が無限待機しないよう gh 呼び出しに 5s timeout を入れる。既存 `PRStatus` は picker 用に温存し、新関数を追加する。

**Files:**

- Modify: `internal/gh/gh.go`
- Modify: `internal/gh/gh_test.go`

- [ ] **Step 4.1: Read existing test setup**

```bash
go test ./internal/gh/ -run TestPRStatus -v
```

(failure を確認するためでなく、既存パターン把握)

- [ ] **Step 4.2: Write the failing test**

`internal/gh/gh_test.go` に追加:

```go
func TestPRStatusWithTimeout_Success(t *testing.T) {
 r := fakeRunner{json: `[{"number":1,"title":"x","state":"OPEN","headRefName":"feat/a"}]`}
 got, err := PRStatusWithTimeout(r, 1*time.Second, []string{"feat/a"})
 if err != nil {
  t.Fatalf("PRStatusWithTimeout: %v", err)
 }
 if got["feat/a"].Number != 1 {
  t.Errorf("got = %+v", got)
 }
}

func TestPRStatusWithTimeout_RunnerErrorReturnsError(t *testing.T) {
 r := fakeRunner{err: errors.New("boom")}
 if _, err := PRStatusWithTimeout(r, 1*time.Second, []string{"x"}); err == nil {
  t.Fatal("want error, got nil")
 }
}
```

> 既存 `gh_test.go` に `fakeRunner` の定義があるか確認。無ければ次のように同ファイル冒頭に追加:

```go
type fakeRunner struct {
 json string
 err  error
}

func (f fakeRunner) LookPath() error      { return nil }
func (f fakeRunner) AuthStatus() error    { return nil }
func (f fakeRunner) PRListJSON() (string, error) {
 if f.err != nil {
  return "", f.err
 }
 return f.json, nil
}
```

import に `"errors"` と `"time"` を追加。

- [ ] **Step 4.3: Run test to verify failure**

```bash
go test ./internal/gh/ -run TestPRStatusWithTimeout -v
```

Expected: FAIL — `undefined: PRStatusWithTimeout`.

- [ ] **Step 4.4: Implement**

`internal/gh/gh.go` の末尾に追加:

```go
// PRStatusWithTimeout wraps PRStatusWith with a deadline. Currently the timeout
// only bounds Runner.PRListJSON's duration when it honors context; the default
// runner does not, so we run it in a goroutine and abandon on timeout.
func PRStatusWithTimeout(r Runner, timeout time.Duration, branches []string) (map[string]PRInfo, error) {
 type result struct {
  m   map[string]PRInfo
  err error
 }
 ch := make(chan result, 1)
 go func() {
  m, err := PRStatusWith(r, branches)
  ch <- result{m, err}
 }()
 select {
 case res := <-ch:
  return res.m, res.err
 case <-time.After(timeout):
  return nil, fmt.Errorf("gh pr list: timeout after %s", timeout)
 }
}
```

import に `"time"` を追加。

- [ ] **Step 4.5: Run test to verify pass**

```bash
go test ./internal/gh/ -v
```

Expected: PASS (新規 2 テスト + 既存全テスト).

- [ ] **Step 4.6: Commit**

```bash
git add internal/gh/gh.go internal/gh/gh_test.go
git commit -m "feat(gh): add PRStatusWithTimeout for non-interactive callers"
```

---

## Task 5: Define `internal/listmode` types

新規パッケージの型を先に固める。テストはこの段階では型の Marshal だけ。

**Files:**

- Create: `internal/listmode/types.go`
- Create: `internal/listmode/types_test.go`

- [ ] **Step 5.1: Write the failing test**

`internal/listmode/types_test.go`:

```go
package listmode

import (
 "encoding/json"
 "strings"
 "testing"
 "time"
)

func TestOutput_JSONShape(t *testing.T) {
 ts := time.Date(2026, 4, 26, 4, 28, 0, 0, time.UTC)
 out := Output{
  Version: 1,
  Repo: RepoInfo{
   Owner:         "tqer39",
   Name:          "ccw-cli",
   DefaultBranch: "main",
   MainPath:      "/abs",
  },
  Worktrees: []WorktreeEntry{{
   Name:          "ccw-foo",
   Path:          "/abs/.claude/worktrees/ccw-foo",
   Branch:        "worktree-ccw-foo",
   Status:        "pushed",
   Ahead:         0,
   Behind:        0,
   Dirty:         false,
   DefaultBranch: "main",
   CreatedAt:     &ts,
   LastCommit: &CommitInfo{
    SHA:     "9d3dc6e",
    Subject: "feat: x",
    Time:    ts,
   },
   PR: &PRInfo{
    State:  "OPEN",
    Number: 42,
    URL:    "https://github.com/tqer39/ccw-cli/pull/42",
    Title:  "feat: ...",
   },
   Session: SessionInfo{Exists: true, LogPath: stringPtr("/log.jsonl")},
  }},
 }

 b, err := json.Marshal(out)
 if err != nil {
  t.Fatalf("Marshal: %v", err)
 }
 s := string(b)
 for _, want := range []string{
  `"version":1`,
  `"owner":"tqer39"`,
  `"default_branch":"main"`,
  `"worktrees":[`,
  `"status":"pushed"`,
  `"ahead":0`,
  `"dirty":false`,
  `"pr":{`,
  `"state":"OPEN"`,
  `"session":{"exists":true,"log_path":"/log.jsonl"}`,
 } {
  if !strings.Contains(s, want) {
   t.Errorf("JSON missing %q\nfull: %s", want, s)
  }
 }
}

func TestOutput_EmptyWorktreesIsArrayNotNull(t *testing.T) {
 out := Output{Version: 1, Repo: RepoInfo{}, Worktrees: []WorktreeEntry{}}
 b, _ := json.Marshal(out)
 if !strings.Contains(string(b), `"worktrees":[]`) {
  t.Errorf("want empty array, got %s", string(b))
 }
}

func TestPRInfoNullsCleanly(t *testing.T) {
 w := WorktreeEntry{Name: "x", PR: nil}
 b, _ := json.Marshal(w)
 if !strings.Contains(string(b), `"pr":null`) {
  t.Errorf("want pr:null, got %s", string(b))
 }
}

func stringPtr(s string) *string { return &s }
```

- [ ] **Step 5.2: Run test to verify failure**

```bash
go test ./internal/listmode/ -v
```

Expected: FAIL — package not found / undefined types.

- [ ] **Step 5.3: Implement types**

`internal/listmode/types.go`:

```go
// Package listmode produces machine-readable summaries of ccw-managed
// worktrees for the `ccw -L` non-interactive list command.
package listmode

import "time"

// Output is the top-level JSON shape, version-pinned to 1.
type Output struct {
 Version   int             `json:"version"`
 Repo      RepoInfo        `json:"repo"`
 Worktrees []WorktreeEntry `json:"worktrees"`
}

// RepoInfo describes the main repository the listing came from.
type RepoInfo struct {
 Owner         string `json:"owner"`
 Name          string `json:"name"`
 DefaultBranch string `json:"default_branch"`
 MainPath      string `json:"main_path"`
}

// WorktreeEntry is one ccw-managed worktree.
type WorktreeEntry struct {
 Name          string      `json:"name"`
 Path          string      `json:"path"`
 Branch        string      `json:"branch"`
 Status        string      `json:"status"` // pushed | local-only | dirty | prunable
 Ahead         int         `json:"ahead"`
 Behind        int         `json:"behind"`
 Dirty         bool        `json:"dirty"`
 DefaultBranch string      `json:"default_branch"`
 CreatedAt     *time.Time  `json:"created_at"`
 LastCommit    *CommitInfo `json:"last_commit"`
 PR            *PRInfo     `json:"pr"`
 Session       SessionInfo `json:"session"`
}

// CommitInfo describes a worktree's HEAD commit.
type CommitInfo struct {
 SHA     string    `json:"sha"`
 Subject string    `json:"subject"`
 Time    time.Time `json:"time"`
}

// PRInfo is the GitHub pull request associated with a worktree's branch.
type PRInfo struct {
 State  string `json:"state"`
 Number int    `json:"number"`
 URL    string `json:"url"`
 Title  string `json:"title"`
}

// SessionInfo summarizes Claude Code session presence.
type SessionInfo struct {
 Exists  bool    `json:"exists"`
 LogPath *string `json:"log_path"`
}

// Options control optional data gathering during Build.
type Options struct {
 NoPR      bool
 NoSession bool
}

// Warning is a non-fatal diagnostic emitted during Build (printed to stderr
// by the caller).
type Warning struct {
 Message string
}
```

- [ ] **Step 5.4: Run test to verify pass**

```bash
go test ./internal/listmode/ -v
```

Expected: PASS (3 tests).

- [ ] **Step 5.5: Commit**

```bash
git add internal/listmode/types.go internal/listmode/types_test.go
git commit -m "feat(listmode): define Output / RepoInfo / WorktreeEntry / SessionInfo types"
```

---

## Task 6: `listmode.Build`

`Build(mainRepo, opts)` で Output を組立てる。fake 注入のため `gh` 呼び出し関数と `worktree.List` 呼び出し関数を変数化する。

**Files:**

- Create: `internal/listmode/build.go`
- Create: `internal/listmode/build_test.go`

- [ ] **Step 6.1: Write the failing test**

`internal/listmode/build_test.go`:

```go
package listmode

import (
 "errors"
 "testing"
 "time"

 "github.com/tqer39/ccw-cli/internal/gh"
 "github.com/tqer39/ccw-cli/internal/worktree"
)

func TestBuild_HappyPath(t *testing.T) {
 ts := time.Now()
 wts := []worktree.Info{{
  Path:        "/abs/.claude/worktrees/ccw-x",
  Branch:      "worktree-ccw-x",
  Status:      worktree.StatusPushed,
  AheadCount:  0,
  BehindCount: 0,
  HasSession:  true,
  SessionPath: "/log.jsonl",
  CreatedAt:   &ts,
  LastCommit:  &worktree.CommitInfo{SHA: "abc1234", Subject: "init", Time: ts},
 }}
 prs := map[string]gh.PRInfo{
  "worktree-ccw-x": {Number: 42, Title: "feat", State: "OPEN"},
 }

 b := Builder{
  ListWorktrees: func(string) ([]worktree.Info, error) { return wts, nil },
  ResolveRepo:   func(string) (RepoInfo, error) { return RepoInfo{Owner: "tqer39", Name: "ccw-cli", DefaultBranch: "main", MainPath: "/abs"}, nil },
  FetchPRs:      func([]string) (map[string]gh.PRInfo, error) { return prs, nil },
  GhAvailable:   func() bool { return true },
 }
 out, warns, err := b.Build("/abs", Options{})
 if err != nil {
  t.Fatalf("Build: %v", err)
 }
 if len(warns) != 0 {
  t.Errorf("warns = %v, want none", warns)
 }
 if out.Version != 1 {
  t.Errorf("Version = %d", out.Version)
 }
 if len(out.Worktrees) != 1 {
  t.Fatalf("Worktrees len = %d", len(out.Worktrees))
 }
 w := out.Worktrees[0]
 if w.Name != "ccw-x" {
  t.Errorf("Name = %q", w.Name)
 }
 if w.Status != "pushed" {
  t.Errorf("Status = %q", w.Status)
 }
 if w.PR == nil || w.PR.Number != 42 {
  t.Errorf("PR = %+v", w.PR)
 }
 if w.PR.URL != "https://github.com/tqer39/ccw-cli/pull/42" {
  t.Errorf("PR.URL = %q", w.PR.URL)
 }
 if !w.Session.Exists || w.Session.LogPath == nil || *w.Session.LogPath != "/log.jsonl" {
  t.Errorf("Session = %+v", w.Session)
 }
}

func TestBuild_GhUnavailable_PRNullPlusWarning(t *testing.T) {
 b := Builder{
  ListWorktrees: func(string) ([]worktree.Info, error) {
   return []worktree.Info{{Path: "/a/.claude/worktrees/x", Branch: "b", Status: worktree.StatusPushed}}, nil
  },
  ResolveRepo: func(string) (RepoInfo, error) { return RepoInfo{Owner: "o", Name: "r", MainPath: "/a"}, nil },
  FetchPRs:    func([]string) (map[string]gh.PRInfo, error) { return nil, nil },
  GhAvailable: func() bool { return false },
 }
 out, warns, err := b.Build("/a", Options{})
 if err != nil {
  t.Fatalf("Build: %v", err)
 }
 if len(warns) != 1 {
  t.Errorf("warns = %v, want exactly 1", warns)
 }
 if out.Worktrees[0].PR != nil {
  t.Errorf("PR = %+v, want nil", out.Worktrees[0].PR)
 }
}

func TestBuild_NoPROptionSkipsFetch(t *testing.T) {
 called := false
 b := Builder{
  ListWorktrees: func(string) ([]worktree.Info, error) { return nil, nil },
  ResolveRepo:   func(string) (RepoInfo, error) { return RepoInfo{}, nil },
  FetchPRs:      func([]string) (map[string]gh.PRInfo, error) { called = true; return nil, nil },
  GhAvailable:   func() bool { return true },
 }
 if _, _, err := b.Build("/a", Options{NoPR: true}); err != nil {
  t.Fatalf("Build: %v", err)
 }
 if called {
  t.Error("FetchPRs called despite NoPR=true")
 }
}

func TestBuild_NoSessionOptionForcesEmpty(t *testing.T) {
 b := Builder{
  ListWorktrees: func(string) ([]worktree.Info, error) {
   return []worktree.Info{{Path: "/a/.claude/worktrees/x", Branch: "b", Status: worktree.StatusPushed, HasSession: true, SessionPath: "/p"}}, nil
  },
  ResolveRepo: func(string) (RepoInfo, error) { return RepoInfo{}, nil },
  FetchPRs:    func([]string) (map[string]gh.PRInfo, error) { return nil, nil },
  GhAvailable: func() bool { return true },
 }
 out, _, _ := b.Build("/a", Options{NoSession: true})
 if out.Worktrees[0].Session.Exists {
  t.Error("Session.Exists = true despite NoSession=true")
 }
}

func TestBuild_PRFetchErrorBecomesWarning(t *testing.T) {
 b := Builder{
  ListWorktrees: func(string) ([]worktree.Info, error) {
   return []worktree.Info{{Path: "/a/.claude/worktrees/x", Branch: "b", Status: worktree.StatusPushed}}, nil
  },
  ResolveRepo: func(string) (RepoInfo, error) { return RepoInfo{Owner: "o", Name: "r"}, nil },
  FetchPRs:    func([]string) (map[string]gh.PRInfo, error) { return nil, errors.New("rate limit") },
  GhAvailable: func() bool { return true },
 }
 out, warns, err := b.Build("/a", Options{})
 if err != nil {
  t.Fatalf("Build: %v", err)
 }
 if len(warns) != 1 {
  t.Errorf("want 1 warning, got %v", warns)
 }
 if out.Worktrees[0].PR != nil {
  t.Errorf("PR not nil on fetch failure")
 }
}

func TestBuild_PrunableSkipsAheadAndCommit(t *testing.T) {
 b := Builder{
  ListWorktrees: func(string) ([]worktree.Info, error) {
   return []worktree.Info{{Path: "/a/.claude/worktrees/p", Branch: "b", Status: worktree.StatusPrunable}}, nil
  },
  ResolveRepo: func(string) (RepoInfo, error) { return RepoInfo{}, nil },
  FetchPRs:    func([]string) (map[string]gh.PRInfo, error) { return nil, nil },
  GhAvailable: func() bool { return true },
 }
 out, _, _ := b.Build("/a", Options{})
 w := out.Worktrees[0]
 if w.Status != "prunable" {
  t.Errorf("Status = %q", w.Status)
 }
 if w.LastCommit != nil || w.CreatedAt != nil {
  t.Errorf("commit/created should be nil for prunable, got %+v %+v", w.LastCommit, w.CreatedAt)
 }
}
```

- [ ] **Step 6.2: Run test to verify failure**

```bash
go test ./internal/listmode/ -run TestBuild -v
```

Expected: FAIL — `undefined: Builder` 等。

- [ ] **Step 6.3: Implement**

`internal/listmode/build.go`:

```go
package listmode

import (
 "fmt"
 "path/filepath"

 "github.com/tqer39/ccw-cli/internal/gh"
 "github.com/tqer39/ccw-cli/internal/worktree"
)

// Builder bundles the dependencies Build needs. Tests inject fakes; production
// callers use NewBuilder() for the default real wiring.
type Builder struct {
 ListWorktrees func(mainRepo string) ([]worktree.Info, error)
 ResolveRepo   func(mainRepo string) (RepoInfo, error)
 FetchPRs      func(branches []string) (map[string]gh.PRInfo, error)
 GhAvailable   func() bool
}

// Build assembles an *Output from the given main repo. Returns warnings
// for non-fatal degradations (gh missing, PR fetch failures, etc).
func (b Builder) Build(mainRepo string, opts Options) (*Output, []Warning, error) {
 repo, err := b.ResolveRepo(mainRepo)
 if err != nil {
  return nil, nil, fmt.Errorf("resolve repo: %w", err)
 }

 infos, err := b.ListWorktrees(mainRepo)
 if err != nil {
  return nil, nil, fmt.Errorf("list worktrees: %w", err)
 }

 var warns []Warning
 prs := map[string]gh.PRInfo{}
 if !opts.NoPR {
  switch {
  case !b.GhAvailable():
   warns = append(warns, Warning{Message: "gh not available, PR info disabled"})
  default:
   branches := make([]string, 0, len(infos))
   for _, info := range infos {
    if info.Branch != "" {
     branches = append(branches, info.Branch)
    }
   }
   fetched, err := b.FetchPRs(branches)
   if err != nil {
    warns = append(warns, Warning{Message: fmt.Sprintf("gh pr fetch failed: %v", err)})
   } else {
    prs = fetched
   }
  }
 }

 out := &Output{
  Version:   1,
  Repo:      repo,
  Worktrees: make([]WorktreeEntry, 0, len(infos)),
 }
 for _, info := range infos {
  out.Worktrees = append(out.Worktrees, buildEntry(info, repo, prs, opts))
 }
 return out, warns, nil
}

func buildEntry(info worktree.Info, repo RepoInfo, prs map[string]gh.PRInfo, opts Options) WorktreeEntry {
 entry := WorktreeEntry{
  Name:          filepath.Base(info.Path),
  Path:          info.Path,
  Branch:        info.Branch,
  Status:        info.Status.String(),
  Ahead:         info.AheadCount,
  Behind:        info.BehindCount,
  Dirty:         info.Status == worktree.StatusDirty,
  DefaultBranch: repo.DefaultBranch,
  CreatedAt:     info.CreatedAt,
 }
 if info.LastCommit != nil {
  entry.LastCommit = &CommitInfo{
   SHA:     info.LastCommit.SHA,
   Subject: info.LastCommit.Subject,
   Time:    info.LastCommit.Time,
  }
 }
 if pr, ok := prs[info.Branch]; ok && info.Branch != "" {
  entry.PR = &PRInfo{
   State:  pr.State,
   Number: pr.Number,
   URL:    prURL(repo, pr.Number),
   Title:  pr.Title,
  }
 }
 if !opts.NoSession && info.HasSession {
  path := info.SessionPath
  entry.Session = SessionInfo{Exists: true, LogPath: &path}
 }
 return entry
}

func prURL(repo RepoInfo, number int) string {
 if repo.Owner == "" || repo.Name == "" {
  return ""
 }
 return fmt.Sprintf("https://github.com/%s/%s/pull/%d", repo.Owner, repo.Name, number)
}
```

- [ ] **Step 6.4: Run test to verify pass**

```bash
go test ./internal/listmode/ -v
```

Expected: PASS (Build 6 tests + types 3 tests).

- [ ] **Step 6.5: Commit**

```bash
git add internal/listmode/build.go internal/listmode/build_test.go
git commit -m "feat(listmode): implement Builder.Build with fail-soft PR/session handling"
```

---

## Task 7: `listmode.RenderJSON`

整形済み JSON を `io.Writer` に書き出す。

**Files:**

- Create: `internal/listmode/render_json.go`
- Create: `internal/listmode/render_json_test.go`

- [ ] **Step 7.1: Write the failing test**

`internal/listmode/render_json_test.go`:

```go
package listmode

import (
 "bytes"
 "encoding/json"
 "strings"
 "testing"
)

func TestRenderJSON_RoundTrip(t *testing.T) {
 out := &Output{
  Version:   1,
  Repo:      RepoInfo{Owner: "o", Name: "r", DefaultBranch: "main", MainPath: "/p"},
  Worktrees: []WorktreeEntry{{Name: "x", Path: "/p/.claude/worktrees/x", Status: "pushed"}},
 }
 var buf bytes.Buffer
 if err := RenderJSON(out, &buf); err != nil {
  t.Fatalf("RenderJSON: %v", err)
 }
 var got Output
 if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
  t.Fatalf("Unmarshal: %v", err)
 }
 if got.Version != 1 || got.Worktrees[0].Name != "x" {
  t.Errorf("round trip failed: %+v", got)
 }
}

func TestRenderJSON_Indented(t *testing.T) {
 out := &Output{Version: 1, Worktrees: []WorktreeEntry{}}
 var buf bytes.Buffer
 if err := RenderJSON(out, &buf); err != nil {
  t.Fatal(err)
 }
 if !strings.Contains(buf.String(), "\n") {
  t.Errorf("output not indented: %s", buf.String())
 }
}

func TestRenderJSON_TrailingNewline(t *testing.T) {
 var buf bytes.Buffer
 _ = RenderJSON(&Output{Version: 1, Worktrees: []WorktreeEntry{}}, &buf)
 if !strings.HasSuffix(buf.String(), "\n") {
  t.Errorf("missing trailing newline")
 }
}
```

- [ ] **Step 7.2: Run test to verify failure**

```bash
go test ./internal/listmode/ -run TestRenderJSON -v
```

Expected: FAIL — `undefined: RenderJSON`.

- [ ] **Step 7.3: Implement**

`internal/listmode/render_json.go`:

```go
package listmode

import (
 "encoding/json"
 "io"
)

// RenderJSON writes out as indented JSON followed by a trailing newline.
func RenderJSON(out *Output, w io.Writer) error {
 enc := json.NewEncoder(w)
 enc.SetIndent("", "  ")
 return enc.Encode(out) // Encode emits a trailing newline.
}
```

- [ ] **Step 7.4: Run test to verify pass**

```bash
go test ./internal/listmode/ -run TestRenderJSON -v
```

Expected: PASS (3 tests).

- [ ] **Step 7.5: Commit**

```bash
git add internal/listmode/render_json.go internal/listmode/render_json_test.go
git commit -m "feat(listmode): RenderJSON emits indented JSON"
```

---

## Task 8: `listmode.RenderText`

table 形式の text 出力。spec のフォーマット通り。

**Files:**

- Create: `internal/listmode/render_text.go`
- Create: `internal/listmode/render_text_test.go`

- [ ] **Step 8.1: Write the failing test**

`internal/listmode/render_text_test.go`:

```go
package listmode

import (
 "bytes"
 "strings"
 "testing"
)

func TestRenderText_Header(t *testing.T) {
 var buf bytes.Buffer
 if err := RenderText(&Output{Version: 1, Worktrees: []WorktreeEntry{}}, &buf); err != nil {
  t.Fatal(err)
 }
 first := strings.SplitN(buf.String(), "\n", 2)[0]
 for _, col := range []string{"NAME", "STATUS", "AHEAD/BEHIND", "PR", "SESSION", "BRANCH"} {
  if !strings.Contains(first, col) {
   t.Errorf("header missing %q\nheader: %s", col, first)
  }
 }
}

func TestRenderText_PushedRow(t *testing.T) {
 out := &Output{
  Version: 1,
  Worktrees: []WorktreeEntry{{
   Name:   "ccw-x",
   Branch: "worktree-ccw-x",
   Status: "pushed",
   Ahead:  0, Behind: 0,
   PR:      &PRInfo{Number: 42, State: "OPEN"},
   Session: SessionInfo{Exists: true},
  }},
 }
 var buf bytes.Buffer
 _ = RenderText(out, &buf)
 got := buf.String()
 for _, want := range []string{"ccw-x", "pushed", "0/0", "#42 OPEN", "RESUME", "worktree-ccw-x"} {
  if !strings.Contains(got, want) {
   t.Errorf("missing %q\nfull:\n%s", want, got)
  }
 }
}

func TestRenderText_NoSessionRendersNEW(t *testing.T) {
 out := &Output{Worktrees: []WorktreeEntry{{Name: "x", Status: "pushed", Session: SessionInfo{Exists: false}}}}
 var buf bytes.Buffer
 _ = RenderText(out, &buf)
 if !strings.Contains(buf.String(), "NEW") {
  t.Errorf("expected NEW, got: %s", buf.String())
 }
}

func TestRenderText_NoPRRendersDash(t *testing.T) {
 out := &Output{Worktrees: []WorktreeEntry{{Name: "x", Status: "pushed", PR: nil}}}
 var buf bytes.Buffer
 _ = RenderText(out, &buf)
 // PR セルは "-" 単独
 lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
 if len(lines) < 2 {
  t.Fatalf("not enough lines: %s", buf.String())
 }
 if !strings.Contains(lines[1], " - ") {
  t.Errorf("expected '-' in row, got: %s", lines[1])
 }
}

func TestRenderText_PrunableShowsDashAhead(t *testing.T) {
 out := &Output{Worktrees: []WorktreeEntry{{Name: "p", Status: "prunable"}}}
 var buf bytes.Buffer
 _ = RenderText(out, &buf)
 if !strings.Contains(buf.String(), "prunable") {
  t.Errorf("missing prunable: %s", buf.String())
 }
 // AHEAD/BEHIND セルが "-"（"0/0" でないこと）
 if strings.Contains(buf.String(), "0/0") {
  t.Errorf("prunable row should not show 0/0: %s", buf.String())
 }
}

func TestRenderText_NoANSIEscapes(t *testing.T) {
 out := &Output{Worktrees: []WorktreeEntry{{Name: "x", Status: "dirty"}}}
 var buf bytes.Buffer
 _ = RenderText(out, &buf)
 if strings.Contains(buf.String(), "\x1b[") {
  t.Errorf("ANSI escape leaked: %q", buf.String())
 }
}
```

- [ ] **Step 8.2: Run test to verify failure**

```bash
go test ./internal/listmode/ -run TestRenderText -v
```

Expected: FAIL — `undefined: RenderText`.

- [ ] **Step 8.3: Implement**

`internal/listmode/render_text.go`:

```go
package listmode

import (
 "fmt"
 "io"
 "text/tabwriter"
)

// RenderText writes a column-aligned, ANSI-free table to w.
// Columns: NAME / STATUS / AHEAD/BEHIND / PR / SESSION / BRANCH.
// Header is always written; trailing newline included.
func RenderText(out *Output, w io.Writer) error {
 tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
 if _, err := fmt.Fprintln(tw, "NAME\tSTATUS\tAHEAD/BEHIND\tPR\tSESSION\tBRANCH"); err != nil {
  return err
 }
 for _, e := range out.Worktrees {
  ab := fmt.Sprintf("%d/%d", e.Ahead, e.Behind)
  if e.Status == "prunable" {
   ab = "-"
  }
  pr := "-"
  if e.PR != nil {
   pr = fmt.Sprintf("#%d %s", e.PR.Number, e.PR.State)
  }
  session := "NEW"
  if e.Session.Exists {
   session = "RESUME"
  }
  if e.Status == "prunable" {
   session = "-"
  }
  if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
   e.Name, e.Status, ab, pr, session, e.Branch); err != nil {
   return err
  }
 }
 return tw.Flush()
}
```

- [ ] **Step 8.4: Run test to verify pass**

```bash
go test ./internal/listmode/ -v
```

Expected: PASS (全 listmode テスト).

- [ ] **Step 8.5: Commit**

```bash
git add internal/listmode/render_text.go internal/listmode/render_text_test.go
git commit -m "feat(listmode): RenderText emits column-aligned ANSI-free table"
```

---

## Task 9: CLI フラグ追加 (`-L`, `-d`, `--json`, `--no-pr`, `--no-session`)

`Flags` 構造体と排他チェックを実装。

**Files:**

- Modify: `internal/cli/parse.go`
- Modify: `internal/cli/parse_test.go`

- [ ] **Step 9.1: Write the failing tests**

`internal/cli/parse_test.go` の末尾に追加:

```go
func TestParse_ListShortFlag(t *testing.T) {
 f, err := Parse([]string{"-L"})
 if err != nil {
  t.Fatalf("Parse: %v", err)
 }
 if !f.List {
  t.Error("List = false, want true")
 }
}

func TestParse_ListLongFlag(t *testing.T) {
 f, _ := Parse([]string{"--list"})
 if !f.List {
  t.Error("List = false on --list")
 }
}

func TestParse_ListWithJSON(t *testing.T) {
 f, _ := Parse([]string{"-L", "--json"})
 if !f.List || !f.JSON {
  t.Errorf("List=%v JSON=%v", f.List, f.JSON)
 }
}

func TestParse_ListWithDir(t *testing.T) {
 f, _ := Parse([]string{"-L", "-d", "/tmp/repo"})
 if f.TargetDir != "/tmp/repo" {
  t.Errorf("TargetDir = %q", f.TargetDir)
 }
}

func TestParse_ListWithNoPRAndNoSession(t *testing.T) {
 f, _ := Parse([]string{"-L", "--no-pr", "--no-session"})
 if !f.NoPR || !f.NoSession {
  t.Errorf("NoPR=%v NoSession=%v", f.NoPR, f.NoSession)
 }
}

func TestParse_DirWithoutListErrors(t *testing.T) {
 if _, err := Parse([]string{"-d", "/x"}); err == nil {
  t.Fatal("want error: -d without -L")
 }
}

func TestParse_ListWithNewIsExclusive(t *testing.T) {
 if _, err := Parse([]string{"-L", "-n"}); err == nil {
  t.Fatal("want error: -L with -n")
 }
}

func TestParse_ListWithSuperpowersIsExclusive(t *testing.T) {
 if _, err := Parse([]string{"-L", "-s"}); err == nil {
  t.Fatal("want error: -L with -s")
 }
}

func TestParse_ListWithCleanAllIsExclusive(t *testing.T) {
 if _, err := Parse([]string{"-L", "--clean-all"}); err == nil {
  t.Fatal("want error: -L with --clean-all")
 }
}

func TestParse_ListWithPassthroughIsExclusive(t *testing.T) {
 if _, err := Parse([]string{"-L", "--", "--model", "x"}); err == nil {
  t.Fatal("want error: -L with -- passthrough")
 }
}
```

- [ ] **Step 9.2: Run test to verify failure**

```bash
go test ./internal/cli/ -run TestParse_List -v
```

Expected: FAIL — `f.List undefined` 等。

- [ ] **Step 9.3: Implement**

`internal/cli/parse.go` の `Flags` 構造体を以下に置換:

```go
// Flags is the parsed representation of ccw's command-line arguments.
type Flags struct {
 Help         bool
 Version      bool
 NewWorktree  bool
 Superpowers  bool
 CleanAll     bool
 StatusFilter string
 Force        bool
 DryRun       bool
 AssumeYes    bool
 List         bool
 TargetDir    string
 JSON         bool
 NoPR         bool
 NoSession    bool
 Passthrough  []string
}
```

`Parse` 内、既存 `BoolVarP` 群の直後（`AssumeYes` の下）に追加:

```go
 fs.BoolVarP(&f.List, "list", "L", false, "non-interactive list of ccw worktrees (text by default)")
 fs.StringVarP(&f.TargetDir, "dir", "d", "", "target directory for --list (defaults to cwd)")
 fs.BoolVar(&f.JSON, "json", false, "use JSON output for --list")
 fs.BoolVar(&f.NoPR, "no-pr", false, "skip gh PR lookup for --list")
 fs.BoolVar(&f.NoSession, "no-session", false, "skip session log lookup for --list")
```

`f.Passthrough = post` の前に排他チェックを追加:

```go
 if f.TargetDir != "" && !f.List {
  return Flags{}, fmt.Errorf("--dir/-d requires --list/-L")
 }
 if f.List {
  switch {
  case f.NewWorktree:
   return Flags{}, fmt.Errorf("--list cannot be combined with --new")
  case f.Superpowers:
   return Flags{}, fmt.Errorf("--list cannot be combined with --superpowers")
  case f.CleanAll:
   return Flags{}, fmt.Errorf("--list cannot be combined with --clean-all")
  case post != nil:
   return Flags{}, fmt.Errorf("--list does not accept passthrough args after --")
  }
 }
```

(`post` はここで参照可能。`splitAtDoubleDash` の戻り値)

- [ ] **Step 9.4: Run test to verify pass**

```bash
go test ./internal/cli/ -v
```

Expected: PASS (新規 10 + 既存全).

- [ ] **Step 9.5: Commit**

```bash
git add internal/cli/parse.go internal/cli/parse_test.go
git commit -m "feat(cli): add -L/--list, -d, --json, --no-pr, --no-session with exclusivity"
```

---

## Task 10: `--help` 文言更新

**Files:**

- Modify: `internal/cli/help.go`

- [ ] **Step 10.1: Modify usage**

`internal/cli/help.go` の `usage` 定数を以下に置換:

```go
const usage = `Usage: ccw [options] [-- <claude-args>...]

Options:
  -n, --new            Always start a new worktree (skip picker)
  -s, --superpowers    Inject superpowers preamble (implies -n)
  -v, --version        Show version
  -h, --help           Show this help

List mode (non-interactive):
  -L, --list           Print ccw worktrees and exit (text table)
  -d, --dir <path>     Target directory for --list (defaults to cwd)
      --json           Emit JSON instead of the text table
      --no-pr          Skip gh PR lookup
      --no-session     Skip session log lookup

Bulk delete:
      --clean-all        Bulk delete mode
      --status=<filter>  all | pushed | local-only | dirty (default: all)
      --force            Delete dirty worktrees with --force
      --dry-run          Print targets and exit
  -y, --yes              Skip confirmation prompts (--clean-all, -s plugin install)

Arguments after ` + "`--`" + ` are forwarded to ` + "`claude`" + ` verbatim.

Environment:
  NO_COLOR=1           Disable colored output
  CCW_DEBUG=1          Verbose debug logging

Exit codes:
  0  success
  1  user error / cancellation
  2  system error (git failure, etc.)
  *  passthrough from ` + "`claude`" + `

Repository: https://github.com/tqer39/ccw-cli
`
```

- [ ] **Step 10.2: Sanity build**

```bash
go build ./...
go test ./internal/cli/ -v
```

Expected: PASS.

- [ ] **Step 10.3: Commit**

```bash
git add internal/cli/help.go
git commit -m "docs(cli): document -L list mode in --help"
```

---

## Task 11: `cmd/ccw/main.go` で list 分岐を実装

**Files:**

- Modify: `cmd/ccw/main.go`

- [ ] **Step 11.1: Add import and dispatcher**

`cmd/ccw/main.go` の import に追加:

```go
 "github.com/tqer39/ccw-cli/internal/gh"
 "github.com/tqer39/ccw-cli/internal/listmode"
 "time"
```

`run` 関数内の冒頭、`mainRepo` 取得直後（既存の `if flags.CleanAll` の前）に list 分岐を挿入:

```go
 if flags.List {
  return runList(flags)
 }
```

`mainRepo` 解決を list 経路でも自前でやりたいので、`runList` は `flags` を直接受け取り `mainRepo` を内部で解決する。`run` の冒頭の `resolveMainRepo()` よりも前で分岐させる必要がある。よって配置を以下のように調整:

```go
func run(flags cli.Flags) int {
 if flags.List {
  return runList(flags)
 }

 mainRepo, err := resolveMainRepo()
 if err != nil {
  return 1
 }
 // ... 以下既存
}
```

ファイル末尾に `runList` を追加:

```go
func runList(flags cli.Flags) int {
 startDir := flags.TargetDir
 if startDir == "" {
  cwd, err := os.Getwd()
  if err != nil {
   ui.Error("getwd: %v", err)
   return 1
  }
  startDir = cwd
 }

 if err := gitx.RequireRepo(startDir); err != nil {
  ui.Error("ccw -L: not a git repository: %s", startDir)
  return 1
 }
 mainRepo, err := gitx.ResolveMainRepo(startDir)
 if err != nil {
  ui.Error("ccw -L: resolve main repo: %v", err)
  return 2
 }

 b := listmode.Builder{
  ListWorktrees: worktree.List,
  ResolveRepo:   resolveListRepo,
  FetchPRs: func(branches []string) (map[string]gh.PRInfo, error) {
   return gh.PRStatusWithTimeout(gh.DefaultRunner{}, 5*time.Second, branches)
  },
  GhAvailable: gh.Available,
 }

 out, warns, err := b.Build(mainRepo, listmode.Options{NoPR: flags.NoPR, NoSession: flags.NoSession})
 if err != nil {
  ui.Error("ccw -L: %v", err)
  return 2
 }
 for _, w := range warns {
  ui.Warn(w.Message)
 }
 if flags.JSON {
  if err := listmode.RenderJSON(out, os.Stdout); err != nil {
   ui.Error("render json: %v", err)
   return 2
  }
  return 0
 }
 if err := listmode.RenderText(out, os.Stdout); err != nil {
  ui.Error("render text: %v", err)
  return 2
 }
 return 0
}

func resolveListRepo(mainRepo string) (listmode.RepoInfo, error) {
 repo := listmode.RepoInfo{MainPath: mainRepo}
 if rawURL, err := gitx.OriginURL(mainRepo); err == nil && rawURL != "" {
  if owner, name, err := gitx.ParseOriginURL(rawURL); err == nil {
   repo.Owner = owner
   repo.Name = name
  }
 }
 if repo.Owner == "" {
  repo.Owner = "local"
 }
 if repo.Name == "" {
  repo.Name = filepath.Base(mainRepo)
 }
 if db, err := gitx.DefaultBranch(mainRepo); err == nil {
  repo.DefaultBranch = db
 }
 return repo, nil
}
```

import に `"path/filepath"` を追加。

- [ ] **Step 11.2: Build and vet**

```bash
go build ./...
go vet ./...
```

Expected: no errors.

- [ ] **Step 11.3: Smoke run inside this repo**

```bash
go run ./cmd/ccw -L
```

Expected: テーブルヘッダーと現在の worktrees が表示される。エラー無し。

```bash
go run ./cmd/ccw -L --json | head -5
```

Expected: 整形 JSON が始まる。

- [ ] **Step 11.4: Commit**

```bash
git add cmd/ccw/main.go
git commit -m "feat(ccw): wire -L list mode through listmode.Builder"
```

---

## Task 12: 統合テスト (`tests/list_mode_test.go`)

binary を build して実行し、stdout の形を確認する。

**Files:**

- Create: `tests/list_mode_test.go`

- [ ] **Step 12.1: Inspect existing helpers**

```bash
go doc ./tests
```

`buildBinary`, `initRepo`, `runCcw` といった既存ヘルパが `tests/resume_flow_test.go` 等にある。同パッケージ内なので再利用可。

- [ ] **Step 12.2: Write the test**

`tests/list_mode_test.go`:

```go
package tests

import (
 "encoding/json"
 "os"
 "os/exec"
 "path/filepath"
 "strings"
 "testing"
)

func TestListMode_EmptyRepoEmitsHeaderOnly(t *testing.T) {
 binDir := t.TempDir()
 buildBinary(t, "../cmd/ccw", filepath.Join(binDir, "ccw"))

 repo := t.TempDir()
 initRepo(t, repo)

 cmd := exec.Command(filepath.Join(binDir, "ccw"), "-L", "--no-pr", "--no-session")
 cmd.Dir = repo
 cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
 out, err := cmd.CombinedOutput()
 if err != nil {
  t.Fatalf("run ccw -L: %v\n%s", err, out)
 }
 if !strings.Contains(string(out), "NAME") {
  t.Errorf("expected header, got: %s", out)
 }
}

func TestListMode_JSONShape(t *testing.T) {
 binDir := t.TempDir()
 buildBinary(t, "../cmd/ccw", filepath.Join(binDir, "ccw"))

 repo := t.TempDir()
 initRepo(t, repo)

 cmd := exec.Command(filepath.Join(binDir, "ccw"), "-L", "--json", "--no-pr", "--no-session")
 cmd.Dir = repo
 cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
 out, err := cmd.Output()
 if err != nil {
  t.Fatalf("run ccw -L --json: %v", err)
 }
 var parsed map[string]any
 if err := json.Unmarshal(out, &parsed); err != nil {
  t.Fatalf("Unmarshal: %v\n%s", err, out)
 }
 if parsed["version"].(float64) != 1 {
  t.Errorf("version = %v", parsed["version"])
 }
 if _, ok := parsed["repo"]; !ok {
  t.Error("missing repo key")
 }
 if wts, ok := parsed["worktrees"].([]any); !ok || wts == nil {
  t.Errorf("worktrees missing or nil: %v", parsed["worktrees"])
 }
}

func TestListMode_DirOverridesCwd(t *testing.T) {
 binDir := t.TempDir()
 buildBinary(t, "../cmd/ccw", filepath.Join(binDir, "ccw"))

 repo := t.TempDir()
 initRepo(t, repo)

 cwd := t.TempDir() // not a repo
 cmd := exec.Command(filepath.Join(binDir, "ccw"), "-L", "-d", repo, "--no-pr", "--no-session")
 cmd.Dir = cwd
 cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
 if out, err := cmd.CombinedOutput(); err != nil {
  t.Fatalf("run ccw -L -d: %v\n%s", err, out)
 }
}

func TestListMode_DirInvalidExits1(t *testing.T) {
 binDir := t.TempDir()
 buildBinary(t, "../cmd/ccw", filepath.Join(binDir, "ccw"))

 cmd := exec.Command(filepath.Join(binDir, "ccw"), "-L", "-d", "/nonexistent-dir-xyz")
 cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
 out, err := cmd.CombinedOutput()
 if err == nil {
  t.Fatalf("want non-zero exit, got success\n%s", out)
 }
 if exitErr, ok := err.(*exec.ExitError); ok {
  if exitErr.ExitCode() != 1 {
   t.Errorf("exit code = %d, want 1", exitErr.ExitCode())
  }
 }
}
```

- [ ] **Step 12.3: Run integration tests**

```bash
go test ./tests/ -run TestListMode -v
```

Expected: PASS (4 tests).

- [ ] **Step 12.4: Run full test suite**

```bash
go test ./...
```

Expected: PASS (全テスト、リグレッション無し).

- [ ] **Step 12.5: Commit**

```bash
git add tests/list_mode_test.go
git commit -m "test(integration): cover ccw -L empty/json/dir-override/invalid-dir"
```

---

## Task 13: README updates (英 + 日)

**Files:**

- Modify: `README.md`
- Modify: `docs/README.ja.md`

- [ ] **Step 13.1: Read current Features and Usage sections**

```bash
go run ./cmd/ccw --help
```

(出力を確認しつつ README に反映)

- [ ] **Step 13.2: Edit `README.md`**

Features セクション末尾に追加:

```markdown
- 📋 **Machine-readable list** — `ccw -L --json` aggregates worktree × git × PR × session info in one shot, ideal for scripts and Claude Code agent use
```

Usage セクション、`ccw -- --model …` の下に追加:

```markdown
ccw -L                                    # list ccw worktrees (text table)
ccw -L --json                             # same, JSON for scripts / agents
ccw -L -d ~/repo --no-pr --no-session     # target a specific repo, skip gh and session lookup
```

- [ ] **Step 13.3: Edit `docs/README.ja.md`**

Features セクションに同様の項目を日本語で追加:

```markdown
- 📋 **機械可読リスト** — `ccw -L --json` で worktree × git × PR × session 情報を一括取得。スクリプトや Claude Code のエージェント用途に最適
```

Usage セクションにも対応する例を追加（英語版と同じコマンド、コメントは日本語）:

```markdown
ccw -L                                    # ccw worktree を表形式で一覧
ccw -L --json                             # 同上、JSON 出力（スクリプト・エージェント用）
ccw -L -d ~/repo --no-pr --no-session     # 特定の repo を対象、gh / session 探索を skip
```

- [ ] **Step 13.4: Run readme-sync sanity check**

```bash
ls docs/README.ja.md README.md
```

両ファイルが揃っていることを確認。textlint があれば:

```bash
just lint || true
```

(lint エラーが出たら原因を確認、ただし list セクション以外の既存問題は別 PR とする)

- [ ] **Step 13.5: Commit**

```bash
git add README.md docs/README.ja.md
git commit -m "docs(readme): document ccw -L list mode (en + ja)"
```

---

## Task 14: 最終チェック

- [ ] **Step 14.1: Full test pass**

```bash
go test ./...
go vet ./...
```

Expected: all green.

- [ ] **Step 14.2: Manual smoke**

```bash
go run ./cmd/ccw -L
go run ./cmd/ccw -L --json | head -20
go run ./cmd/ccw -L --json --no-pr --no-session
go run ./cmd/ccw -L -d "$(pwd)"
go run ./cmd/ccw -L -n  # should error
go run ./cmd/ccw -d /tmp  # should error (no -L)
```

Expected:

- 1〜4: 正常出力
- 5, 6: stderr にエラー、exit 1

- [ ] **Step 14.3: Verify spec doc reference**

```bash
grep -l "2026-04-26-ccw-list-mode-design" docs/superpowers/plans/2026-04-26-ccw-list-mode.md
```

Expected: ヒット（plan からの spec 参照が活きていること）。

- [ ] **Step 14.4: Final lint pass**

```bash
just lint || go vet ./...
```

Expected: clean (or only pre-existing issues unrelated to this PR).

- [ ] **Step 14.5: No further commit needed**

このタスクはチェックのみ。コミットは task 13 までで完了している。

---

## 完了基準

- [ ] `go test ./...` 全 PASS
- [ ] `ccw -L` / `ccw -L --json` / `ccw -L -d <path>` / `--no-pr` / `--no-session` が動作
- [ ] 排他フラグ違反は exit 1 + stderr メッセージ
- [ ] `--help` に list セクションあり
- [ ] README.md / docs/README.ja.md に list mode の説明あり
- [ ] picker / new / clean-all 経路にリグレッションなし

## 注意点

- `internal/worktree.List` に新フィールド取得を足したぶん、picker 起動時間が微増する。実測して数百 ms 単位の劣化が出れば `worktree.ListOptions` を導入してフィールド取得を opt-in 化（本 plan のスコープ外）。
- gh の timeout 5s は固定値。長すぎる / 短すぎるとフィードバックが来たら別 PR で `--gh-timeout` フラグ化検討。
- text 出力の `BRANCH` 列は worktree 名と被りやすいが、機械パース用途では `--json` を使う想定なので冗長性は許容。
