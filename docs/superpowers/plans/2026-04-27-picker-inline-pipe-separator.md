# Picker ヘッダー行のインライン化 (`|` 区切り) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** picker の worktree 行ヘッダーを「右寄せ status」から「`|` 区切りで左寄せインライン」に変更し、広いターミナルでの視線移動を解消する。

**Architecture:** `internal/picker/style.go` に `Separator()` / `MissingOnDisk()` を追加し、`internal/picker/delegate.go::renderRow` を `padBetween` ベースから `Separator()` 連結に書き換える。prunable 行は `(missing on disk)` 1 セルに集約する。`padBetween` 関数は削除。

**Tech Stack:** Go, charm.land/lipgloss/v2, charm.land/bubbles/v2 (既存)

**Spec:** `docs/superpowers/specs/2026-04-27-picker-inline-pipe-separator-design.md`

---

## File Structure

- **Modify** `internal/picker/style.go` — `Separator()` と `MissingOnDisk()` を追加
- **Modify** `internal/picker/style_test.go` — 上記ヘルパのテストを追加
- **Modify** `internal/picker/delegate.go` — `renderRow` のヘッダー組み立てをインライン化、`padBetween` を削除
- **Modify** `internal/picker/delegate_test.go` — 新フォーマット用のテストを追加、既存テストは Contains 系なので原則そのまま

その他のファイル (`run.go`, `update.go`, `view.go`, `model.go`, `bulk.go`, README, tape) は触らない。

---

### Task 1: `Separator()` ヘルパを追加

`internal/picker/style.go` に dim grey の `" | "` を返すヘルパを TDD で追加する。

**Files:**

- Modify: `internal/picker/style.go`
- Modify: `internal/picker/style_test.go`

- [ ] **Step 1: 失敗するテストを追加 (NO_COLOR 経路)**

`internal/picker/style_test.go` の末尾に追加:

```go
func TestSeparator_NoColor(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 if got := Separator(); got != " | " {
  t.Errorf("Separator() NO_COLOR = %q, want %q", got, " | ")
 }
}

func TestSeparator_ColoredContainsPipe(t *testing.T) {
 t.Setenv("NO_COLOR", "")
 got := Separator()
 if !strings.Contains(got, "|") {
  t.Errorf("Separator() = %q, want substring |", got)
 }
 if !strings.Contains(got, "\x1b[") {
  t.Errorf("Separator() expected ANSI escape when NO_COLOR unset, got %q", got)
 }
}
```

- [ ] **Step 2: テストを実行して失敗を確認**

Run: `go test ./internal/picker/ -run TestSeparator -v`
Expected: FAIL — `Separator` undefined。

- [ ] **Step 3: `Separator()` を実装**

`internal/picker/style.go` の `noColor()` 定義の直後あたりに追加:

```go
// Separator returns the dim-grey vertical bar used to join inline header
// segments (resume, name, status, indicators). Falls back to plain " | "
// when noColor() is true.
func Separator() string {
 if noColor() {
  return " | "
 }
 return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" | ")
}
```

- [ ] **Step 4: テストを実行して PASS を確認**

Run: `go test ./internal/picker/ -run TestSeparator -v`
Expected: PASS (両方)。

- [ ] **Step 5: コミット**

```bash
git add internal/picker/style.go internal/picker/style_test.go
git commit -m "$(cat <<'EOF'
feat(picker): Separator() を追加 (dim grey " | ")

ヘッダー行のインライン化で使用する dim grey の `" | "` を返すヘルパを style.go に追加。
NO_COLOR 時は素の " | " を返す。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: `MissingOnDisk()` ヘルパを追加

prunable 行で status バッジの代わりに表示する `(missing on disk)` を、dim grey で返すヘルパを TDD で追加する。

**Files:**

- Modify: `internal/picker/style.go`
- Modify: `internal/picker/style_test.go`

- [ ] **Step 1: 失敗するテストを追加**

`internal/picker/style_test.go` の末尾に追加:

```go
func TestMissingOnDisk_NoColor(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 if got := MissingOnDisk(); got != "(missing on disk)" {
  t.Errorf("MissingOnDisk() NO_COLOR = %q, want %q", got, "(missing on disk)")
 }
}

func TestMissingOnDisk_ColoredContainsLabel(t *testing.T) {
 t.Setenv("NO_COLOR", "")
 got := MissingOnDisk()
 if !strings.Contains(got, "(missing on disk)") {
  t.Errorf("MissingOnDisk() = %q, want substring (missing on disk)", got)
 }
 if !strings.Contains(got, "\x1b[") {
  t.Errorf("MissingOnDisk() expected ANSI escape when NO_COLOR unset, got %q", got)
 }
}
```

- [ ] **Step 2: テストを実行して失敗を確認**

Run: `go test ./internal/picker/ -run TestMissingOnDisk -v`
Expected: FAIL — `MissingOnDisk` undefined。

- [ ] **Step 3: `MissingOnDisk()` を実装**

`internal/picker/style.go` の `Separator()` の直後に追加:

```go
// MissingOnDisk returns the dim-grey "(missing on disk)" label used in place
// of status / indicators for prunable worktrees whose physical directory is
// gone. Falls back to plain text when noColor() is true.
func MissingOnDisk() string {
 const label = "(missing on disk)"
 if noColor() {
  return label
 }
 return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(label)
}
```

- [ ] **Step 4: テストを実行して PASS を確認**

Run: `go test ./internal/picker/ -run TestMissingOnDisk -v`
Expected: PASS (両方)。

- [ ] **Step 5: コミット**

```bash
git add internal/picker/style.go internal/picker/style_test.go
git commit -m "$(cat <<'EOF'
feat(picker): MissingOnDisk() を追加

prunable 行で status / indicators の代わりに表示する dim grey の
"(missing on disk)" を返すヘルパを style.go に追加。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: `renderRow` の非 prunable ヘッダーをインライン化

`internal/picker/delegate.go::renderRow` のヘッダー組み立てを `padBetween` から `Separator()` ベースに置き換える。prunable は次のタスクで処理。

**Files:**

- Modify: `internal/picker/delegate.go:35-94`
- Modify: `internal/picker/delegate_test.go`

- [ ] **Step 1: 失敗するテスト 1 (pipe separator) を追加**

`internal/picker/delegate_test.go` の末尾に追加:

```go
func TestRenderRow_HeaderUsesPipeSeparator(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 li := listItem{
  tag: tagWorktree,
  wt: &worktree.Info{
   Path:       "/repo/.claude/worktrees/foo",
   Branch:     "feat/login",
   Status:     worktree.StatusPushed,
   HasSession: true,
  },
 }
 got := renderRow(li, 200, true, false)
 header := strings.SplitN(got, "\n", 2)[0]
 // Expected segments joined by " | ":
 //   prefix "  " + [RESUME] | 🌲 foo | [pushed] | (indicators)
 for _, want := range []string{
  "[RESUME] | 🌲 foo",
  "🌲 foo | [pushed]",
 } {
  if !strings.Contains(header, want) {
   t.Errorf("header missing %q:\n%s", want, header)
  }
 }
 // `·` should no longer appear (replaced by `|`)
 if strings.Contains(header, "·") {
  t.Errorf("header should not contain `·` (replaced by `|`):\n%s", header)
 }
}
```

- [ ] **Step 2: 失敗するテスト 2 (wide width で右側余白なし) を追加**

同じファイルの末尾に追加:

```go
func TestRenderRow_HeaderHasNoLargeRightPadding(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 li := listItem{
  tag: tagWorktree,
  wt: &worktree.Info{
   Path:       "/repo/.claude/worktrees/foo",
   Branch:     "feat/login",
   Status:     worktree.StatusPushed,
   HasSession: true,
  },
 }
 got := renderRow(li, 200, true, false)
 header := strings.SplitN(got, "\n", 2)[0]
 // With inline layout, header visible width should be much smaller than 200.
 // Generous upper bound (80) catches the regression where padBetween fills
 // the gap to width.
 if w := lipgloss.Width(header); w > 80 {
  t.Errorf("header visible width %d > 80 at terminal width 200; left/right alignment leaked back in:\n%s", w, header)
 }
}
```

- [ ] **Step 3: テストを実行して失敗を確認**

Run: `go test ./internal/picker/ -run "TestRenderRow_HeaderUsesPipeSeparator|TestRenderRow_HeaderHasNoLargeRightPadding" -v`
Expected: FAIL — 旧実装は `·` を使い、wide width で空白詰めするため両方失敗。

- [ ] **Step 4: `renderRow` の非 prunable パスを書き換える**

`internal/picker/delegate.go` の `renderRow` を以下に置き換える:

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
 sep := Separator()

 // Reserve a 4-cell right-edge margin so IDE-embedded terminals (Cursor,
 // cmux, ...) that report Width slightly larger than the visible area
 // don't clip the right edge. Falls back to the raw width when it is too
 // small to shrink meaningfully.
 effectiveWidth := width
 if width > 4 {
  effectiveWidth = width - 4
 }

 var header string
 if wt.Status == worktree.StatusPrunable {
  header = prefix + resume + sep + "🌲 " + name + sep + MissingOnDisk()
 } else {
  status := Badge(wt.Status)
  indicators := wt.Indicators()
  header = prefix + resume + sep + "🌲 " + name + sep + status + sep + indicators
 }

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

そして `padBetween` 関数 (現在の `delegate.go:85-94`) を **削除** する。`strings` import は `truncateToWidth` で使われていないか確認 — 今後不要なら import 自体は他用途で使われているか同ファイルを再読して判断 (`strings.Repeat` の唯一の使用箇所が `padBetween` 内なので、削除後は import も削除)。

- [ ] **Step 5: テストを実行して PASS を確認**

Run: `go test ./internal/picker/ -v`
Expected: 全テスト PASS。新規 2 テストも通る。`TestRenderRow_RightMargin` (visible width <= width-4) も継続して通る。

落ちる既存テストがあれば、そのテストが対象としているレイアウト前提を確認した上で、Contains 系の assert は緩いはずなので原則そのまま通るはず。万一落ちる場合は assert 文字列を新フォーマット (`|` 区切り) に合わせて最小修正。

- [ ] **Step 6: ビルドチェック**

Run: `go build ./...`
Expected: 成功 (削除した `padBetween` への参照が残っていないこと)。

- [ ] **Step 7: コミット**

```bash
git add internal/picker/delegate.go internal/picker/delegate_test.go
git commit -m "$(cat <<'EOF'
feat(picker): ヘッダー行を `|` 区切りでインライン化

`padBetween` による右寄せをやめ、`Separator()` (`" | "` dim grey) で
連結する左寄せインラインレイアウトに変更。広いターミナルで worktree 名と
status バッジの距離が広がる問題を解消。`padBetween` 関数は削除。

prunable 行は `(missing on disk)` 1 セルに集約する分岐を含む。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: prunable 行のテストを追加

prunable パスは Task 3 のコード変更で実装済みだが、振る舞いを固定する専用テストを追加する。

**Files:**

- Modify: `internal/picker/delegate_test.go`

- [ ] **Step 1: prunable 用テストを追加**

`internal/picker/delegate_test.go` の末尾に追加:

```go
func TestRenderRow_PrunableShowsMissingOnDisk(t *testing.T) {
 t.Setenv("NO_COLOR", "1")
 li := listItem{
  tag: tagWorktree,
  wt: &worktree.Info{
   Path:        "/repo/.claude/worktrees/stale",
   Branch:      "stale/feature",
   Status:      worktree.StatusPrunable,
   AheadCount:  3,
   BehindCount: 2,
   DirtyCount:  7,
   HasSession:  true,
  },
 }
 got := renderRow(li, 120, true, false)
 header := strings.SplitN(got, "\n", 2)[0]
 if !strings.Contains(header, "(missing on disk)") {
  t.Errorf("prunable header should contain (missing on disk):\n%s", header)
 }
 // status badge / indicators should NOT appear
 for _, unwanted := range []string{"[prune]", "[pushed]", "[dirty]", "[local]", "↑3", "↓2", "✎7"} {
  if strings.Contains(header, unwanted) {
   t.Errorf("prunable header should not contain %q:\n%s", unwanted, header)
  }
 }
}
```

- [ ] **Step 2: テストを実行して PASS を確認**

Run: `go test ./internal/picker/ -run TestRenderRow_PrunableShowsMissingOnDisk -v`
Expected: PASS (Task 3 の実装で既に成立しているはず)。

- [ ] **Step 3: コミット**

```bash
git add internal/picker/delegate_test.go
git commit -m "$(cat <<'EOF'
test(picker): prunable 行が (missing on disk) を出すことを固定

prunable では status バッジ / indicators の代わりに (missing on disk) を
1 セルに集約することを assert する単体テストを追加。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: 全体テスト & lint 確認

`internal/picker` 以外のテストにも影響がないことを確認し、lint / cspell も通す。

**Files:** なし (検証のみ)

- [ ] **Step 1: 全テストを実行**

Run: `go test ./...`
Expected: 全 PASS。失敗があれば該当テストを確認し、レイアウト依存の assert なら新フォーマットに合わせて修正する。

- [ ] **Step 2: ビルド**

Run: `go build ./...`
Expected: 成功。

- [ ] **Step 3: vet**

Run: `go vet ./...`
Expected: 問題なし。

- [ ] **Step 4: pre-commit hook (lefthook) 相当の確認**

Run: `lefthook run pre-commit` (利用可能なら) もしくは個別に:

```bash
# markdownlint (該当ファイルなし — picker は go のみ)
# cspell — plan/spec の追加日本語語彙のみ
git diff --name-only HEAD~5..HEAD | xargs -I {} sh -c 'echo "--- {}"' || true
```

Expected: 既存 hook で fail しないこと。

- [ ] **Step 5: 手元目視確認 (オプショナル)**

可能なら `go run ./cmd/ccw` で picker を起動し、80 桁 / 120 桁 / 200 桁のターミナルでヘッダーが `|` 区切りで連続表示されることを確認。実機確認できない場合はスキップ可。

---

### Task 6: PR 作成

レビュー対象として PR を出す。

**Files:** なし

- [ ] **Step 1: ブランチを push**

```bash
git push -u origin spec/picker-inline-pipe-separator
```

- [ ] **Step 2: PR 作成**

```bash
gh pr create --title "feat(picker): ヘッダー行を `|` 区切りでインライン化" --body "$(cat <<'EOF'
## Summary
- 広いターミナルで右寄せ status バッジが worktree 名から離れすぎる問題を解消
- `padBetween` ベースの右寄せをやめ、`Separator()` (`" | "` dim grey) で連結する左寄せインラインに変更
- prunable 行は `(missing on disk)` 1 セルに集約

## Spec / Plan
- Spec: `docs/superpowers/specs/2026-04-27-picker-inline-pipe-separator-design.md`
- Plan: `docs/superpowers/plans/2026-04-27-picker-inline-pipe-separator.md`

## Test plan
- [ ] `go test ./...`
- [ ] 80 / 120 / 200 桁ターミナルで picker を起動し、`|` 区切りで連続表示されること
- [ ] prunable な worktree で `(missing on disk)` が status の代わりに表示されること
- [ ] NO_COLOR=1 で起動して `|` が dim 化されないこと
- [ ] vhs tape / GIF の更新は別 PR で対応

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Expected: PR URL が表示される。

---

## Self-Review (writing-plans skill)

**Spec coverage:** spec の各セクションを確認:

- 「`|` 区切りの仕様 (色 / 前後 space / NO_COLOR)」 → Task 1
- 「`(missing on disk)` の dim 化」 → Task 2
- 「`renderRow` 非 prunable インライン化」 → Task 3
- 「`renderRow` prunable 専用パス」 → Task 3 (実装) + Task 4 (テスト)
- 「`padBetween` 削除」 → Task 3 Step 4
- 「4-cell 右マージン保持」 → Task 3 Step 4 (`effectiveWidth` を `truncateToWidth` で使用)
- 「テスト一覧 (5 種)」 → Task 1 (Separator x2), Task 2 (MissingOnDisk x2), Task 3 (HeaderUsesPipeSeparator + HeaderHasNoLargeRightPadding), Task 4 (PrunableShowsMissingOnDisk) で全カバー
- 「触らないもの (`update.go`, `run.go`, branch/pr 行, タグ系メニュー行)」 → Task 3 のコード差分が `renderRow` 内のヘッダー部分に限定されているため遵守
- 「vhs tape / GIF」 → Task 6 PR 本文で「別 PR」と明記

ギャップなし。

**Placeholder scan:** TBD / TODO / "適切に" / "似たように" などのプレースホルダ無し。全 step で実コード / 実コマンドを記載済み。

**Type consistency:** `Separator()` `MissingOnDisk()` の関数名はタスク間で一貫。`renderRow` のシグネチャ変更なし。`padBetween` は完全削除。
