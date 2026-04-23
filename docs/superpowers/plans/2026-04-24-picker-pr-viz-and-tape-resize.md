# picker の PR 視覚強化と tape サイズ縮小 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** picker 行の PR 部分に状態バッジ／薄背景／動線グリフを導入し、worktree ↔ PR の視覚的な紐づけを強化する。さらにデモ GIF をコンパクトなサイズで再録画する。

**Architecture:** `internal/picker/style.go` に PR 状態 → 配色マッピングと NO_COLOR フォールバックを追加。`delegate.go` の `renderRow` を、PR セル描画を `renderPRCell` に切り出して `→` グリフを差し込む形にリファクタ。既存の 2 行構成（上段メタ情報 / 下段 path）は維持。tape は `1024x640 / FontSize 24 / Padding 20` にリサイズ。

**Tech Stack:** Go 1.25, charmbracelet/bubbletea + bubbles/list + lipgloss, vhs (tape ファイル).

**Spec:** `docs/superpowers/specs/2026-04-24-picker-pr-viz-and-tape-resize-design.md`

---

## File Structure

| File | Role | Change type |
|---|---|---|
| `internal/picker/style.go` | 状態バッジ / PR バッジ / PR セル背景の lipgloss スタイル生成と NO_COLOR フォールバック | Modify |
| `internal/picker/delegate.go` | `renderRow` から PR 列を `renderPRCell` に切り出し、動線グリフ `→` を挟む | Modify |
| `internal/picker/delegate_test.go` | 各 PR 状態 / prUnavailable / PR nil / NO_COLOR の組合せを検証 | Modify |
| `docs/assets/picker-demo.tape` | サイズ・フォント定数の差し替え | Modify |
| `docs/assets/picker-demo.gif` | `vhs` による再生成（再生成コマンドは手動ステップ） | Regenerate |

---

## Task 1: PR 状態 → 配色マッピングと PR バッジ描画を style.go に追加

**Files:**

- Modify: `internal/picker/style.go`
- Test: `internal/picker/style_test.go` (new)

**背景:** 現在 `style.go` は worktree の status 用の `Badge()` のみを提供する。PR 状態（`OPEN/DRAFT/MERGED/CLOSED`）用に独立した関数を追加し、`NO_COLOR=1` では小文字プレーンのフォールバックを返す。

- [ ] **Step 1: `internal/picker/style_test.go` を新規作成して失敗するテストを書く**

```go
package picker

import (
 "strings"
 "testing"
)

func TestPRBadge_NoColorLowercase(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 cases := map[string]string{
  "OPEN":   "[open]",
  "DRAFT":  "[draft]",
  "MERGED": "[merged]",
  "CLOSED": "[closed]",
 }
 for in, want := range cases {
  got := PRBadge(in)
  if got != want {
   t.Errorf("PRBadge(%q) = %q, want %q", in, got, want)
  }
 }
}

func TestPRBadge_ColoredContainsLabel(t *testing.T) {
 t.Setenv("NO_COLOR", "")
 for _, state := range []string{"OPEN", "DRAFT", "MERGED", "CLOSED"} {
  got := PRBadge(state)
  if !strings.Contains(got, "["+state+"]") {
   t.Errorf("PRBadge(%q) should contain [%s], got %q", state, state, got)
  }
  // colored output should include an ANSI escape
  if !strings.Contains(got, "\x1b[") {
   t.Errorf("PRBadge(%q) expected ANSI escape when NO_COLOR unset, got %q", state, got)
  }
 }
}

func TestPRBadge_UnknownState(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 got := PRBadge("WEIRD")
 if got != "[weird]" {
  t.Errorf("PRBadge(WEIRD) = %q, want [weird]", got)
 }
}
```

- [ ] **Step 2: テストを実行して失敗することを確認**

Run: `go test ./internal/picker/ -run TestPRBadge -v`
Expected: FAIL with "undefined: PRBadge"

- [ ] **Step 3: `PRBadge` / PR セル背景スタイル関数を `style.go` に追加**

`internal/picker/style.go` に以下を追記:

```go
// PRBadge renders a fixed-width PR state badge. The upstream state strings
// come from `gh pr list` (OPEN / DRAFT / MERGED / CLOSED); any other value
// falls back to the lowercased bracketed label.
func PRBadge(state string) string {
 label := "[" + strings.ToLower(state) + "]"
 if noColor() {
  return label
 }
 fg, bg, ok := prBadgeColor(state)
 if !ok {
  return label
 }
 return lipgloss.NewStyle().
  Bold(true).
  Foreground(lipgloss.Color(fg)).
  Background(lipgloss.Color(bg)).
  Render(label)
}

// PRCellStyle returns the lipgloss style used to wrap the whole PR cell
// (state badge + PR number + title) with a dim state-tinted background.
// Returns an empty style when NO_COLOR is set or state is unknown.
func PRCellStyle(state string) lipgloss.Style {
 if noColor() {
  return lipgloss.NewStyle()
 }
 bg, ok := prCellBackground(state)
 if !ok {
  return lipgloss.NewStyle()
 }
 return lipgloss.NewStyle().Background(lipgloss.Color(bg))
}

// prBadgeColor returns (foreground, background, ok) for the strong badge.
func prBadgeColor(state string) (fg, bg string, ok bool) {
 switch state {
 case "OPEN":
  return "0", "2", true
 case "DRAFT":
  return "15", "8", true
 case "MERGED":
  return "15", "5", true
 case "CLOSED":
  return "15", "1", true
 }
 return "", "", false
}

// prCellBackground returns the dim state-tinted background used around the
// whole PR cell.
func prCellBackground(state string) (string, bool) {
 switch state {
 case "OPEN":
  return "22", true
 case "DRAFT":
  return "237", true
 case "MERGED":
  return "53", true
 case "CLOSED":
  return "52", true
 }
 return "", false
}
```

そして `style.go` の先頭の import に `strings` を追加:

```go
import (
 "os"
 "strings"

 "github.com/charmbracelet/lipgloss"
 "github.com/tqer39/ccw-cli/internal/worktree"
)
```

- [ ] **Step 4: テストを実行して通ることを確認**

Run: `go test ./internal/picker/ -run TestPRBadge -v`
Expected: PASS (3 tests)

- [ ] **Step 5: コミット**

```bash
git add internal/picker/style.go internal/picker/style_test.go
git commit -m "picker: add PR state badge and cell background styles"
```

---

## Task 2: `renderRow` から PR セルを切り出し、動線グリフ `→` を挟む

**Files:**

- Modify: `internal/picker/delegate.go`
- Modify: `internal/picker/delegate_test.go`

**背景:** 既存 `renderRow` は PR 部を `fmt.Sprintf("#%d %s \"%s\"", ...)` で組み立てている。新仕様では `[STATE]` バッジと `#N "title"` を分離し、全体を `PRCellStyle` で包み、先頭に `→` 区切りを挟む。

- [ ] **Step 1: delegate_test.go に新レンダリング用の失敗テストを追加**

`internal/picker/delegate_test.go` に以下の新テストを追加（既存テストは次ステップで書き換える）:

```go
func TestRenderRow_ContainsArrowAndPRBadge(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 row := renderRow(listItem{
  tag: tagWorktree,
  wt: &worktree.Info{
   Branch: "feat/login",
   Path:   "/tmp/x",
   Status: worktree.StatusPushed,
  },
  pr: &gh.PRInfo{Number: 42, State: "OPEN", Title: "add login page"},
 }, 120, false, false)
 if !strings.Contains(row, "->") {
  t.Errorf("want arrow separator `->` in NO_COLOR mode, got:\n%s", row)
 }
 if !strings.Contains(row, "[open]") {
  t.Errorf("want PR state badge [open], got:\n%s", row)
 }
 if !strings.Contains(row, "#42") || !strings.Contains(row, "add login page") {
  t.Errorf("want PR number + title, got:\n%s", row)
 }
 // state string should no longer be duplicated outside the badge
 if strings.Count(row, "open") != 1 {
  t.Errorf("state label should appear exactly once, got:\n%s", row)
 }
}

func TestRenderRow_ArrowOmittedWhenPRUnavailable(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 row := renderRow(listItem{
  tag: tagWorktree,
  wt: &worktree.Info{
   Branch: "nebula",
   Path:   "/tmp/n",
   Status: worktree.StatusDirty,
  },
  pr: nil,
 }, 120, true, false)
 if strings.Contains(row, "->") {
  t.Errorf("arrow should be omitted when prUnavailable, got:\n%s", row)
 }
}

func TestRenderRow_ArrowWithNoPRPlaceholder(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 row := renderRow(listItem{
  tag: tagWorktree,
  wt: &worktree.Info{
   Branch: "lonely",
   Path:   "/tmp/l",
   Status: worktree.StatusLocalOnly,
  },
  pr: nil,
 }, 120, false, false)
 if !strings.Contains(row, "->") {
  t.Errorf("arrow should appear even when PR is absent, got:\n%s", row)
 }
 if !strings.Contains(row, "(no PR)") {
  t.Errorf("want (no PR) placeholder, got:\n%s", row)
 }
}
```

また、既存の `TestRenderRow_PushedNoColor` のアサーションを新仕様に合わせて更新する:

```go
func TestRenderRow_PushedNoColor(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 row := renderRow(listItem{
  tag: tagWorktree,
  wt: &worktree.Info{
   Branch:      "kahan",
   Path:        "/tmp/x",
   Status:      worktree.StatusPushed,
   AheadCount:  0,
   BehindCount: 0,
  },
  pr: &gh.PRInfo{Number: 12, State: "MERGED", Title: "Add picker mod"},
 }, 120, false, false)
 if !strings.Contains(row, "[pushed]") {
  t.Errorf("want [pushed] badge, got:\n%s", row)
 }
 if !strings.Contains(row, "kahan") || !strings.Contains(row, "↑0 ↓0") {
  t.Errorf("missing branch/counts:\n%s", row)
 }
 if !strings.Contains(row, "#12") || !strings.Contains(row, "[merged]") {
  t.Errorf("missing PR badge/number:\n%s", row)
 }
}
```

- [ ] **Step 2: テストを実行して失敗することを確認**

Run: `go test ./internal/picker/ -run TestRenderRow -v`
Expected: FAIL — `[open]` / `->` / `(no PR)` が出ない、もしくは既存の `"#12 merged"` が変わってアサーションずれ。

- [ ] **Step 3: `renderRow` を書き換えて PR セル切り出しと動線グリフ挿入を実装**

`internal/picker/delegate.go` を以下で置き換える:

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

// rowDelegate renders items as two lines: meta (badge/branch/indicators/→/PR)
// on top, path below.
type rowDelegate struct {
 prUnavailable bool
}

func (d rowDelegate) Height() int                             { return 2 }
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

// arrowGlyph returns the worktree→PR separator.
// Uses `→` by default, and a plain ASCII `->` in NO_COLOR mode for
// width-predictable output.
func arrowGlyph() string {
 if noColor() {
  return "->"
 }
 return "→"
}

// renderRow is a pure function used by the delegate and tests.
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
 badge := Badge(wt.Status)
 indicators := fmt.Sprintf("↑%d ↓%d", wt.AheadCount, wt.BehindCount)
 if wt.Status == worktree.StatusDirty {
  indicators += fmt.Sprintf(" ✎%d", wt.DirtyCount)
 }

 meta := strings.TrimRight(fmt.Sprintf("%s%s  %-24s %s", prefix, badge, wt.Branch, indicators), " ")
 top := meta
 if !prUnavailable {
  top = meta + "  " + arrowGlyph() + "  " + renderPRCell(li.pr)
 }
 if width > 0 && lipgloss.Width(top) > width {
  top = truncateToWidth(top, width)
 }
 return top + "\n  " + wt.Path
}

// renderPRCell builds the PR portion of the row: either a styled
// `[STATE] #N "title"` cell or a dim `(no PR)` placeholder.
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

// truncateToWidth trims the visible width of s to n cells.
func truncateToWidth(s string, n int) string {
 if lipgloss.Width(s) <= n {
  return s
 }
 // Naive byte-trim fallback: good enough for ASCII-heavy rows.
 for len(s) > 0 && lipgloss.Width(s) > n {
  s = s[:len(s)-1]
 }
 return s
}
```

ポイント:

- `%q` で title を `"..."` で囲む（以前は `\"%s\"` 手書き）
- PR セル全体を `PRCellStyle(state).Render(inner)` で薄背景に包み、その中で `PRBadge` が自前の濃背景で描画される（lipgloss のネスト）
- `arrowGlyph()` で NO_COLOR 時に ASCII フォールバック

- [ ] **Step 4: picker 全テストを走らせてパスすることを確認**

Run: `go test ./internal/picker/...`
Expected: PASS（既存テスト + 新規テスト 3 本 + Task 1 の 3 本）

- [ ] **Step 5: `go vet` を実行**

Run: `go vet ./...`
Expected: no output

- [ ] **Step 6: コミット**

```bash
git add internal/picker/delegate.go internal/picker/delegate_test.go
git commit -m "picker: wrap PR cell with state background and arrow separator"
```

---

## Task 3: tape ファイルの解像度を縮小

**Files:**

- Modify: `docs/assets/picker-demo.tape`

**背景:** 現状 `1280x780 / FontSize 28 / Padding 28`。README 埋め込み時に余白が目立つため `1024x640 / FontSize 24 / Padding 20` に縮小する。`TypingSpeed` / `PlaybackSpeed` / `Sleep` はそのまま。

- [ ] **Step 1: tape のサイズ・フォント定数を差し替える**

`docs/assets/picker-demo.tape` を次の差分で編集:

```text
Output "/tmp/ccw-demo.gif"

Set Shell "bash"
Set FontSize 24
Set Width 1024
Set Height 640
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 110ms
Set PlaybackSpeed 0.92
```

それ以外（`Hide` / `Type` / `Sleep` / ナビゲーション）は変更しない。

- [ ] **Step 2: diff を確認**

Run: `git diff docs/assets/picker-demo.tape`
Expected: `FontSize`, `Width`, `Height`, `Padding` の 4 行だけが変わっている。

- [ ] **Step 3: コミット**

```bash
git add docs/assets/picker-demo.tape
git commit -m "docs: shrink picker demo tape to 1024x640 / fontSize 24"
```

---

## Task 4: GIF 再生成（手動実行）

**Files:**

- Regenerate: `docs/assets/picker-demo.gif`

**背景:** picker 行の見た目と tape サイズの両方が変わったため、デモ GIF を作り直す。再生成コマンドはすでに `picker-demo-setup.sh` にスクリプト化されている。

- [ ] **Step 1: `vhs` と前提ツールが揃っているか確認**

Run: `which vhs && which go`
Expected: 両方のパスが出る（無ければ `brew install vhs` を案内）。

- [ ] **Step 2: デモ環境をセットアップ**

Run: `bash docs/assets/picker-demo-setup.sh`
Expected: 末尾に `ready. now run: vhs docs/assets/picker-demo.tape` が出る。

- [ ] **Step 3: `vhs` で GIF を生成**

Run: `vhs docs/assets/picker-demo.tape`
Expected: `/tmp/ccw-demo.gif` が生成される。コンソールに各フレームの進捗。

- [ ] **Step 4: 新しい GIF を `docs/assets/` にコピー**

Run: `cp /tmp/ccw-demo.gif docs/assets/picker-demo.gif`

- [ ] **Step 5: 目視確認**

Run: `open docs/assets/picker-demo.gif`

確認ポイント:

- picker 行に `→` セパレータと `[OPEN]` 等の PR バッジが出ている
- PR セル全体が薄い状態色で塗られている
- フォント・ウィンドウが以前より小さく、README 埋め込みで違和感がないサイズ感
- PlaybackSpeed 0.92 で速度感は変わっていない

- [ ] **Step 6: コミット**

```bash
git add docs/assets/picker-demo.gif
git commit -m "docs: regenerate picker demo GIF with PR badges and smaller tape"
```

---

## Post-Implementation Checklist

- [ ] `go test ./...` がグリーン
- [ ] `go vet ./...` がクリーン
- [ ] `docs/assets/picker-demo.gif` が新デザインで差し替わっている
- [ ] README には spec/plan の変更起因の追記は不要（機能変更でなく見た目調整のため）

---

## Commit 一覧（予定）

1. `picker: add PR state badge and cell background styles`
2. `picker: wrap PR cell with state background and arrow separator`
3. `docs: shrink picker demo tape to 1024x640 / fontSize 24`
4. `docs: regenerate picker demo GIF with PR badges and smaller tape`
