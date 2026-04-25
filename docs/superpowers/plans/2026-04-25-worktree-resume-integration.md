# Worktree ↔ Claude Code Session Resume Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** ccw が作る worktree の名前を Claude Code のセッション名と 1:1 にし、picker から既存 worktree を選んだら過去会話を resume できるようにする。

**Architecture:**

- ccw は worktree 名を生成し、`claude --worktree <name> -n <name>` で起動する。
- picker 表示時に `~/.claude/projects/<encoded-cwd>/` を読み、`*.jsonl` の有無で resume 可否を判定する。
- 既存 worktree を選んで `[r] run` した場合、デフォルトで `claude --continue` を呼び、失敗時のみ `claude -n <name>` フォールバック。

**Tech Stack:** Go 1.25 / bubbletea / lipgloss / Claude Code CLI `>= 2.1.118`

**Spec:** `docs/superpowers/specs/2026-04-25-worktree-resume-integration-design.md`

---

## File Structure

新規ファイル:

- `internal/namegen/namegen.go` — worktree / session 名のジェネレータ
- `internal/namegen/namegen_test.go`
- `internal/worktree/has_session.go` — `EncodeProjectPath` / `HasSession`
- `internal/worktree/has_session_test.go`
- `internal/tips/tips.go` — TIPS 文字列セット + `PickRandom`
- `internal/tips/tips_test.go`
- `tests/resume_flow_test.go` — fake claude バイナリでの統合テスト

修正:

- `internal/claude/claude.go` — `BuildNewArgs(name, preamble, extra)`、`BuildContinueArgs`、`Continue`
- `internal/claude/claude_test.go`
- `internal/worktree/worktree.go` — `Info.HasSession`、`List` で埋める
- `internal/worktree/worktree_test.go`
- `internal/picker/style.go` — `ResumeBadge(hasSession bool)`
- `internal/picker/style_test.go`
- `internal/picker/delegate.go` — L2 4 行レイアウト
- `internal/picker/delegate_test.go`
- `internal/picker/view.go` — footer に random TIPS
- `internal/picker/model.go` — `New` で TIPS 選択
- `internal/picker/model_test.go`（必要なら）
- `cmd/ccw/main.go` — 名前生成、`Continue` 呼び出し、フォールバック
- `docs/README.md` / `docs/README.ja.md` — 旧警告撤去、命名規約・TIPS 追記
- `.claude/settings.local.json` — `Read(~/.claude/projects/**)` 許可

---

## Task 0: 前提環境チェック（実装前）

**目的:** spec の前提検証項目を満たすローカル環境を確認する。

**Files:**

- なし（手動検証）

- [ ] **Step 1: claude のバージョン確認**

Run: `claude --version`
Expected: `2.1.118` 以上

不足していたら `npm i -g @anthropic-ai/claude-code` などで更新。

- [ ] **Step 2: フラグ併用確認**

Run（任意のサンドボックス repo で）:

```bash
mkdir -p /tmp/ccw-flag-check && cd /tmp/ccw-flag-check && git init -q && \
  echo x > a && git add . && git commit -q -m init
claude --worktree foo -n foo --print "ok"
```

Expected: エラーなく終了し、`/tmp/ccw-flag-check/.claude/worktrees/foo/` が作成されている。

- [ ] **Step 3: `--continue` の no-session 挙動を観察**

Run（セッションログが無いディレクトリで）:

```bash
cd /tmp && claude --continue --print "ok"; echo "exit=$?"
```

Expected: 非ゼロ exit、または picker 起動後にユーザ操作で抜ける。挙動を実装中の参考にする。

- [ ] **Step 4: パスエンコード規則の確認**

Run:

```bash
ls ~/.claude/projects/ | head
```

Expected: `<absolute-path>` の `/` と `.` が `-` に置換されたディレクトリが存在する。`*.jsonl` ファイルが入っている。

実装と仕様が一致するか目視確認。一致しなければ `EncodeProjectPath` の実装を後の Task で調整する。

---

## Task 1: 名前ジェネレータ `internal/namegen` を追加

**目的:** ccw が worktree 名と session 名に同じ値を渡すための、決定的かつ衝突しにくい名前生成器を作る。

**Files:**

- Create: `internal/namegen/namegen.go`
- Test: `internal/namegen/namegen_test.go`

- [ ] **Step 1: テストを書く**

`internal/namegen/namegen_test.go`:

```go
package namegen

import (
 "regexp"
 "testing"
)

func TestGenerate_FormatAndUniqueness(t *testing.T) {
 re := regexp.MustCompile(`^[a-z]+-[a-z]+-[0-9a-f]{4}$`)
 seen := map[string]struct{}{}
 for i := 0; i < 100; i++ {
  got := Generate()
  if !re.MatchString(got) {
   t.Fatalf("Generate() = %q, want match %s", got, re)
  }
  seen[got] = struct{}{}
 }
 if len(seen) < 90 {
  t.Errorf("Generate() collisions too high: %d/100 unique", len(seen))
 }
}

func TestGenerateWithSeed_Deterministic(t *testing.T) {
 a := generateWithSeed(42)
 b := generateWithSeed(42)
 if a != b {
  t.Errorf("generateWithSeed(42): non-deterministic %q vs %q", a, b)
 }
}

func TestGenerate_NoSpacesNoUppercase(t *testing.T) {
 for i := 0; i < 50; i++ {
  got := Generate()
  for _, r := range got {
   if r == ' ' {
    t.Fatalf("Generate() = %q contains space", got)
   }
   if r >= 'A' && r <= 'Z' {
    t.Fatalf("Generate() = %q contains uppercase", got)
   }
  }
 }
}
```

- [ ] **Step 2: テストを実行して fail を確認**

Run: `go test ./internal/namegen/...`
Expected: `package namegen: no Go files` または `undefined: Generate`

- [ ] **Step 3: 実装**

`internal/namegen/namegen.go`:

```go
// Package namegen generates short slug names like "quick-falcon-7bd2"
// to use as both worktree directory name and Claude Code session name.
package namegen

import (
 "fmt"
 "math/rand/v2"
 "time"
)

var adjectives = []string{
 "quick", "lazy", "happy", "brave", "calm", "eager", "fancy", "glad",
 "jolly", "kind", "lively", "merry", "nice", "polite", "quiet", "silly",
 "witty", "zany", "bright", "clever", "daring", "fierce", "gentle", "mighty",
 "nimble", "proud", "rapid", "shiny", "sturdy", "tame",
}

var nouns = []string{
 "falcon", "otter", "lion", "tiger", "wolf", "panda", "eagle", "shark",
 "crane", "fox", "raven", "owl", "lynx", "bison", "moose", "hawk",
 "orca", "puma", "yak", "ibex", "robin", "swan", "gecko", "mantis",
 "koala", "badger", "heron", "jaguar", "lemur", "mole",
}

// Generate returns a slug like "quick-falcon-7bd2".
// Uses time.Now().UnixNano() as the seed.
func Generate() string {
 return generateWithSeed(uint64(time.Now().UnixNano()))
}

func generateWithSeed(seed uint64) string {
 r := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
 adj := adjectives[r.IntN(len(adjectives))]
 noun := nouns[r.IntN(len(nouns))]
 suffix := fmt.Sprintf("%04x", r.IntN(0x10000))
 return fmt.Sprintf("%s-%s-%s", adj, noun, suffix)
}
```

- [ ] **Step 4: テストが通ることを確認**

Run: `go test ./internal/namegen/...`
Expected: PASS

- [ ] **Step 5: コミット**

```bash
git add internal/namegen
git commit -m "feat(namegen): worktree/session 名ジェネレータを追加"
```

---

## Task 2: `claude` パッケージにフラグ生成ロジックを更新

**目的:** `BuildNewArgs` に name 引数を追加し `--worktree <name> -n <name>` を生成。`BuildResumeArgs` を `BuildContinueArgs` にリネームして `--continue` を付与。`Resume` を `Continue` にリネーム。

**Files:**

- Modify: `internal/claude/claude.go`
- Modify: `internal/claude/claude_test.go`

- [ ] **Step 1: テストを更新**

`internal/claude/claude_test.go` を以下で**置き換え**:

```go
package claude

import (
 "reflect"
 "testing"
)

func TestBuildNewArgs_NameOnly(t *testing.T) {
 got := BuildNewArgs("foo", "", nil)
 want := []string{"--permission-mode", "auto", "--worktree", "foo", "-n", "foo"}
 if !reflect.DeepEqual(got, want) {
  t.Errorf("BuildNewArgs:\n got  = %v\n want = %v", got, want)
 }
}

func TestBuildNewArgs_WithExtra(t *testing.T) {
 got := BuildNewArgs("foo", "", []string{"--model", "claude-opus-4-7"})
 want := []string{"--permission-mode", "auto", "--worktree", "foo", "-n", "foo", "--model", "claude-opus-4-7"}
 if !reflect.DeepEqual(got, want) {
  t.Errorf("BuildNewArgs extra:\n got  = %v\n want = %v", got, want)
 }
}

func TestBuildNewArgs_WithPreamble(t *testing.T) {
 got := BuildNewArgs("foo", "hello", nil)
 want := []string{"--permission-mode", "auto", "--worktree", "foo", "-n", "foo", "--", "hello"}
 if !reflect.DeepEqual(got, want) {
  t.Errorf("BuildNewArgs preamble:\n got  = %v\n want = %v", got, want)
 }
}

func TestBuildNewArgs_WithExtraAndPreamble(t *testing.T) {
 got := BuildNewArgs("foo", "hi", []string{"--model", "x"})
 want := []string{"--permission-mode", "auto", "--worktree", "foo", "-n", "foo", "--model", "x", "--", "hi"}
 if !reflect.DeepEqual(got, want) {
  t.Errorf("BuildNewArgs both:\n got  = %v\n want = %v", got, want)
 }
}

func TestBuildContinueArgs_Empty(t *testing.T) {
 got := BuildContinueArgs(nil)
 want := []string{"--permission-mode", "auto", "--continue"}
 if !reflect.DeepEqual(got, want) {
  t.Errorf("BuildContinueArgs:\n got  = %v\n want = %v", got, want)
 }
}

func TestBuildContinueArgs_WithExtra(t *testing.T) {
 got := BuildContinueArgs([]string{"--model", "x"})
 want := []string{"--permission-mode", "auto", "--continue", "--model", "x"}
 if !reflect.DeepEqual(got, want) {
  t.Errorf("BuildContinueArgs extra:\n got  = %v\n want = %v", got, want)
 }
}
```

- [ ] **Step 2: テストを実行して fail 確認**

Run: `go test ./internal/claude/...`
Expected: コンパイルエラー（`BuildNewArgs` シグネチャ不一致 / `BuildContinueArgs` 未定義）

- [ ] **Step 3: 実装を更新**

`internal/claude/claude.go`:

```go
// Package claude wraps launching the `claude` CLI in ccw-appropriate ways
// (new worktree session vs. continue existing worktree).
package claude

import (
 "errors"
 "fmt"
 "os"
 "os/exec"
)

// BuildNewArgs constructs argv (excluding the program name) for
// `claude --permission-mode auto --worktree <name> -n <name> [extra...] [-- preamble]`.
func BuildNewArgs(name, preamble string, extra []string) []string {
 args := make([]string, 0, 6+len(extra)+2)
 args = append(args, "--permission-mode", "auto", "--worktree", name, "-n", name)
 args = append(args, extra...)
 if preamble != "" {
  args = append(args, "--", preamble)
 }
 return args
}

// BuildContinueArgs constructs argv for `claude --permission-mode auto --continue [extra...]`.
func BuildContinueArgs(extra []string) []string {
 args := make([]string, 0, 3+len(extra))
 args = append(args, "--permission-mode", "auto", "--continue")
 return append(args, extra...)
}

// LaunchNew execs claude with BuildNewArgs in cwd. Returns claude's exit code
// (0 on success, the child exit code on non-zero exit, -1 on exec error).
func LaunchNew(cwd, name, preamble string, extra []string) (int, error) {
 return runClaude(cwd, BuildNewArgs(name, preamble, extra))
}

// Continue execs claude with BuildContinueArgs in cwd.
func Continue(cwd string, extra []string) (int, error) {
 return runClaude(cwd, BuildContinueArgs(extra))
}

func runClaude(cwd string, args []string) (int, error) {
 cmd := exec.Command("claude", args...)
 cmd.Dir = cwd
 cmd.Stdin = os.Stdin
 cmd.Stdout = os.Stdout
 cmd.Stderr = os.Stderr
 err := cmd.Run()
 if err == nil {
  return 0, nil
 }
 var exitErr *exec.ExitError
 if errors.As(err, &exitErr) {
  return exitErr.ExitCode(), nil
 }
 return -1, fmt.Errorf("run claude: %w", err)
}
```

- [ ] **Step 4: パッケージ単独テストが通ることを確認**

Run: `go test ./internal/claude/...`
Expected: PASS

- [ ] **Step 5: 全体ビルドが壊れていることを確認（呼び出し側未修正）**

Run: `go build ./...`
Expected: `cmd/ccw/main.go` で `claude.Resume` 未定義、`claude.LaunchNew` 引数不一致のエラー

これは Task 6 で直すので一旦そのままでよい。次の Task に進む前にコミットだけ済ませる。

- [ ] **Step 6: コミット**

```bash
git add internal/claude
git commit -m "feat(claude): -n フラグと --continue を組み込み Continue にリネーム"
```

---

## Task 3: `internal/worktree/has_session.go` を追加

**目的:** worktree path から `~/.claude/projects/<encoded>/` のセッションログ有無を判定する。

**Files:**

- Create: `internal/worktree/has_session.go`
- Test: `internal/worktree/has_session_test.go`

- [ ] **Step 1: テストを書く**

`internal/worktree/has_session_test.go`:

```go
package worktree

import (
 "os"
 "path/filepath"
 "testing"
)

func TestEncodeProjectPath(t *testing.T) {
 cases := []struct {
  in, want string
 }{
  {"/Users/foo/repo/.claude/worktrees/bar", "-Users-foo-repo--claude-worktrees-bar"},
  {"/a.b/c", "-a-b-c"},
  {"/", "-"},
 }
 for _, tc := range cases {
  if got := EncodeProjectPath(tc.in); got != tc.want {
   t.Errorf("EncodeProjectPath(%q) = %q, want %q", tc.in, got, tc.want)
  }
 }
}

func TestHasSession_True(t *testing.T) {
 home := t.TempDir()
 t.Setenv("HOME", home)

 wt := "/Users/foo/repo/.claude/worktrees/bar"
 dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(wt))
 if err := os.MkdirAll(dir, 0o755); err != nil {
  t.Fatal(err)
 }
 if err := os.WriteFile(filepath.Join(dir, "abc.jsonl"), []byte("{}\n"), 0o644); err != nil {
  t.Fatal(err)
 }

 if !HasSession(wt) {
  t.Error("HasSession() = false, want true")
 }
}

func TestHasSession_FalseWhenNoJsonl(t *testing.T) {
 home := t.TempDir()
 t.Setenv("HOME", home)

 wt := "/Users/foo/repo/.claude/worktrees/bar"
 dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(wt))
 if err := os.MkdirAll(dir, 0o755); err != nil {
  t.Fatal(err)
 }
 if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0o644); err != nil {
  t.Fatal(err)
 }

 if HasSession(wt) {
  t.Error("HasSession() = true, want false (no .jsonl)")
 }
}

func TestHasSession_FalseWhenDirMissing(t *testing.T) {
 home := t.TempDir()
 t.Setenv("HOME", home)
 if HasSession("/nonexistent/path") {
  t.Error("HasSession() = true, want false (dir missing)")
 }
}

func TestHasSession_FalseWhenHomeUnset(t *testing.T) {
 t.Setenv("HOME", "")
 if HasSession("/Users/foo/repo/.claude/worktrees/bar") {
  t.Error("HasSession() = true, want false (HOME empty)")
 }
}
```

- [ ] **Step 2: テストを実行して fail を確認**

Run: `go test ./internal/worktree/...`
Expected: `undefined: EncodeProjectPath` / `undefined: HasSession`

- [ ] **Step 3: 実装**

`internal/worktree/has_session.go`:

```go
package worktree

import (
 "os"
 "path/filepath"
 "strings"
)

// EncodeProjectPath converts an absolute worktree path to the directory name
// Claude Code uses under ~/.claude/projects/. Both '/' and '.' map to '-'.
// This rule is observed from claude's behavior; it is not part of any
// public contract and may change.
func EncodeProjectPath(absPath string) string {
 return strings.NewReplacer("/", "-", ".", "-").Replace(absPath)
}

// HasSession reports whether ~/.claude/projects/<encoded(absPath)>/ contains
// at least one *.jsonl file. Returns false on any error (HOME unset, dir
// missing, read failure) so callers can use it as a UI hint without
// branching on errors.
func HasSession(absPath string) bool {
 home, err := os.UserHomeDir()
 if err != nil || home == "" {
  return false
 }
 dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(absPath))
 entries, err := os.ReadDir(dir)
 if err != nil {
  return false
 }
 for _, e := range entries {
  if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
   return true
  }
 }
 return false
}
```

- [ ] **Step 4: テストが通ることを確認**

Run: `go test ./internal/worktree/...`
Expected: PASS

- [ ] **Step 5: コミット**

```bash
git add internal/worktree/has_session.go internal/worktree/has_session_test.go
git commit -m "feat(worktree): ~/.claude/projects 参照で HasSession を判定"
```

---

## Task 4: `worktree.Info` に `HasSession` フィールドを追加し `List` で埋める

**目的:** picker が描画時に resume 可否を判定できるよう、`Info` に状態をキャッシュする。

**Files:**

- Modify: `internal/worktree/worktree.go`
- Modify: `internal/worktree/worktree_test.go`

- [ ] **Step 1: 既存テストを確認**

Run: `cat internal/worktree/worktree_test.go | head -60`

`TestList_*` 系のテストがあれば、`Info` の比較に `HasSession: false` が現れることを許容する形に直す必要がある。なければスキップ。

- [ ] **Step 2: 失敗するテストを追加**

`internal/worktree/worktree_test.go` の末尾に追記:

```go
func TestList_PopulatesHasSession(t *testing.T) {
 home := t.TempDir()
 t.Setenv("HOME", home)

 main := setupRepo(t) // 既存のテストヘルパ。無ければ TestList_* のヘルパを参照
 wt := filepath.Join(main, ".claude", "worktrees", "alpha")
 createWorktree(t, main, wt, "alpha") // 同上

 dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(wt))
 if err := os.MkdirAll(dir, 0o755); err != nil {
  t.Fatal(err)
 }
 if err := os.WriteFile(filepath.Join(dir, "x.jsonl"), []byte("{}\n"), 0o644); err != nil {
  t.Fatal(err)
 }

 infos, err := List(main)
 if err != nil {
  t.Fatal(err)
 }
 var found bool
 for _, in := range infos {
  if in.Path == wt {
   if !in.HasSession {
    t.Errorf("Info.HasSession = false, want true for %s", wt)
   }
   found = true
  }
 }
 if !found {
  t.Fatalf("worktree %s not in List() output", wt)
 }
}
```

> 既存の `worktree_test.go` に `setupRepo` / `createWorktree` 系ヘルパが無い場合は、`TestList_*` の既存テストを参考に、git init + ブランチ作成 + worktree add のヘルパを共通化してから本テストを追加する。

- [ ] **Step 3: テストが fail することを確認**

Run: `go test ./internal/worktree/ -run TestList_PopulatesHasSession`
Expected: コンパイルエラー（`Info.HasSession` 未定義）

- [ ] **Step 4: `Info` 定義と `List` を更新**

`internal/worktree/worktree.go` の `Info` を:

```go
type Info struct {
 Path        string
 Branch      string
 Status      Status
 AheadCount  int
 BehindCount int
 DirtyCount  int
 HasSession  bool
}
```

`List` 内 `info := Info{...}` の直後を:

```go
  info := Info{Path: e.Path, Branch: e.Branch, Status: st}
  ahead, behind, _ := gitx.AheadBehind(e.Path)
  info.AheadCount = ahead
  info.BehindCount = behind
  if st == StatusDirty {
   n, _ := gitx.DirtyCount(e.Path)
   info.DirtyCount = n
  }
  info.HasSession = HasSession(e.Path)
  result = append(result, info)
```

に変更（最終行 `info.HasSession = HasSession(e.Path)` を追加）。

- [ ] **Step 5: テストが通ることを確認**

Run: `go test ./internal/worktree/...`
Expected: PASS

- [ ] **Step 6: コミット**

```bash
git add internal/worktree/worktree.go internal/worktree/worktree_test.go
git commit -m "feat(worktree): Info.HasSession を List で埋める"
```

---

## Task 5: picker のスタイル — RESUME / NEW バッジを追加

**目的:** worktree 行に表示する RESUME / NEW バッジのスタイルを定義する。

**Files:**

- Modify: `internal/picker/style.go`
- Modify: `internal/picker/style_test.go`

- [ ] **Step 1: テストを追加**

`internal/picker/style_test.go` の末尾に追記:

```go
func TestResumeBadge_HasSession(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 if got := ResumeBadge(true); got != "[RESUME]" {
  t.Errorf("ResumeBadge(true) NO_COLOR = %q, want [RESUME]", got)
 }
 if got := ResumeBadge(false); got != "[NEW]   " {
  t.Errorf("ResumeBadge(false) NO_COLOR = %q, want [NEW]   ", got)
 }
}

func TestResumeBadge_Colored(t *testing.T) {
 t.Setenv("NO_COLOR", "")
 got := ResumeBadge(true)
 if !strings.Contains(got, "RESUME") {
  t.Errorf("ResumeBadge(true) = %q, want substring RESUME", got)
 }
 got = ResumeBadge(false)
 if !strings.Contains(got, "NEW") {
  t.Errorf("ResumeBadge(false) = %q, want substring NEW", got)
 }
}
```

`strings` の import を追加（既存ならそのまま）。

- [ ] **Step 2: fail 確認**

Run: `go test ./internal/picker/ -run TestResumeBadge`
Expected: `undefined: ResumeBadge`

- [ ] **Step 3: 実装**

`internal/picker/style.go` の末尾に追記:

```go
// ResumeBadge renders a RESUME / NEW badge.
// hasSession=true → 💬 RESUME (green-cyan, prominent)
// hasSession=false → ⚡ NEW (dim grey)
// Under NO_COLOR, returns plain "[RESUME]" / "[NEW]   " (padded to same width).
func ResumeBadge(hasSession bool) string {
 if noColor() {
  if hasSession {
   return "[RESUME]"
  }
  return "[NEW]   "
 }
 if hasSession {
  return lipgloss.NewStyle().
   Padding(0, 1).Bold(true).
   Background(lipgloss.Color("14")).
   Foreground(lipgloss.Color("0")).
   Render("💬 RESUME")
 }
 return lipgloss.NewStyle().
  Padding(0, 1).
  Background(lipgloss.Color("240")).
  Foreground(lipgloss.Color("15")).
  Render("⚡ NEW   ")
}
```

- [ ] **Step 4: テストが通ることを確認**

Run: `go test ./internal/picker/ -run TestResumeBadge`
Expected: PASS

- [ ] **Step 5: コミット**

```bash
git add internal/picker/style.go internal/picker/style_test.go
git commit -m "feat(picker): RESUME / NEW バッジのスタイル"
```

---

## Task 6: picker delegate を L2 4 行レイアウトに拡張

**目的:** worktree 行を 4 行表示に変更し、RESUME バッジを最も目立つ位置に配置する。

**Files:**

- Modify: `internal/picker/delegate.go`
- Modify: `internal/picker/delegate_test.go`

レイアウト（spec 161 行付近）:

```text
> 💬 RESUME · foo                            [PUSHED]  ↑0 ↓0
    branch:  feature/auth
    pr:      [OPEN] #123 "feat: add auth"
    path:    ~/.claude/worktrees/foo
```

- [ ] **Step 1: 既存テストを読む**

Run: `cat internal/picker/delegate_test.go`

既存の `renderRow` テストの前提（高さ 2、フォーマット）が変わるので、置き換える。

- [ ] **Step 2: テストを更新（fail 状態に）**

`internal/picker/delegate_test.go` で `renderRow` を呼んでいる箇所すべてを以下のように書き換える（最低 1 ケースは変更：4 行のうち先頭にバッジが付く + path が `path:` ラベル付きで描画される）:

```go
func TestRenderRow_ResumeBadge(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 li := listItem{
  tag: tagWorktree,
  wt: &worktree.Info{
   Path:       "/repo/.claude/worktrees/foo",
   Branch:     "feature/auth",
   Status:     worktree.StatusPushed,
   HasSession: true,
  },
 }
 got := renderRow(li, 120, true /*prUnavailable*/, false)
 if !strings.Contains(got, "[RESUME]") {
  t.Errorf("missing RESUME badge:\n%s", got)
 }
 if !strings.Contains(got, "foo") {
  t.Errorf("missing worktree name foo:\n%s", got)
 }
 if !strings.Contains(got, "branch:  feature/auth") {
  t.Errorf("missing branch line:\n%s", got)
 }
 if !strings.Contains(got, "path:    /repo/.claude/worktrees/foo") {
  t.Errorf("missing path line:\n%s", got)
 }
 lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
 if len(lines) != 4 {
  t.Errorf("got %d lines, want 4:\n%s", len(lines), got)
 }
}

func TestRenderRow_NewBadge(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 li := listItem{
  tag: tagWorktree,
  wt: &worktree.Info{
   Path:       "/repo/.claude/worktrees/bar",
   Branch:     "bar",
   Status:     worktree.StatusLocalOnly,
   HasSession: false,
  },
 }
 got := renderRow(li, 120, true, true /*selected*/)
 if !strings.Contains(got, "[NEW]") {
  t.Errorf("missing NEW badge:\n%s", got)
 }
 if !strings.HasPrefix(got, "> ") {
  t.Errorf("selected row should start with '> ': %q", got[:2])
 }
}
```

既存の `TestRenderRow_*` で 2 行レイアウトを期待しているケースは削除または書き換える。

- [ ] **Step 3: テストが fail することを確認**

Run: `go test ./internal/picker/ -run TestRenderRow`
Expected: FAIL（古いレイアウト）

- [ ] **Step 4: 実装**

`internal/picker/delegate.go` を以下に置き換え:

```go
package picker

import (
 "fmt"
 "io"
 "strings"

 "github.com/charmbracelet/bubbles/list"
 tea "github.com/charmbracelet/bubbletea"
 "github.com/charmbracelet/lipgloss"
 "github.com/tqer39/ccw-cli/internal/gh"
 "github.com/tqer39/ccw-cli/internal/worktree"
)

// rowDelegate renders worktree items as four lines:
//   line 1: [resume-badge] · <name>            [status] indicators
//   line 2:   branch:  <branch>
//   line 3:   pr:      <pr cell or (no PR)>
//   line 4:   path:    <path>
type rowDelegate struct {
 prUnavailable bool
}

func (d rowDelegate) Height() int                             { return 4 }
func (d rowDelegate) Spacing() int                            { return 1 }
func (d rowDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d rowDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
 li, ok := item.(listItem)
 if !ok {
  return
 }
 selected := index == m.Index()
 _, _ = fmt.Fprint(w, renderRow(li, m.Width(), d.prUnavailable, selected))
}

func renderRow(li listItem, width int, prUnavailable bool, selected bool) string {
 prefix := "  "
 if selected {
  prefix = "> "
 }
 switch li.tag {
 case tagNew, tagQuit, tagDeleteAll, tagCleanPushed, tagCustomSelect:
  return prefix + li.title + "\n  " + li.desc
 }
 wt := li.wt
 name := worktreeName(wt.Path)
 resume := ResumeBadge(wt.HasSession)
 status := Badge(wt.Status)
 indicators := fmt.Sprintf("↑%d ↓%d", wt.AheadCount, wt.BehindCount)
 if wt.Status == worktree.StatusDirty {
  indicators += fmt.Sprintf(" ✎%d", wt.DirtyCount)
 }

 header := fmt.Sprintf("%s%s · %s", prefix, resume, name)
 right := fmt.Sprintf("%s  %s", status, indicators)
 header = padBetween(header, right, width)

 branchLine := fmt.Sprintf("    branch:  %s", wt.Branch)
 prLine := "    pr:      " + renderPRForLine(li.pr, prUnavailable)
 pathLine := fmt.Sprintf("    path:    %s", wt.Path)

 if width > 0 {
  header = truncateToWidth(header, width)
  branchLine = truncateToWidth(branchLine, width)
  prLine = truncateToWidth(prLine, width)
  pathLine = truncateToWidth(pathLine, width)
 }

 return header + "\n" + branchLine + "\n" + prLine + "\n" + pathLine
}

// padBetween places left and right on the same line, padding spaces between
// so that right is right-aligned at width. If width is 0 or too small, falls
// back to "left  right".
func padBetween(left, right string, width int) string {
 if width <= 0 {
  return left + "  " + right
 }
 gap := width - lipgloss.Width(left) - lipgloss.Width(right)
 if gap < 2 {
  gap = 2
 }
 return left + strings.Repeat(" ", gap) + right
}

func renderPRForLine(pr *gh.PRInfo, prUnavailable bool) string {
 if prUnavailable {
  return ""
 }
 return renderPRCell(pr)
}

func worktreeName(path string) string {
 idx := strings.LastIndex(path, "/")
 if idx < 0 {
  return path
 }
 return path[idx+1:]
}

func arrowGlyph() string {
 if noColor() {
  return "->"
 }
 return "→"
}

func renderPRCell(pr *gh.PRInfo) string {
 if pr == nil {
  if noColor() {
   return "(no PR)"
  }
  return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(no PR)")
 }
 title := pr.Title
 if len(title) > 30 {
  title = title[:29] + "…"
 }
 inner := fmt.Sprintf("%s #%d %q", PRBadge(pr.State), pr.Number, title)
 return PRCellStyle(pr.State).Render(inner)
}

func truncateToWidth(s string, n int) string {
 if lipgloss.Width(s) <= n {
  return s
 }
 for len(s) > 0 && lipgloss.Width(s) > n {
  s = s[:len(s)-1]
 }
 return s
}
```

注: `arrowGlyph` は他で使っていれば残す。使われていなければ削除可。

- [ ] **Step 5: テストが通ることを確認**

Run: `go test ./internal/picker/...`
Expected: PASS

既存テストが旧レイアウトを期待していて壊れていたら、上記新レイアウトに合わせて修正してからパスさせる。

- [ ] **Step 6: コミット**

```bash
git add internal/picker/delegate.go internal/picker/delegate_test.go
git commit -m "feat(picker): worktree 行を L2 4 行レイアウトに拡張"
```

---

## Task 7: `internal/tips` パッケージを追加

**目的:** picker footer に表示するランダム TIPS 文字列を提供する。

**Files:**

- Create: `internal/tips/tips.go`
- Test: `internal/tips/tips_test.go`

- [ ] **Step 1: テストを書く**

`internal/tips/tips_test.go`:

```go
package tips

import (
 "strings"
 "testing"
)

func TestPickRandom_FromDefaultSet(t *testing.T) {
 got := PickRandom(42)
 if got == "" {
  t.Fatal("PickRandom(42) = empty string")
 }
 found := false
 for _, c := range Defaults() {
  if got == c {
   found = true
   break
  }
 }
 if !found {
  t.Errorf("PickRandom(42) = %q, not in Defaults()", got)
 }
}

func TestPickRandom_Deterministic(t *testing.T) {
 if PickRandom(7) != PickRandom(7) {
  t.Error("PickRandom(7) is non-deterministic")
 }
}

func TestPickFrom_Empty(t *testing.T) {
 if got := pickFrom(nil, 1); got != "" {
  t.Errorf("pickFrom(nil) = %q, want empty", got)
 }
 if got := pickFrom([]string{}, 1); got != "" {
  t.Errorf("pickFrom([]) = %q, want empty", got)
 }
}

func TestDefaults_NonEmpty(t *testing.T) {
 d := Defaults()
 if len(d) == 0 {
  t.Error("Defaults() empty")
 }
 for _, s := range d {
  if strings.TrimSpace(s) == "" {
   t.Errorf("empty TIPS string in defaults")
  }
 }
}
```

- [ ] **Step 2: fail 確認**

Run: `go test ./internal/tips/...`
Expected: `package tips: no Go files`

- [ ] **Step 3: 実装**

`internal/tips/tips.go`:

```go
// Package tips provides short rotating tip strings shown in the picker footer.
package tips

import "math/rand/v2"

var defaults = []string{
 "worktree 名 = session 名。手で /rename しても ccw は何もしません",
 "claude --from-pr <番号> で PR 連携セッションを直接 resume できます",
 "--clean-all で push 済 worktree を一括削除",
 "ccw -- --model <id> で claude にフラグを素通し",
 "picker の RESUME バッジは ~/.claude/projects/ から判定しています",
}

// Defaults returns a copy of the built-in TIPS set.
func Defaults() []string {
 out := make([]string, len(defaults))
 copy(out, defaults)
 return out
}

// PickRandom returns a single tip from the defaults using seed.
func PickRandom(seed uint64) string {
 return pickFrom(defaults, seed)
}

func pickFrom(set []string, seed uint64) string {
 if len(set) == 0 {
  return ""
 }
 r := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
 return set[r.IntN(len(set))]
}
```

- [ ] **Step 4: テストが通ることを確認**

Run: `go test ./internal/tips/...`
Expected: PASS

- [ ] **Step 5: コミット**

```bash
git add internal/tips
git commit -m "feat(tips): random TIPS パッケージを追加"
```

---

## Task 8: picker footer に random TIPS を出す

**目的:** Model の構築時に TIPS を 1 件選び、`View` の footer で表示する。

**Files:**

- Modify: `internal/picker/model.go`
- Modify: `internal/picker/view.go`
- Modify: `internal/picker/model_test.go`（必要なら）

- [ ] **Step 1: テストを追加**

`internal/picker/view_test.go`（無ければ新規）に:

```go
package picker

import (
 "strings"
 "testing"

 "github.com/tqer39/ccw-cli/internal/worktree"
)

func TestView_FooterShowsTips(t *testing.T) {
 m := New([]worktree.Info{{Path: "/x/.claude/worktrees/a", Branch: "a"}})
 m.ghAvailable = true
 m.tip = "test tip line"
 m.state = stateList
 out := m.View()
 if !strings.Contains(out, "💡 Tip: test tip line") {
  t.Errorf("View footer missing tip:\n%s", out)
 }
}

func TestView_FooterShowsGhHintWhenUnavailable(t *testing.T) {
 m := New([]worktree.Info{{Path: "/x/.claude/worktrees/a", Branch: "a"}})
 m.ghAvailable = false
 m.tip = "should-not-show"
 m.state = stateList
 out := m.View()
 if !strings.Contains(out, "Install gh") {
  t.Errorf("View should show gh hint when gh unavailable:\n%s", out)
 }
 if strings.Contains(out, "should-not-show") {
  t.Errorf("View should not show tip when gh unavailable:\n%s", out)
 }
}
```

- [ ] **Step 2: fail 確認**

Run: `go test ./internal/picker/ -run TestView_Footer`
Expected: `m.tip undefined`

- [ ] **Step 3: 実装 — Model に tip フィールド追加**

`internal/picker/model.go` の `Model` 構造体に追加:

```go
type Model struct {
 state         state
 infos         []worktree.Info
 list          list.Model
 selIdx        int
 action        Action
 selection     Selection
 width         int
 height        int
 ghAvailable   bool
 prs           map[string]gh.PRInfo
 prUnavailable bool
 bulkFilter    map[worktree.Status]bool
 bulkTargets   []int
 bulkForce     bool
 tip           string
}
```

`New(infos)` の戻り行を:

```go
 return Model{
  state:       stateList,
  infos:       infos,
  list:        l,
  ghAvailable: gh.Available(),
  tip:         tips.PickRandom(uint64(time.Now().UnixNano())),
 }
```

import に `"time"` と `"github.com/tqer39/ccw-cli/internal/tips"` を追加。

- [ ] **Step 4: View() の footer を更新**

`internal/picker/view.go` の `case stateList` を:

```go
 case stateList:
  base := m.list.View()
  footer := ""
  switch {
  case !m.ghAvailable:
   footer = "💡 Install gh to see PR titles here"
  case m.tip != "":
   footer = "💡 Tip: " + m.tip
  }
  if footer == "" {
   return base
  }
  return base + "\n\n" + footer
```

- [ ] **Step 5: テストが通ることを確認**

Run: `go test ./internal/picker/...`
Expected: PASS

- [ ] **Step 6: コミット**

```bash
git add internal/picker
git commit -m "feat(picker): footer に random TIPS を表示"
```

---

## Task 9: `cmd/ccw/main.go` を新シグネチャに合わせる

**目的:** `LaunchNew(name, ...)` / `Continue(...)` の新呼び出しに移行し、HasSession で resume / new を分岐、失敗時にフォールバック。

**Files:**

- Modify: `cmd/ccw/main.go`

- [ ] **Step 1: 既存呼び出しを確認**

Run: `grep -n "claude.Launch\|claude.Resume" cmd/ccw/main.go`
Expected: 既存ヒット（74 行付近、103 行付近）

- [ ] **Step 2: import 追加**

`cmd/ccw/main.go` の import に:

```go
 "github.com/tqer39/ccw-cli/internal/namegen"
```

を追加。

- [ ] **Step 3: 新規 worktree 起動を更新**

`run` 関数内の以下のブロックを:

```go
 if flags.NewWorktree {
  code, err := claude.LaunchNew(mainRepo, preamble, flags.Passthrough)
  if err != nil {
   ui.Error("%v", err)
   return 1
  }
  return code
 }
```

↓

```go
 if flags.NewWorktree {
  name := namegen.Generate()
  code, err := claude.LaunchNew(mainRepo, name, preamble, flags.Passthrough)
  if err != nil {
   ui.Error("%v", err)
   return 1
  }
  return code
 }
```

- [ ] **Step 4: picker [new] の起動を更新**

`runPicker` 内の `case picker.ActionNew:` を:

```go
  case picker.ActionNew:
   name := namegen.Generate()
   code, err := claude.LaunchNew(mainRepo, name, "", passthrough)
   if err != nil {
    ui.Error("%v", err)
    return 1
   }
   return code
```

- [ ] **Step 5: picker [r] run を `Continue` に変更し、フォールバック追加**

`case picker.ActionResume:` を:

```go
  case picker.ActionResume:
   code, err := claude.Continue(sel.Path, passthrough)
   if err != nil {
    ui.Error("%v", err)
    return 1
   }
   if code != 0 && !sel.HasSession {
    // HasSession=false なのに Resume パスを通った（picker から強制 run など）。
    // セッション無し → --continue は失敗し得るので -n <name> でフォールバック。
    name := worktreeName(sel.Path)
    code, err = claude.LaunchNew(sel.Path, name, "", passthrough)
    if err != nil {
     ui.Error("%v", err)
     return 1
    }
   }
   return code
```

`worktreeName` ヘルパは `cmd/ccw/main.go` 末尾に追加:

```go
func worktreeName(path string) string {
 for i := len(path) - 1; i >= 0; i-- {
  if path[i] == '/' {
   return path[i+1:]
  }
 }
 return path
}
```

- [ ] **Step 6: `Selection` に `HasSession` を持たせる**

`internal/picker/model.go` の `Selection` 構造体に追記:

```go
type Selection struct {
 Path        string
 Branch      string
 Status      worktree.Status
 HasSession  bool
 ForceDelete bool
}
```

`internal/picker/update.go` か `internal/picker/run.go` で `Selection{...}` を構築している箇所をすべて grep し、`HasSession: w.HasSession` を埋める。

Run: `grep -n "Selection{" internal/picker/*.go`
各ヒット箇所に `HasSession` を追加。

- [ ] **Step 7: ビルドが通ることを確認**

Run: `go build ./...`
Expected: 成功

- [ ] **Step 8: ユニットテストが通ることを確認**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 9: コミット**

```bash
git add cmd/ccw/main.go internal/picker/model.go internal/picker/run.go internal/picker/update.go
git commit -m "feat(ccw): worktree 名 = session 名 を確立、resume を既存 worktree のデフォルトに"
```

---

## Task 10: 統合テスト `tests/resume_flow_test.go`

**目的:** fake `claude` バイナリで end-to-end の引数遷移を検証する。

**Files:**

- Create: `tests/resume_flow_test.go`
- 必要なら: `tests/testdata/fake-claude.go` のようなビルド対象

- [ ] **Step 1: 既存 `tests/` 構造を確認**

Run: `ls tests/ 2>/dev/null && find tests -name '*.go' 2>/dev/null`

無ければ作成して進む。

- [ ] **Step 2: fake claude バイナリのソースを作る**

注: Go の `testdata/` は build 対象外なので、`tests/fakes/fake_claude/main.go` に置く（通常の package として認識される）。

`tests/fakes/fake_claude/main.go`:

```go
// fake claude binary used in resume_flow_test.go.
// Writes os.Args[1:] to $CCW_FAKE_CLAUDE_LOG (newline-separated) and exits
// with the code in $CCW_FAKE_CLAUDE_EXIT (default 0).
package main

import (
 "os"
 "strconv"
 "strings"
)

func main() {
 logPath := os.Getenv("CCW_FAKE_CLAUDE_LOG")
 if logPath != "" {
  f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
  if err == nil {
   _, _ = f.WriteString(strings.Join(os.Args[1:], "\n") + "\n---\n")
   _ = f.Close()
  }
 }
 exit := 0
 if v := os.Getenv("CCW_FAKE_CLAUDE_EXIT"); v != "" {
  if n, err := strconv.Atoi(v); err == nil {
   exit = n
  }
 }
 os.Exit(exit)
}
```

- [ ] **Step 3: 統合テストを書く**

`tests/resume_flow_test.go`:

```go
package tests

import (
 "os"
 "os/exec"
 "path/filepath"
 "strings"
 "testing"
)

func buildBinary(t *testing.T, target, out string) {
 t.Helper()
 cmd := exec.Command("go", "build", "-o", out, target)
 if output, err := cmd.CombinedOutput(); err != nil {
  t.Fatalf("build %s: %v\n%s", target, err, output)
 }
}

func setupFakeEnv(t *testing.T) (binDir, logPath, home string) {
 t.Helper()
 binDir = t.TempDir()
 home = t.TempDir()
 buildBinary(t, "./fakes/fake_claude", filepath.Join(binDir, "claude"))
 buildBinary(t, "../cmd/ccw", filepath.Join(binDir, "ccw"))
 logPath = filepath.Join(t.TempDir(), "claude.log")
 return
}

func initRepo(t *testing.T, dir string) {
 t.Helper()
 for _, args := range [][]string{
  {"init", "-q"},
  {"commit", "--allow-empty", "-q", "-m", "init"},
 } {
  c := exec.Command("git", args...)
  c.Dir = dir
  if out, err := c.CombinedOutput(); err != nil {
   t.Fatalf("git %v: %v\n%s", args, err, out)
  }
 }
}

func runCcw(t *testing.T, binDir, repo, log, home string, args ...string) string {
 t.Helper()
 cmd := exec.Command(filepath.Join(binDir, "ccw"), args...)
 cmd.Dir = repo
 cmd.Env = append(os.Environ(),
  "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
  "HOME="+home,
  "CCW_FAKE_CLAUDE_LOG="+log,
  "NO_COLOR=1",
 )
 out, err := cmd.CombinedOutput()
 if err != nil {
  t.Fatalf("ccw %v: %v\n%s", args, err, out)
 }
 return string(out)
}

func readLog(t *testing.T, p string) []string {
 t.Helper()
 b, err := os.ReadFile(p)
 if err != nil {
  t.Fatalf("read log: %v", err)
 }
 return strings.Split(strings.TrimSpace(string(b)), "\n---\n")
}

func TestResumeFlow_NewWorktreePassesNameToBoth(t *testing.T) {
 binDir, log, home := setupFakeEnv(t)
 repo := t.TempDir()
 initRepo(t, repo)

 _ = runCcw(t, binDir, repo, log, home, "-n")

 calls := readLog(t, log)
 if len(calls) < 1 {
  t.Fatalf("expected at least 1 claude call, got %d", len(calls))
 }
 first := calls[0]
 if !strings.Contains(first, "--worktree\n") {
  t.Errorf("first call missing --worktree:\n%s", first)
 }
 if !strings.Contains(first, "\n-n\n") {
  t.Errorf("first call missing -n:\n%s", first)
 }
 // argument after --worktree should equal the argument after -n
 args := strings.Split(first, "\n")
 idxWT := indexOf(args, "--worktree")
 idxN := indexOf(args, "-n")
 if idxWT < 0 || idxN < 0 || idxWT+1 >= len(args) || idxN+1 >= len(args) {
  t.Fatalf("malformed args:\n%s", first)
 }
 if args[idxWT+1] != args[idxN+1] {
  t.Errorf("--worktree %q != -n %q", args[idxWT+1], args[idxN+1])
 }
}

func indexOf(s []string, target string) int {
 for i, v := range s {
  if v == target {
   return i
  }
 }
 return -1
}
```

注: picker 経由（ActionResume）の統合テストは TUI なので非インタラクティブ fallback パスでは `--continue` を直接通せない。fallback フローに `Continue` 経路が無い場合は、本テストは新規パス（`-n`）のみカバーで OK。後続テストは手動検証チェックリストに任せる。

- [ ] **Step 4: テストを実行**

Run: `go test ./tests/...`
Expected: PASS

- [ ] **Step 5: コミット**

```bash
git add tests
git commit -m "test: 新規 worktree で --worktree と -n が同一名であることを e2e で確認"
```

---

## Task 11: `.claude/settings.local.json` で読み取り権限を許可

**目的:** Claude Code 経由で本リポジトリを開発するときに `~/.claude/projects/` の読み取りプロンプトが出ないようにする。

**Files:**

- Modify or Create: `.claude/settings.local.json`

- [ ] **Step 1: 既存ファイルを確認**

Run: `cat .claude/settings.local.json 2>/dev/null || echo MISSING`

- [ ] **Step 2: 権限を追加**

ファイルが無い場合、新規作成:

```json
{
  "permissions": {
    "allow": ["Read(~/.claude/projects/**)"]
  }
}
```

ファイルが存在する場合、既存の `permissions.allow` 配列に `"Read(~/.claude/projects/**)"` を追加（重複していなければ）。

- [ ] **Step 3: gitignore 状態を確認**

Run: `git check-ignore .claude/settings.local.json && echo "ignored" || echo "tracked"`
Expected: `ignored` （`.claude/settings.local.json` は通常 gitignore 対象）

`tracked` の場合は明示的に追加コミットする。`ignored` の場合は本ファイルはローカル設定として残し、コミットしない（本タスクの spec 上の意図に沿う）。

- [ ] **Step 4: コミットの要否を判断**

ignored ならコミット不要 — Step 5 に進まずスキップ。

tracked ならコミット:

```bash
git add .claude/settings.local.json
git commit -m "chore: ~/.claude/projects/ への Read 許可を追加"
```

---

## Task 12: README を更新

**目的:** 旧 `--resume` 警告を削除し、新動作・命名規約を追記する。

**Files:**

- Modify: `docs/README.md`
- Modify: `docs/README.ja.md`

- [ ] **Step 1: 旧警告を削除**

`docs/README.ja.md` の `> ⚠️` から始まる `--resume ID` 関連ブロック（README.ja.md:79-80 付近）を削除。`docs/README.md` の対応する英語ブロックも削除。

- [ ] **Step 2: picker のサブメニュー説明を更新**

`docs/README.ja.md` の picker 説明で「`run` は選択した worktree で `claude --permission-mode auto` を新規起動するもので、Claude Code のセッション ID を引き継ぐ（`--resume` 相当の）操作は**行いません**。」の段を、新しい挙動に書き換える:

```markdown
worktree を選択すると `[r] run` / `[d] delete` / `[b] back` のサブメニューに遷移。`run` はセッションログが残っていれば `claude --continue` で**過去会話を復元**、無ければ `claude -n <worktree名>` で新規起動します。`[delete all]` / `[clean pushed]` / `[custom select]` は一括削除のショートカットで、dirty を含む場合は `--force` か、または 3 択確認 (`y` force · `s` dirty を除外 · `N` キャンセル) を経由します。
```

`docs/README.md` の対応箇所も同様に更新。

- [ ] **Step 3: RESUME / NEW バッジ表を追加**

`docs/README.ja.md` の Worktree 状態バッジ表のあとに追記:

```markdown
セッションバッジ:

| バッジ | 意味 |
|---|---|
| 💬 `RESUME` | 過去のセッションログがあり、`run` で会話を復元できる |
| ⚡ `NEW`    | セッションログ無し。`run` は新規起動 |
```

- [ ] **Step 4: 命名規約セクションを追加**

`## 📖 使い方` のあと（または picker 説明の直後）に挿入:

```markdown
### 命名規約

ccw は新規 worktree を作るとき、worktree 名と Claude Code のセッション名を 1:1 で揃えます:

- ディレクトリ: `<repo>/.claude/worktrees/<name>/`
- ブランチ: `worktree-<name>`
- セッション名: `<name>`（`claude -n <name>` で設定）

`<name>` は `quick-falcon-7bd2` のようなジェネレータ生成。手動で `/rename` した場合 ccw は何もしませんが、`--continue` は cwd 基準で動くので会話復元には影響しません。
```

英語版（`docs/README.md`）も対応訳で同様に追加。

- [ ] **Step 5: 依存欄の最低バージョンを引き上げ**

`docs/README.ja.md` の依存欄:

```markdown
- [Claude Code](https://docs.claude.com/claude-code) `>= 2.1.49`
```

を:

```markdown
- [Claude Code](https://docs.claude.com/claude-code) `>= 2.1.118` — ccw が利用する `--worktree <name>` と `-n <name>` の併用は 2.1.118 以降で動作確認済み
```

に変更。`docs/README.md` も同様に。

- [ ] **Step 6: textlint / markdownlint を通す**

Run: `pnpm dlx markdownlint-cli2 docs/README.md docs/README.ja.md` または lefthook が走る場合は staging 後 `git commit` で検出。

問題があれば修正。

- [ ] **Step 7: コミット**

```bash
git add docs/README.md docs/README.ja.md
git commit -m "docs: resume integration の挙動と命名規約を README に反映"
```

---

## Task 13: 全体検証 + PR 作成準備

**目的:** すべてのテスト・lint を通し、手動検証チェックリストを実行してから PR を作る準備をする。

**Files:**

- なし（検証のみ）

- [ ] **Step 1: 全 Go テスト**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 2: go vet**

Run: `go vet ./...`
Expected: 何も出力なし

- [ ] **Step 3: ビルド**

Run: `go build ./cmd/ccw`
Expected: `ccw` バイナリ生成

- [ ] **Step 4: lefthook pre-commit ローカル実行**

Run: `lefthook run pre-commit --all-files`
Expected: 成功

- [ ] **Step 5: 手動検証チェックリスト**

実機の sandbox repo で以下を確認:

```bash
mkdir -p /tmp/ccw-resume-check && cd /tmp/ccw-resume-check && git init -q && \
  git commit --allow-empty -q -m init
~/path/to/built/ccw -n
```

- [ ] picker から起動した worktree で `claude --worktree foo -n foo` 相当が走る
- [ ] 同 worktree で 2 回目の `ccw` (picker から `[r] run`) が前回会話を復元
- [ ] picker で RESUME / NEW バッジが分かれる
- [ ] `/rename` 後も `--continue` で復元できる
- [ ] `~/.claude/projects/<encoded>/` が実機で同じ規則
- [ ] `NO_COLOR=1 ccw` で表示崩れなし
- [ ] 80 cols 端末で L2 4 行が読める
- [ ] `CCW_DEBUG=1 ccw` で encoded path がログ出力される（実装していなければスキップ）
- [ ] TIPS が起動ごとに変わる

- [ ] **Step 6: ブランチを push して PR 作成**

```bash
git push -u origin HEAD
gh pr create --title "feat: worktree ↔ Claude Code session resume integration" --body "$(cat <<'EOF'
## Summary

- worktree 名と Claude Code セッション名を 1:1 マッピング
- picker から既存 worktree を選んだ時、デフォルトで `claude --continue` で過去会話を復元
- picker に `💬 RESUME` / `⚡ NEW` バッジ、L2 4 行レイアウト、random TIPS footer
- 旧 `--resume` 警告を撤去、命名規約と最低バージョンを更新

Spec: `docs/superpowers/specs/2026-04-25-worktree-resume-integration-design.md`
Plan: `docs/superpowers/plans/2026-04-25-worktree-resume-integration.md`

## Test plan

- [ ] `go test ./...` passes
- [ ] `lefthook run pre-commit --all-files` passes
- [ ] 手動: 新規 worktree → 同 worktree で再起動 → 会話復元
- [ ] 手動: NO_COLOR=1 で表示崩れなし
- [ ] 手動: TIPS が起動ごとに変わる

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## Self-Review Notes

- spec のすべてのコンポーネント（`namegen` は spec に明記されていなかったが `--worktree <name> -n <name>` を成立させる必須前提として追加。spec 側にもメモを追記すべきだが本計画の Task 1 内で明示的に補足）
- E1〜E9 のうち実装で直接触れるのは E2（`--continue` フォールバック）/ E3〜E4（`HasSession` の安全な false 返却）/ E5（`/rename` は cwd 基準で問題なし）。E1 はバージョンチェック（Task 0）でカバー。E6〜E9 は実装上の追加コードを必要としない
- Continue が `Resume` という名前から変わるため、`internal/picker/model.go` の `ActionResume` という列挙名は混乱を生むが、UI 上の意味は「選択 worktree で claude を起動」で本質は変わらないので**今回はリネームしない**（YAGNI / 関係ない変更を混ぜない）
- README で `worktree-<name>` ブランチ命名は、現状 `claude --worktree <name>` がどう命名するかに依存。Task 0 Step 2 で観察した実際のブランチ名に合わせて Task 12 Step 4 を最終調整する
