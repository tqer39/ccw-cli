# `make bootstrap` + `Brewfile` による開発環境セットアップ集約

- 作成日: 2026-04-24
- ステータス: ドラフト
- 対象: ccw-cli コントリビューター向けの開発環境セットアップ

## 背景

ccw-cli のコントリビューター向け手順は現在 README の Development セクションに散らばっており、以下を手で実行する必要がある。

```bash
brew install lefthook yamllint actionlint
lefthook install
```

一方で `lefthook.yml` は README に書かれていない `shellcheck` / `shfmt` / `golangci-lint` / `markdownlint-cli2`（npm 経由）/ `renovate-config-validator`（npm 経由）にも依存しており、README と実体が乖離している。リリース/デモ用途では `goreleaser` / `gh` / `vhs` も必要。

セットアップの入口を **`make bootstrap` の 1 コマンド**に集約し、必要ツールを **`Brewfile`** で宣言的に管理することで、乖離を解消し新規コントリビューターの導入を単純化する。

## ゴール

- `make bootstrap` を実行すれば pre-commit / lint / build / release / デモ生成に必要なローカル環境がすべて整う
- Brewfile を見れば「このリポジトリが必要とするシステムツール」が一目でわかる
- README の Development セクションは `make bootstrap` 一行に縮約される（詳細は Brewfile / Makefile を正とする）
- 既存の CI には影響を与えない（CI は Ubuntu + apt-get / setup-go を使い続ける）

## 非ゴール

- CI の Brewfile 化（CI は独自導線で高速）
- Windows / Linux 非 brew 環境のサポート（将来 Roadmap で検討、今回は macOS / Linuxbrew 前提）
- 初回セットアップ以外の更新系（`brew bundle --upgrade` 等）の自動化

## 責務分離

| レイヤー | ファイル | 責務 |
|---|---|---|
| システムツール | `Brewfile` | CLI バイナリ群（lefthook, linter, goreleaser 等） |
| 言語ランタイム | `mise.toml` | Go / Node のバージョン固定 |
| Git フック | `lefthook.yml` | pre-commit で走るチェック定義（既存） |
| オーケストレーション | `Makefile` | 上記 3 つを正しい順序で呼ぶ `bootstrap` ターゲット |

## 設計

### Brewfile（新規）

ルート直下に以下を配置する。

```ruby
# Runtime manager
brew "mise"

# Git hooks
brew "lefthook"

# Linters / formatters
brew "yamllint"
brew "actionlint"
brew "shellcheck"
brew "shfmt"
brew "golangci-lint"

# Release / demo
brew "goreleaser"
brew "gh"
brew "vhs"
```

`markdownlint-cli2` と `renovate-config-validator` は `npm exec -y` 経由で実行されるため Node.js ランタイムがあれば足りる（ローカル install 不要）。Node.js は mise 管理。

### `mise.toml`（編集）

```toml
[tools]
go = "1.25"
node = "lts"
```

`node = "lts"` を追加。lefthook の markdownlint / renovate-config-validator 呼び出しで使用される。

### `Makefile`（編集）

既存ターゲットは維持し、`bootstrap` を追加する。

```make
.PHONY: build test lint tidy run clean release-check release-snapshot release-clean bootstrap

bootstrap:
 @command -v brew >/dev/null 2>&1 || { \
   echo "❌ Homebrew is required. See https://brew.sh"; exit 1; }
 brew bundle --file=Brewfile
 mise install
 lefthook install
 go mod download
 @echo "✅ bootstrap complete"
```

挙動:

1. `brew` が PATH に無ければメッセージを出して `exit 1`
2. `brew bundle` で Brewfile のツールをインストール（既存は skip）
3. `mise install` で Go / Node を `mise.toml` に従って導入
4. `lefthook install` で `.git/hooks/pre-commit` を配置
5. `go mod download` で Go モジュールをローカルキャッシュに取得
6. 完了メッセージを出力

途中で失敗した場合は make が自然に中断し、以降のステップは走らない。

### README 更新

#### `README.md`

Development セクションの pre-commit 関連を以下に置換する。

変更前（該当箇所）:

```markdown
Pre-commit hooks are managed by [lefthook](https://github.com/evilmartians/lefthook):
```

`` `brew install lefthook yamllint actionlint` `` + `` `lefthook install` `` のコードブロックが続く。

変更後（該当箇所）:

```markdown
Set up the full dev environment (Homebrew required) with:
```

`` `make bootstrap` `` のコードブロックが続き、以下の説明文を添える:

> This installs Homebrew packages listed in `Brewfile`, provisions Go / Node via `mise`, and enables [lefthook](https://github.com/evilmartians/lefthook) pre-commit hooks.

#### `docs/README.ja.md`

同じ位置を日本語で置換。見出し「開発」以下、次のプレーン文 + 一行コードブロック + 説明に置き換える:

- プレーン文: 「開発環境は 1 コマンドで整います（Homebrew 必須）:」
- コードブロック: `` `make bootstrap` ``
- 説明: 「`Brewfile` の Homebrew パッケージをインストールし、`mise` で Go / Node を用意、[lefthook](https://github.com/evilmartians/lefthook) の pre-commit フックを有効化します。」

`readme-sync` スキル対象なので両ファイルを必ず同期する。

## データフロー

```text
contributor
    │
    ▼
make bootstrap
    │
    ├─▶ command -v brew  ──── 無ければ exit 1
    │
    ├─▶ brew bundle   ──▶  Brewfile に列挙された CLI ツール
    │
    ├─▶ mise install  ──▶  mise.toml (go / node)
    │
    ├─▶ lefthook install ─▶  .git/hooks/pre-commit
    │
    └─▶ go mod download ──▶  $GOPATH/pkg/mod
```

## テスト / 検証

`make bootstrap` 自体は 1 回限りの整備コマンドなので自動テストは追加しない。代わりに以下を手動で確認する。

1. ツールが入った状態で `make bootstrap` を実行 → 全ステップが成功し完了メッセージが出る
2. `brew bundle check --file=Brewfile` が `The Brewfile's dependencies are satisfied.` を返す
3. `lefthook` インストール後に `.git/hooks/pre-commit` が存在する
4. `go mod download` が成功し、`go build ./cmd/ccw` が通る
5. `brew` が PATH に無い擬似状況（例: `PATH=/usr/bin make bootstrap`）で「❌ Homebrew is required.」を出して `exit 1` する

## エラーハンドリング

| ケース | 挙動 |
|---|---|
| `brew` 未インストール | 明示的メッセージで `exit 1` |
| `brew bundle` 途中失敗 | make が非ゼロ終了、以降は走らない |
| `mise install` 失敗 | 同上 |
| `lefthook install` 失敗 | 同上 |
| `go mod download` 失敗 | 同上 |

途中失敗時は `make bootstrap` を再実行すれば冪等に再開する（`brew bundle` は既存インストールを skip、`mise install` / `lefthook install` / `go mod download` も冪等）。

## 影響範囲

- 新規: `Brewfile`, `docs/superpowers/specs/2026-04-24-make-bootstrap-brewfile-design.md`
- 編集: `Makefile`, `mise.toml`, `README.md`, `docs/README.ja.md`
- CI: 変更なし
- `lefthook.yml`: 変更なし（既に参照済みのツールを Brewfile 化するだけ）

## オープン項目

なし。
