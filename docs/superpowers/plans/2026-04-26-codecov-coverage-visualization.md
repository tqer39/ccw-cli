# Codecov によるテストカバレッジ可視化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Codecov を導入して Go テストカバレッジをバッジ・PR コメント・行単位ヒートマップ・auto しきい値で可視化し、ローカルでも `make coverage-html` 一発で HTML レポートを開けるようにする。

**Architecture:** 既存 `go-test` ジョブが生成している `coverage.out` を `codecov/codecov-action` で Codecov にアップロードする 1 ステップを追加。リポジトリルートの `codecov.yml` で auto しきい値・PR コメント・無視ルールを宣言。Makefile にローカル用 `coverage` / `coverage-html` ターゲットを追加。README 英日両方に Codecov バッジを追加（readme-sync 規約）。Go 本体（`cmd/`, `internal/`）には変更なし。

**Tech Stack:** GitHub Actions, codecov/codecov-action v5, Codecov SaaS, Go 1.25 (`go test -coverprofile`, `go tool cover`), make, yamllint, actionlint, pinact, lefthook.

**Spec:** [docs/superpowers/specs/2026-04-26-codecov-coverage-visualization-design.md](../specs/2026-04-26-codecov-coverage-visualization-design.md)

---

## File Structure

| ファイル | 種別 | 責務 |
|---|---|---|
| `codecov.yml` | 新規 | Codecov の status (project/patch auto)・PR コメントレイアウト・ignore ルールを宣言 |
| `.github/workflows/ci.yml` | 変更 | `go-test` ジョブ末尾に `codecov/codecov-action` ステップを追加 |
| `Makefile` | 変更 | `.PHONY` に `coverage` / `coverage-html` を追加。新規ターゲット 2 件と `clean` の rm 行を更新 |
| `README.md` | 変更 | バッジ列に Codecov バッジを 1 個追加 |
| `docs/README.ja.md` | 変更 | 同上の日本語版（`readme-sync` 規約で英文ソースから同期） |

---

## Task 1: `codecov.yml` を新規作成し yamllint を通す

**Files:**

- Create: `codecov.yml`

- [ ] **Step 1: `codecov.yml` を書く**

リポジトリルートに以下の内容で新規作成:

```yaml
codecov:
  require_ci_to_pass: true

coverage:
  status:
    project:
      default:
        target: auto
        threshold: 1%
    patch:
      default:
        target: auto

comment:
  layout: "reach,diff,flags,files"
  behavior: default
  require_changes: false

ignore:
  - "tests/**"
  - "**/*_test.go"
```

- [ ] **Step 2: yamllint を通す**

```bash
yamllint --no-warnings codecov.yml
```

期待: 何も出力されず exit 0。

エラーが出る場合の典型: `truthy` ルール違反（`yes`/`no` を書いた等）。`.yamllint.yml` は `truthy.allowed-values: [true, false, on]` のため `true` / `false` のみ使うこと。

- [ ] **Step 3: コミット**

```bash
git add codecov.yml
git commit -m "feat(codecov): add codecov.yml with auto thresholds and PR comment layout"
```

---

## Task 2: CI workflow に Codecov アップロードステップを追加

**Files:**

- Modify: `.github/workflows/ci.yml`（`go-test` ジョブ。現状 25-41 行目あたり）

- [ ] **Step 1: `go-test` ジョブの末尾に Codecov ステップを追加**

`.github/workflows/ci.yml` の `go-test` ジョブで、既存の `actions/upload-artifact@...` ステップの **直後** に以下を追加:

```yaml
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v5
        with:
          token: ${{ secrets.CODECOV_TOKEN }}
          files: ./coverage.out
          fail_ci_if_error: false
          verbose: true
```

追加後の `go-test` ジョブ全体は以下のような構造になる（参考、既存部分は変更しない）:

```yaml
  go-test:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
      - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
      - uses: actions/setup-go@4a3601121dd01d1626a1e23e37211e3254c1c06c # v6.4.0
        with:
          go-version-file: go.mod
          cache: true
      - name: Run tests
        run: go test ./... -race -coverprofile=coverage.out
      - uses: actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a # v7.0.1
        with:
          name: coverage
          path: coverage.out
          if-no-files-found: error
          retention-days: 7
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v5
        with:
          token: ${{ secrets.CODECOV_TOKEN }}
          files: ./coverage.out
          fail_ci_if_error: false
          verbose: true
```

注意点:

- `workflow-result` ジョブの `needs` リストは触らない。Codecov ステップは `go-test` の中の最終ステップとして扱う。
- インデントは既存ステップに合わせて半角スペース 6 個（`steps:` の子）。

- [ ] **Step 2: actionlint を通す**

```bash
actionlint .github/workflows/ci.yml
```

期待: 何も出力されず exit 0。

- [ ] **Step 3: pinact で `@v5` を SHA に展開**

```bash
pinact run .github/workflows/ci.yml
```

期待: `codecov/codecov-action@v5` が `codecov/codecov-action@<40桁SHA> # v5.x.y` の形に書き換わる。書き換え結果を `git diff .github/workflows/ci.yml` で確認。

- [ ] **Step 4: yamllint を通す**

```bash
yamllint --no-warnings .github/workflows/ci.yml
```

期待: 何も出力されず exit 0。

- [ ] **Step 5: コミット**

```bash
git add .github/workflows/ci.yml
git commit -m "ci(codecov): upload coverage.out from go-test job"
```

注: lefthook の pre-commit が `actionlint` / `pinact` / `yamllint` を再実行する。Step 2-4 で先に通しておくのは早期失敗のため。

---

## Task 3: Makefile に `coverage` / `coverage-html` ターゲットを追加

**Files:**

- Modify: `Makefile`

- [ ] **Step 1: `.PHONY` 行を更新**

`Makefile` の 1 行目を以下に置換:

変更前:

```makefile
.PHONY: bootstrap build test lint tidy run clean release-check release-snapshot release-clean
```

変更後:

```makefile
.PHONY: bootstrap build test lint tidy run clean release-check release-snapshot release-clean coverage coverage-html
```

- [ ] **Step 2: `coverage` / `coverage-html` ターゲットを追加**

`Makefile` の `release-clean:` ターゲットの直後（ファイル末尾、`bootstrap` の直前）に以下を追加:

```makefile

coverage:
 go test ./... -race -coverprofile=coverage.out
 go tool cover -func=coverage.out | tail -n 1

coverage-html: coverage
 go tool cover -html=coverage.out -o coverage.html
 @echo "open coverage.html"
```

**重要**: Makefile はインデントに **タブ文字** が必須。スペースだとエラーになる。`go test` / `go tool` / `@echo` の各行頭は必ずタブ 1 個。

- [ ] **Step 3: `clean` ターゲットの `rm` 行を更新**

変更前:

```makefile
clean:
 rm -f ccw coverage.out
```

変更後:

```makefile
clean:
 rm -f ccw coverage.out coverage.html
```

- [ ] **Step 4: 動作確認 — `make coverage` が通ること**

```bash
make coverage
```

期待:

- `coverage.out` が生成される
- 最終行に `total: (statements) XX.X%` のようなサマリが表示される
- 終了ステータス 0

- [ ] **Step 5: 動作確認 — `make coverage-html` が `coverage.html` を生成すること**

```bash
make coverage-html
ls -la coverage.html
```

期待:

- `coverage.html` ファイルが存在する（数百 KB の HTML）
- 標準出力末尾に `open coverage.html` が表示される

任意で `open coverage.html`（macOS）/ `xdg-open coverage.html`（Linux）でブラウザで見て、ファイル単位のドリルダウンと未カバー行のハイライト（赤色）が出ていれば成功。

- [ ] **Step 6: 動作確認 — `make clean` が両ファイルを消すこと**

```bash
make clean
ls coverage.out coverage.html 2>&1
```

期待: `ls: coverage.out: No such file or directory` 等が出て、両ファイルが消えていること。

- [ ] **Step 7: コミット**

```bash
git add Makefile
git commit -m "feat(make): add coverage and coverage-html targets"
```

---

## Task 4: README 英日両方に Codecov バッジを追加

**Files:**

- Modify: `README.md`（badge 行、現状 7-11 行目）
- Modify: `docs/README.ja.md`（badge 行、現状 7-11 行目）

- [ ] **Step 1: `README.md` にバッジを追加**

11 行目の `brew-audit` バッジの **直後** に 1 行追加:

変更前:

```markdown
[![brew-audit](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml/badge.svg)](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml)

[🇺🇸 English](README.md) · [🇯🇵 日本語](docs/README.ja.md)
```

変更後:

```markdown
[![brew-audit](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml/badge.svg)](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml)
[![codecov](https://codecov.io/gh/tqer39/ccw-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/tqer39/ccw-cli)

[🇺🇸 English](README.md) · [🇯🇵 日本語](docs/README.ja.md)
```

- [ ] **Step 2: `docs/README.ja.md` にバッジを追加**

同じ位置（11 行目 `brew-audit` の直後）に 1 行追加:

変更前:

```markdown
[![brew-audit](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml/badge.svg)](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml)

[🇺🇸 English](../README.md) · [🇯🇵 日本語](README.ja.md)
```

変更後:

```markdown
[![brew-audit](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml/badge.svg)](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml)
[![codecov](https://codecov.io/gh/tqer39/ccw-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/tqer39/ccw-cli)

[🇺🇸 English](../README.md) · [🇯🇵 日本語](README.ja.md)
```

注: 日本語版のバッジ URL も Codecov の動的 SVG なので英語版と同一でよい（リダイレクト先は同じダッシュボード）。

- [ ] **Step 3: markdownlint を通す**

```bash
npm exec -y --package=markdownlint-cli2 -- markdownlint-cli2 README.md docs/README.ja.md
```

期待: 全ファイル PASS。

- [ ] **Step 4: コミット**

```bash
git add README.md docs/README.ja.md
git commit -m "docs(readme): add Codecov badge to en/ja READMEs"
```

---

## Task 5: PR を作成して Codecov 連携を実地検証

**Files:** なし（GitHub 上での確認）

- [ ] **Step 1: ブランチを push して PR を作成**

```bash
git push -u origin "$(git branch --show-current)"
gh pr create --title "feat(coverage): visualize Go test coverage via Codecov" --body "$(cat <<'EOF'
## Summary

- `codecov/codecov-action` を `go-test` ジョブに追加し、`coverage.out` を Codecov へアップロード
- `codecov.yml` を追加（`status.project: auto + threshold 1%`、`status.patch: auto`、PR コメント有効、`tests/**` と `*_test.go` を ignore）
- Makefile に `coverage` / `coverage-html` ターゲットを追加し、ローカルで HTML レポートを開けるように
- README 英日両方に Codecov バッジを追加

Spec: docs/superpowers/specs/2026-04-26-codecov-coverage-visualization-design.md

## Test plan

- [ ] `go-test` ジョブの "Upload coverage to Codecov" ステップが成功する
- [ ] 既存 `coverage` artifact upload も引き続き成功する
- [ ] PR に Codecov bot のコメントが投稿される（`reach,diff,flags,files` レイアウト）
- [ ] GitHub Checks に `codecov/project` と `codecov/patch` のステータスが現れる
- [ ] ローカルで `make coverage-html` を実行し `coverage.html` が生成される
- [ ] README のバッジ画像が表示される（main マージ後に確定）

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 2: CI 完走を確認**

```bash
gh pr checks
```

期待: `go-test`、`go-build`、`go-lint`、`shellcheck`、`shfmt`、`bats`、`workflow-result` すべてが pass（緑）。

`go-test` のログを開いて "Upload coverage to Codecov" ステップが成功していること（HTTP 200 系のレスポンス、`uploaded` メッセージ等）を確認:

```bash
gh run view --log $(gh run list --branch "$(git branch --show-current)" --limit 1 --json databaseId -q '.[0].databaseId') | grep -A 20 "Upload coverage to Codecov"
```

期待: アップロード成功ログ。失敗の場合でも `fail_ci_if_error: false` のため CI 自体は通っている前提。

- [ ] **Step 3: Codecov ダッシュボード/コメントの確認（手動）**

PR ページを GitHub で開き、以下を目視確認:

1. Codecov bot からのコメントが投稿されている（数分のラグあり）
2. コメントに `reach`, `diff`, `flags`, `files` セクションが含まれる
3. GitHub Checks 一覧に `codecov/project` および `codecov/patch` が出現し、いずれも success（初回はベースラインがないため pass する想定）

`https://codecov.io/gh/tqer39/ccw-cli` にアクセスし、リポジトリのダッシュボードが作られていることを確認。

- [ ] **Step 4: 必要に応じて修正コミット**

何か問題があれば Task 1-4 の対応セクションに戻って修正し、追加コミットを push。Codecov bot のコメントが想定通りでない場合は `codecov.yml` を調整してから再 push。

- [ ] **Step 5: マージ後の確認（PR マージ後に実施）**

main にマージされた後:

1. README の Codecov バッジ画像が表示され、数値（％）が出ていることを確認
2. `https://codecov.io/gh/tqer39/ccw-cli` の dashboard で main の coverage が記録されていることを確認
3. ファイル単位ヒートマップが閲覧できることを確認
