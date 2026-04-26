# actions/create-github-app-token client-id migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `release` ワークフローの `actions/create-github-app-token` 入力 `app-id` を deprecation 警告のない `client-id` に置き換え、annotations を解消する。

**Architecture:** `.github/workflows/release.yml` の `Mint Homebrew tap token` ステップを 1 行差し替えるだけ。Secret `GHA_APP_CLIENT_ID` は既に Client ID 値を保持しているため、Secret 側の変更は不要。Go コードへの変更なし。検証は lefthook の静的フック（`yamllint` / `actionlint` / `pinact`）と post-merge での `workflow_dispatch` 実行で行う。

**Tech Stack:** GitHub Actions / `actions/create-github-app-token@v3` / lefthook (yamllint / actionlint / pinact) / gh CLI

---

## File Structure

| Path | 操作 | 責務 |
|---|---|---|
| `.github/workflows/release.yml` | Modify | `app-id:` → `client-id:` の 1 行差し替え |
| `docs/superpowers/plans/2026-04-26-create-github-app-token-client-id-migration.md` | Create | このプラン本体 |

Spec: `docs/superpowers/specs/2026-04-26-create-github-app-token-client-id-migration-design.md`

---

### Task 1: ベースライン確認

**Files:**

- Read only: `.github/workflows/release.yml`

- [ ] **Step 1: 該当行を確認**

Run:

```bash
grep -n "app-id\|client-id" .github/workflows/release.yml
```

Expected output（1 行のみマッチ）:

```text
64:          app-id: ${{ secrets.GHA_APP_CLIENT_ID }}
```

行番号がずれている場合はその後の Task で番号を読み替える。マッチが 2 行以上ある場合は事前確認した spec と乖離しているため、ユーザーに確認する。

- [ ] **Step 2: 現在の deprecation 警告を再現確認（任意）**

Run:

```bash
gh run view 24938206359 --repo tqer39/ccw-cli --log 2>&1 | grep "deprecated"
```

Expected output（少なくとも 1 行含む）:

```text
##[warning]Input 'app-id' has been deprecated with message: Use 'client-id' instead.
```

これは現状確認のためで、ネットワーク不通なら省略してよい。

- [ ] **Step 3: ブランチを確認**

Run:

```bash
git branch --show-current
```

Expected output:

```text
feat/create-github-app-token-client-id
```

spec コミット時に既にこのブランチに切り替わっている前提。`main` に居る場合は次を実行してから次のタスクへ進む。

```bash
git checkout -b feat/create-github-app-token-client-id
```

---

### Task 2: workflow YAML を修正

**Files:**

- Modify: `.github/workflows/release.yml`（`Mint Homebrew tap token` ステップ内、現在 line 64）

- [ ] **Step 1: 該当行を `client-id` に書き換える**

`.github/workflows/release.yml` の以下の差分を適用する。

```diff
       - name: Mint Homebrew tap token
         id: tap-token
         uses: actions/create-github-app-token@1b10c78c7865c340bc4f6099eb2f838309f1e8c3 # v3.1.1
         with:
-          app-id: ${{ secrets.GHA_APP_CLIENT_ID }}
+          client-id: ${{ secrets.GHA_APP_CLIENT_ID }}
           private-key: ${{ secrets.GHA_APP_PRIVATE_KEY }}
           owner: tqer39
           repositories: homebrew-tap
```

その他のフィールド（`uses` の SHA pin / `private-key` / `owner` / `repositories`）は一切変更しない。

- [ ] **Step 2: 差分を確認**

Run:

```bash
git diff .github/workflows/release.yml
```

Expected output（該当行のみ変わっている）:

```diff
-          app-id: ${{ secrets.GHA_APP_CLIENT_ID }}
+          client-id: ${{ secrets.GHA_APP_CLIENT_ID }}
```

差分が 1 行（削除 1 / 追加 1）以外の場合は意図しない変更があるので revert して Step 1 をやり直す。

---

### Task 3: 静的検証

**Files:**

- Read only: `.github/workflows/release.yml`

- [ ] **Step 1: yamllint を実行**

Run:

```bash
yamllint --no-warnings .github/workflows/release.yml
```

Expected exit code: `0`（出力なし）

`yamllint: command not found` の場合は `mise install` か、`brew install yamllint` で導入してから再試行。

- [ ] **Step 2: actionlint を実行**

Run:

```bash
actionlint .github/workflows/release.yml
```

Expected exit code: `0`（出力なし）

`actionlint: command not found` の場合は `mise install` か `brew install actionlint`。

`create-github-app-token` の `client-id` 入力が未対応バージョンの actionlint だと警告が出る可能性があるが、本リポジトリの lefthook で同コマンドが既に運用されているので通る想定。仮に warning が出た場合はメッセージを記録の上ユーザーに報告する。

- [ ] **Step 3: pinact で SHA pin の整合を確認**

Run:

```bash
pinact run --check .github/workflows/release.yml
```

Expected exit code: `0`

未導入の場合は `brew install pinact`。pin そのものは変更していないので失敗しない想定。

- [ ] **Step 4: lefthook の pre-commit を空打ちして全フック確認**

ファイルをステージしてからフック相当を流す。

Run:

```bash
git add .github/workflows/release.yml
lefthook run pre-commit --files .github/workflows/release.yml
```

Expected exit code: `0`

`lefthook: command not found` の場合は `brew install lefthook`。フックがコミット時に走る構成なので、ここで通せば次の Task のコミットも通る見込み。

---

### Task 4: コミット

**Files:**

- Stage: `.github/workflows/release.yml`

- [ ] **Step 1: ステージ状態を確認**

Run:

```bash
git status --short
```

Expected output（`M` のみ。`??` などは含まない）:

```text
M  .github/workflows/release.yml
```

未ステージ（`M`）になっていたら `git add .github/workflows/release.yml` を再実行。

- [ ] **Step 2: コミット**

Run:

```bash
git commit -m "$(cat <<'EOF'
ci(release): use client-id input for create-github-app-token

actions/create-github-app-token v3 で deprecation 警告となっていた
`app-id` 入力を `client-id` に置き換えて警告を解消する。Secret
GHA_APP_CLIENT_ID は既に Client ID 値を保持しているため値の変更は
不要。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

Expected: 通常のコミット成功出力（lefthook の pre-commit が green）。pre-commit が落ちた場合は出力を確認し、Task 3 に戻る。`--no-verify` は使わない。

- [ ] **Step 3: ログを確認**

Run:

```bash
git log --oneline -3
```

Expected output（直近 1 件目に新コミット、2 件目が spec コミット）:

```text
<sha> ci(release): use client-id input for create-github-app-token
eaf5cc3 docs(specs): add design for create-github-app-token client-id migration
9d3dc6e feat(namegen): deterministic worktree names (ccw-<owner>-<repo>-<shorthash6>) (#48)
```

---

### Task 5: PR の作成

**Files:** （変更なし。GitHub 側の操作のみ）

- [ ] **Step 1: ブランチを push**

Run:

```bash
git push -u origin feat/create-github-app-token-client-id
```

Expected: push 成功。リモートが既に存在する場合は `--force-with-lease` ではなく通常 push で済む（ローカル先行のはずなので）。

- [ ] **Step 2: PR を作成**

Run:

```bash
gh pr create --title "ci(release): migrate create-github-app-token to client-id" --body "$(cat <<'EOF'
## Summary

- `.github/workflows/release.yml` の `Mint Homebrew tap token` ステップで `actions/create-github-app-token` の入力を `app-id` から `client-id` に切り替える。
- `actions/create-github-app-token` v3 で `app-id` が deprecated 化しており、`run 24938206359` の annotations に `Input 'app-id' has been deprecated with message: Use 'client-id' instead.` が出ていた件への対応。
- Secret `GHA_APP_CLIENT_ID` には既に Client ID 値が格納されているため、ワークフロー側の入力フィールド名のみを更新する。

Spec: `docs/superpowers/specs/2026-04-26-create-github-app-token-client-id-migration-design.md`

## Test plan

- [ ] lefthook の pre-commit（yamllint / actionlint / pinact）が green
- [ ] PR の Actions 上で workflow 構文エラーが出ない
- [ ] マージ後に `gh workflow run release.yml -f version=...` で実行し、`Mint Homebrew tap token` ステップに `app-id` deprecation warning が出ないこと、Homebrew tap への push が成功することを確認

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Expected: PR URL が返る。出力された URL をユーザーに伝える。

- [ ] **Step 3: PR URL を控える**

Run:

```bash
gh pr view --json url --jq .url
```

Expected: `https://github.com/tqer39/ccw-cli/pull/<n>` 形式。実装報告にこの URL を含める。

---

### Task 6: post-merge 検証メモ（実行しない）

このタスクは PR マージ後にユーザー側で行う検証手順を記録するもの。エンジニアはこの段階では何もコマンドを走らせない。

- [ ] **Step 1: 検証手順をユーザーに伝える**

PR マージ後、ユーザーに以下を依頼する旨を最終報告に含める。

1. 次回のリリース機会、または手動で:

   ```bash
   gh workflow run release.yml -f version=vX.Y.Z --repo tqer39/ccw-cli
   ```

2. 直近 run の log を取得:

   ```bash
   gh run list --workflow release.yml --limit 1 --repo tqer39/ccw-cli
   gh run view <run-id> --repo tqer39/ccw-cli --log | grep -i "deprecated\|app-id\|client-id"
   ```

3. Expected: `Input 'app-id' has been deprecated` が **出力されないこと**。`Mint Homebrew tap token` ステップが success し、Homebrew tap リポジトリへの push が完了していること。

---

## Self-Review チェック

- spec の「変更内容」「検証方法」「スコープ外」「リスク / ロールバック」が Task 2〜6 で全てカバーされている
- placeholder（TBD / TODO）なし
- 1 ファイル 1 行の変更につき、ファイル名・行範囲・差分・期待出力すべて具体化済み
- goreleaser の brews → homebrew_casks 移行は spec と同じくスコープ外として明示
