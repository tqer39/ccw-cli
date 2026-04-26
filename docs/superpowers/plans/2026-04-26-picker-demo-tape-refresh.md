# picker-demo.tape Refresh Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refresh `docs/assets/picker-demo.tape` (and regenerate the GIF) so it reflects the post-PR-#71 picker TUI (3-line rows, 🌲 prefix, right-margin reservation).

**Architecture:** Minimum-diff refresh. Comments inside the tape are updated to the new TUI vocabulary, hover Sleeps are unified to 3000ms, and `picker-demo-setup.sh` gets a small assertion that `chore/cleanup` is recognized as DIRTY (and not prunable). The GIF is regenerated with `vhs`. Picker source code is not touched.

**Tech Stack:** [vhs](https://github.com/charmbracelet/vhs) tape DSL, bash, git.

---

## File Structure

| File | Responsibility | Action |
|---|---|---|
| `docs/assets/picker-demo.tape` | vhs script that drives the picker walkthrough | Modify (comments + Sleep values) |
| `docs/assets/picker-demo-setup.sh` | Build sandbox repo + 4 worktrees + mock `gh` + fake `HOME` | Modify (add 1 sanity assertion) |
| `docs/assets/picker-demo.gif` | Demo GIF embedded in `README.md` and `docs/README.ja.md` | Regenerate |
| `docs/superpowers/specs/2026-04-26-picker-demo-tape-refresh-design.md` | Source spec | Read-only reference |

No new files. No source-code changes.

---

## Task 1: Update tape comments and Sleep values

**Files:**

- Modify: `docs/assets/picker-demo.tape`

- [ ] **Step 1: Validate the current tape parses**

Run:

```bash
vhs validate docs/assets/picker-demo.tape
```

Expected: exit 0, no output (or `valid` summary). Establishes a clean baseline before edits.

- [ ] **Step 2: Replace the file contents with the refreshed tape**

Open `docs/assets/picker-demo.tape` and replace the entire file with:

```text
Output "/tmp/ccw-demo.gif"

Env CCW_LANG en

Set Shell "bash"
Set FontSize 24
Set Width 1440
Set Height 640
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 110ms
Set PlaybackSpeed 1.0

Hide
Type "cd /tmp/ccw-demo && export HOME=/tmp/ccw-demo-home PATH=/tmp/ccw-demo-bin:/tmp/fake-gh:$PATH && clear"
Enter
Sleep 1000ms
Show

Type "ccw"
Enter
# initial picker view — rows render as 3 lines (🌲 name + status / branch / pr)
# top row hovered: 🌲 feat-login · [PUSHED] + RESUME + [OPEN] #42
Sleep 7000ms

# walk through every worktree row to show the 4 status × 4 PR-state combos:
# 🌲 feat-login [PUSHED] [OPEN]   →  🌲 feat-dashboard [PUSHED] [MERGED]
Down
Sleep 3000ms
# 🌲 feat-dashboard [PUSHED] [MERGED]  →  🌲 feat-picker [LOCAL] [DRAFT]
Down
Sleep 3000ms
# 🌲 feat-picker [LOCAL] [DRAFT]  →  🌲 chore-cleanup [DIRTY] [CLOSED]
Down
Sleep 3000ms

# return to feat-dashboard ([PUSHED] + [MERGED] + RESUME) for the submenu demo
Up
Sleep 2500ms
Up
Sleep 2500ms

# submenu shows the full worktree path (replaces the path row that used to be in the list)
# [r] run / [d] delete / [b] back
Enter
Sleep 5000ms
Type "b"
Sleep 2500ms

# jump to [clean pushed] — past chore-cleanup and [delete all]
Down 4
Sleep 2500ms
Enter
# bulk-confirm screen — preview of pushed targets
Sleep 5500ms

# cancel
Type "N"
Sleep 2500ms

# quit picker
Type "q"
Sleep 1500ms
```

- [ ] **Step 3: Validate the refreshed tape parses**

Run:

```bash
vhs validate docs/assets/picker-demo.tape
```

Expected: exit 0, no errors. Catches typos in `Down` / `Type` / `Sleep` etc. before Task 3.

- [ ] **Step 4: Commit**

```bash
git add docs/assets/picker-demo.tape
git commit -m "docs(tape): refresh picker-demo.tape comments for new TUI"
```

---

## Task 2: Add a DIRTY/prunable sanity assertion to setup.sh

**Files:**

- Modify: `docs/assets/picker-demo-setup.sh` (insert block after the chore/cleanup worktree creation)

- [ ] **Step 1: Insert the assertion**

After the `chore/cleanup` worktree creation block (the line that creates `stray.txt` / `extra.md`) and before the `# 4. mock gh` block, insert this snippet:

```bash
# 3.1 sanity-check: chore/cleanup must be DIRTY and no worktree may be prunable.
# Both invariants are required for the demo GIF to render the expected badges
# ([DIRTY] for chore-cleanup, no `(missing on disk)` lines anywhere).
chore_status=$(cd /tmp/ccw-demo/.claude/worktrees/chore-cleanup && git status --porcelain)
if [ -z "$chore_status" ]; then
  echo "picker-demo-setup: chore/cleanup is unexpectedly clean — abort." >&2
  exit 1
fi
if git -C /tmp/ccw-demo worktree list --porcelain | grep -q '^prunable'; then
  echo "picker-demo-setup: a worktree is prunable — abort." >&2
  exit 1
fi
```

The exact insertion point:

```bash
#    chore/cleanup: DIRTY (untracked files) + CLOSED PR #45
git -C /tmp/ccw-demo worktree add -q -b chore/cleanup .claude/worktrees/chore-cleanup
(cd /tmp/ccw-demo/.claude/worktrees/chore-cleanup && echo untracked >stray.txt && echo more >extra.md)

# 3.1 sanity-check: chore/cleanup must be DIRTY and no worktree may be prunable.
# Both invariants are required for the demo GIF to render the expected badges
# ([DIRTY] for chore-cleanup, no `(missing on disk)` lines anywhere).
chore_status=$(cd /tmp/ccw-demo/.claude/worktrees/chore-cleanup && git status --porcelain)
if [ -z "$chore_status" ]; then
  echo "picker-demo-setup: chore/cleanup is unexpectedly clean — abort." >&2
  exit 1
fi
if git -C /tmp/ccw-demo worktree list --porcelain | grep -q '^prunable'; then
  echo "picker-demo-setup: a worktree is prunable — abort." >&2
  exit 1
fi

# 4. mock gh that returns canned PR rows covering every PR state
```

- [ ] **Step 2: Run shellcheck on the script**

Run:

```bash
shellcheck docs/assets/picker-demo-setup.sh
```

Expected: exit 0, no errors. (If `shellcheck` is unavailable locally, lefthook will run it on commit.)

- [ ] **Step 3: Run the setup script end-to-end to confirm the assertions pass**

Run:

```bash
bash docs/assets/picker-demo-setup.sh
```

Expected:

- exit 0
- final stdout line: `ready. now run: vhs docs/assets/picker-demo.tape`
- no `picker-demo-setup: ... — abort.` message on stderr

- [ ] **Step 4: Commit**

```bash
git add docs/assets/picker-demo-setup.sh
git commit -m "docs(tape): assert chore/cleanup is DIRTY and no worktree is prunable"
```

---

## Task 3: Regenerate the demo GIF

**Files:**

- Modify: `docs/assets/picker-demo.gif`

- [ ] **Step 1: Run setup**

Run:

```bash
bash docs/assets/picker-demo-setup.sh
```

Expected: exit 0, ends with `ready. now run: vhs docs/assets/picker-demo.tape`. Required even if Task 2 already ran the script — the demo dirs are throwaway.

- [ ] **Step 2: Render the GIF**

Run:

```bash
vhs docs/assets/picker-demo.tape
```

Expected:

- exit 0
- `/tmp/ccw-demo.gif` exists and is non-empty

If the command fails because vhs is not installed: install it first (`brew install vhs`) and re-run.

- [ ] **Step 3: Copy the GIF into the repo**

Run:

```bash
cp /tmp/ccw-demo.gif docs/assets/picker-demo.gif
```

- [ ] **Step 4: Visual verification**

Open `docs/assets/picker-demo.gif` in an image viewer and confirm each item below. Stop and re-record (back to Step 1) if any fails.

1. Each worktree row renders as 3 lines: header (`🌲 <name>` + status + indicators), `branch:`, `pr:`.
2. The 🌲 prefix is visible on every worktree row.
3. `↑0 ↓0` and other right-aligned indicators are not clipped at the right edge.
4. The walk-through hovers all four worktrees in this order:
   - `feat/login`: `[PUSHED]` + `[OPEN] #42` + RESUME
   - `feat/dashboard`: `[PUSHED]` + `[MERGED] #44` + RESUME
   - `feat/picker`: `[LOCAL]` + `[DRAFT] #43` + NEW
   - `chore/cleanup`: `[DIRTY]` + `[CLOSED] #45` + NEW
5. The submenu shows the **full worktree path** along with `[r] run / [d] delete / [b] back`.
6. The bulk-confirm screen (after `[clean pushed]`) lists the pushed targets as a preview.
7. `N` cancels back to the picker; `q` quits cleanly.

- [ ] **Step 5: Commit**

```bash
git add docs/assets/picker-demo.gif
git commit -m "docs(tape): regenerate picker-demo.gif against new TUI"
```

---

## Self-Review Checklist (already run)

- **Spec coverage:** Comments updated (Task 1), Sleep adjusted to 3000ms (Task 1), setup.sh sanity check (Task 2), regeneration steps (Task 3), all 7 spec verification points covered in Task 3 Step 4. Terminal size unchanged (Task 1 keeps `Width 1440 / Height 640 / FontSize 24`).
- **Placeholders:** None — every step contains the exact file content, command, or checklist item.
- **Type / name consistency:** `picker-demo.tape`, `picker-demo-setup.sh`, `picker-demo.gif`, `/tmp/ccw-demo.gif` are spelled identically across tasks.
