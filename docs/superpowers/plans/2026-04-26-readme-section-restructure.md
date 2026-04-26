# README Section Restructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Promote `### Worktree picker` and `### Naming convention` (currently H3 inside `## 📖 Usage`) to standalone H2 sections in both `README.md` (English) and `docs/README.ja.md` (Japanese), with no text changes — move only.

**Architecture:** Pure documentation refactor. Two files edited in lockstep. No code, no tests. Verification via `markdownlint-cli2` and visual diff.

**Tech Stack:** Markdown, markdownlint-cli2 (via lefthook).

**Spec:** [`docs/superpowers/specs/2026-04-26-readme-section-restructure-design.md`](../specs/2026-04-26-readme-section-restructure-design.md)

---

## File Structure

| File | Change |
|---|---|
| `README.md` | Modify lines 59–97 area: split off two H3 → H2 sections after `## 📖 Usage` |
| `docs/README.ja.md` | Same restructure for the Japanese translation |

No new files. No deleted files.

---

### Task 1: Restructure `README.md` (English)

**Files:**

- Modify: `README.md` (move H3 sections out of Usage, promote to H2)

The current `## 📖 Usage` section ends at line 57 (`Run \`ccw --help\` for the full flag reference.`). After it, two H3 blocks follow:

- `### Worktree picker` (lines 59–87)
- `### Naming convention` (lines 89–97)

These two H3 blocks must be promoted to standalone H2 sections placed directly after Usage and before `## 📦 Installation`.

- [ ] **Step 1: Read the current `README.md` to confirm exact line ranges**

Run: `wc -l README.md && sed -n '57,99p' README.md`

Expected: confirms `## 📖 Usage` ends at line 57, `### Worktree picker` is at line 59, `### Naming convention` is at line 89, `## 📦 Installation` starts at line 99.

- [ ] **Step 2: Replace `### Worktree picker` heading with `## 🎯 Picker reference`**

In `README.md`, replace the exact line:

```text
### Worktree picker
```

with:

```text
## 🎯 Picker reference
```

The body content beneath (the three badge tables and the two trailing paragraphs starting `Selecting a worktree opens...` and `Without \`gh\`, the picker stays...`) stays unchanged.

- [ ] **Step 3: Replace `### Naming convention` heading with `## 🏷️ Naming`**

In `README.md`, replace the exact line:

```text
### Naming convention
```

with:

```text
## 🏷️ Naming
```

The body content (`When ccw creates a new worktree...` paragraph + 3 bullets + the `\`<name>\` is generated as...` paragraph) stays unchanged.

- [ ] **Step 4: Verify the new heading hierarchy**

Run:

```bash
grep -n '^##' README.md
```

Expected output (in order):

```text
## ⚡ Quick Start
## ✨ Features
## 🎬 Demo
## 📖 Usage
## 🎯 Picker reference
## 🏷️ Naming
## 📦 Installation
## 🪝 Auto-prompt on new worktree sessions
## ⚙️ Environment
## 🛠️ Development
## 🤖 Built With
## 📄 License
```

Also confirm no `### Worktree picker` or `### Naming convention` remain:

```bash
grep -nE '^### (Worktree picker|Naming convention)' README.md
```

Expected: no matches (exit code 1).

- [ ] **Step 5: Verify Usage body shrank to commands + one trailing line**

Run:

```bash
awk '/^## 📖 Usage/,/^## 🎯 Picker reference/' README.md | grep -c '^##'
```

Expected: `2` (just the two H2 headings; everything between is the code block + the `Run \`ccw --help\` ...` line).

Run also:

```bash
awk '/^## 📖 Usage/,/^## 🎯 Picker reference/' README.md
```

Expected: shows `## 📖 Usage`, blank line, the ```bash code block (8 command lines), blank line, `Run \`ccw --help\` for the full flag reference.`, blank line,`## 🎯 Picker reference`. No badge tables, no`### Worktree picker`, no`### Naming convention`.

---

### Task 2: Restructure `docs/README.ja.md` (Japanese)

**Files:**

- Modify: `docs/README.ja.md` (same restructure mirrored to Japanese)

- [ ] **Step 1: Read the current `docs/README.ja.md` to confirm exact line ranges**

Run: `wc -l docs/README.ja.md && sed -n '57,99p' docs/README.ja.md`

Expected: confirms `## 📖 使い方` ends at line 57, `### Worktree picker` at line 59, `### 命名規約` at line 89, `## 📦 インストール` at line 99.

- [ ] **Step 2: Replace `### Worktree picker` heading with `## 🎯 Picker リファレンス`**

In `docs/README.ja.md`, replace the exact line:

```text
### Worktree picker
```

with:

```text
## 🎯 Picker リファレンス
```

Body unchanged.

- [ ] **Step 3: Replace `### 命名規約` heading with `## 🏷️ 命名規約`**

In `docs/README.ja.md`, replace the exact line:

```text
### 命名規約
```

with:

```text
## 🏷️ 命名規約
```

Body unchanged.

- [ ] **Step 4: Verify the new heading hierarchy**

Run:

```bash
grep -n '^##' docs/README.ja.md
```

Expected output (in order):

```text
## ⚡ Quick Start
## ✨ 特長
## 🎬 デモ
## 📖 使い方
## 🎯 Picker リファレンス
## 🏷️ 命名規約
## 📦 インストール
## 🪝 新規 worktree セッションでの自動プロンプト注入
## ⚙️ 環境変数
## 🛠️ 開発
## 🤖 作成ツール
## 📄 ライセンス
```

Also confirm no `### Worktree picker` or `### 命名規約` remain:

```bash
grep -nE '^### (Worktree picker|命名規約)' docs/README.ja.md
```

Expected: no matches (exit code 1).

- [ ] **Step 5: Confirm 1:1 H2 count between EN and JA**

Run:

```bash
en=$(grep -c '^## ' README.md); ja=$(grep -c '^## ' docs/README.ja.md); echo "en=$en ja=$ja"
```

Expected: `en=12 ja=12`.

---

### Task 3: Lint and commit

**Files:** none modified — verification + commit.

- [ ] **Step 1: Run markdownlint-cli2**

Run:

```bash
npm exec -y --package=markdownlint-cli2 -- markdownlint-cli2 README.md docs/README.ja.md
```

Expected: exit code 0, no errors. (Config is `.markdownlint-cli2.jsonc` at the repo root.)

If errors appear, read the error, fix the offending line in place, re-run until clean.

- [ ] **Step 2: Visual diff sanity check**

Run:

```bash
git diff --stat README.md docs/README.ja.md
git diff README.md docs/README.ja.md | head -80
```

Expected: only the four lines containing `### Worktree picker`, `### Naming convention` / `### 命名規約` are changed (replaced by their new H2 forms). No other content modified. Total of ~4 lines changed per file (2 deletions + 2 additions).

- [ ] **Step 3: Stage and commit**

Run:

```bash
git add README.md docs/README.ja.md
git commit -m "$(cat <<'EOF'
docs: README の Picker / Naming を H2 に昇格

Usage 配下の H3 (Worktree picker / Naming convention) を独立 H2
(🎯 Picker reference / 🏷️ Naming) に昇格。Usage はコマンド例のみに縮小。
英 (README.md) と日 (docs/README.ja.md) を 1:1 で同期。文言の追加・削除
はなく、見出しの移動のみ。

Refs: docs/superpowers/specs/2026-04-26-readme-section-restructure-design.md
EOF
)"
```

Expected: commit succeeds; lefthook `markdownlint` hook runs on `README.md` / `docs/README.ja.md` and passes.

If the hook auto-fixes anything, re-stage and re-run. If the hook fails with errors that aren't auto-fixable, read the error and fix in place — do **not** use `--no-verify`.

- [ ] **Step 4: Confirm clean tree**

Run: `git status`

Expected: `nothing to commit, working tree clean`.
