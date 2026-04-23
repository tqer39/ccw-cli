---
name: readme-sync
description: Keeps ccw-cli's two README files in sync — `README.md` (English, repo root) and `docs/README.ja.md` (Japanese). Whenever either file is edited, or the user asks for a README change without naming a file, this skill must fire and propagate the semantically equivalent change to the other file. Trigger on any edit touching `README.md` or `docs/README.ja.md`, and on any user request about "the README" (singular) since the siblings drift silently if you don't.
---

# README Sync (ccw-cli)

This repo ships two README files that must stay in sync:

- `README.md` — English, repo root, authoritative structure
- `docs/README.ja.md` — Japanese translation, same structure, localized prose

When you edit one, always edit the other in the same change. Drifting siblings are the exact failure mode this skill exists to prevent.

## When to fire

Any of these is enough:

1. An Edit/Write/MultiEdit targets `README.md` or `docs/README.ja.md`.
2. The user asks to update, translate, reword, add a section to, or fix something in "the README" without naming a file.
3. A diff being reviewed touches only one of the two.

Do not wait for the user to say "and update the other one." Treat "update the README" as "update both READMEs."

## Workflow

1. **Identify the source of truth** — the file the user (or you) just modified. Everything else is a follow-up to the sibling.

2. **Classify the change:**
   - **Content change** (wording, new section, version bump, link change, new bullet): propagate the semantic change into the sibling, matching its existing voice and terminology. Don't machine-translate mechanically — read the surrounding paragraph and keep its tone. Code blocks, shell commands, URLs, and badge markup stay byte-identical.
   - **Structural change** (heading reorder, section added/removed, table row added): mirror the structure exactly. Heading text is localized; heading order and hierarchy match.
   - **Cosmetic/meta change** (badge updates, anchor fixes, image paths, the `🇺🇸 English · 🇯🇵 日本語` switcher row): apply identically, adjusting only relative paths where siblings sit at different depths (`README.md` at root needs `docs/README.ja.md`; `docs/README.ja.md` needs `../README.md`).

3. **Preserve legitimate divergence.** Japanese-specific phrasing, localized examples, or commentary that already differs between the files should stay as-is unless the change you're propagating directly affects it. Translate only the delta.

4. **Keep the language switcher links correct.** If a section move changes relative paths, verify both switcher rows still resolve:
   - `README.md` → `[🇺🇸 English](README.md) · [🇯🇵 日本語](docs/README.ja.md)`
   - `docs/README.ja.md` → `[🇺🇸 English](../README.md) · [🇯🇵 日本語](README.ja.md)`

5. **Report what you synced** in one or two sentences. The user can read the diff themselves.

## Why this matters for ccw-cli

ccw-cli is distributed via Homebrew to a bilingual audience. A typo in the English "Requirements" section that never lands in Japanese means JP users install the wrong Claude Code version for months. A new `-s` flag documented only in Japanese means English readers never discover it. The individual miss is small; the cumulative effect is a docs pair that's untrustworthy in at least one language.

## Anti-patterns

- **"I'll do the Japanese in a follow-up commit."** You won't. Do it now.
- **Re-translating the whole paragraph** when only one sentence changed. Touch only the delta — the user already carefully worded the rest.
- **Mechanically translating the English into Japanese** ignoring existing terminology. Look at how `docs/README.ja.md` already phrases similar concepts and match that.
- **Copying code blocks with their English comments** into the Japanese file when the existing Japanese file has localized comments inside the same block — keep the localized comments.
- **Touching the flag-emoji language switcher on every edit.** Leave it alone unless a file moved.
- **Creating a third language file (e.g. `README.zh.md`)** without being asked. Sync what exists.

## Examples

### Example 1: Version bump

`README.md` edit:

```diff
- Go >= 1.24
+ Go >= 1.25
```

Sync to `docs/README.ja.md`:

```diff
- Go >= 1.24
+ Go >= 1.25
```

(Surrounding Japanese prose is untouched.)

### Example 2: New section

`README.md` appends:

```markdown
## 🗺️ Roadmap

- Shell completion (bash / zsh)
- Windows support
```

`docs/README.ja.md` appends at the matching location:

```markdown
## 🗺️ ロードマップ

- シェル補完 (bash / zsh)
- Windows サポート
```

### Example 3: Link-only fix

Source swaps `https://old/...` → `https://new/...`. Sync: update the same URL in the sibling; do not retranslate surrounding prose.
