# cspell を lefthook + CI に導入する Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `lefthook.yml` の pre-commit と GitHub Actions の `ci.yml` の両方で cspell による綴りチェックを実行できるようにする。

**Architecture:** `.cspell/cspell.json` を中心に設定を集約し、判定（除外含む）は cspell 側に一元化する。lefthook では `{staged_files}` を素のまま渡し、CI では `**/*` を全件スキャンする。cspell 本体は `npm exec -y --package=cspell` 方式で都度取得し、Brewfile/mise を増やさない。

**Tech Stack:** cspell (npm), lefthook, GitHub Actions, pinact (action SHA pinning)

**Spec:** `docs/superpowers/specs/2026-04-26-cspell-lefthook-design.md`

**Branch:** `feat/cspell-lefthook`（spec コミット済み）

---

## File Structure

| パス | 役割 |
|---|---|
| `.cspell/cspell.json` | cspell 設定。`ignorePaths` でバイナリ・superpowers 関連・ノイズ源を除外 |
| `.cspell/project-words.txt` | プロジェクト固有語の辞書。1 行 1 語 |
| `.cspell/README.md` | 偽陽性が出た際の運用手順（追記してコミット） |
| `lefthook.yml` | `pre-commit.commands.cspell` を追加 |
| `.github/workflows/ci.yml` | `cspell` ジョブと `workflow-result` の依存・チェックを追加 |

---

## Task 1: `.cspell/cspell.json` を作成

**Files:**

- Create: `.cspell/cspell.json`

- [ ] **Step 1: ディレクトリと設定ファイルを作成**

`.cspell/cspell.json` を以下の内容で作成する。

```json
{
  "$schema": "https://raw.githubusercontent.com/streetsidesoftware/cspell/main/cspell.schema.json",
  "version": "0.2",
  "language": "en",
  "dictionaryDefinitions": [
    {
      "name": "project-words",
      "path": "./project-words.txt",
      "addWords": true
    }
  ],
  "dictionaries": ["project-words"],
  "ignorePaths": [
    ".git/**",
    "node_modules/**",
    "dist/**",
    "vendor/**",
    "coverage.*",
    "*.prof",
    "go.sum",
    "go.mod",
    ".cspell/**",
    ".claude/**",
    "docs/superpowers/**",
    "internal/superpowers/preamble_*.txt",
    "tests/fixtures/**",
    ".goreleaser.yaml",
    "Formula/**",
    "*.png",
    "*.jpg",
    "*.jpeg",
    "*.gif",
    "*.ico",
    "*.tape",
    "*.gz",
    "*.zip",
    "*.exe",
    "*.dll",
    "*.so",
    "*.dylib"
  ]
}
```

- [ ] **Step 2: JSON 構文確認**

Run: `node -e "JSON.parse(require('fs').readFileSync('.cspell/cspell.json','utf8')); console.log('ok')"`
Expected: `ok`

- [ ] **Step 3: コミットしない（次タスクと一緒にコミットする）**

辞書ファイルが無いと cspell は警告するので、Task 2 を経てから commit する。

---

## Task 2: `.cspell/project-words.txt` を作成（初期辞書）

**Files:**

- Create: `.cspell/project-words.txt`

- [ ] **Step 1: 初期辞書ファイルを作成**

`.cspell/project-words.txt` を以下の内容で作成する。1 行 1 語、空行や末尾改行 OK。

```text
ccw
tqer
goreleaser
golangci
shfmt
shellcheck
yamllint
actionlint
pinact
lefthook
markdownlint
renovate
codecov
charmbracelet
bubbletea
teatest
homebrew
brewfile
mise
goimports
preamble
superpowers
worktrees
worktree
kakehashi
ooyama
takeru
```

- [ ] **Step 2: cspell が走ることを単体確認**

`cspell` を一度走らせて起動できるか確認する（ヒット 0 / 1 件以上どちらでもよい）。

Run:

```bash
npm exec -y --package=cspell -- cspell lint \
  --no-progress --config .cspell/cspell.json \
  "README.md"
```

Expected: 終了ステータス 0 か非 0。**起動エラー（"unknown option" / "config not found" 等）が出ないこと**が成功条件。

---

## Task 3: `.cspell/README.md` を作成

**Files:**

- Create: `.cspell/README.md`

- [ ] **Step 1: 運用手順を書く**

`.cspell/README.md` を以下の内容で作成する。

````markdown
# cspell 設定

このディレクトリは pre-commit / CI で動作する `cspell`（綴りチェッカ）の設定一式です。

## ファイル

- `cspell.json` — 本体設定。除外パス（バイナリ・superpowers 関連・ノイズ源）を `ignorePaths` で集約管理しています。
- `project-words.txt` — プロジェクト固有語の辞書。1 行 1 語。

## 偽陽性が出たとき

cspell が正当な語をエラー扱いした場合、その語を `project-words.txt` に 1 行 1 語で追記してコミットしてください。

```bash
echo "yourword" >> .cspell/project-words.txt
git add .cspell/project-words.txt
git commit -m "chore(cspell): add yourword to dictionary"
```

## ローカル全件チェック

```bash
npm exec -y --package=cspell -- cspell lint \
  --no-progress --config .cspell/cspell.json \
  "**/*"
```
````

- [ ] **Step 2: 3 ファイルまとめてコミット**

```bash
git add .cspell/cspell.json .cspell/project-words.txt .cspell/README.md
git commit -m "chore(cspell): add base configuration and project dictionary

- .cspell/cspell.json: ignorePaths でバイナリ・superpowers 関連・
  Formula / .goreleaser.yaml などのノイズ源を除外
- .cspell/project-words.txt: プロジェクト固有語の初期辞書
- .cspell/README.md: 偽陽性が出たときの運用手順"
```

---

## Task 4: 全件スキャンで辞書を確定する

**Files:**

- Modify: `.cspell/project-words.txt`

cspell を全ファイル対象で走らせ、出てきた未知語のうち**正当なもの**を辞書に追加する。誤字・タイポは別タスクで（あるいは現状維持で）対処。

- [ ] **Step 1: 全件スキャン**

Run:

```bash
npm exec -y --package=cspell -- cspell lint \
  --no-progress --config .cspell/cspell.json \
  --words-only --unique \
  "**/*" 2>/dev/null | sort -u > /tmp/cspell-unknown.txt
wc -l /tmp/cspell-unknown.txt
```

Expected: 候補語の一覧（行数 = 未知語の種類数）。

- [ ] **Step 2: 候補語を確認**

Run: `cat /tmp/cspell-unknown.txt`

各語について以下を判定:

- **正当な固有名詞・略語・コード由来語** → `.cspell/project-words.txt` に追記
- **明らかなタイポ** → 修正は本タスク外。残しておくと CI が落ちるので、当面は辞書に追加し、後追いで修正タスクを切る（もしくは即座に修正コミットを別途立てる）
- **判断つかないもの** → 一旦辞書に入れる方針（保守的）

- [ ] **Step 3: 辞書更新**

`.cspell/project-words.txt` を編集して必要な語を追記。アルファベット順を強制する必要はないが、既存順序を崩しすぎないこと。

- [ ] **Step 4: 再スキャンで 0 件を確認**

Run:

```bash
npm exec -y --package=cspell -- cspell lint \
  --no-progress --config .cspell/cspell.json \
  "**/*"
echo "exit=$?"
```

Expected: `exit=0`

0 件にならない場合、Step 2-3 を繰り返す。

- [ ] **Step 5: コミット**

```bash
git add .cspell/project-words.txt
git commit -m "chore(cspell): expand project dictionary to cover repo content"
```

---

## Task 5: `lefthook.yml` に cspell コマンドを追加

**Files:**

- Modify: `lefthook.yml`

- [ ] **Step 1: 追加位置を確認**

`lefthook.yml` の `pre-commit.commands` 配下、`renovate-config-validator` の **下**に追加する。

- [ ] **Step 2: cspell コマンドを追加**

`lefthook.yml` の最後の `renovate-config-validator` ブロックの**直後**（行 71 の直後、コメント `# commit-msg / pre-push は現状不要。` の前）に以下を挿入する。

```yaml
    # ── Spell check ──
    cspell:
      run: |
        npm exec -y --package=cspell -- cspell lint \
          --no-progress --no-summary --no-must-find-files \
          --config .cspell/cspell.json \
          {staged_files}
```

挿入後、`lefthook.yml` の該当付近は以下のようになる。

```yaml
    # ── Renovate ──
    renovate-config-validator:
      glob: "renovate.json5"
      run: npm exec -y --package=renovate -- renovate-config-validator {staged_files}

    # ── Spell check ──
    cspell:
      run: |
        npm exec -y --package=cspell -- cspell lint \
          --no-progress --no-summary --no-must-find-files \
          --config .cspell/cspell.json \
          {staged_files}

# commit-msg / pre-push は現状不要。必要になったらここへ追加。
```

- [ ] **Step 3: yamllint 単体で検証**

Run: `yamllint --no-warnings lefthook.yml`
Expected: エラーなし、出力なし。

- [ ] **Step 4: lefthook を実行**

`README.md` のような既存ファイルだけ stage してフックが通るか確認する。

Run:

```bash
git add lefthook.yml
lefthook run pre-commit
```

Expected:

- `cspell` が `(skip) no files for inspection` ではなく実際に走り、エラーなし。
- 他のリンタも green か skip。

注: lefthook は staged_files を見るので、`git add lefthook.yml` 済みの状態で `cspell` が `lefthook.yml` をチェックする。0 件で抜けるはず。

- [ ] **Step 5: コミット**

```bash
git commit -m "chore(lefthook): add cspell pre-commit hook"
```

---

## Task 6: CI ジョブを追加（pinact 解決込み）

**Files:**

- Modify: `.github/workflows/ci.yml`

CI に新規ジョブ `cspell` を追加し、`workflow-result` の依存と判定にも組み込む。`actions/setup-node` は仮 SHA で書き、`pinact run` で正しい SHA に解決させる。

- [ ] **Step 1: cspell ジョブを `bats` ジョブの直後に追加**

`.github/workflows/ci.yml` の行 103（`bats` ジョブの最終行 `run: bats tests/`）の直後、`workflow-result:` の前に以下を挿入する。

```yaml
  cspell:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
      - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
      - uses: actions/setup-node@v4
        with:
          node-version: 'lts/*'
      - name: Run cspell
        run: |
          npm exec -y --package=cspell -- cspell lint \
            --no-progress --config .cspell/cspell.json \
            "**/*"
```

注意: `actions/setup-node@v4` は **仮タグ**。Step 4 で pinact が SHA に解決する。

- [ ] **Step 2: `workflow-result` の `needs` と判定に `cspell` を追加**

行 106 の `needs` 配列を以下に変更:

変更前:

```yaml
    needs: [go-lint, go-test, go-build, shellcheck, shfmt, bats]
```

変更後:

```yaml
    needs: [go-lint, go-test, go-build, shellcheck, shfmt, bats, cspell]
```

さらに `bats` の result チェック直後（`echo "All jobs completed successfully"` の前）に以下のチェックを追加:

```yaml
          if [ "${{ needs.cspell.result }}" != "success" ]; then
            echo "cspell failed: ${{ needs.cspell.result }}"
            exit 1
          fi
```

- [ ] **Step 3: yamllint と actionlint で単体検証**

Run:

```bash
yamllint --no-warnings .github/workflows/ci.yml
actionlint .github/workflows/ci.yml
```

Expected: エラーなし、出力なし。

- [ ] **Step 4: pinact で SHA を解決**

Run:

```bash
pinact run .github/workflows/ci.yml
git diff .github/workflows/ci.yml
```

Expected: `actions/setup-node@v4` が `actions/setup-node@<40 桁 SHA> # v4.x.x` 形式に書き換わる。

- [ ] **Step 5: 再度 actionlint で検証**

Run: `actionlint .github/workflows/ci.yml`
Expected: エラーなし。

- [ ] **Step 6: lefthook で関連 hook を全部走らせる**

Run:

```bash
git add .github/workflows/ci.yml
lefthook run pre-commit
```

Expected: yamllint / actionlint / pinact / cspell すべて green。

- [ ] **Step 7: コミット**

```bash
git commit -m "ci: add cspell job and include in workflow-result gate"
```

---

## Task 7: 動作確認（pre-commit と CI 両方）

- [ ] **Step 1: ローカル全件チェック**

Run:

```bash
npm exec -y --package=cspell -- cspell lint \
  --no-progress --config .cspell/cspell.json \
  "**/*"
echo "exit=$?"
```

Expected: `exit=0`

非 0 の場合、未知語が残っている。Task 4 の Step 2-4 を再度実行して辞書を整備する。

- [ ] **Step 2: lefthook 全 hook をローカル実行**

Run: `lefthook run pre-commit --all-files`
Expected: 全て green（または skip）。`cspell` が 0 件で通ること。

- [ ] **Step 3: PR を作成して CI 確認**

```bash
git push -u origin feat/cspell-lefthook
gh pr create --title "feat: cspell を lefthook + CI に導入" --body "$(cat <<'EOF'
## Summary
- pre-commit と CI で cspell による綴りチェックを実行
- 設定は `.cspell/cspell.json` に集約、辞書は `.cspell/project-words.txt`
- バイナリと superpowers 関連 (`internal/superpowers/preamble_*.txt`, `docs/superpowers/**`, `.claude/**`) は除外

設計ドキュメント: `docs/superpowers/specs/2026-04-26-cspell-lefthook-design.md`

## Test plan
- [x] `lefthook run pre-commit --all-files` がローカルで green
- [x] `cspell lint --config .cspell/cspell.json "**/*"` がローカルで 0 件
- [ ] CI の `cspell` ジョブが green
- [ ] CI の `workflow-result` が green

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Expected: PR URL が表示される。

- [ ] **Step 4: CI green を待つ**

Run: `gh pr checks --watch`
Expected: 全 check が success、特に `cspell` ジョブと `workflow-result`。

赤になった場合は、ログを `gh run view --log-failed` で確認し、未知語が出ていれば Task 4 と同様に辞書追加でコミットする。

---

## Self-Review メモ

- [x] Spec の決定事項表 6 項目すべてに対応するタスクがある
- [x] `.cspell/cspell.json` の `ignorePaths` は spec と一致
- [x] `lefthook.yml` の追加コマンドは spec のサンプルと一致
- [x] CI ジョブの構造（setup-node + npm exec）は spec のサンプルと一致
- [x] `workflow-result` の更新も含めている（spec で明記）
- [x] `actions/setup-node` の SHA は pinact が解決する手順を明示
- [x] プレースホルダ語句（TBD/TODO）なし
- [x] 各 Task に正確なファイルパスとコマンド、期待出力を記載
