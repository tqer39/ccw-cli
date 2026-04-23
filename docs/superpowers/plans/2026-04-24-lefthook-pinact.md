# lefthook で pinact を実行する — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** lefthook の pre-commit フックで pinact を自動実行し、GitHub Actions の `uses:` 参照を常に「SHA 固定 + バージョンコメント付き」に保つ。

**Architecture:** 3 ファイル変更のみの小規模導入。Brewfile に pinact を追加して `make bootstrap` で入るようにし、最小の `.pinact.yaml` を置き、lefthook.yml に `stage_fixed: true` 付きの pinact コマンドを追加する。既存 workflow は全て SHA 化済みのため初回適用の書き換え差分は出ない想定。

**Tech Stack:** lefthook, pinact (v3, Homebrew), Brewfile, GitHub Actions workflow files

**Reference:** `docs/superpowers/specs/2026-04-24-lefthook-pinact-design.md`

---

## File Structure

| ファイル | 状態 | 責務 |
|---|---|---|
| `Brewfile` | Modify | `pinact` を `# Linters / formatters` セクションに追記（`actionlint` の直下）|
| `.pinact.yaml` | Create | pinact の設定最小構成（`version: 3` のみ）|
| `lefthook.yml` | Modify | `pre-commit.commands` の YAML/Actions セクションに `pinact` command を追加 |

テストコードは追加しない（ツール設定の導入であり、検証は手動ステップで行う）。

---

## Task 1: 事前検証 — pinact のローカル導入と動作確認

**Files:** なし（ローカル環境の準備のみ）

- [ ] **Step 1.1: pinact をインストール**

Run:

```bash
brew install pinact
```

Expected: インストール成功。

- [ ] **Step 1.2: バージョンと対象 glob の動作を確認**

Run:

```bash
pinact --version
pinact run --help | head -20
```

Expected: `pinact version 3.x.x` が表示される。`run --help` で使えるオプション（`--check`, `--verify` など）が表示される。

- [ ] **Step 1.3: 既存 workflow に対して pinact run を空打ちして差分が出ないことを確認**

Run:

```bash
git status
pinact run .github/workflows/*.yml
git status
```

Expected: `pinact run` 実行前後でワーキングツリーに差分が出ない（既存は全て SHA 化済みのため）。

---

## Task 2: Brewfile に pinact を追加

**Files:**

- Modify: `Brewfile`

- [ ] **Step 2.1: Brewfile の linters セクションに pinact を追加**

`Brewfile` の `# Linters / formatters (referenced by lefthook.yml and CI parity)` コメント直下、`brew "actionlint"` の次の行に `brew "pinact"` を追加する。

変更後の該当セクションは以下の形:

```ruby
# Linters / formatters (referenced by lefthook.yml and CI parity)
brew "yamllint"
brew "actionlint"
brew "pinact"
brew "shellcheck"
brew "shfmt"
brew "golangci-lint"
```

- [ ] **Step 2.2: Brewfile を bundle で検証**

Run:

```bash
brew bundle check --file=Brewfile --verbose
```

Expected: `pinact` を含む全 brew が `installed` と表示される（Task 1.1 で既に導入済みのため）。

- [ ] **Step 2.3: この時点ではまだ commit しない**

`.pinact.yaml` と `lefthook.yml` をまとめて 1 コミットにするため保留する。

---

## Task 3: `.pinact.yaml` を追加

**Files:**

- Create: `.pinact.yaml`

- [ ] **Step 3.1: リポジトリルートに `.pinact.yaml` を作成**

内容:

```yaml
version: 3
```

- [ ] **Step 3.2: pinact が設定を認識することを確認**

Run:

```bash
pinact run --check .github/workflows/*.yml
echo "exit: $?"
```

Expected: `exit: 0`（全 uses が既に SHA 化されているため check が通る）。設定ファイル未検出のエラーや `unknown version` エラーが出ないこと。

---

## Task 4: `lefthook.yml` に pinact command を追加

**Files:**

- Modify: `lefthook.yml`

- [ ] **Step 4.1: `# ── YAML / Actions ──` セクションに pinact ブロックを追加**

`lefthook.yml` の `actionlint:` ブロック直後（48〜54 行目の後ろ）に以下を挿入する。

```yaml
    pinact:
      glob: ".github/{workflows,actions}/**/*.{yml,yaml}"
      run: pinact run {staged_files}
      stage_fixed: true
```

変更後、該当セクション全体は以下の形:

```yaml
    # ── YAML / Actions ──
    yamllint:
      glob: "*.{yml,yaml}"
      run: yamllint --no-warnings {staged_files}
    actionlint:
      glob: ".github/workflows/*.{yml,yaml}"
      run: actionlint {staged_files}
    pinact:
      glob: ".github/{workflows,actions}/**/*.{yml,yaml}"
      run: pinact run {staged_files}
      stage_fixed: true
```

- [ ] **Step 4.2: YAML 構文を lefthook 側で検証**

Run:

```bash
lefthook validate
```

Expected: エラーなく終了。

- [ ] **Step 4.3: 全ファイル対象で pre-commit を dry-run**

Run:

```bash
lefthook run pre-commit --all-files
```

Expected: 全コマンドが成功（`pinact` を含む）。`pinact` が既存 workflow に対して差分を出さない。

---

## Task 5: フック動作の実機検証 — 正常パス

**Files:** なし（検証シナリオ）

- [ ] **Step 5.1: 一時的に workflow を tag 参照に退行させる**

Run:

```bash
cp .github/workflows/ci.yml /tmp/ci.yml.bak
# ci.yml の先頭の checkout だけ tag 参照に書き換える（手動編集）
```

`ci.yml:16` を以下のように一時変更:

- 変更前: `- uses: actions/checkout@93cb6efe18208431cddfb8368fd83d5badbf9bfd # v5.0.2`
- 変更後: `- uses: actions/checkout@v5.0.2`

- [ ] **Step 5.2: 退行させた状態を stage してコミットを試みる**

Run:

```bash
git add .github/workflows/ci.yml
git commit -m "test: verify pinact lefthook integration"
```

Expected:

- lefthook の `pinact` フックが走り、`actions/checkout@v5.0.2` が SHA + コメント形式に書き換えられる
- `stage_fixed: true` により書き換え結果が自動で再 stage される
- コミットが成立する
- コミット後の `.github/workflows/ci.yml:16` が元どおり `<sha> # v5.0.2` 形式に戻っている

- [ ] **Step 5.3: 検証コミットを巻き戻す**

Run:

```bash
git reset --hard HEAD~1
cp /tmp/ci.yml.bak .github/workflows/ci.yml
rm /tmp/ci.yml.bak
git status
```

Expected: ワーキングツリーがクリーン。

---

## Task 6: フック動作の実機検証 — 既存ファイル更新で誤介入しない

**Files:** なし（検証シナリオ）

- [ ] **Step 6.1: 既存 workflow にコメント追加だけの変更をする**

`.github/workflows/ci.yml` 末尾に 1 行コメント（例: `# trigger pinact noop test`）を追加し、stage する。

- [ ] **Step 6.2: コミットしてフックが無害に通ることを確認**

Run:

```bash
git add .github/workflows/ci.yml
git commit -m "test: verify pinact noop on non-uses change"
```

Expected: lefthook が成功し、`pinact` が不要な書き換えを一切加えずコミットが通る。

- [ ] **Step 6.3: 検証コミットを巻き戻す**

Run:

```bash
git reset --hard HEAD~1
git status
```

Expected: ワーキングツリーがクリーン。

---

## Task 7: 3 ファイルをまとめてコミット

**Files:** `Brewfile`, `.pinact.yaml`, `lefthook.yml`（Task 2〜4 で変更済み）

- [ ] **Step 7.1: 現在の差分を確認**

Run:

```bash
git status
git diff Brewfile lefthook.yml
git diff --cached
cat .pinact.yaml
```

Expected:

- `Brewfile` と `lefthook.yml` に意図した追記のみ
- 新規 `.pinact.yaml` が `version: 3` の 1 行（+ 改行）

- [ ] **Step 7.2: 3 ファイルを add してコミット**

Run:

```bash
git add Brewfile .pinact.yaml lefthook.yml
git commit -m "$(cat <<'EOF'
ci: pin github actions via pinact in lefthook pre-commit

Brewfile に pinact を追加し、.pinact.yaml（version: 3）を新設して
lefthook.yml の pre-commit に pinact コマンド（stage_fixed: true）を
追加する。コミット時点で GitHub Actions の uses: 参照が SHA 固定 +
バージョンコメント形式に保たれることを機械的に保証する。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

Expected: lefthook pre-commit 全コマンドが成功し、コミットが成立する。

- [ ] **Step 7.3: コミット内容を確認**

Run:

```bash
git log --oneline -3
git show --stat HEAD
```

Expected: 直近コミットが 3 ファイル変更（`Brewfile` +1 行、`.pinact.yaml` 新規、`lefthook.yml` +4 行）になっている。

---

## Task 8: 全体スモーク検証

**Files:** なし

- [ ] **Step 8.1: `lefthook run pre-commit --all-files` を最終実行**

Run:

```bash
lefthook run pre-commit --all-files
echo "exit: $?"
```

Expected: `exit: 0`。全コマンドが pass し、ワーキングツリーに差分が残らない。

- [ ] **Step 8.2: `make bootstrap` 相当の流れが壊れていないか確認（任意）**

Run:

```bash
brew bundle check --file=Brewfile --verbose
```

Expected: 全 brew が installed 表示。

---

## 成功条件（plan レベル）

- `Brewfile`, `.pinact.yaml`, `lefthook.yml` の 3 ファイル変更が 1 コミットに含まれている
- Task 5（退行→自動修復）および Task 6（無害な変更）が両方期待通りに動作する
- `lefthook run pre-commit --all-files` が exit 0 で終わる
- 既存 lint（actionlint / yamllint / golangci-lint 等）に回帰がない

---

## Self-Review メモ

- Spec の「変更点（3 ファイル）」「動作フロー」「エッジケース」「テスト観点」はすべて Task 2〜8 で網羅している
- placeholder（TBD/TODO等）なし
- `glob`, `stage_fixed`, `version: 3` のキー名・値が Task 間で一貫している
- Task 5 の退行検証に使う `/tmp/ci.yml.bak` のバックアップ→復元手順が明示されている（検証失敗時のリカバリ経路）
