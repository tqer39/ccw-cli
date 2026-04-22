# Phase 0: Repository Quality Infrastructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Go 書き換え着手の前提として、ccw-cli リポジトリに品質基盤（`.gitignore` 更新 / Renovate / lefthook pre-commit / auto-assign workflow）を導入する。

**Architecture:** 設定ファイル追加・更新のみで完結させる。既存の `bin/ccw`（bash）と既存 CI（shellcheck + shfmt + bats）には手を入れない。lefthook は開発者がローカルで `lefthook install` して有効化する（リポジトリ側では強制しない）。Phase 1 以降（Go 化）は別プランで扱う。

**Tech Stack:** Renovate (`github>tqer39/renovate-config`), Lefthook, GitHub Actions (`kentaro-m/auto-assign-action`), gitignore.io (`git:gitignore` skill).

**関連 spec:** `docs/superpowers/specs/2026-04-23-go-rewrite-and-brew-design.md` §「リポジトリ品質基盤」

---

## File Structure

本プランで触るファイル:

- Create: `renovate.json5` — Renovate 設定（extends のみ）
- Create: `lefthook.yml` — pre-commit フック定義
- Create: `.github/auto_assign.yml` — auto-assign-action の動作設定
- Create: `.github/workflows/auto-assign.yml` — PR open / ready_for_review で assign
- Modify: `.gitignore` — Go / goreleaser / coverage / lefthook を追加
- Modify: `README.md` — lefthook インストール手順 + 前提ツール一覧を追記

触らないファイル: `bin/ccw`, `tests/*.bats`, `.github/workflows/ci.yml`（あれば）, `.editorconfig`。

---

## Task 1: `.gitignore` を Go / goreleaser 向けに更新

**Files:**
- Modify: `.gitignore`

- [ ] **Step 1: 現状の `.gitignore` を確認**

Run:
```bash
cat .gitignore
```

期待（現状）:
```
# OS
.DS_Store
Thumbs.db

# editors
.vscode/
.idea/
*.swp

# logs / tmp
*.log
tmp/
.tmp/
```

- [ ] **Step 2: `git:gitignore` skill を起動**

ユーザー明示指示「`/gitignore` を使って更新」に従う。

Skill 呼び出し:
- skill: `git:gitignore`
- args: `Go macOS Linux VisualStudioCode Emacs`

スキルは gitignore.io から該当テンプレートを取得し、既存 `.gitignore` にマージする（重複排除して末尾追記）。

- [ ] **Step 3: プロジェクト固有のエントリを手動追記**

スキルが入れない ccw-cli 固有のビルド成果物を追記する。

`.gitignore` の末尾に以下ブロックを付け足す:

```
# goreleaser / local build
/dist/
/ccw
/ccw.exe

# Go coverage
coverage.out
coverage.html
*.prof

# lefthook local overrides
lefthook-local.yml
.lefthook-local/
```

- [ ] **Step 4: 追加行が実際に ignore されるか検証**

Run:
```bash
mkdir -p dist && touch dist/.keep coverage.out ccw
git check-ignore -v dist/.keep coverage.out ccw
rm -rf dist coverage.out ccw
```

期待: 各パスが `.gitignore:<lineno>:<pattern>\t<path>` 形式で出力される。出力が空ならパターンが効いていないので Step 3 を再確認。

- [ ] **Step 5: コミット**

Run:
```bash
git add .gitignore
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
🔧 .gitignore: Go / goreleaser / coverage / lefthook を追加

Phase 0 品質基盤整備の一環として、Go バイナリ成果物・goreleaser の
dist/ ディレクトリ・Go coverage ファイル・lefthook のローカル設定を
ignore 対象に追加。gitignore.io からは Go / macOS / Linux /
VSCode / Emacs のテンプレートを取り込んだ。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Renovate 設定を追加

**Files:**
- Create: `renovate.json5`

- [ ] **Step 1: `renovate.json5` を作成**

作成内容（完全体）:
```json5
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": [
    "github>tqer39/renovate-config"
  ]
}
```

ルールは `tqer39/renovate-config` 側で一元管理。ccw-cli 側は extends のみ。

- [ ] **Step 2: Renovate 設定の構文検証**

Run:
```bash
npx --yes --package=renovate -- renovate-config-validator renovate.json5
```

期待出力: 末尾に `Config validated successfully` もしくは同等の成功メッセージ。ネットワーク環境により初回 npm fetch に時間がかかる点に留意。

npm が無い場合は先に `brew install node` で入れておく。

- [ ] **Step 3: コミット**

Run:
```bash
git add renovate.json5
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
🔧 Renovate: tqer39/renovate-config を有効化

共有 Renovate 設定 (github>tqer39/renovate-config) を extends する
最小構成を追加。Go モジュール・GitHub Actions の依存ピンを
自動 PR 化できるようになる。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: `lefthook.yml` を追加（pre-commit フック）

**Files:**
- Create: `lefthook.yml`

### 前提

- lefthook 本体はローカルに別途インストール必要（`brew install lefthook`）
- 一部 hook は外部ツールを要求: `shellcheck`, `shfmt`, `yamllint`, `actionlint`, `gofmt`（Go 本体に付属）, `golangci-lint`。`markdownlint-cli2` / `renovate-config-validator` は `npx --yes` で都度取得するので事前 install 不要
- ccw-cli は現状 Go ファイルを持たないが、Phase 1 以降で追加されるため glob ベースで先行定義しておく（該当ファイル不在時は lefthook が skip する）

### 採用ポリシー

参考 `terraform-github/lefthook.yml` から Go プロジェクトに必要な hook のみ抽出:
- 安全系: `check-added-large-files`, `detect-private-key`（inline 実装。外部スクリプト不要）
- Go: `gofmt`, `golangci-lint`
- Shell: `shellcheck`, `shfmt`（既存 `bin/ccw` / `tests/*.bats` をカバー）
- YAML / Actions: `yamllint`, `actionlint`
- Markdown: `markdownlint-cli2`
- Renovate: `renovate-config-validator`

除外: `cspell`, `textlint`（辞書整備が Phase 0 の範囲を超える）, `terraform-fmt`, `biome-format`（対象ファイルなし）, `detect-aws-credentials`（ccw-cli スコープ外）。`trim-trailing-whitespace` / `end-of-file-fixer` は `.editorconfig` + 各 linter でカバーされるため省略。

- [ ] **Step 1: `lefthook.yml` を作成**

作成内容（完全体）:

```yaml
# Lefthook configuration.
# See https://lefthook.dev/configuration/ for details.
pre-commit:
  parallel: true
  commands:
    # ── セーフティ系 ──
    check-added-large-files:
      run: |
        fail=0
        for f in {staged_files}; do
          [ -f "$f" ] || continue
          sz=$(wc -c <"$f" 2>/dev/null || echo 0)
          if [ "$sz" -gt 524288 ]; then
            printf '❌ %s exceeds 512KB (%s bytes)\n' "$f" "$sz" >&2
            fail=1
          fi
        done
        exit "$fail"
    detect-private-key:
      run: |
        hits=$(grep -lE '-----BEGIN [A-Z ]*PRIVATE KEY-----' {staged_files} 2>/dev/null || true)
        if [ -n "$hits" ]; then
          printf '❌ private key detected in:\n%s\n' "$hits" >&2
          exit 1
        fi

    # ── Go ──
    gofmt:
      glob: "*.go"
      run: |
        out=$(gofmt -l {staged_files})
        if [ -n "$out" ]; then
          printf '❌ gofmt violations:\n%s\nRun: gofmt -w {staged_files}\n' "$out" >&2
          exit 1
        fi
    golangci-lint:
      glob: "*.go"
      run: golangci-lint run {staged_files}

    # ── Shell ──
    shellcheck:
      glob: "{*.sh,*.bash,*.bats,bin/ccw}"
      run: shellcheck {staged_files}
    shfmt:
      glob: "{*.sh,*.bash,bin/ccw}"
      run: shfmt -d -i 2 -ci -bn {staged_files}

    # ── YAML / Actions ──
    yamllint:
      glob: "*.{yml,yaml}"
      run: yamllint --no-warnings {staged_files}
    actionlint:
      glob: ".github/workflows/*.{yml,yaml}"
      run: actionlint {staged_files}

    # ── Markdown ──
    markdownlint:
      glob: "*.{md,markdown}"
      run: npx --yes markdownlint-cli2 --fix {staged_files}
      stage_fixed: true

    # ── Renovate ──
    renovate-config-validator:
      glob: "renovate.json5"
      run: npx --yes --package=renovate -- renovate-config-validator {staged_files}

# commit-msg / pre-push は現状不要。必要になったらここへ追加。
```

- [ ] **Step 2: lefthook をインストール & 有効化**

Run:
```bash
command -v lefthook >/dev/null 2>&1 || brew install lefthook
lefthook install
```

期待: `.git/hooks/pre-commit` が作成される（lefthook 経由の shim）。

- [ ] **Step 3: lefthook 設定の YAML 構文検証**

Run:
```bash
lefthook dump
```

期待: YAML の全 command が展開されて出力され、非 0 終了しない。

- [ ] **Step 4: 手動で pre-commit を試走（ダミーコミット作成）**

Run:
```bash
# 既存 bin/ccw に空行 touch して差分を作る (直後に revert)
printf '' >> bin/ccw
git add bin/ccw
lefthook run pre-commit
git restore --staged bin/ccw
git checkout -- bin/ccw
```

期待: `shellcheck` / `shfmt` / `check-added-large-files` / `detect-private-key` 等が並列実行され全て PASS。Go 関連は該当ファイルなしで skip。

失敗する場合:
- `shellcheck: command not found` → `brew install shellcheck`
- `shfmt: command not found` → `brew install shfmt`
- `yamllint: command not found` → `brew install yamllint`
- `actionlint: command not found` → `brew install actionlint`

- [ ] **Step 5: コミット**

Run:
```bash
git add lefthook.yml
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
🔧 lefthook: pre-commit フックを導入

bin/ccw (bash) と今後追加予定の Go コード両方をカバーする
lefthook.yml を追加。Phase 0 では設定ファイルのみ導入し、
開発者は \`brew install lefthook && lefthook install\` で有効化する。

採用 hook: check-added-large-files, detect-private-key,
gofmt, golangci-lint, shellcheck, shfmt, yamllint, actionlint,
markdownlint-cli2, renovate-config-validator。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: auto-assign 設定とワークフローを追加

**Files:**
- Create: `.github/auto_assign.yml`
- Create: `.github/workflows/auto-assign.yml`

- [ ] **Step 1: `.github/` ディレクトリを確認**

Run:
```bash
ls -la .github/ .github/workflows/ 2>/dev/null || true
```

`.github/workflows/` は既存想定（CI 用）。無ければこの Task で作成される。

- [ ] **Step 2: `.github/auto_assign.yml` を作成**

作成内容（完全体、参考 `terraform-github/.github/auto_assign.yml` と同一）:

```yaml
# see: https://github.com/kentaro-m/auto-assign-action
addAssignees: author
```

- [ ] **Step 3: `.github/workflows/auto-assign.yml` を作成**

作成内容（完全体、参考 `terraform-github` をコピーして SHA ピンもそのまま採用。Renovate が以降の更新を担う）:

```yaml
name: Auto Assign

on:
  pull_request:
    types: [opened, ready_for_review]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  add-reviews:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    timeout-minutes: 5
    steps:
      - name: Auto Assign
        uses: kentaro-m/auto-assign-action@0a2c53d3721e4c5cfcce6afd4014aecc337979f6 # v2.0.2
        if: ${{ github.event.pull_request.assignee == null && join(github.event.pull_request.assignees) == '' }}
        with:
          configuration-path: .github/auto_assign.yml
```

- [ ] **Step 4: actionlint で workflow を検証**

Run:
```bash
actionlint .github/workflows/auto-assign.yml
```

期待: 出力なし（= エラーなし）。`actionlint` が無ければ `brew install actionlint`。

- [ ] **Step 5: auto_assign 設定側の YAML 構文検証**

Run:
```bash
yamllint --no-warnings .github/auto_assign.yml
```

期待: 出力なし。

- [ ] **Step 6: コミット**

Run:
```bash
git add .github/auto_assign.yml .github/workflows/auto-assign.yml
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
👷 CI: auto-assign ワークフローを追加

kentaro-m/auto-assign-action を使い、PR の opened /
ready_for_review イベントで author を assignee に自動設定。
SHA ピンは Renovate が以降のバージョン追従を担う。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: README に lefthook インストール手順を追記

**Files:**
- Modify: `README.md`

- [ ] **Step 1: 挿入位置を確認**

Run:
```bash
grep -n '^## ' README.md
```

`## Future work` と `## License` の直前に `## Development` セクションを差し込む。

- [ ] **Step 2: `## Development` セクションを追加**

`README.md` の `## Future work` 直前に以下ブロックを挿入する。本プラン文書上のフェンス衝突を避けるため外側を `~~~` で囲って提示する。README 本体には `~~~` を除いた中身（先頭の `## Development` から末尾のリストまで）を貼り付けること。

~~~markdown
## Development

### Prerequisites

ローカルで lefthook pre-commit フックを有効化するため、以下を事前に用意してください。

```bash
brew install lefthook shellcheck shfmt yamllint actionlint
lefthook install
```

### Hooks

- `check-added-large-files`: 512KB 超のファイルをブロック
- `detect-private-key`: 秘密鍵の混入を検出
- `gofmt` / `golangci-lint`: Go ファイル対象（Phase 1 以降で有効化）
- `shellcheck` / `shfmt`: bash スクリプト対象（`bin/ccw` / `tests/*.bats`）
- `yamllint` / `actionlint`: YAML / GitHub Actions ワークフロー対象
- `markdownlint-cli2`: Markdown 対象（`npx` で都度取得）
- `renovate-config-validator`: `renovate.json5` のみ対象
~~~

- [ ] **Step 3: markdownlint-cli2 で整形検査**

Run:
```bash
npx --yes markdownlint-cli2 README.md
```

期待: 既存の lint 違反がなければ追加ブロックも PASS。既存違反がある場合は `--fix` 付きで再実行して差分確認。

- [ ] **Step 4: コミット**

Run:
```bash
git add README.md
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
📝 README: Development セクションを追加

lefthook インストール手順と pre-commit hook の一覧を明記。
コントリビュータが brew で依存ツールを揃え lefthook install
するだけで開発環境が整う状態を目指す。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Final Verification

全タスク完了後、リポジトリ全体での整合を確認する。

- [ ] **Step 1: 全 hook の drybrun**

Run:
```bash
git ls-files | xargs -I{} sh -c '[ -f "{}" ] && echo "{}"' | head -100 > /tmp/ccw-all-files.txt
# 代表ファイルを stage 相当にして pre-commit 全体を試走
git add -N lefthook.yml renovate.json5 README.md .github/auto_assign.yml .github/workflows/auto-assign.yml .gitignore
lefthook run pre-commit --all-files
```

期待: 全 hook が PASS。

- [ ] **Step 2: git log で 5 コミットが順に並ぶことを確認**

Run:
```bash
git log --oneline -6
```

期待（新しい順）:
```
<sha> 📝 README: Development セクションを追加
<sha> 👷 CI: auto-assign ワークフローを追加
<sha> 🔧 lefthook: pre-commit フックを導入
<sha> 🔧 Renovate: tqer39/renovate-config を有効化
<sha> 🔧 .gitignore: Go / goreleaser / coverage / lefthook を追加
<sha> (直前の既存コミット)
```

- [ ] **Step 3: PR 作成 (オプション)**

本プランは main から直接の変更ではなく、worktree 上で進めている前提。完了したら:

```bash
git push -u origin HEAD
gh pr create --title "Phase 0: repository quality infrastructure" --body "$(cat <<'EOF'
## Summary
- `.gitignore` を Go / goreleaser 向けに更新
- Renovate 設定 (`tqer39/renovate-config` extends) を追加
- lefthook で pre-commit フックを導入 (bin/ccw + 今後の Go コード両対応)
- auto-assign ワークフローで PR 作成時に author を assignee に設定
- README に Development セクションを追加

## Test plan
- [ ] `lefthook install` + `lefthook run pre-commit --all-files` で全 hook PASS
- [ ] `npx --yes --package=renovate -- renovate-config-validator renovate.json5` が成功
- [ ] `actionlint .github/workflows/auto-assign.yml` が成功
- [ ] PR 作成後、auto-assign workflow が起動し author が assignee に追加される

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## 次のプラン

本プラン完了後、以下のプランを順次作成予定（本プランの範囲外）:

1. **Phase 1 plan**: `go.mod` 初期化 + `cmd/ccw/main.go` スケルトン + `internal/cli` + `internal/version` + `internal/ui`
2. **Phase 2 plan**: `internal/gitx` + `internal/worktree` + `internal/claude` + `internal/superpowers`
3. **Phase 3 plan**: `internal/picker`（bubbletea）+ teatest
4. **Phase 4 plan**: `.goreleaser.yaml` + release workflow + `tqer39/homebrew-tap` 初期化
5. **Phase 5 plan**: `v0.1.0` タグ → brew install 検証 → README 更新
