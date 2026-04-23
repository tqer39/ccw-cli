# README `--resume` Passthrough Caveat (PR-C) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `README.md` と `docs/README.ja.md` の `### Worktree picker` セクション内に、`ccw -- --resume ID` / `ccw -n -- --resume ID` / `ccw -s -- --resume ID` の併用が意図と合わない旨を示す warning blockquote（EN / JA）を追加する。

**Architecture:** Docs-only change. 2 ファイル編集のみ。コードパスは触らない。挿入位置は picker 説明の末尾、`PR display requires gh` 段落の**直後**（`## 📦 Installation` 直前）。picker の挙動説明と一括して warning が読まれる配置。

**Tech Stack:** Markdown, lefthook, markdownlint-cli2, cspell, textlint, プロジェクトローカルの `readme-sync` skill。

**Spec:** `docs/superpowers/specs/2026-04-24-readme-resume-passthrough-caveat-design.md`

---

## File Structure

- Modify: `README.md` — `### Worktree picker` セクション末尾（`PR display requires gh` 段落と `## 📦 Installation` の間）に warning blockquote を 1 つ挿入
- Modify: `docs/README.ja.md` — 対応位置に JA 版 warning blockquote を挿入
- No new files, no deletions.

---

## Task 1: `README.md`（EN）に warning blockquote を追加

**Files:**

- Modify: `README.md`（現状の 66 行目と 68 行目の間に挿入）

- [ ] **Step 1: 現状の該当箇所を確認**

```bash
sed -n '56,70p' README.md
```

Expected: 56 行目が `### Worktree picker`、66 行目が `PR display requires [\`gh\`](<https://cli.github.com/>). ...`、68 行目が`## 📦 Installation`（67 行目は空行）であることを確認。

- [ ] **Step 2: `PR display requires gh` 段落の直後に warning blockquote を追加**

`README.md` の 66 行目（`PR display requires [\`gh\`](<https://cli.github.com/>). Without \`gh\`, the picker stays functional and shows a hint; rate-limit / network failures hide the PR column silently.`）と 68 行目（`## 📦 Installation`）の間にある空行を使い、その後に以下のブロックと空行を挿入する（結果として`PR display` 段落 → 空行 → 新 blockquote → 空行 → `## 📦 Installation` の順になる）。

挿入する本文（Edit ツールの old_string / new_string で行う場合、`old_string` には 66 行目の段落全体と直後の空行＋見出し行を含めて一意化する）:

```md
PR display requires [`gh`](https://cli.github.com/). Without `gh`, the picker stays functional and shows a hint; rate-limit / network failures hide the PR column silently.

> ⚠️ **Passing `--resume` through `--` is unsupported.**
> `ccw -n -- --resume ID` and `ccw -s -- --resume ID` combine `claude --worktree` (new worktree) with `--resume` (continue a prior session); the resumed transcript's file references won't match the freshly-created worktree. Even the picker's re-entry path suffers the same mismatch if the selected worktree differs from the session's original. If a resumed session is what you want, run `claude --resume ID` directly — bypass ccw.

## 📦 Installation
```

Edit ツールでの具体的な置換:

- `old_string`:

  ```md
  PR display requires [`gh`](https://cli.github.com/). Without `gh`, the picker stays functional and shows a hint; rate-limit / network failures hide the PR column silently.

  ## 📦 Installation
  ```

- `new_string`:

  ```md
  PR display requires [`gh`](https://cli.github.com/). Without `gh`, the picker stays functional and shows a hint; rate-limit / network failures hide the PR column silently.

  > ⚠️ **Passing `--resume` through `--` is unsupported.**
  > `ccw -n -- --resume ID` and `ccw -s -- --resume ID` combine `claude --worktree` (new worktree) with `--resume` (continue a prior session); the resumed transcript's file references won't match the freshly-created worktree. Even the picker's re-entry path suffers the same mismatch if the selected worktree differs from the session's original. If a resumed session is what you want, run `claude --resume ID` directly — bypass ccw.

  ## 📦 Installation
  ```

- [ ] **Step 3: 差分を視覚確認**

```bash
git diff -- README.md
```

Expected: 追加行のみ（2 行の blockquote + 直後の空行）。削除行なし。`## 📦 Installation` の位置は下にずれるだけで中身は変わらない。

---

## Task 2: `docs/README.ja.md`（JA）に warning blockquote を追加

**Files:**

- Modify: `docs/README.ja.md`（EN と同じ相対位置）

- [ ] **Step 1: 現状の該当箇所を確認**

```bash
sed -n '56,70p' docs/README.ja.md
```

Expected: 56 行目が `### Worktree picker`、66 行目が `PR 表示には [\`gh\`](<https://cli.github.com/>) が必要です。...`、68 行目が`## 📦 インストール`。

- [ ] **Step 2: `PR 表示には gh が必要です` 段落の直後に warning blockquote を追加**

Edit ツールでの置換:

- `old_string`:

  ```md
  PR 表示には [`gh`](https://cli.github.com/) が必要です。`gh` が無い場合も picker は動作し、ヒントを下部に表示。rate limit / ネットワークエラー時は PR 列だけを静かに隠します。

  ## 📦 インストール
  ```

- `new_string`:

  ```md
  PR 表示には [`gh`](https://cli.github.com/) が必要です。`gh` が無い場合も picker は動作し、ヒントを下部に表示。rate limit / ネットワークエラー時は PR 列だけを静かに隠します。

  > ⚠️ **`-- --resume ID` のパススルーは非推奨です。**
  > `ccw -n -- --resume ID` や `ccw -s -- --resume ID` は `claude --worktree`（新 worktree 作成）と `--resume`（過去セッション継続）を同時に使うことになり、resume された会話中のファイル参照が新 worktree の実体と合いません。picker 経由で既存 worktree に再入場する場合も、選んだ worktree と session 元の worktree が違えば同様のズレが出ます。過去セッションを resume したいときは ccw を介さず直接 `claude --resume ID` を呼んでください。

  ## 📦 インストール
  ```

- [ ] **Step 3: 差分を視覚確認**

```bash
git diff -- docs/README.ja.md
```

Expected: 追加行のみ（2 行の blockquote + 直後の空行）。削除行なし。

---

## Task 3: `readme-sync` skill で EN / JA parity を確認

**Files:**

- Read: `README.md`, `docs/README.ja.md`
- (追加の編集は skill が divergence を報告した場合のみ)

- [ ] **Step 1: `readme-sync` skill を起動**

`Skill` ツールで `readme-sync` を呼び出し、返ってきたチェックリストに従う。

- [ ] **Step 2: skill が差分を報告した場合の対処**

EN / JA で blockquote の位置・行数・リンクが揃っているかを skill が確認する。

- EN と JA で blockquote が**同じ段落の直後**（`PR display requires gh` / `PR 表示には gh が必要です`）に入っていること
- 両 blockquote とも 2 行で構成されていること（見出し行 + 本文行）
- 参照 URL・コマンドサンプル (`ccw -n -- --resume ID` など) の表記揺れが無いこと

問題があれば該当側を修正し、再度 Step 1 を実行して clean になるまで繰り返す。
skill が no issues を返したら Task 4 に進む。

---

## Task 4: `lefthook run pre-commit` で lint を事前実行

**Files:** none（ツール実行のみ）

- [ ] **Step 1: 変更ファイルをステージ**

```bash
git add README.md docs/README.ja.md
git status --short
```

Expected:

```text
M  README.md
M  docs/README.ja.md
```

- [ ] **Step 2: pre-commit を commit なしで実行**

```bash
lefthook run pre-commit
```

Expected: `.md` にかかるフック（`markdownlint`, `detect-private-key`, `check-added-large-files`, `cspell`, `textlint`）が全て pass。Go / yaml / shell 向けのフックは skip されて OK。

発生しがちな失敗とインライン修正方針:

- **markdownlint MD013（line-length）**: blockquote の 2 行目は本文がやや長め。必要なら自然な区切り（`;` の後、`— bypass ccw` の前など）で改行して `>` プレフィックスを維持した複数行 blockquote に分割する。ただし意味を変えない。
- **cspell unknown word**: `transcript's` / `mismatch` / `preamble` などの英単語は既存辞書に含まれる想定。新規未登録語が出たら `.cspell/` 配下の辞書（存在する場合）に追記するか、表現を既存語に置き換える。
- **textlint**: JA 側の読点・半角スペース周りで警告が出た場合は、既存ファイルの書き方に合わせる（例: 全角 → 半角スペースの有無、`()` の全角半角）。

- [ ] **Step 3: 修正が必要だった場合は再 stage して再実行**

```bash
git add README.md docs/README.ja.md
lefthook run pre-commit
```

全フックが `.md` に関して pass するまで繰り返す。

---

## Task 5: Commit

**Files:** none（git commit のみ）

- [ ] **Step 1: stage 済み diff を最終確認**

```bash
git diff --staged
```

Expected: 2 ファイルのみ変更。各ファイルとも追加のみ（blockquote 2 行 + 空行）。無関係な空白変更・改行コード変更なし。

- [ ] **Step 2: 最近の commit スタイルに合わせてコミット**

直近の履歴は lowercase type prefix（`docs:`, `fix:`, `skill:`）。以下をそのまま実行:

```bash
git commit -m "$(cat <<'EOF'
docs: warn that `-- --resume ID` passthrough is unsupported

- ccw -n / -s は claude --worktree を付けるため、--resume された
  会話の file 参照と新 worktree の実体が合わない
- picker 経由で別 worktree に再入場する場合も同様のズレが出る
- resume したい場合は ccw を介さず直接 `claude --resume ID` を呼ぶ
  旨を README.md / docs/README.ja.md の Worktree picker 節末尾に追記

Spec: docs/superpowers/specs/2026-04-24-readme-resume-passthrough-caveat-design.md

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

pre-commit フックがここで再度走り、失敗した場合は `--amend` や `--no-verify` は**使わない**。問題を直して re-stage し、**新しい** commit を作る。

- [ ] **Step 2b: フックが失敗した場合の追加コミット**

```bash
git add README.md docs/README.ja.md
git commit -m "docs: fix lint on --resume caveat"
```

- [ ] **Step 3: log 確認**

```bash
git log --oneline -3
```

Expected: 新 commit が最上段。2 段目以降は現在の `main` HEAD（`d32c886` またはそれ以降）につながっている。

---

## Task 6: Push と PR-C を作成

**Files:** none（git + gh のみ）

- [ ] **Step 1: 現在のブランチを push**

このワークツリーのブランチは `worktree-noble-wandering-kettle`（ccw 生成）。`git branch --show-current` で念のため確認してから origin に push する。

```bash
git branch --show-current
git push -u origin worktree-noble-wandering-kettle
```

- [ ] **Step 2: PR を作成**

```bash
gh pr create --title "docs: warn that \`-- --resume ID\` passthrough is unsupported" --body "$(cat <<'EOF'
## Summary

- `ccw -n -- --resume ID` / `ccw -s -- --resume ID` は `claude --worktree`（新 worktree）と `--resume`（過去セッション継続）を同時に使うことになり、resume された会話中のファイル参照が新 worktree の実体とズレる。
- picker 経由で既存 worktree に再入場する場合も、選んだ worktree が session 元と違えば同じズレが出る。
- resume したい場合は ccw を介さず直接 `claude --resume ID` を呼ぶ運用を推奨、という注意書きを `README.md` と `docs/README.ja.md` の Worktree picker 節末尾に追加する。
- コード側の検出・警告は本 PR のスコープ外（別 PR 候補）。

Spec: `docs/superpowers/specs/2026-04-24-readme-resume-passthrough-caveat-design.md`

## Test plan

- [ ] `lefthook run pre-commit` が 2 ファイルに対して全フック pass
- [ ] `readme-sync` skill が EN / JA divergence を報告しない
- [ ] GitHub 上での Markdown レンダリングで blockquote が意図通り表示される（PR open 後に目視確認）

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 3: PR URL を控える**

`gh pr create` が出力する PR URL をユーザーに報告する。

---

## Out of scope for PR-C

- `ccw` コード側での `-- --resume` 検出と警告出力（別 PR 候補）
- `--resume` 機能自体のサポート追加
- 他の claude フラグ（例: `--continue`）の扱い
