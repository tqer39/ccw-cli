# prunable worktree 対応 実装計画

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `git worktree list --porcelain` の `prunable` エントリで ccw が起動失敗 (exit 128) する問題を解消し、prunable な worktree を picker から `git worktree prune` で掃除できるようにする。

**Architecture:** `internal/gitx` の porcelain パーサに `prunable` 行を認識させ、`internal/worktree` に新ステータス `StatusPrunable` と `Prune` ヘルパーを追加。picker は prunable 行を新バッジで表示し、単独削除では `git worktree prune` を呼ぶ確認フローに分岐、bulk 削除では通常 `Remove` のあとに `Prune` を 1 回追加実行する。`cmd/ccw/main.go` 側で `Selection.IsPrunable` / `BulkDeletion.RunPrune` を見て `worktree.Prune` を呼び分ける。

**Tech Stack:** Go 1.22+, `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, 標準 `os/exec` (内部 `gitx`)。

**Reference Spec:** `docs/superpowers/specs/2026-04-25-prunable-worktree-handling-design.md`

---

## File Structure

| File | Responsibility | Change |
| --- | --- | --- |
| `internal/gitx/worktree.go` | `git worktree list --porcelain` のパースと `worktree remove` ラッパ | `WorktreeEntry.Prunable` 追加、`ParsePorcelain` 拡張、`Prune` 関数追加 |
| `internal/gitx/worktree_test.go` | 上記の単体テスト | prunable 行のテスト追加 |
| `internal/worktree/worktree.go` | ccw 視点の worktree 一覧と分類 | `StatusPrunable` 追加、`List` で prunable 短絡、`Prune` 追加 |
| `internal/worktree/worktree_test.go` | 上記の単体テスト | `String()` と `List` の prunable パス追加 |
| `internal/picker/model.go` | bubbletea Model / Selection / Bulk 型 | `Selection.IsPrunable`, `BulkDeletion.RunPrune` を追加 |
| `internal/picker/bulk.go` | bulk 選択ヘルパー | `HasPrunable` 追加 |
| `internal/picker/style.go` | ステータスバッジ (`Badge`, `Icon`) | `StatusPrunable` を扱う分岐追加 |
| `internal/picker/delegate.go` | リスト 1 行のレンダリング | prunable 行は ahead/behind 等を省略 |
| `internal/picker/update.go` | キー入力ハンドラ | `currentSelection` に `IsPrunable` を載せる、`Bulk()` で `RunPrune` を立てる |
| `internal/picker/view.go` | 各 state のビュー | `deleteConfirmView` を prunable 分岐、`bulkConfirmView` で prune の追加注記 |
| `internal/picker/model_test.go`, `bulk_test.go`, `style_test.go`, `view_test.go` | 上記のテスト | 必要に応じて追加 |
| `cmd/ccw/main.go` | picker の戻り値ハンドリング | `IsPrunable` / `RunPrune` で `worktree.Prune` を呼び分け |

---

## Task 1: `gitx` パーサに `prunable` 行と `Prune` 関数を追加

**Files:**

- Modify: `internal/gitx/worktree.go`
- Test: `internal/gitx/worktree_test.go`

- [ ] **Step 1.1: 失敗するテストを書く (`Prunable` フィールド + `prunable` 行のパース)**

`internal/gitx/worktree_test.go` に以下を追加:

```go
func TestParsePorcelain_Prunable(t *testing.T) {
 in := strings.Join([]string{
  "worktree /a/main",
  "HEAD abc123",
  "branch refs/heads/main",
  "",
  "worktree /a/.claude/worktrees/missing",
  "HEAD def456",
  "branch refs/heads/feature",
  "prunable gitdir file points to non-existent location",
  "",
 }, "\n")

 got := ParsePorcelain(in)
 want := []WorktreeEntry{
  {Path: "/a/main", Branch: "main", Prunable: false},
  {Path: "/a/.claude/worktrees/missing", Branch: "feature", Prunable: true},
 }
 if !reflect.DeepEqual(got, want) {
  t.Errorf("ParsePorcelain prunable:\n got  = %+v\n want = %+v", got, want)
 }
}

func TestParsePorcelain_PrunableNoReason(t *testing.T) {
 in := strings.Join([]string{
  "worktree /a/.claude/worktrees/missing",
  "HEAD def456",
  "branch refs/heads/feature",
  "prunable",
  "",
 }, "\n")

 got := ParsePorcelain(in)
 if len(got) != 1 || !got[0].Prunable {
  t.Errorf("ParsePorcelain prunable (no reason): got %+v", got)
 }
}
```

- [ ] **Step 1.2: テストを走らせて失敗することを確認**

Run: `go test ./internal/gitx/ -run TestParsePorcelain_Prunable -v`
Expected: 既存 `WorktreeEntry` に `Prunable` フィールドが無いためコンパイルエラー。

- [ ] **Step 1.3: `WorktreeEntry` と `ParsePorcelain` を実装**

`internal/gitx/worktree.go` を以下のように変更:

```go
// WorktreeEntry represents one record from `git worktree list --porcelain`.
type WorktreeEntry struct {
 Path     string
 Branch   string // without "refs/heads/" prefix; empty for detached HEAD
 Prunable bool   // true when git tagged this entry with `prunable`
}

// ListRaw returns every worktree attached to mainRepo. Caller is responsible
// for filtering (e.g. ccw-managed paths only).
func ListRaw(mainRepo string) ([]WorktreeEntry, error) {
 out, err := Output(mainRepo, "worktree", "list", "--porcelain")
 if err != nil {
  return nil, fmt.Errorf("git worktree list: %w", err)
 }
 return ParsePorcelain(out), nil
}

// ParsePorcelain parses `git worktree list --porcelain` output.
func ParsePorcelain(s string) []WorktreeEntry {
 var entries []WorktreeEntry
 var cur WorktreeEntry
 flush := func() {
  if cur.Path != "" {
   entries = append(entries, cur)
  }
  cur = WorktreeEntry{}
 }
 for _, line := range strings.Split(s, "\n") {
  switch {
  case strings.HasPrefix(line, "worktree "):
   flush()
   cur.Path = strings.TrimPrefix(line, "worktree ")
  case strings.HasPrefix(line, "branch "):
   cur.Branch = strings.TrimPrefix(
    strings.TrimPrefix(line, "branch "),
    "refs/heads/",
   )
  case line == "prunable" || strings.HasPrefix(line, "prunable "):
   cur.Prunable = true
  case line == "":
   flush()
  }
 }
 flush()
 return entries
}
```

- [ ] **Step 1.4: 新 `Prune` 関数のテストを書く**

`internal/gitx/worktree_test.go` に追加:

```go
func TestPrune_Integration(t *testing.T) {
 main := initRepo(t)
 mustRun(t, main, "git", "commit", "--allow-empty", "-m", "init")
 wt := filepath.Join(main, ".claude", "worktrees", "doomed")
 mustRun(t, main, "git", "worktree", "add", "-b", "doomed-branch", wt)

 // Manually delete the worktree directory so git marks it prunable.
 if err := os.RemoveAll(wt); err != nil {
  t.Fatalf("rm worktree dir: %v", err)
 }

 // Sanity: the entry should now be prunable in the porcelain output.
 entries, err := ListRaw(main)
 if err != nil {
  t.Fatalf("ListRaw: %v", err)
 }
 var foundPrunable bool
 for _, e := range entries {
  if e.Path == wt && e.Prunable {
   foundPrunable = true
  }
 }
 if !foundPrunable {
  t.Fatalf("expected prunable entry for %s, got %+v", wt, entries)
 }

 if err := Prune(main); err != nil {
  t.Fatalf("Prune: %v", err)
 }

 entries, err = ListRaw(main)
 if err != nil {
  t.Fatalf("ListRaw after prune: %v", err)
 }
 for _, e := range entries {
  if e.Path == wt {
   t.Errorf("worktree %q still present after Prune", wt)
  }
 }
}
```

注: `os` import を追加してください (まだ未 import の場合)。

- [ ] **Step 1.5: テストを走らせて失敗することを確認**

Run: `go test ./internal/gitx/ -run TestPrune_Integration -v`
Expected: `Prune` 未定義のためコンパイルエラー。

- [ ] **Step 1.6: `Prune` を実装**

`internal/gitx/worktree.go` の末尾に追加:

```go
// Prune calls `git -C mainRepo worktree prune`. This removes admin files for
// worktrees whose working directory has been deleted (i.e. those flagged as
// `prunable` in `git worktree list --porcelain`).
func Prune(mainRepo string) error {
 if err := Run(mainRepo, "worktree", "prune"); err != nil {
  return fmt.Errorf("git worktree prune: %w", err)
 }
 return nil
}
```

- [ ] **Step 1.7: パッケージ全体のテストが通ることを確認**

Run: `go test ./internal/gitx/ -v`
Expected: PASS (既存テスト含めグリーン)。

- [ ] **Step 1.8: コミット**

```bash
git add internal/gitx/worktree.go internal/gitx/worktree_test.go
git commit -m "feat(gitx): parse prunable line and add Prune helper

git worktree list --porcelain lists prunable entries (working directory
gone but admin files remain). ParsePorcelain now records this on
WorktreeEntry.Prunable, and the new Prune helper wraps
git worktree prune for callers that want to clean them up."
```

---

## Task 2: `worktree` パッケージに `StatusPrunable` と `Prune` を追加

**Files:**

- Modify: `internal/worktree/worktree.go`
- Test: `internal/worktree/worktree_test.go`

- [ ] **Step 2.1: `String()` のテストを追加**

`internal/worktree/worktree_test.go` の `TestStatus_String` を拡張:

```go
func TestStatus_String(t *testing.T) {
 cases := []struct {
  s    Status
  want string
 }{
  {StatusPushed, "pushed"},
  {StatusLocalOnly, "local-only"},
  {StatusDirty, "dirty"},
  {StatusPrunable, "prunable"},
 }
 for _, tc := range cases {
  if got := tc.s.String(); got != tc.want {
   t.Errorf("Status(%d).String() = %q, want %q", tc.s, got, tc.want)
  }
 }
}
```

- [ ] **Step 2.2: テストを走らせて失敗することを確認**

Run: `go test ./internal/worktree/ -run TestStatus_String -v`
Expected: `StatusPrunable` 未定義のためコンパイルエラー。

- [ ] **Step 2.3: `StatusPrunable` を実装**

`internal/worktree/worktree.go` の `Status` 定数群と `String()` を更新:

```go
const (
 // StatusPushed means clean, upstream exists, ahead == 0.
 StatusPushed Status = iota
 // StatusLocalOnly means clean, but either no upstream or ahead > 0.
 StatusLocalOnly
 // StatusDirty means the working tree has untracked or modified files.
 StatusDirty
 // StatusPrunable means the working directory is gone but git still
 // keeps admin files for it. Cleared by `git worktree prune`.
 StatusPrunable
)

// String returns the short lowercase label used in picker UI.
func (s Status) String() string {
 switch s {
 case StatusPushed:
  return "pushed"
 case StatusLocalOnly:
  return "local-only"
 case StatusDirty:
  return "dirty"
 case StatusPrunable:
  return "prunable"
 default:
  return "unknown"
 }
}
```

- [ ] **Step 2.4: 走らせて Step 2.1 が通ることを確認**

Run: `go test ./internal/worktree/ -run TestStatus_String -v`
Expected: PASS。

- [ ] **Step 2.5: `List` の prunable 短絡テストを書く**

`internal/worktree/worktree_test.go` に追加:

```go
func TestList_PrunableEntryDoesNotTouchDisk(t *testing.T) {
 main := initMainRepo(t)
 wt := addWorktree(t, main, "doomed")

 // Delete the worktree dir so git marks it prunable.
 if err := os.RemoveAll(wt); err != nil {
  t.Fatalf("RemoveAll: %v", err)
 }

 infos, err := List(main)
 if err != nil {
  t.Fatalf("List: %v", err)
 }
 var found bool
 for _, in := range infos {
  if in.Branch == "doomed-branch" {
   found = true
   if in.Status != StatusPrunable {
    t.Errorf("Status = %s, want prunable", in.Status)
   }
   if in.AheadCount != 0 || in.BehindCount != 0 || in.DirtyCount != 0 {
    t.Errorf("counts should be zero for prunable, got %+v", in)
   }
   if in.HasSession {
    t.Errorf("HasSession should be false for prunable")
   }
  }
 }
 if !found {
  t.Fatalf("doomed-branch not in List() output: %+v", infos)
 }
}
```

注: `os` import を追加してください (まだ未 import の場合)。

- [ ] **Step 2.6: テストを走らせて失敗することを確認**

Run: `go test ./internal/worktree/ -run TestList_PrunableEntryDoesNotTouchDisk -v`
Expected: `List` が `Classify` を呼び存在しないパスで `git status` を実行 → エラーで FAIL。

- [ ] **Step 2.7: `List` を更新して prunable を短絡**

`internal/worktree/worktree.go` の `List` を以下に置き換え:

```go
// List returns ccw-managed worktrees under mainRepo, each classified.
func List(mainRepo string) ([]Info, error) {
 entries, err := gitx.ListRaw(mainRepo)
 if err != nil {
  return nil, fmt.Errorf("list worktrees: %w", err)
 }
 var result []Info
 for _, e := range entries {
  if !strings.Contains(e.Path, ccwPathMarker) {
   continue
  }
  if e.Prunable {
   result = append(result, Info{
    Path:   e.Path,
    Branch: e.Branch,
    Status: StatusPrunable,
   })
   continue
  }
  st, err := Classify(e.Path)
  if err != nil {
   return nil, err
  }
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
 }
 return result, nil
}
```

- [ ] **Step 2.8: テストを走らせて通ることを確認**

Run: `go test ./internal/worktree/ -run TestList_PrunableEntryDoesNotTouchDisk -v`
Expected: PASS。

- [ ] **Step 2.9: `Prune` ラッパを追加**

`internal/worktree/worktree.go` の末尾に追加:

```go
// Prune cleans up admin files for prunable worktrees attached to mainRepo.
// Wraps `git -C mainRepo worktree prune`.
func Prune(mainRepo string) error {
 if err := gitx.Prune(mainRepo); err != nil {
  return fmt.Errorf("prune worktrees: %w", err)
 }
 return nil
}
```

- [ ] **Step 2.10: パッケージ全体テスト**

Run: `go test ./internal/worktree/ -v`
Expected: PASS。

- [ ] **Step 2.11: コミット**

```bash
git add internal/worktree/worktree.go internal/worktree/worktree_test.go
git commit -m "feat(worktree): add StatusPrunable and Prune helper

When git worktree list reports a prunable entry, surface it as
StatusPrunable instead of running git status on a missing path
(which fails with exit 128 and broke ccw startup). The new Prune
helper wraps gitx.Prune so the picker can clean these up."
```

---

## Task 3: `picker.Selection` / `picker.BulkDeletion` に prunable シグナルを追加

**Files:**

- Modify: `internal/picker/model.go`
- Modify: `internal/picker/bulk.go`
- Test: `internal/picker/bulk_test.go`

- [ ] **Step 3.1: `HasPrunable` ヘルパーのテストを書く**

`internal/picker/bulk_test.go` に追加 (ファイル存在を確認: `internal/picker/bulk_test.go`):

```go
func TestHasPrunable_True(t *testing.T) {
 infos := []worktree.Info{
  {Path: "/a", Status: worktree.StatusPushed},
  {Path: "/b", Status: worktree.StatusPrunable},
 }
 if !HasPrunable(infos, []int{0, 1}) {
  t.Error("HasPrunable should return true when a prunable index is included")
 }
}

func TestHasPrunable_False(t *testing.T) {
 infos := []worktree.Info{
  {Path: "/a", Status: worktree.StatusPushed},
  {Path: "/b", Status: worktree.StatusDirty},
 }
 if HasPrunable(infos, []int{0, 1}) {
  t.Error("HasPrunable should return false without prunable")
 }
}
```

注: 既存の `bulk_test.go` の import 句に `"github.com/tqer39/ccw-cli/internal/worktree"` と `"testing"` が無ければ追加。

- [ ] **Step 3.2: テストを走らせて失敗することを確認**

Run: `go test ./internal/picker/ -run TestHasPrunable -v`
Expected: `HasPrunable` 未定義でコンパイルエラー。

- [ ] **Step 3.3: `HasPrunable` を実装**

`internal/picker/bulk.go` の末尾に追加:

```go
// HasPrunable reports whether any of the given indices references a prunable worktree.
func HasPrunable(infos []worktree.Info, indices []int) bool {
 for _, i := range indices {
  if infos[i].Status == worktree.StatusPrunable {
   return true
  }
 }
 return false
}
```

- [ ] **Step 3.4: テストを通す**

Run: `go test ./internal/picker/ -run TestHasPrunable -v`
Expected: PASS。

- [ ] **Step 3.5: `Selection.IsPrunable` / `BulkDeletion.RunPrune` を追加**

`internal/picker/model.go` の該当型を以下に置き換え:

```go
// Selection identifies the worktree the user picked.
type Selection struct {
 Path        string
 Branch      string
 Status      worktree.Status
 HasSession  bool
 ForceDelete bool
 IsPrunable  bool
}

// BulkDeletion describes the set of worktrees to remove in a bulk delete.
type BulkDeletion struct {
 Paths    []string
 Force    bool
 RunPrune bool
}
```

- [ ] **Step 3.6: `Bulk()` を更新して `RunPrune` を立てる**

`internal/picker/model.go` の `Bulk()` を以下に置き換え:

```go
// Bulk returns the bulk-delete descriptor (valid after ActionBulkDelete).
func (m Model) Bulk() BulkDeletion {
 paths := make([]string, 0, len(m.bulkTargets))
 hasPrunable := false
 for _, i := range m.bulkTargets {
  w := m.infos[i]
  if w.Status == worktree.StatusPrunable {
   hasPrunable = true
   // Skip path: git worktree remove cannot operate on prunable
   // entries (path no longer exists). They are handled by Prune.
   continue
  }
  paths = append(paths, w.Path)
 }
 return BulkDeletion{Paths: paths, Force: m.bulkForce, RunPrune: hasPrunable}
}
```

理由のコメントが冗長な場合は短縮可。WHY が伝わる長さで残してください。

- [ ] **Step 3.7: パッケージビルドを通す**

Run: `go build ./...`
Expected: PASS (この時点では呼び出し側 `cmd/ccw/main.go` は新フィールドを参照しないので OK)。

- [ ] **Step 3.8: コミット**

```bash
git add internal/picker/bulk.go internal/picker/bulk_test.go internal/picker/model.go
git commit -m "feat(picker): add prunable signals to Selection/BulkDeletion

Selection.IsPrunable and BulkDeletion.RunPrune let the caller decide
when to invoke worktree.Prune instead of (or in addition to)
worktree.Remove. Bulk paths now exclude prunable entries since
git worktree remove cannot act on them."
```

---

## Task 4: ステータスバッジに `prunable` を追加

**Files:**

- Modify: `internal/picker/style.go`
- Modify: `internal/picker/delegate.go`
- Test: `internal/picker/style_test.go`, `internal/picker/model_test.go`

- [ ] **Step 4.1: `Icon` のテストを更新**

`internal/picker/model_test.go` の `TestIcon` の cases に prunable を追加:

```go
func TestIcon(t *testing.T) {
 cases := []struct {
  s    worktree.Status
  want string
 }{
  {worktree.StatusPushed, "✅"},
  {worktree.StatusLocalOnly, "⚠"},
  {worktree.StatusDirty, "⛔"},
  {worktree.StatusPrunable, "🧹"},
  {worktree.Status(99), "•"},
 }
 for _, tc := range cases {
  if got := Icon(tc.s); got != tc.want {
   t.Errorf("Icon(%s) = %q, want %q", tc.s, got, tc.want)
  }
 }
}
```

- [ ] **Step 4.2: `Badge` の `NO_COLOR` パスをテスト**

`internal/picker/style_test.go` を確認し、無ければ追加。`NO_COLOR=1` 下で `Badge(StatusPrunable)` が `[prunable]` を返すことを確認するテストを足す:

```go
func TestBadge_PrunableNoColor(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 got := Badge(worktree.StatusPrunable)
 if got != "[prune] " {
  t.Errorf("Badge(prunable) NO_COLOR = %q, want %q", got, "[prune] ")
 }
}
```

(`[prune]` は他のラベル `[pushed]` `[local]` `[dirty]` と同じ 8 文字幅にそろえた値です。)

- [ ] **Step 4.3: テストを走らせて失敗することを確認**

Run: `go test ./internal/picker/ -run "TestIcon|TestBadge_Prunable" -v`
Expected: 期待値ミスマッチで FAIL。

- [ ] **Step 4.4: `Icon` / `Badge` / `badgeLabel` に prunable 分岐を追加**

`internal/picker/style.go` の `Badge` を以下に置き換え (色は灰色系: `8` 背景 / `15` 前景):

```go
// Badge renders a fixed-width status badge for a Status.
// Respects NO_COLOR=1 by returning a plain-text bracketed label.
func Badge(s worktree.Status) string {
 label, plain := badgeLabel(s)
 if noColor() {
  return plain
 }
 style := lipgloss.NewStyle().Padding(0, 1).Bold(true)
 switch s {
 case worktree.StatusPushed:
  style = style.Background(lipgloss.Color("10")).Foreground(lipgloss.Color("0"))
 case worktree.StatusLocalOnly:
  style = style.Background(lipgloss.Color("11")).Foreground(lipgloss.Color("0"))
 case worktree.StatusDirty:
  style = style.Background(lipgloss.Color("9")).Foreground(lipgloss.Color("15"))
 case worktree.StatusPrunable:
  style = style.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
 }
 return style.Render(label)
}

func badgeLabel(s worktree.Status) (colored, plain string) {
 switch s {
 case worktree.StatusPushed:
  return "PUSHED", "[pushed]"
 case worktree.StatusLocalOnly:
  return "LOCAL ", "[local] "
 case worktree.StatusDirty:
  return "DIRTY ", "[dirty] "
 case worktree.StatusPrunable:
  return "PRUNE ", "[prune] "
 default:
  return "??????", "[?]     "
 }
}
```

`Icon` を更新:

```go
func Icon(s worktree.Status) string {
 switch s {
 case worktree.StatusPushed:
  return "✅"
 case worktree.StatusLocalOnly:
  return "⚠"
 case worktree.StatusDirty:
  return "⛔"
 case worktree.StatusPrunable:
  return "🧹"
 default:
  return "•"
 }
}
```

- [ ] **Step 4.5: テストを通す**

Run: `go test ./internal/picker/ -run "TestIcon|TestBadge_Prunable" -v`
Expected: PASS。

- [ ] **Step 4.6: delegate (行レンダラ) で prunable は ahead/behind を出さないようにする**

`internal/picker/delegate.go` の `renderRow` を以下に置き換え:

```go
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

 var indicators string
 switch wt.Status {
 case worktree.StatusPrunable:
  indicators = "(missing on disk)"
 case worktree.StatusDirty:
  indicators = fmt.Sprintf("↑%d ↓%d ✎%d", wt.AheadCount, wt.BehindCount, wt.DirtyCount)
 default:
  indicators = fmt.Sprintf("↑%d ↓%d", wt.AheadCount, wt.BehindCount)
 }

 header := fmt.Sprintf("%s%s · %s", prefix, resume, name)
 right := fmt.Sprintf("%s  %s", status, indicators)
 header = padBetween(header, right, width)

 branchLine := fmt.Sprintf("    branch:  %s", wt.Branch)
 prCell := ""
 if !prUnavailable {
  prCell = renderPRCell(li.pr)
 }
 prLine := "    pr:      " + prCell
 pathLine := fmt.Sprintf("    path:    %s", wt.Path)

 if width > 0 {
  header = truncateToWidth(header, width)
  branchLine = truncateToWidth(branchLine, width)
  prLine = truncateToWidth(prLine, width)
  pathLine = truncateToWidth(pathLine, width)
 }

 return header + "\n" + branchLine + "\n" + prLine + "\n" + pathLine
}
```

- [ ] **Step 4.7: パッケージ全体テスト**

Run: `go test ./internal/picker/ -v`
Expected: PASS。

- [ ] **Step 4.8: コミット**

```bash
git add internal/picker/style.go internal/picker/delegate.go internal/picker/style_test.go internal/picker/model_test.go
git commit -m "feat(picker): render prunable status with gray badge

Adds the prunable badge ([prune]/PRUNE) and the 🧹 icon in the legacy
Icon API. Prunable rows replace ahead/behind/dirty indicators with
a (missing on disk) hint since the working directory is gone."
```

---

## Task 5: 単独削除フローを prunable で分岐

**Files:**

- Modify: `internal/picker/update.go`
- Modify: `internal/picker/view.go`
- Test: `internal/picker/model_test.go`

- [ ] **Step 5.1: prunable 単独削除のテストを書く**

`internal/picker/model_test.go` に追加:

```go
func TestDeleteConfirm_Prunable_SetsIsPrunable(t *testing.T) {
 m := New([]worktree.Info{
  {Path: "/a/.claude/worktrees/missing", Branch: "missing", Status: worktree.StatusPrunable},
 })
 m = goToDeleteConfirm(m)
 next, _ := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
 mm := next.(Model)
 if mm.Action() != ActionDelete {
  t.Errorf("Action = %s, want delete", mm.Action())
 }
 if !mm.Selection().IsPrunable {
  t.Error("Selection.IsPrunable should be true for prunable row")
 }
 if mm.Selection().ForceDelete {
  t.Error("ForceDelete should be false for prunable (no remove --force)")
 }
}

func TestDeleteConfirmView_PrunableSinglePromptsPrune(t *testing.T) {
 m := New([]worktree.Info{
  {Path: "/a/.claude/worktrees/missing", Branch: "missing", Status: worktree.StatusPrunable},
 })
 m = goToDeleteConfirm(m)
 out := m.View().Content
 if !strings.Contains(out, "git worktree prune") {
  t.Errorf("prunable confirm view must mention git worktree prune:\n%s", out)
 }
 // only one prunable in the list -> short prompt, no enumeration
 if strings.Contains(out, "following") || strings.Contains(out, "以下") {
  t.Errorf("single-prunable view should not enumerate, got:\n%s", out)
 }
}

func TestDeleteConfirmView_PrunableMultipleEnumerates(t *testing.T) {
 m := New([]worktree.Info{
  {Path: "/a/.claude/worktrees/p1", Branch: "p1", Status: worktree.StatusPrunable},
  {Path: "/a/.claude/worktrees/p2", Branch: "p2", Status: worktree.StatusPrunable},
 })
 m = goToDeleteConfirm(m)
 out := m.View().Content
 for _, want := range []string{"git worktree prune", "/a/.claude/worktrees/p1", "/a/.claude/worktrees/p2"} {
  if !strings.Contains(out, want) {
   t.Errorf("multi-prunable confirm view missing %q:\n%s", want, out)
  }
 }
}
```

- [ ] **Step 5.2: テストを走らせて失敗することを確認**

Run: `go test ./internal/picker/ -run "TestDeleteConfirm_Prunable|TestDeleteConfirmView_Prunable" -v`
Expected: 表示文言と `IsPrunable` 不在で FAIL。

- [ ] **Step 5.3: `currentSelection` を更新して `IsPrunable` を載せる**

`internal/picker/update.go` の `currentSelection` を以下に置き換え:

```go
func (m Model) currentSelection() Selection {
 w := m.infos[m.selIdx]
 return Selection{
  Path:       w.Path,
  Branch:     w.Branch,
  Status:     w.Status,
  HasSession: w.HasSession,
  IsPrunable: w.Status == worktree.StatusPrunable,
 }
}
```

- [ ] **Step 5.4: `updateDeleteConfirm` の `y` 分岐で prunable は force を立てない**

既に上記の `currentSelection` 経由で Status は伝わるので、現状の `updateDeleteConfirm` は

```go
sel.ForceDelete = sel.Status == worktree.StatusDirty
```

のままで十分 (prunable は dirty 判定にはならないので false 維持)。変更不要。

- [ ] **Step 5.5: `deleteConfirmView` を prunable 分岐に書き換え**

`internal/picker/view.go` の `deleteConfirmView` を以下に置き換え:

```go
func (m Model) deleteConfirmView() string {
 w := m.infos[m.selIdx]
 if w.Status == worktree.StatusPrunable {
  return m.prunableConfirmView()
 }
 cmd := fmt.Sprintf("git worktree remove %q", w.Path)
 if w.Status == worktree.StatusDirty {
  cmd = fmt.Sprintf("git worktree remove --force %q", w.Path)
 }
 return fmt.Sprintf(
  "Delete worktree %s?\n  path:   %s\n  status: %s\n\nThis will run: %s\n\nConfirm? [y/N]\n",
  w.Branch, w.Path, w.Status, cmd,
 )
}

func (m Model) prunableConfirmView() string {
 var prunables []worktree.Info
 for _, in := range m.infos {
  if in.Status == worktree.StatusPrunable {
   prunables = append(prunables, in)
  }
 }
 if len(prunables) <= 1 {
  w := m.infos[m.selIdx]
  return fmt.Sprintf(
   "Prune worktree %s?\n  path:   %s\n\nThis will run: git worktree prune\n\nConfirm? [y/N]\n",
   w.Branch, w.Path,
  )
 }
 var b strings.Builder
 fmt.Fprintf(&b, "Prune %d prunable worktrees? (git worktree prune removes all of them at once)\n\n", len(prunables))
 for _, p := range prunables {
  fmt.Fprintf(&b, "  %s %s\n", p.Branch, p.Path)
 }
 b.WriteString("\nThis will run: git worktree prune\n\nConfirm? [y/N]\n")
 return b.String()
}
```

`internal/picker/view.go` の import に `"github.com/tqer39/ccw-cli/internal/worktree"` が無ければ追加 (既に存在)。

- [ ] **Step 5.6: テストを通す**

Run: `go test ./internal/picker/ -run "TestDeleteConfirm_Prunable|TestDeleteConfirmView_Prunable" -v`
Expected: PASS。

- [ ] **Step 5.7: パッケージ全体テスト**

Run: `go test ./internal/picker/ -v`
Expected: PASS (既存の `TestDeleteConfirm_*` も green のまま)。

- [ ] **Step 5.8: コミット**

```bash
git add internal/picker/update.go internal/picker/view.go internal/picker/model_test.go
git commit -m "feat(picker): branch single delete confirm for prunable

Selecting a prunable row and pressing 'd' now shows a Prune prompt
that runs git worktree prune. When more than one prunable exists,
the confirm view enumerates all affected entries so the user knows
the prune is global, not just the selected one."
```

---

## Task 6: bulk 削除フローで prunable を許容

**Files:**

- Modify: `internal/picker/view.go`
- Test: `internal/picker/model_test.go`

- [ ] **Step 6.1: prunable を含む bulk のテストを書く**

`internal/picker/model_test.go` に追加:

```go
func TestUpdate_DeleteAll_WithPrunable_SetsRunPrune(t *testing.T) {
 m := New([]worktree.Info{
  {Branch: "a", Path: "/a", Status: worktree.StatusPushed},
  {Branch: "p", Path: "/p", Status: worktree.StatusPrunable},
 })
 m.bulkTargets = []int{0, 1}
 m.state = stateBulkConfirm
 got, _ := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
 mm := got.(Model)
 if mm.Action() != ActionBulkDelete {
  t.Errorf("Action = %s, want bulk-delete", mm.Action())
 }
 b := mm.Bulk()
 if !b.RunPrune {
  t.Error("RunPrune should be true when a prunable target is selected")
 }
 if len(b.Paths) != 1 || b.Paths[0] != "/a" {
  t.Errorf("Paths = %v, want only /a (prunable excluded)", b.Paths)
 }
}

func TestBulkConfirmView_ShowsPruneNote(t *testing.T) {
 m := New([]worktree.Info{
  {Branch: "a", Path: "/a", Status: worktree.StatusPushed},
  {Branch: "p", Path: "/p", Status: worktree.StatusPrunable},
 })
 m.bulkTargets = []int{0, 1}
 m.state = stateBulkConfirm
 out := m.View().Content
 if !strings.Contains(out, "git worktree prune") {
  t.Errorf("bulkConfirmView with prunable target must mention git worktree prune:\n%s", out)
 }
}
```

- [ ] **Step 6.2: テストを走らせて確認**

Run: `go test ./internal/picker/ -run "TestUpdate_DeleteAll_WithPrunable|TestBulkConfirmView_ShowsPruneNote" -v`
Expected: 1 つ目はおそらく既に PASS する (Task 3 の `Bulk()` で実装済み)、2 つ目は注記文言が無いので FAIL。

- [ ] **Step 6.3: `bulkConfirmView` に prune 注記を追加**

`internal/picker/view.go` の `bulkConfirmView` を以下に置き換え:

```go
func (m Model) bulkConfirmView() string {
 var b strings.Builder
 fmt.Fprintf(&b, "Delete %d worktrees?\n\n", len(m.bulkTargets))
 hasDirty := HasDirty(m.infos, m.bulkTargets)
 hasPrunable := HasPrunable(m.infos, m.bulkTargets)
 dirtyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
 for _, i := range m.bulkTargets {
  w := m.infos[i]
  line := fmt.Sprintf("  %s %s  %s\n", Badge(w.Status), w.Branch, w.Path)
  if w.Status == worktree.StatusDirty && !noColor() {
   line = dirtyStyle.Render(line)
  }
  b.WriteString(line)
 }
 if hasPrunable {
  b.WriteString("\nℹ Prunable entries will be cleaned up via `git worktree prune` after the removals.\n")
 }
 if hasDirty {
  b.WriteString("\n⚠ Dirty worktrees are included. `git worktree remove --force` is required.\n")
  b.WriteString("  [y] yes (include dirty, use --force)\n")
  b.WriteString("  [s] skip dirty (remove clean only)\n")
  b.WriteString("  [N] cancel\n")
 } else {
  b.WriteString("\nConfirm? [y/N]\n")
 }
 return b.String()
}
```

- [ ] **Step 6.4: テストを通す**

Run: `go test ./internal/picker/ -run "TestUpdate_DeleteAll_WithPrunable|TestBulkConfirmView_ShowsPruneNote" -v`
Expected: PASS。

- [ ] **Step 6.5: パッケージ全体テスト**

Run: `go test ./internal/picker/ -v`
Expected: PASS。

- [ ] **Step 6.6: コミット**

```bash
git add internal/picker/view.go internal/picker/model_test.go
git commit -m "feat(picker): support prunable entries in bulk delete

Bulk confirm view notes when a prune step will follow the removals,
and the existing flow lets users mix prunable + normal targets without
hitting the dirty-confirmation branch."
```

---

## Task 7: `cmd/ccw/main.go` で `Prune` を呼び分け

**Files:**

- Modify: `cmd/ccw/main.go`

- [ ] **Step 7.1: `runPicker` の delete アクションで `IsPrunable` を分岐**

`cmd/ccw/main.go` の `case picker.ActionDelete:` ブロックを以下に置き換え:

```go
case picker.ActionDelete:
 if sel.IsPrunable {
  if err := worktree.Prune(mainRepo); err != nil {
   ui.Error("%v", err)
   return 1
  }
  ui.Success("Pruned worktree admin files")
  continue
 }
 if err := worktree.Remove(mainRepo, sel.Path, sel.ForceDelete); err != nil {
  ui.Error("%v", err)
  return 1
 }
 ui.Success("Removed %s", sel.Path)
```

注: 既存ループは `for {}` なので `continue` で次の picker iteration に戻ります。`Removed %s` を出していた既存パスは保持。

- [ ] **Step 7.2: `applyBulkDelete` で `RunPrune` を最後に処理**

`cmd/ccw/main.go` の `applyBulkDelete` を以下に置き換え:

```go
func applyBulkDelete(mainRepo string, bulk picker.BulkDeletion) int {
 errs := 0
 for _, p := range bulk.Paths {
  if err := worktree.Remove(mainRepo, p, bulk.Force); err != nil {
   ui.Error("remove %s: %v", p, err)
   errs++
   continue
  }
  ui.Success("Removed %s", p)
 }
 if bulk.RunPrune {
  if err := worktree.Prune(mainRepo); err != nil {
   ui.Error("prune: %v", err)
   errs++
  } else {
   ui.Success("Pruned worktree admin files")
  }
 }
 if errs > 0 {
  return 1
 }
 return 0
}
```

- [ ] **Step 7.3: ビルドとリポジトリ全体テスト**

Run: `go build ./... && go test ./...`
Expected: PASS。

- [ ] **Step 7.4: コミット**

```bash
git add cmd/ccw/main.go
git commit -m "feat(ccw): wire up prunable delete and bulk prune

Selecting a prunable row in the picker now runs worktree.Prune,
and bulk delete invocations follow the removals with a single
prune call when prunable targets were selected."
```

---

## Task 8: 手動受け入れテスト

**Files:** なし (手動確認)。

- [ ] **Step 8.1: ローカルバイナリをビルド**

Run (ccw-cli リポジトリ ルートから):

```bash
go build -o /tmp/ccw-prunable ./cmd/ccw
```

Expected: バイナリが生成される。

- [ ] **Step 8.2: 再現リポジトリで起動を確認**

Run:

```bash
/tmp/ccw-prunable
```

を `~/workspace/tqer39/terraform-github` で実行。

Expected:

- `✖ list worktrees: classify status: git status --porcelain: exit status 128` が出ない。
- picker に `synchronous-finding-wave` が `[prune]` バッジ付きで表示される。

- [ ] **Step 8.3: 単独 prune を確認**

picker で `synchronous-finding-wave` を選択 → `d` → 確認画面が `git worktree prune` を案内 → `y`。

Expected:

- 一覧から該当行が消える。
- `/usr/bin/git -C ~/workspace/tqer39/terraform-github worktree list --porcelain | grep prunable` が空。

- [ ] **Step 8.4: 複数 prunable のケース (任意)**

可能であれば `~/workspace/tqer39/terraform-github` 直下に worktree を一時的に 2 つ追加 → ディレクトリ実体を `rm -rf` で消して 2 件 prunable を作る → ccw 起動 → 1 件選択 → 確認画面が両方の path を列挙することを確認 → `y` → 両方消えること、起動が成功すること。

- [ ] **Step 8.5: 既存 worktree への非影響**

通常 worktree (例: `cryptic-stirring-catmull`) を選んで `d` → 既存どおり `git worktree remove` が走り削除されること。bulk delete も既存どおり動くこと (prunable 無しのケース)。

---

## Self-Review チェック (writer 用メモ)

実装完了後、以下を確認:

1. **spec カバレッジ**:
   - `WorktreeEntry.Prunable` / `ParsePorcelain` 拡張 → Task 1
   - `gitx.Prune` → Task 1
   - `StatusPrunable` / `String()` → Task 2
   - `worktree.List` の prunable 短絡 → Task 2
   - `worktree.Prune` → Task 2
   - picker の prunable バッジ → Task 4
   - `delegate` ahead/behind 抑制 → Task 4
   - 単独削除 prunable 分岐 + 1/N 件出し分け → Task 5
   - bulk delete に prunable 含める + 末尾 prune → Tasks 3, 6, 7
   - `cmd/ccw/main.go` 結線 → Task 7
   - 手動受け入れ → Task 8
2. **placeholder スキャン**: TBD/TODO/省略コードはなし、全 step に実コードあり。
3. **型整合**: `Selection.IsPrunable`, `BulkDeletion.RunPrune`, `HasPrunable`, `worktree.Prune`, `gitx.Prune` の名前が全 task で一致。

---

**Plan complete and saved to `docs/superpowers/plans/2026-04-25-prunable-worktree-handling.md`. Two execution options:**

1. **Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. **Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
