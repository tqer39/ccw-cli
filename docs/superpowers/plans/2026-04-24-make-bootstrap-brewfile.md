# `make bootstrap` + `Brewfile` 実装計画

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** ccw-cli のコントリビューター向けセットアップを `make bootstrap` の 1 コマンドに集約し、必要な CLI ツールを `Brewfile` で宣言的に管理する。

**Architecture:** システムツールは `Brewfile` (Homebrew)、言語ランタイムは `mise.toml` (mise)、Git フックは `lefthook.yml` (lefthook) に分離し、`Makefile` の `bootstrap` ターゲットが順番にオーケストレートする。CI は既存の apt-get / setup-go フローのまま変更しない。

**Tech Stack:** Homebrew (`brew bundle`), mise, lefthook, GNU Make, Go toolchain

**Spec:** [`docs/superpowers/specs/2026-04-24-make-bootstrap-brewfile-design.md`](../specs/2026-04-24-make-bootstrap-brewfile-design.md)

**Branch:** `spec/make-bootstrap-brewfile`（spec コミット `e425011` が先頭）

**Working directory:** リポジトリルート（worktree の場合は `ccw-cli/.claude/worktrees/<name>` がルート）

---

## ファイル構成

| ファイル | 操作 | 責務 |
|---|---|---|
| `Brewfile` | 新規作成 | Homebrew で入れるシステム CLI ツールを宣言 |
| `mise.toml` | 編集 | Go に加えて Node.js (LTS) のバージョン指定を追加 |
| `Makefile` | 編集 | `bootstrap` ターゲット追加（brew 検証 → `brew bundle` → `mise install` → `lefthook install` → `go mod download`） |
| `README.md` | 編集 | Development セクションの pre-commit 手順を `make bootstrap` に置換 |
| `docs/README.ja.md` | 編集 | 同上（日本語版、内容を英語版と同期） |

---

## Task 1: Brewfile を新規作成

**Files:**

- Create: `Brewfile`

- [ ] **Step 1: ルート直下に `Brewfile` を作成**

以下を完全な内容として書き込む。

```ruby
# System tools required to contribute to ccw-cli.
# Run `make bootstrap` to install everything.

# Runtime manager (Go / Node are pinned in mise.toml)
brew "mise"

# Git hooks
brew "lefthook"

# Linters / formatters (referenced by lefthook.yml and CI parity)
brew "yamllint"
brew "actionlint"
brew "shellcheck"
brew "shfmt"
brew "golangci-lint"

# Release / demo tooling
brew "goreleaser"
brew "gh"
brew "vhs"
```

- [ ] **Step 2: `brew bundle` のパースが通ることを検証**

Run: `brew bundle check --file=Brewfile --verbose || true`

Expected: `The Brewfile's dependencies are satisfied.`（既存 install 済みの場合）または `would install X` のような missing 一覧が並ぶ。いずれも Ruby 構文エラーが出なければ OK。`ArgumentError` / `SyntaxError` が出たら該当行を直す。

- [ ] **Step 3: コミット**

```bash
git add Brewfile
git commit -m "chore: add Brewfile listing system tools required for local dev"
```

---

## Task 2: `mise.toml` に Node.js を追加

**Files:**

- Modify: `mise.toml`

- [ ] **Step 1: 現状を確認**

Run: `cat mise.toml`

Expected:

```toml
[tools]
go = "1.25"
```

- [ ] **Step 2: `node = "lts"` を追加**

ファイル全体を次の内容に置き換える（末尾改行を含める）。

```toml
[tools]
go = "1.25"
node = "lts"
```

- [ ] **Step 3: `mise` が新しい設定を解釈できることを検証**

Run: `mise ls --current 2>&1 | head -20`

Expected: `go` と `node` の両方が列挙される（system / installed / missing のいずれかで OK。解釈できずにパースエラーを出さなければ合格）。`mise` 未導入の環境であれば `mise: command not found` が返るので、その場合は `grep -E '^(go|node) =' mise.toml` で行の存在だけ確認する。

- [ ] **Step 4: コミット**

```bash
git add mise.toml
git commit -m "chore(mise): add Node.js LTS for lefthook markdownlint/renovate hooks"
```

---

## Task 3: Makefile に `bootstrap` ターゲットを追加

**Files:**

- Modify: `Makefile`

**重要:** Makefile のレシピ行は **タブ** でインデントする必要がある。スペースに置換されると `Makefile:N: *** missing separator. Stop.` になる。エディタがタブを自動展開する設定になっていないか確認すること。

- [ ] **Step 1: 現状を確認**

Run: `cat Makefile`

Expected:

```make
.PHONY: build test lint tidy run clean release-check release-snapshot release-clean

build:
 go build -o ccw ./cmd/ccw

test:
 go test ./... -race -coverprofile=coverage.out

lint:
 golangci-lint run

tidy:
 go mod tidy

run:
 go run ./cmd/ccw $(ARGS)

clean:
 rm -f ccw coverage.out

release-check:
 goreleaser check

release-snapshot:
 HOMEBREW_TAP_GITHUB_TOKEN=dummy goreleaser release --snapshot --clean --skip=publish

release-clean:
 rm -rf dist/
```

- [ ] **Step 2: `.PHONY` 行に `bootstrap` を追加**

1 行目を次に置換（既存ターゲット列の末尾に `bootstrap` を追加）。

変更前:

```make
.PHONY: build test lint tidy run clean release-check release-snapshot release-clean
```

変更後:

```make
.PHONY: bootstrap build test lint tidy run clean release-check release-snapshot release-clean
```

- [ ] **Step 3: ファイル末尾に `bootstrap` ターゲットを追加**

末尾（`release-clean` の次）に以下を追加（各レシピ行は**タブ**開始）。

```make

bootstrap:
 @command -v brew >/dev/null 2>&1 || { \
   echo "❌ Homebrew is required. See https://brew.sh"; \
   exit 1; \
 }
 brew bundle --file=Brewfile
 mise install
 lefthook install
 go mod download
 @echo "✅ bootstrap complete"
```

- [ ] **Step 4: Makefile の構文を検証**

Run: `make -n bootstrap`

Expected: 実行ログとしてコマンド列（`command -v brew ...` / `brew bundle --file=Brewfile` / `mise install` / `lefthook install` / `go mod download` / `echo "✅ bootstrap complete"` 相当）がエラーなく順に表示される。`missing separator` や `*** No rule to make target` が出たら該当行のタブを修正。

- [ ] **Step 5: brew 不在ケースの挙動を手動検証**

Run: `PATH=/usr/bin:/bin make bootstrap; echo "exit=$?"`

Expected:

```text
❌ Homebrew is required. See https://brew.sh
make: *** [Makefile:N: bootstrap] Error 1
exit=2
```

`exit=0` になった場合は Step 3 の `exit 1` 行が欠落しているので修正。

- [ ] **Step 6: 正常系の挙動を実環境で手動検証**

Run: `make bootstrap`

Expected: `brew bundle` が進行（既インストール分は skip）→ `mise install` が Go / Node を確認/導入 → `lefthook install` が `.git/hooks/pre-commit` を配置 → `go mod download` が成功 → `✅ bootstrap complete` が出力されて終了コード 0。

- [ ] **Step 7: `.git/hooks/pre-commit` の配置を確認**

Run: `ls -l .git/hooks/pre-commit`

Expected: シンボリックリンクまたはスクリプトファイルが存在する（lefthook が配置したもの）。

- [ ] **Step 8: コミット**

```bash
git add Makefile
git commit -m "feat(make): add bootstrap target orchestrating brew/mise/lefthook/go setup"
```

---

## Task 4: README.md (English) を更新

**Files:**

- Modify: `README.md`

- [ ] **Step 1: 現在の該当セクションを確認**

Run: `sed -n '102,118p' README.md`

Expected（`🛠️ Development` 見出しに続いて 3 行の `go test/vet/build` コードブロック、続けて「Pre-commit hooks are managed by lefthook:」の説明文、続けて以下のコードブロック）:

```bash
brew install lefthook yamllint actionlint
lefthook install
```

- [ ] **Step 2: pre-commit セクションを `make bootstrap` 説明に置換**

変更前の該当ブロック（説明文 1 行 + `brew install` / `lefthook install` の 2 行の bash fenced code block）を以下に置き換える。

置換対象の文字列:

```text
Pre-commit hooks are managed by [lefthook](https://github.com/evilmartians/lefthook):
```

のあとに続く `bash` コードブロック（`brew install lefthook yamllint actionlint` と `lefthook install` の 2 行）を含めて丸ごと削除し、次の内容を差し込む:

- 説明文: `Set up the full dev environment (Homebrew required) with:`
- `bash` コードブロック（1 行のみ）: `make bootstrap`
- 説明文: ``This installs the Homebrew packages listed in [`Brewfile`](Brewfile), provisions Go / Node via [`mise`](https://mise.jdx.dev/), and enables [lefthook](https://github.com/evilmartians/lefthook) pre-commit hooks.``

- [ ] **Step 3: 差分を確認**

Run: `git diff README.md`

Expected: 旧ブロック（`brew install lefthook yamllint actionlint` と `lefthook install` の 2 行を含む fenced block）が削除され、新ブロック（`make bootstrap` の 1 行 fenced block と Brewfile / mise 言及の説明文）が追加されている。

- [ ] **Step 4: markdownlint を通す**

Run: `npm exec -y --package=markdownlint-cli2 -- markdownlint-cli2 README.md`

Expected: `Summary: 0 error(s)`。MD040 / MD046 等が出たら fence の言語指定や blank line を調整。

---

## Task 5: `docs/README.ja.md`（日本語）を同期

**Files:**

- Modify: `docs/README.ja.md`

`readme-sync` スキルに従い、英語版と対になる位置を同じ内容で更新する。Task 4 と同じコミットに含める（sync 乖離を防ぐため）。

- [ ] **Step 1: 現在の該当セクションを確認**

Run: `sed -n '102,118p' docs/README.ja.md`

Expected（`🛠️ 開発` 見出しに続いて 3 行の `go test/vet/build` bash コードブロック、続けて「pre-commit は lefthook で管理:」の説明文、続けて以下のコードブロック）:

```bash
brew install lefthook yamllint actionlint
lefthook install
```

- [ ] **Step 2: pre-commit セクションを `make bootstrap` 説明に置換**

変更前の該当ブロック（説明文 1 行 + `brew install` / `lefthook install` の 2 行の bash fenced code block）を以下に置き換える。

置換対象の文字列:

```text
pre-commit は [lefthook](https://github.com/evilmartians/lefthook) で管理:
```

のあとに続く `bash` コードブロック（`brew install lefthook yamllint actionlint` と `lefthook install` の 2 行）を含めて丸ごと削除し、次の内容を差し込む:

- 説明文: `開発環境は 1 コマンドで整います（Homebrew 必須）:`
- `bash` コードブロック（1 行のみ）: `make bootstrap`
- 説明文: ``[`Brewfile`](../Brewfile) の Homebrew パッケージをインストールし、[`mise`](https://mise.jdx.dev/) で Go / Node を用意、[lefthook](https://github.com/evilmartians/lefthook) の pre-commit フックを有効化します。``

- [ ] **Step 3: 差分を確認**

Run: `git diff docs/README.ja.md`

Expected: 英語版と同じ意図の差分（古い 2 行のコードブロックが 1 行 `make bootstrap` に置換、説明文が Brewfile / mise 言及に更新）。

- [ ] **Step 4: markdownlint を通す**

Run: `npm exec -y --package=markdownlint-cli2 -- markdownlint-cli2 docs/README.ja.md`

Expected: `Summary: 0 error(s)`。

- [ ] **Step 5: 英日同期コミット**

```bash
git add README.md docs/README.ja.md
git commit -m "docs: replace manual lefthook setup with make bootstrap (en/ja)"
```

---

## Task 6: 統合検証

**Files:** なし（確認のみ）

- [ ] **Step 1: 既存ターゲットが壊れていないことを確認**

Run: `make build && make test`

Expected: 既存どおり `ccw` バイナリ生成と `go test ./... -race` の成功（既存テストの現状維持）。

- [ ] **Step 2: `make bootstrap` が冪等であることを確認**

Run: `make bootstrap`

Expected: 再実行しても全ステップが成功（`brew bundle` は skip が増えるだけ、`mise install` / `lefthook install` / `go mod download` はいずれも冪等）。

- [ ] **Step 3: lefthook の pre-commit が機能することを確認**

Run: `git commit --allow-empty -m "chore: verify lefthook hook activation"` → 直後に `git reset --soft HEAD~1`

Expected: コミット時に `🥊 lefthook` バナーが出て各チェック（detect-private-key / check-added-large-files など）が走る。走らない場合は `lefthook install --force` を実行。`git reset --soft HEAD~1` で検証用の空コミットを取り消す。

- [ ] **Step 4: コミット不要（検証のみ）**

本タスクはファイル変更を伴わない。失敗があれば該当 Task に戻って修正する。

---

## 完了基準

- [ ] `Brewfile` がルートに存在し、`brew bundle check --file=Brewfile` がパースエラーを出さない
- [ ] `mise.toml` に `go = "1.25"` と `node = "lts"` が両方ある
- [ ] `make bootstrap` が実環境で成功し、5 ステップ（brew チェック → `brew bundle` → `mise install` → `lefthook install` → `go mod download`）を順に実行し完了メッセージで終わる
- [ ] `PATH=/usr/bin:/bin make bootstrap` が `❌ Homebrew is required.` を出して非ゼロ終了する
- [ ] `README.md` と `docs/README.ja.md` の Development セクションが `make bootstrap` に置換され、英日で内容が同期している
- [ ] markdownlint / 既存 `make build` / `make test` が全て通る
- [ ] 新規コミットが `spec/make-bootstrap-brewfile` 上に合計 4 つ積まれている（Task 1 / 2 / 3 / 5。Task 4-5 は 1 コミットにまとめる）

---

## 実行上のメモ

- **作業ブランチ:** `spec/make-bootstrap-brewfile`（既に仕様コミット `e425011` が先頭に存在）
- **コミット粒度:** 各 Task 末尾で 1 コミット（Task 4 と 5 は README 同期のため 1 コミット）
- **PR:** 全 Task 完了後にユーザーに確認してから作成
- **ロールバック:** Task ごとに独立したコミットなので、問題が見つかれば該当コミットを revert するだけで戻せる
