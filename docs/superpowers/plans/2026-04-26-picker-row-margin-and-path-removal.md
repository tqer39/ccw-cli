# Picker 行の右端マージン確保と path 行削除 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** picker 一覧の右端見切れを解消し、`path:` 行を削除して情報密度を上げ、worktree 名の前に 🌲 を付け、README の Demo 直下に「RESUME 名は表示しない」注意書きを追加する。

**Architecture:** `internal/picker/delegate.go` の `renderRow` と `rowDelegate.Height` を修正する。レイアウトは行レベルで縮め、bubbles list 側 (`update.go`) には触らない。`internal/picker/delegate_test.go` の期待値も同時に更新する。`README.md` / `docs/README.ja.md` の Demo 直下に RESUME バッジの意味を補足する短い注意書きを追加する。

**Tech Stack:** Go (`charm.land/bubbletea/v2` / `charm.land/bubbles/v2/list` / `charm.land/lipgloss/v2`)、`testing`、`lefthook` + `markdownlint-cli2` + `cspell`。

**Spec:** `docs/superpowers/specs/2026-04-26-picker-row-margin-and-path-removal-design.md`

---

## File Structure

- **Modify:** `internal/picker/delegate.go`
  - `rowDelegate.Height`: `4 → 3`
  - `renderRow`: `pathLine` 削除、header に `🌲` prefix、右マージン 4 文字確保
- **Modify:** `internal/picker/delegate_test.go`
  - 既存テストの path assert と行数 assert (4 → 3) を更新
  - header の `🌲` を assert
  - 右マージンの境界テストを 1 つ追加
- **Modify:** `README.md`
  - `## 🎬 Demo` 直下に RESUME 名非表示の注意書き
- **Modify:** `docs/README.ja.md`
  - `## 🎬 デモ` 直下に同等の注意書き
- **Do not touch:** `internal/picker/update.go` (`list.SetSize` の引数を縮めると bubbles list 側で副作用が出る懸念)
- **Do not touch:** `internal/picker/run.go` (フォールバック出力)
- **Do not touch:** `internal/picker/view_test.go` (path / Height への参照無しを確認済み)

---

## Task 1: `path:` 行削除 + `rowDelegate.Height` を 3 に + header に 🌲 prefix

**Files:**

- Modify: `internal/picker/delegate_test.go`
- Modify: `internal/picker/delegate.go:22` (`Height` の戻り値)
- Modify: `internal/picker/delegate.go:35-73` (`renderRow`)

### Step 1.1: テストを Red 化する

- [ ] `internal/picker/delegate_test.go` の `TestRenderRow_ResumeBadge` を以下のように書き換える。

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
 got := renderRow(li, 120, true, false)
 if !strings.Contains(got, "[RESUME]") {
  t.Errorf("missing RESUME badge:\n%s", got)
 }
 if !strings.Contains(got, "🌲 foo") {
  t.Errorf("missing tree icon + worktree name '🌲 foo':\n%s", got)
 }
 if !strings.Contains(got, "branch:  feature/auth") {
  t.Errorf("missing branch line:\n%s", got)
 }
 if strings.Contains(got, "path:") {
  t.Errorf("path: line should be removed:\n%s", got)
 }
 lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
 if len(lines) != 3 {
  t.Errorf("got %d lines, want 3:\n%s", len(lines), got)
 }
}
```

- [ ] `Height()` の期待値テストを新規追加。`TestRenderRow_ResumeBadge` の直下に追加する。

```go
func TestRowDelegateHeight(t *testing.T) {
 if got := (rowDelegate{}).Height(); got != 3 {
  t.Errorf("rowDelegate.Height() = %d, want 3", got)
 }
}
```

### Step 1.2: テストを走らせて Red を確認

- [ ] 以下のコマンドを実行する。

```bash
go test ./internal/picker/ -run 'TestRenderRow_ResumeBadge|TestRowDelegateHeight' -v
```

期待: `TestRenderRow_ResumeBadge` は `path: line should be removed` または `missing tree icon` で fail。`TestRowDelegateHeight` は `Height() = 4, want 3` で fail。

### Step 1.3: `Height` を 3 に変更し `renderRow` から path 行を削除し header に 🌲 を入れる (Green)

- [ ] `internal/picker/delegate.go` の `Height()` 行を変更する。

```go
func (d rowDelegate) Height() int                             { return 3 }
```

- [ ] 同ファイルの `renderRow` 関数を以下に置き換える (header の `· %s` → `· 🌲 %s`、`pathLine` 関連の組み立て・truncate・連結を削除)。

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
 name := filepath.Base(wt.Path)
 resume := ResumeBadge(wt.HasSession)
 status := Badge(wt.Status)
 indicators := wt.Indicators()
 if wt.Status == worktree.StatusPrunable {
  indicators = "(missing on disk)"
 }

 header := fmt.Sprintf("%s%s · 🌲 %s", prefix, resume, name)
 right := fmt.Sprintf("%s  %s", status, indicators)
 header = padBetween(header, right, width)

 branchLine := fmt.Sprintf("    branch:  %s", wt.Branch)
 prCell := ""
 if !prUnavailable {
  prCell = renderPRCell(li.pr)
 }
 prLine := "    pr:      " + prCell

 if width > 0 {
  header = truncateToWidth(header, width)
  branchLine = truncateToWidth(branchLine, width)
  prLine = truncateToWidth(prLine, width)
 }

 return header + "\n" + branchLine + "\n" + prLine
}
```

- [ ] `internal/picker/delegate.go` 上部のコメント (`// rowDelegate renders worktree items as four lines: header (resume + name + ...`) を 3 行構成に合わせて更新する。

```go
// rowDelegate renders worktree items as three lines: header (resume + tree
// icon + worktree name + status badge + indicators), branch, pr.
type rowDelegate struct {
 prUnavailable bool
}
```

### Step 1.4: 関連テストを走らせて Green を確認

- [ ] 以下を実行。

```bash
go test ./internal/picker/ -v
```

期待: 既存の `TestRenderRow_NewBadge` / `TestRenderRow_StatusBadgeAndIndicators` / `TestRenderRow_PRLineWithPR` / `TestRenderRow_PRLineNoPR` / `TestRenderRow_PRUnavailableHidesPRContent` も含めて全 PASS。

### Step 1.5: コミット

- [ ] 変更を 1 コミットにまとめる。

```bash
git add internal/picker/delegate.go internal/picker/delegate_test.go
git commit -m "$(cat <<'EOF'
feat(picker): drop path row, height 4→3, prefix worktree name with 🌲

- 一覧から path: 行を削除し Height を 4→3 に
- header の worktree 名前に 🌲 を付けて識別性を上げる
- フルパスは menu / delete confirm 画面で従来通り表示される
EOF
)"
```

---

## Task 2: 右端マージン 4 文字を確保

**Files:**

- Modify: `internal/picker/delegate_test.go` (境界テストを追加)
- Modify: `internal/picker/delegate.go:35-73` (`renderRow` 内で `effectiveWidth = width - 4` を導入)

### Step 2.1: 右マージンの境界テストを追加 (Red)

- [ ] `internal/picker/delegate_test.go` の末尾に以下のテストを追加する。

```go
func TestRenderRow_RightMargin(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 li := listItem{
  tag: tagWorktree,
  wt: &worktree.Info{
   Path:        "/repo/.claude/worktrees/foo",
   Branch:      "feature/right-margin",
   Status:      worktree.StatusLocalOnly,
   AheadCount:  0,
   BehindCount: 0,
  },
 }
 const width = 80
 const margin = 4
 got := renderRow(li, width, true, false)
 lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
 for i, line := range lines {
  if w := lipgloss.Width(line); w > width-margin {
   t.Errorf("line %d has visible width %d > %d (width %d - margin %d):\n%s",
    i, w, width-margin, width, margin, line)
  }
 }
}
```

- [ ] テストファイル冒頭の import に `"charm.land/lipgloss/v2"` を追加する。既に他 import が並んでいるので alphabetical 順で挿入する。

```go
import (
 "strings"
 "testing"

 "charm.land/lipgloss/v2"
 "github.com/tqer39/ccw-cli/internal/gh"
 "github.com/tqer39/ccw-cli/internal/worktree"
)
```

### Step 2.2: テストを走らせて Red を確認

- [ ] 以下を実行。

```bash
go test ./internal/picker/ -run TestRenderRow_RightMargin -v
```

期待: `header` の visible width が `width=80` (= 76 でなく 80) で算出されるため `> 76` で fail する。

### Step 2.3: `renderRow` に effective width を導入 (Green)

- [ ] `internal/picker/delegate.go` の `renderRow` を以下に置き換える (Task 1 から effective width 部分のみ差分追加)。

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
 name := filepath.Base(wt.Path)
 resume := ResumeBadge(wt.HasSession)
 status := Badge(wt.Status)
 indicators := wt.Indicators()
 if wt.Status == worktree.StatusPrunable {
  indicators = "(missing on disk)"
 }

 // Reserve a 4-cell right-edge margin so IDE-embedded terminals (Cursor,
 // cmux, ...) that report Width slightly larger than the visible area
 // don't clip the right-aligned status/indicators. Falls back to the raw
 // width when it is too small to shrink meaningfully.
 effectiveWidth := width
 if width > 4 {
  effectiveWidth = width - 4
 }

 header := fmt.Sprintf("%s%s · 🌲 %s", prefix, resume, name)
 right := fmt.Sprintf("%s  %s", status, indicators)
 header = padBetween(header, right, effectiveWidth)

 branchLine := fmt.Sprintf("    branch:  %s", wt.Branch)
 prCell := ""
 if !prUnavailable {
  prCell = renderPRCell(li.pr)
 }
 prLine := "    pr:      " + prCell

 if width > 0 {
  header = truncateToWidth(header, effectiveWidth)
  branchLine = truncateToWidth(branchLine, effectiveWidth)
  prLine = truncateToWidth(prLine, effectiveWidth)
 }

 return header + "\n" + branchLine + "\n" + prLine
}
```

### Step 2.4: テストを走らせて Green を確認

- [ ] 以下を実行。

```bash
go test ./internal/picker/ -v
```

期待: `TestRenderRow_RightMargin` を含め全 PASS。

### Step 2.5: コミット

- [ ] 変更を 1 コミットにまとめる。

```bash
git add internal/picker/delegate.go internal/picker/delegate_test.go
git commit -m "$(cat <<'EOF'
feat(picker): reserve 4-cell right margin to avoid right-edge clipping

Cursor / cmux などの IDE 内蔵ターミナルは Width 報告が可視幅より
大きい場合があり、`↑0 ↓0` などの右寄せ要素が見切れていた。
renderRow 内で width > 4 のとき effectiveWidth = width - 4 として
padBetween / truncateToWidth に渡すことで、80×24 でもマージンを
確保しつつ IDE 内蔵ターミナルでも見切れない。
EOF
)"
```

---

## Task 3: README の Demo 直下に RESUME 名非表示の注意書きを追加 (EN + JA)

**Files:**

- Modify: `README.md:39-41` (`## 🎬 Demo` セクション)
- Modify: `docs/README.ja.md:39-41` (`## 🎬 デモ` セクション)

### Step 3.1: `README.md` の Demo 直下に注意書きを追加

- [ ] `README.md` の以下の箇所を変更する。

変更前:

```markdown
## 🎬 Demo

![picker demo](docs/assets/picker-demo.gif)

## 📖 Usage
```

変更後:

```markdown
## 🎬 Demo

![picker demo](docs/assets/picker-demo.gif)

> **Note:** the `💬 RESUME` badge only signals that a session log exists for the worktree. The session title or first prompt is **not** previewed in the picker — `ccw` simply runs `claude --continue` and lets the Claude Code CLI pick the most recent session.

## 📖 Usage
```

### Step 3.2: `docs/README.ja.md` の デモ 直下に同等の注意書きを追加

- [ ] `docs/README.ja.md` の以下の箇所を変更する。

変更前:

```markdown
## 🎬 デモ

![picker demo](assets/picker-demo.gif)

## 📖 使い方
```

変更後:

```markdown
## 🎬 デモ

![picker demo](assets/picker-demo.gif)

> **メモ:** `💬 RESUME` バッジは「その worktree に紐づくセッションログがある」ことだけを示します。セッションのタイトルや最初のプロンプトは picker には表示されません。`ccw` は `claude --continue` を実行するだけで、最新セッションの選択は Claude Code CLI 任せです。

## 📖 使い方
```

### Step 3.3: lefthook hook 相当のチェックを手元で先に通す

- [ ] markdownlint と cspell をローカルで走らせる。

```bash
npx --yes markdownlint-cli2 README.md docs/README.ja.md
npx --yes cspell --no-progress README.md docs/README.ja.md
```

期待: 両者ともエラーなしで終了。`cspell` で未知単語があれば `.cspell/project-words.txt` に追加して再実行する (今回追加したのは英文中の "ccw" / "Claude" など既存単語のみのはずなので、新規追加は通常不要)。

### Step 3.4: コミット

- [ ] 変更を 1 コミットにまとめる。

```bash
git add README.md docs/README.ja.md
git commit -m "$(cat <<'EOF'
docs(readme): note that RESUME name is not previewed in the picker

`💬 RESUME` バッジは session ログの有無だけを示し、session タイトルや
最初のプロンプトは picker に表示されない。`ccw` は claude --continue を
実行するのみで、最新 session の選択は Claude Code CLI 任せである旨を
EN/JA Demo セクション直下に追記する。
EOF
)"
```

---

## Task 4: 全テスト + 手動目視で受け入れ基準を確認

**Files:** (検証のみ、変更なし)

### Step 4.1: 全テストを走らせる

- [ ] 以下を実行。

```bash
go test ./...
```

期待: PASS。

### Step 4.2: 80×24 の standalone ターミナルで目視確認

- [ ] バイナリをビルドして実行する。

```bash
go build -o ./bin/ccw ./cmd/ccw
```

- [ ] Terminal.app または好きな standalone ターミナルを 80×24 に揃え、リポジトリ root で `./bin/ccw` を起動。

期待:

- 各 worktree 行が `header / branch / pr` の 3 行構成
- header の左から順に `[⚡ NEW] · 🌲 <worktree-name>` (selected 行は `>` prefix)
- 右側の `[LOCAL] ↑0 ↓0` 等の右に少なくともスペース 4 文字分の余白
- `path:` 行が消えている
- worktree を選択して `r`/`d` を押した後の menu / delete confirm 画面では従来通りフルパスが表示される

`q` で picker を抜ける。

### Step 4.3: Cursor / cmux 内蔵ターミナルで目視確認

- [ ] Cursor / cmux などの IDE 内蔵ターミナルで同じ `./bin/ccw` を起動。

期待: 右端の `↑0 ↓0` が見切れない。

### Step 4.4: 受け入れ基準のチェックリスト

spec の `## 受け入れ基準` を 1 項目ずつ確認する。

- [ ] 一覧画面で各行が `header / branch / pr` の 3 行構成になっている
- [ ] `path:` 行が一覧画面に表示されない
- [ ] header の worktree 名の前に `🌲` が付いている
- [ ] 80×24 の Terminal.app で `↑0 ↓0` の右にスペース 4 文字以上
- [ ] Cursor / cmux 内蔵ターミナルで右端が見切れない
- [ ] menu / delete confirm 画面では従来通りフルパスが見える
- [ ] README.md と docs/README.ja.md の Demo 直下に RESUME 名非表示の注意書き
- [ ] `go test ./...` が通る
- [ ] markdownlint / cspell が通る (lefthook 経由 or 手動)

### Step 4.5: 不要なローカルバイナリを掃除

- [ ] `./bin/ccw` は git ignore されている前提だが念のため `git status` で untracked ファイルが意図したものだけになっていることを確認する。

```bash
git status
```

期待: `working tree clean` または既知の untracked のみ。

---

## 後続 (本プラン外)

- **RESUME 名表示**: 別 spec として `docs/superpowers/specs/2026-04-26-picker-resume-summary-design.md` (仮) で扱う。`~/.claude/projects/<encoded>/<最新 mtime>.jsonl` を読んで最初のユーザープロンプトを要約する案。jsonl format が public contract でないこと、複数 jsonl ある場合の選択ルール、要約の長さ・多言語切り詰めが論点。
