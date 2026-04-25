# Header 画像差し替え + tape の RESUME/NEW 反映 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** ヘッダー画像を新しいデザインに差し替え、demo tape / setup スクリプトを直近マージされた RESUME/NEW セッションバッジ機能に追従させる。

**Architecture:** ドキュメント / アセットのみの変更。Go コードは触らない。setup script に fake HOME (`/tmp/ccw-demo-home`) と fake `~/.claude/projects/<encoded>/dummy.jsonl` を仕込み、tape の export 行に `HOME` を追加することで、picker の `HasSession` 判定（`internal/worktree/has_session.go`）が deterministic に RESUME / NEW を返すようにする。最後に `vhs` で GIF を再生成。

**Tech Stack:** PNG (header asset), bash (setup script), [vhs](https://github.com/charmbracelet/vhs) tape 構文, GIF.

**Files:**

- Modify: `docs/assets/header.png` (binary 上書き)
- Modify: `docs/assets/picker-demo-setup.sh`
- Modify: `docs/assets/picker-demo.tape`
- Modify: `docs/assets/picker-demo.gif` (再生成)

参考:

- spec: `docs/superpowers/specs/2026-04-25-header-and-tape-refresh-design.md`
- session 判定ロジック: `internal/worktree/has_session.go:13-37`

---

## Task 1: ヘッダー画像差し替え

**Files:**

- Modify: `docs/assets/header.png`

- [ ] **Step 1: 現在のヘッダーをバックアップ確認（git で復元可能なので削除コミット込みで進める）**

Run: `file docs/assets/header.png ~/Downloads/header.png`
Expected: 旧 1778×592、新 2172×724 と表示される。

- [ ] **Step 2: 上書き**

Run:

```bash
cp ~/Downloads/header.png docs/assets/header.png
```

- [ ] **Step 3: サイズ・寸法確認**

Run: `file docs/assets/header.png && ls -lh docs/assets/header.png`
Expected: `PNG image data, 2172 x 724, 8-bit/color RGB, non-interlaced`、サイズ 約 1.0 MB。

- [ ] **Step 4: README プレビュー（任意・目視）**

Run: `open README.md` か VS Code でプレビューしてヘッダーが新画像に置き換わっていることを確認。

- [ ] **Step 5: コミット**

```bash
git add docs/assets/header.png
git commit -m "docs: update header.png to new design"
```

---

## Task 2: setup script に fake HOME / session log を追加

**Files:**

- Modify: `docs/assets/picker-demo-setup.sh`

- [ ] **Step 1: ファイル末尾、`echo "ready..."` の直前に fake HOME ブロックを追加**

`docs/assets/picker-demo-setup.sh` の最後の `echo "ready. now run: vhs docs/assets/picker-demo.tape"` の直前に以下を挿入:

```bash
# 5. fake HOME so the picker can detect RESUME / NEW deterministically.
# Encoding mirrors internal/worktree/has_session.go:EncodeProjectPath
# (replaces '/' and '.' with '-').
rm -rf /tmp/ccw-demo-home
PROJECTS=/tmp/ccw-demo-home/.claude/projects
mkdir -p "$PROJECTS"
for wt in feat-login feat-dashboard; do
  enc=$(printf '%s' "/tmp/ccw-demo/.claude/worktrees/$wt" | tr '/.' '--')
  mkdir -p "$PROJECTS/$enc"
  printf '{}\n' >"$PROJECTS/$enc/dummy.jsonl"
done

```

挿入後、`feat-login` と `feat-dashboard` の 2 worktree が RESUME、`feat-picker` と `chore-cleanup` の 2 worktree が NEW になる想定。

- [ ] **Step 2: 構文チェック**

Run: `bash -n docs/assets/picker-demo-setup.sh`
Expected: 何も出力されず exit 0。

- [ ] **Step 3: 一度通しで実行して fake HOME が想定どおりできているか確認**

Run:

```bash
bash docs/assets/picker-demo-setup.sh
ls /tmp/ccw-demo-home/.claude/projects/
```

Expected: 2 つのディレクトリ（feat-login / feat-dashboard をエンコードしたもの）が並び、それぞれに `dummy.jsonl` が存在。例:

```text
-tmp-ccw-demo--claude-worktrees-feat-login
-tmp-ccw-demo--claude-worktrees-feat-dashboard
```

- [ ] **Step 4: 1 worktree だけ手動で `HasSession` 相当のチェックを行う**

Run:

```bash
ls /tmp/ccw-demo-home/.claude/projects/-tmp-ccw-demo--claude-worktrees-feat-login/*.jsonl
```

Expected: `dummy.jsonl` が表示される。

- [ ] **Step 5: コミット**

```bash
git add docs/assets/picker-demo-setup.sh
git commit -m "docs(demo): seed fake \$HOME with session logs for RESUME/NEW demo"
```

---

## Task 3: tape の export 行に HOME を追加

**Files:**

- Modify: `docs/assets/picker-demo.tape:13`

- [ ] **Step 1: 13 行目を差し替える**

旧:

```text
Type "cd /tmp/ccw-demo && export PATH=/tmp/ccw-demo-bin:/tmp/fake-gh:$PATH && clear"
```

新:

```text
Type "cd /tmp/ccw-demo && export HOME=/tmp/ccw-demo-home PATH=/tmp/ccw-demo-bin:/tmp/fake-gh:$PATH && clear"
```

- [ ] **Step 2: tape 構文サニティチェック**

Run: `vhs validate docs/assets/picker-demo.tape`
Expected: 何もエラーが出ない（`vhs validate` 非対応バージョンなら `vhs --help` で一旦確認後 skip）。

> Note: vhs 0.11.0 に `validate` サブコマンドが無い場合、Step 2 は飛ばして Step 3 でまとめて確認する。

- [ ] **Step 3: HOME export が tape 内に確実に入っているか grep で確認**

Run: `grep "HOME=/tmp/ccw-demo-home" docs/assets/picker-demo.tape`
Expected: 該当行が 1 件表示される。

- [ ] **Step 4: コミット**

```bash
git add docs/assets/picker-demo.tape
git commit -m "docs(demo): set HOME in tape so picker shows RESUME/NEW"
```

---

## Task 4: GIF 再生成

**Files:**

- Modify: `docs/assets/picker-demo.gif`

- [ ] **Step 1: setup を最新で流し直す（Task 2 で実行済みでも、tape 直前にもう一度走らせて状態を deterministic にする）**

Run: `bash docs/assets/picker-demo-setup.sh`
Expected: 末尾に `ready. now run: vhs docs/assets/picker-demo.tape`。

- [ ] **Step 2: vhs で tape を実行**

Run: `vhs docs/assets/picker-demo.tape`
Expected: `Creating ...` のログを経て `/tmp/ccw-demo.gif` が生成される。エラーなく終了。

- [ ] **Step 3: 生成された GIF をリポジトリに反映**

Run: `cp /tmp/ccw-demo.gif docs/assets/picker-demo.gif`

- [ ] **Step 4: 寸法 / サイズ確認**

Run: `file docs/assets/picker-demo.gif && ls -lh docs/assets/picker-demo.gif`
Expected: `GIF image data, version 89a, 1440 x 640`。サイズは ~500 KB 前後（既存と同程度）。

- [ ] **Step 5: 目視確認（必須）**

Run: `open docs/assets/picker-demo.gif`

確認項目:

- picker に 4 行表示され、`feat/login` と `feat/dashboard` の行に `💬 RESUME`、`feat/picker` と `chore/cleanup` の行に `⚡ NEW` が見える
- worktree badge / PR badge カラムの整列が崩れていない
- 既存と同じ 60 秒前後の長さで walkthrough → サブメニュー → bulk cancel → quit まで再生される

問題があれば Task 2/3 に戻って修正。

- [ ] **Step 6: コミット**

```bash
git add docs/assets/picker-demo.gif
git commit -m "docs(demo): regenerate picker GIF with RESUME/NEW badges"
```

---

## Task 5: 仕上げ確認

- [ ] **Step 1: ワークツリーが clean か確認**

Run: `git status`
Expected: `nothing to commit, working tree clean`。

- [ ] **Step 2: コミットログを確認**

Run: `git log --oneline -6`
Expected: spec コミット (419f9b6) の上に Task 1-4 の 4 コミットが順番に並ぶ。

- [ ] **Step 3: 既存 Go テストが影響を受けていないか念のため確認**

Run: `go test ./...`
Expected: PASS（このプランは Go コードを触っていないので影響しないはず）。

- [ ] **Step 4: 完了**

このタスクで実装は完了。次は user に PR 化するか、ブランチをそのまま統合するかを確認する（superpowers:finishing-a-development-branch を必要に応じて）。

---

## Self-Review

**Spec 網羅:**

- Section 1（ヘッダー差し替え） → Task 1
- Section 2（setup script） → Task 2
- Section 3（tape） → Task 3
- Section 4（検証） → Task 4 Step 5 + Task 5 Step 3
- Section 5（非スコープ） → 守れている（Go 変更なし、README 文言変更なし、ja README 変更なし、GIF 分割なし）
- Section 6（リスク） → setup / tape ともに `/tmp/ccw-demo-home` のみ操作、実 HOME 非汚染

**Placeholder:** なし。各 step に具体コマンドあり。

**Type / 名称整合:**

- `feat-login` / `feat-dashboard` / `feat-picker` / `chore-cleanup` の表記は spec と plan で一致
- `tr '/.' '--'` で `EncodeProjectPath` を再現する説明 spec と plan で一致
- HOME path `/tmp/ccw-demo-home` 全タスクで一致
