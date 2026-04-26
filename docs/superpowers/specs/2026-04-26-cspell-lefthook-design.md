# cspell を lefthook に導入する設計

- 作成日: 2026-04-26
- ステータス: Approved
- スコープ: pre-commit による綴りチェックの追加 + CI での再チェック

## 背景と目的

現状の `lefthook.yml` には書式・静的解析系のフックは揃っているが、自然言語の綴りチェックが無い。README やドキュメント、コミットメッセージ近傍のコメント等の typo を pre-commit で検出したい。pre-commit のローカル防御に加え、フックをスキップした変更や Web 経由のコミットにも備えるため、CI でも全件スキャンする。

## 決定事項

| 論点 | 決定 |
|---|---|
| 実行タイミング | pre-commit + CI の両方 |
| 対象ファイル指定 | lefthook では glob を付けず `{staged_files}` をそのまま渡し、判定は `cspell` 側の `ignorePaths` / 自動バイナリ検知に一元化 |
| 取得方法 | `npm exec -y --package=cspell -- cspell ...`（既存 `markdownlint-cli2` / `renovate-config-validator` と同方式） |
| 設定ファイル配置 | `.cspell/cspell.json` |
| プロジェクト辞書 | `.cspell/project-words.txt`（1 行 1 語） |
| 失敗時挙動 | ハード fail（既存 lint と同じ） |

## ファイル変更一覧

| パス | 変更 | 内容 |
|---|---|---|
| `.cspell/cspell.json` | 新規 | cspell 本体設定 |
| `.cspell/project-words.txt` | 新規 | プロジェクト固有語辞書 |
| `.cspell/README.md` | 新規 | 偽陽性発生時の運用手順 |
| `lefthook.yml` | 編集 | `cspell` コマンドを `pre-commit.commands` に追加 |
| `.github/workflows/ci.yml` | 編集 | `cspell` ジョブを追加し、`workflow-result` の依存と結果チェックに含める |

## `.cspell/cspell.json`

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

設計上のポイント:

- ユーザー指定の除外（バイナリ・superpowers 関連）を `ignorePaths` で明示。
  - superpowers 関連は `internal/superpowers/preamble_*.txt`、`docs/superpowers/**`、`.claude/**` の 3 系統。
- バイナリは cspell の自動 binary skip にも頼るが、明示拡張子も追加して二重に守る。
- `Formula/**`（Homebrew formula）と `.goreleaser.yaml` はハッシュ・URL の塊で誤検知が多いため除外。
- 辞書ファイル自身（`.cspell/**`）も対象外にして自己参照ループを防ぐ。

## `.cspell/project-words.txt`（初期値）

リポジトリから抽出する想定の固有語。実装時に過不足を調整する。

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

実装フェーズで `cspell` を一度全件実行し、出てきた未知語のうち正当なものを追記してコミットする。

## `lefthook.yml` への追加

`pre-commit.commands` 配下に追加する。

```yaml
    cspell:
      run: |
        npm exec -y --package=cspell -- cspell lint \
          --no-progress --no-summary --no-must-find-files \
          --config .cspell/cspell.json \
          {staged_files}
```

- glob は付けず staged 全件を渡す。除外は `cspell.json` の `ignorePaths` に集約。
- `--no-must-find-files`: 渡されるファイルが空のケース、もしくは全部 `ignorePaths` に該当したケースで fail しない。
- 既存と並行実行（`parallel: true` 下）。

## CI ジョブ（`.github/workflows/ci.yml`）

新規ジョブを追加する。

```yaml
  cspell:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
      - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
      - uses: actions/setup-node@<sha>  # v4.x  (pinact 解決)
        with:
          node-version: 'lts/*'
      - name: Run cspell
        run: |
          npm exec -y --package=cspell -- cspell lint \
            --no-progress --config .cspell/cspell.json \
            "**/*"
```

- `actions/setup-node` の SHA は `pinact run` に解決させる（手書きしない）。
- `workflow-result.needs` に `cspell` を追加。
- `workflow-result` 内の result チェックに `cspell` を追加。

## 偽陽性が出たときの運用

`.cspell/README.md` を新規作成し、以下の運用ルールを記載する:

> 綴りエラーが正当な語の場合は `.cspell/project-words.txt` に 1 行 1 語で追記してコミットしてください。

## テスト戦略

1. ローカル: `lefthook run pre-commit --all-files`（または該当ファイルだけ stage して `git commit`）でパスを確認。
2. 単体: `npm exec -y --package=cspell -- cspell lint --config .cspell/cspell.json "**/*"` で全件 0 件まで辞書を整備。
3. CI: PR を上げて `cspell` ジョブが green になることを確認。

## スコープ外（YAGNI）

- URL / hex / base64 等のカスタム regex 除外（cspell デフォルトで概ね充足。必要になったら追加）。
- 言語別辞書（`en-GB` など）の追加。
- IDE 連携（VSCode 拡張）のセットアップ手順。
- `pre-push` / `commit-msg` への cspell 適用。
- Brewfile / mise.toml への cspell 追加（A 方式採用のため不要）。

## ロールバック手順

問題があれば以下のいずれかで戻せる:

- `lefthook.yml` の `cspell` コマンドを削除（pre-commit のみ無効化）。
- `.github/workflows/ci.yml` の `cspell` ジョブと `workflow-result` の依存を削除（CI のみ無効化）。
- `.cspell/` ディレクトリの削除（設定全体を撤去）。

## 次ステップ

本仕様承認後、`writing-plans` skill で実装計画を作成し、`executing-plans` で実装する。
