# README Features Rewrite (PR-A) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite the Features section of `README.md` and `docs/README.ja.md` so the 6 bullets lead with user-visible value (starting with ccw's "bridge-only" stance), and remove the now-redundant Quick Start blockquote.

**Architecture:** Docs-only change. Edit the two READMEs in-place; verify with `lefthook run pre-commit` (markdownlint / cspell / textlint). No code paths touched.

**Tech Stack:** Markdown, lefthook, markdownlint-cli2, cspell, textlint.

**Spec:** `docs/superpowers/specs/2026-04-24-readme-features-rewrite-design.md`

---

## File Structure

- Modify: `README.md` — Quick Start blockquote (line ~30) and Features section (lines ~32-38)
- Modify: `docs/README.ja.md` — 対応するブロッククォートと特長セクション
- No new files, no deletions.

---

## Task 1: Update `README.md` (EN)

**Files:**

- Modify: `README.md` (Quick Start blockquote + Features section)

- [ ] **Step 1: Delete the Quick Start blockquote**

Open `README.md`. Locate this block (it appears immediately after the Quick Start fenced code block, currently around line 30):

```md
> `ccw` also works from inside a worktree — it resolves the main repo via `git rev-parse --git-common-dir` and operates there, so you don't need to `cd` back to the project root first.
```

Delete the blockquote line AND the blank line before it if one exists (keep exactly one blank line between the Quick Start paragraph and the `## ✨ Features` heading).

- [ ] **Step 2: Replace the Features bullets**

Locate the `## ✨ Features` section, currently:

```md
## ✨ Features

- 🌳 **Isolated sessions** — each `claude` run gets its own git worktree
- 🎯 **Smart picker** — status badges, `↑N ↓M ✎N` indicators, PR info via `gh`
- 🧹 **Bulk delete** — `[clean pushed]` from the picker or `ccw --clean-all`
- 🦸 **Superpowers preamble** — `-s` injects the `brainstorming → writing-plans → executing-plans` workflow
- ➡️ **Transparent passthrough** — anything after `--` reaches `claude` verbatim
```

Replace with:

```md
## ✨ Features

- 🤝 **Hand-off and step aside** — pick (or create) a worktree, launch `claude` in it, then ccw exits. No daemon, no wrapper process, no coupling to tmux/zellij — just the bridge.
- 🧭 **Works from anywhere in the repo** — run `ccw` inside a worktree or subdirectory; ccw resolves the main repo automatically
- 🎯 **Worktree state at a glance** — pushed / ahead / behind / dirty, plus PR info, all in one picker
- 🧹 **Bulk cleanup** — `[clean pushed]` or `ccw --clean-all` sweeps the worktrees you're done with
- 🦸 **"Design first" startup** — `-s` tells claude to follow the brainstorming → writing-plans → executing-plans flow (prompts to install the superpowers plugin if missing)
- ➡️ **claude flags pass through** — anything after `--` goes to claude untouched, so `--model` and friends still work
```

- [ ] **Step 3: Verify the diff visually**

Run:

```bash
git diff -- README.md
```

Expected: only the Quick Start blockquote deletion + the 5→6 bullet replacement shown; no unrelated changes.

---

## Task 2: Update `docs/README.ja.md` (JA)

**Files:**

- Modify: `docs/README.ja.md` (Quick Start blockquote + 特長 section)

- [ ] **Step 1: Delete the Quick Start blockquote (JA)**

Open `docs/README.ja.md`. Locate the block (after the Quick Start code fence, ~line 30):

```md
> `ccw` は worktree 内から起動しても動作します — `git rev-parse --git-common-dir` で main repo を解決してそこを基準に動くので、プロジェクトルートに `cd` し直す必要はありません。
```

Delete this blockquote and the surrounding blank line pattern so exactly one blank line sits between the paragraph above and `## ✨ 特長`.

- [ ] **Step 2: Replace the 特長 bullets (JA)**

Locate:

```md
## ✨ 特長

- 🌳 **セッション分離** — `claude` 起動ごとに専用 git worktree
- 🎯 **スマート picker** — ステータスバッジ、`↑N ↓M ✎N` インジケータ、`gh` 経由の PR 情報表示
- 🧹 **一括削除** — picker の `[clean pushed]` や `ccw --clean-all`
- 🦸 **Superpowers プリアンブル** — `-s` で `brainstorming → writing-plans → executing-plans` ワークフローを注入
- ➡️ **透過的 passthrough** — `--` 以降の引数は `claude` にそのまま渡される
```

Replace with:

```md
## ✨ 特長

- 🤝 **橋渡しまでが仕事** — worktree を選ぶ（or 新規作成）→ その中で `claude` を起動 → ccw は終了。常駐プロセスもラッパーもなく、tmux/zellij にも噛まない。あとは claude の世界
- 🧭 **リポジトリ内のどこからでも起動** — worktree 内やサブディレクトリからでも `ccw` が動く（main repo を自動解決）
- 🎯 **worktree の状態が一目でわかる** — push 済 / ahead・behind / dirty、PR 番号を picker にまとめて表示
- 🧹 **溜まった worktree を一括掃除** — `[clean pushed]` / `ccw --clean-all` で push 済をまとめて削除
- 🦸 **"設計してから書く" 流儀で起動** — `-s` で brainstorming → writing-plans → executing-plans の手順を claude に指示（plugin 未導入なら入れるか確認）
- ➡️ **claude のオプションはそのまま届く** — `--` 以降の引数は素通しするので `--model` などが使える
```

- [ ] **Step 3: Verify the diff visually**

Run:

```bash
git diff -- docs/README.ja.md
```

Expected: only the blockquote deletion + the 5→6 bullet replacement; no unrelated changes.

---

## Task 3: Verify EN / JA parity with `readme-sync` skill

**Files:**

- Read: `README.md`, `docs/README.ja.md`
- (No edits expected unless the skill flags divergence)

- [ ] **Step 1: Invoke the `readme-sync` skill**

This repo has a project-local `readme-sync` skill (commit `ae860cf`) that enforces structural parity between `README.md` and `docs/README.ja.md`. Invoke it via the `Skill` tool with name `readme-sync`. Follow whatever checklist the skill returns.

- [ ] **Step 2: Address any mismatch the skill reports**

If the skill flags a section / bullet count / link mismatch:

- Re-read both files at the flagged location.
- Fix the side that drifted (usually JA mirrors EN; in this plan both were edited together so drift should be zero).
- Re-run Step 1 until clean.

If the skill reports no issues, proceed.

---

## Task 4: Run lefthook pre-commit locally to catch lint issues

**Files:** none (running tools)

- [ ] **Step 1: Stage the two modified files**

```bash
git add README.md docs/README.ja.md
git status --short
```

Expected output:

```text
M  README.md
M  docs/README.ja.md
```

- [ ] **Step 2: Run pre-commit hooks without committing**

```bash
lefthook run pre-commit
```

Expected: all hooks pass (`markdownlint`, `detect-private-key`, `check-added-large-files`, `cspell`, `textlint` — the ones that match `.md` files). Skipped hooks (Go / yaml / shell) are fine.

Common failure modes to fix inline:

- **cspell unknown word** (e.g. a new term like "tmux") → if genuinely a project term, add to the repo's cspell dictionary; otherwise rephrase.
- **textlint** complains about punctuation or spacing in JA → adjust punctuation style to match the rest of the file.
- **markdownlint** MD013 line length → break long bullets at a natural boundary (do not introduce new content).

- [ ] **Step 3: Re-run until clean**

If fixes were needed, re-stage and re-run Step 2 until every hook that touches `.md` passes.

---

## Task 5: Commit

**Files:** none (git commit)

- [ ] **Step 1: Verify staged diff one more time**

```bash
git diff --staged
```

Read through the diff. Confirm: exactly the two files are changed; only the Features sections and the two blockquotes are affected; no stray whitespace or unrelated edits.

- [ ] **Step 2: Commit with a message matching recent repo style**

Recent history uses lowercase type prefixes (`docs:`, `fix:`, `skill:`). Use:

```bash
git commit -m "$(cat <<'EOF'
docs: lead Features with ccw's bridge-only stance, drop jargon

- New 🤝 lead bullet makes the hand-off-and-exit stance explicit
- Rename #1 away from "Isolated sessions" (that's claude --worktree's
  feature, not ccw's) to "Works from anywhere in the repo"
- Rename 🦸 "Superpowers preamble" to "Design first startup"; drop
  the "プリアンブル" jargon on the JA side
- Drop the Quick Start blockquote (its content is now in 🧭)

Spec: docs/superpowers/specs/2026-04-24-readme-features-rewrite-design.md

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

If the pre-commit hook fires again at commit time and fails, **do not** `--amend` or `--no-verify`. Fix the issue, re-stage, and create a new commit.

- [ ] **Step 2b: If the hook failed, make a corrective commit**

```bash
git add README.md docs/README.ja.md
git commit -m "docs: fix lint on Features rewrite"
```

- [ ] **Step 3: Verify log**

```bash
git log --oneline -3
```

Expected: new commit on top of `b0073b1 docs(superpowers): add design specs for README revamp and -s flow`.

---

## Task 6: Push and open PR-A

**Files:** none (git + gh)

- [ ] **Step 1: Push the branch**

Current branch is `worktree-mighty-juggling-peacock` (ccw-generated). Push it to origin:

```bash
git push -u origin worktree-mighty-juggling-peacock
```

- [ ] **Step 2: Open the PR**

```bash
gh pr create --title "docs: lead Features with ccw's bridge-only stance" --body "$(cat <<'EOF'
## Summary

- Reframe the 6 Features bullets around user-visible benefits instead of jargon.
- Lead with 🤝 **Hand-off and step aside** to emphasize ccw's role as a thin launcher that exits after handing off to `claude`.
- Drop the Quick Start blockquote (redundant with the new 🧭 bullet).
- Mirror all changes in `docs/README.ja.md`.

Follows the same rationale as the earlier tagline rewrite (#19): stop describing the mechanism, describe what the user gets.

Spec: `docs/superpowers/specs/2026-04-24-readme-features-rewrite-design.md` (added in the preceding commit on this branch)

## Test plan

- [ ] `lefthook run pre-commit` passes on the two modified files
- [ ] `readme-sync` skill reports no EN/JA divergence
- [ ] Rendered output looks right on GitHub (visual check after PR is open)

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 3: Capture PR URL**

`gh pr create` prints the PR URL. Report it back to the user so they can review.

---

## Out of scope for PR-A

These are handled by their own specs and PRs:

- **PR-B:** `-s` auto-install (`-y` / non-interactive path) — `docs/superpowers/specs/2026-04-24-superpowers-auto-install-design.md`
- **PR-C:** README `-- --resume ID` caveat — `docs/superpowers/specs/2026-04-24-readme-resume-passthrough-caveat-design.md`
- **PR-D:** remove `EnsureGitignore` — `docs/superpowers/specs/2026-04-24-remove-gitignore-interference-design.md`
