# Translate picker tips to English Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** picker フッターの `💡 Tip:` 文言を日本語から英語へ置き換え、demo GIF を再生成し、PR を意味のあるブランチ名で公開する。

**Architecture:** `internal/tips/tips.go` の `defaults` 配列のみ書き換え。テストは内容非依存なので無改修で pass する想定。GIF は `picker-demo-setup.sh` + `vhs picker-demo.tape` で再生成し、`docs/assets/picker-demo.gif` を差し替える。PR は `git:create-branch` skill で `docs/translate-tips-to-english` を切ってから push。

**Tech Stack:** Go 1.25 / Bubbletea v2 picker / vhs (tape) / gh CLI

---

## File Structure

| Path | 操作 | 責務 |
|---|---|---|
| `internal/tips/tips.go` | Modify | `defaults` 配列の翻訳 |
| `internal/tips/tips_test.go` | 無変更 | 内容非依存テスト（変更しない） |
| `docs/assets/picker-demo.gif` | Replace | vhs 再生成で差し替え |
| `docs/assets/picker-demo.tape` | 無変更 | tape スクリプトに日本語含まず |
| `docs/superpowers/plans/2026-04-26-translate-tips-to-english.md` | Create | このプラン本体 |

---

### Task 1: ベースライン確認

**Files:**

- Read only: `internal/tips/tips.go`, `internal/tips/tips_test.go`

- [ ] **Step 1: 既存テスト・lint がグリーンであることを確認**

Run:

```bash
go test ./internal/tips/...
go vet ./internal/tips/...
```

Expected: `ok` / `PASS`、エラーなし。

- [ ] **Step 2: 全体テストもグリーン確認**

Run:

```bash
go test ./...
```

Expected: 全 package PASS。

このタスクは commit を作らない（観測のみ）。

---

### Task 2: tips.go を英訳

**Files:**

- Modify: `internal/tips/tips.go:6-12`

- [ ] **Step 1: `defaults` 配列を英訳版で置換**

`internal/tips/tips.go` を以下のように編集:

```go
var defaults = []string{
 "Worktree name = session name; renaming with /rename is fine, ccw doesn't track it",
 "claude --from-pr <number> resumes a PR-linked session directly",
 "--clean-all sweeps pushed worktrees in bulk",
 "ccw -- --model <id> passes flags through to claude",
 "The RESUME badge is derived from ~/.claude/projects/",
}
```

並び順は元のまま（git diff の対応関係を読みやすくするため）。tabs／インデントは元の `tips.go` に合わせること（go fmt 準拠の tab）。

- [ ] **Step 2: 既存テストが pass することを確認**

Run:

```bash
go test ./internal/tips/... -v
```

Expected: `TestPickRandom_FromDefaultSet`, `TestPickRandom_Deterministic`, `TestPickFrom_Empty`, `TestDefaults_NonEmpty` がすべて PASS。

- [ ] **Step 3: vet / build sanity check**

Run:

```bash
go vet ./...
go build ./cmd/ccw
```

Expected: エラーなし、`./ccw` バイナリが生成される（バイナリは後で消す）。

- [ ] **Step 4: 全体テストでも regression なし**

Run:

```bash
go test ./...
```

Expected: 全 package PASS。

- [ ] **Step 5: 一時バイナリをクリーン**

Run:

```bash
rm -f ./ccw
```

- [ ] **Step 6: コミット**

Run:

```bash
git add internal/tips/tips.go
git commit -m "$(cat <<'EOF'
i18n(tips): 日本語の picker tips を英訳

picker フッターの 5 件の tip を英語に置き換え。tips の選択ロジック
（PickRandom）と tests は変更なし。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: vhs の前提ツールを確認

**Files:** N/A（環境チェックのみ）

- [ ] **Step 1: `vhs` コマンドの存在確認**

Run:

```bash
command -v vhs && vhs --version
```

Expected: パスとバージョンが表示される。

- [ ] **Step 2: 失敗時の対応**

`vhs` が見つからない場合は GIF 再生成 (Task 4) を skip し、Task 5 以降に進む。その場合は PR 本文で「GIF は別 PR で更新予定」と明記する。Step 1 が成功した場合は Task 4 を実施。

---

### Task 4: picker-demo.gif の再生成

**Files:**

- Replace: `docs/assets/picker-demo.gif`

- [ ] **Step 1: demo 環境を再構築**

Run:

```bash
bash docs/assets/picker-demo-setup.sh
```

Expected: 末尾に `ready. now run: vhs docs/assets/picker-demo.tape` と表示。`/tmp/ccw-demo-bin/ccw` には英訳済みの新しい binary が入る（setup.sh が `go build` を含むため）。

- [ ] **Step 2: vhs で GIF 生成**

Run:

```bash
vhs docs/assets/picker-demo.tape
```

Expected: `/tmp/ccw-demo.gif` が作成される。エラーなし。

- [ ] **Step 3: 生成物を repo 内へコピー**

Run:

```bash
cp /tmp/ccw-demo.gif docs/assets/picker-demo.gif
```

- [ ] **Step 4: 視覚確認**

`docs/assets/picker-demo.gif` を任意のビューワで開き、フッター `💡 Tip:` の文字列が英語であることを確認する。VS Code でも開ける。

GIF はランダム 1 件しか映らないため、5 件すべての英訳完了は Task 2 のテストとレビューで担保。GIF は「英語が映っている」ことだけ確認する。

- [ ] **Step 5: コミット**

Run:

```bash
git add docs/assets/picker-demo.gif
git commit -m "$(cat <<'EOF'
docs(demo): tips 英訳に合わせて picker-demo.gif を再生成

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: PR 用ブランチを作成

**Files:** N/A（git operation のみ）

- [ ] **Step 1: 現在のブランチ・HEAD を記録**

Run:

```bash
git branch --show-current
git log --oneline -5
```

Expected: `worktree-rapid-lion-5f9c` 上に Task 2 / Task 4 のコミット、および spec コミット (`docs: tips 英訳タスクの spec を追加`) が積まれている。

- [ ] **Step 2: `git:create-branch` skill を呼び出して PR 用ブランチを作成**

Skill tool 経由で `git:create-branch` を invoke する。引数: ブランチ名 `docs/translate-tips-to-english`、base は `main`。

skill が「base から空のフィーチャーブランチを作る」挙動の場合、後段で worktree-rapid-lion-5f9c のコミットを cherry-pick する必要がある。skill 実行後の状態を確認:

```bash
git branch --show-current
git log --oneline -5
```

挙動分岐:

- (A) skill が新ブランチを作成し、そのブランチに HEAD を移したが Task 2/4 のコミットが乗っていない:
  - 元の `worktree-rapid-lion-5f9c` から該当コミット 3 件（spec、tips 英訳、GIF）を `git cherry-pick` する。

  ```bash
  git cherry-pick worktree-rapid-lion-5f9c~2..worktree-rapid-lion-5f9c
  ```

- (B) skill が現 HEAD から名前だけ変えるタイプの動作で、すでに 3 件乗っている: 何もしない。

どちらでも `git log --oneline -5` で spec / tips / GIF の 3 コミットが先頭にある状態にする。

- [ ] **Step 3: ブランチ最終確認**

Run:

```bash
git branch --show-current
git log --oneline main..HEAD
```

Expected: `docs/translate-tips-to-english` 上で main からの差分 3 コミットが表示される。

---

### Task 6: push & PR 作成

**Files:** N/A（リモート操作）

- [ ] **Step 1: ユーザー確認**

push と PR 作成はリモート影響があるため、auto モードでも実行前に確認を取る。チャットで「push & PR 作成してよいか」を尋ね、ユーザー OK 後に Step 2 へ。

- [ ] **Step 2: push**

Run:

```bash
git push -u origin docs/translate-tips-to-english
```

Expected: GitHub 側に新ブランチが作成される。

- [ ] **Step 3: PR 作成**

Run:

```bash
gh pr create --base main --head docs/translate-tips-to-english --title "i18n(tips): translate picker tips to English" --body "$(cat <<'EOF'
## Summary
- picker フッター `💡 Tip:` の 5 件を日本語から英語に翻訳
- `docs/assets/picker-demo.gif` を再生成（vhs 利用、英訳された tip がフッターに映る）
- 設計と計画は `docs/superpowers/specs/2026-04-25-translate-tips-to-english-design.md` / `docs/superpowers/plans/2026-04-26-translate-tips-to-english.md`

## Test plan
- [ ] `go test ./...` がローカルで PASS
- [ ] `go vet ./...` clean
- [ ] picker を起動してフッター tip が英語表示
- [ ] `docs/assets/picker-demo.gif` の再生で英語の `💡 Tip:` が確認できる

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Expected: PR URL が返る。

- [ ] **Step 4: PR URL を報告**

PR URL をユーザーへ提示して終了。

---

## Acceptance criteria

spec と同じ:

- [ ] `internal/tips/tips.go` の 5 行が英語化されている
- [ ] `go test ./...` が pass
- [ ] `go vet ./...` clean
- [ ] `docs/assets/picker-demo.gif` 差し替え（vhs 不在時は別 PR とする例外あり）
- [ ] PR ブランチ名が `docs/translate-tips-to-english`
- [ ] `docs/README.ja.md` 等は無変更

## Risks

- **vhs 未インストール**: Task 3 でガード。GIF は別 PR にスライドする。
- **cherry-pick 競合**: Task 5 (A) で cherry-pick が必要な場合、競合は理屈上発生しないはず（同じ HEAD からの分岐なので）。発生したら手動 resolve。
- **push 拒否（保護ブランチ等）**: 起きないはずだが、起きたら Step 2 で報告して停止。
