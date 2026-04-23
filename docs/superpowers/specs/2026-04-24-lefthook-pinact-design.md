# lefthook で pinact を実行するデザイン

- 日付: 2026-04-24
- ステータス: 設計確定（ユーザー承認済み）
- 対象リポジトリ: `ccw-cli`

## ゴール

GitHub Actions の `uses:` 参照が常に「SHA 固定 + バージョンコメント付き」に保たれることを、コミット時点で機械的に保証する。人の目視レビューや CI での事後検知に頼らない。

## 前提（調査済み）

- 現行の全 workflow（`.github/workflows/auto-assign.yml` / `ci.yml` / `release.yml`）の `uses:` 17 箇所すべて、既に `<sha> # <tag>` 形式でピン留め済み。pinact 導入による初回書き換え差分は発生しない見込み。
- `.github/actions/` ディレクトリ（composite action）は現時点で存在しないが、pinact のデフォルト対象に含まれているため、将来追加した際も自動で拾われる。
- 既存 lefthook の fixer は `markdownlint` / `shfmt` が `stage_fixed: true` を使った自動整形パターン。pinact もこの体験に揃える。
- 開発ツールは `Brewfile` で一元管理され `make bootstrap` で導入される。新規ツールは Brewfile への追加が自然な導線。

## 決定事項（ブレスト結論）

| 項目 | 決定 |
|---|---|
| 実行モード | 自動修正 + 再ステージ（`pinact run` + `stage_fixed: true`） |
| 対象スコープ | pinact のデフォルト（`.github/workflows/**` + `.github/actions/**`）に任せる |
| インストール手段 | Brewfile に追加し `make bootstrap` で導入 |
| `.pinact.yaml` 構成 | 最小（`version: 3` のみ） |

## 変更点（3 ファイル）

### 1. `Brewfile`

`# Linters / formatters` セクションに `pinact` を追加する（`actionlint` の近く）。

```ruby
# Linters / formatters (referenced by lefthook.yml and CI parity)
brew "yamllint"
brew "actionlint"
brew "pinact"
brew "shellcheck"
brew "shfmt"
brew "golangci-lint"
```

### 2. `.pinact.yaml`（新規）

リポジトリルートに以下を置く。

```yaml
version: 3
```

これで pinact のデフォルト対象（`.github/workflows/**/*.{yml,yaml}` + `.github/actions/**/*.{yml,yaml}`）に対し、全 action を SHA + バージョンコメント形式でピン留めする挙動になる。特定 action の例外は現時点で不要なので `ignore_actions` は書かない（YAGNI）。

### 3. `lefthook.yml`

`pre-commit.commands` の `# ── YAML / Actions ──` セクションに `pinact` を追加する。

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

- `actionlint` とは独立に動く（`parallel: true` の他コマンドと同様）。
- `stage_fixed: true` により、書き換えた差分がコミットに含まれる。

## 動作フロー

```text
git add .github/workflows/ci.yml        # 例: tag 参照で更新したケース
git commit
  └ lefthook pre-commit (parallel)
      ├ actionlint        : 構文チェック
      ├ yamllint          : YAML スタイル
      └ pinact run …      : tag → SHA + コメント に書換、stage_fixed で再 add
  → コミット成立
```

## エッジケース

- **pinact 未インストール環境**: Homebrew 未導入環境では `git commit` がフック失敗で止まる。既存 lefthook 全コマンドと同じ前提（`make bootstrap` 実行が必要）なので、追加負担はない。
- **意図的に tag 固定したい action が将来出た場合**: `.pinact.yaml` に `ignore_actions:` を足す。現状は空でよい。
- **CI 側での整合性チェック**: 本スコープ外。`--no-verify` によるバイパスが問題になった時点で、別タスクとして `pinact run --check` を CI に追加する。

## スコープ外（このタスクで扱わないもの）

- CI での pinact 再実行（`--no-verify` 対策）。
- Renovate 設定の変更（既存の Renovate digest ピン留め運用と pinact は共存するため、今回変更不要）。
- README の開発者向け節への追記（他ツールも個別言及していないため、一貫性のため加えない）。
- ccw 本体コード（picker, worktree, Go 実装）への影響なし。

## テスト観点（手動検証）

1. `make bootstrap` で `pinact` が入ること（`which pinact` / `pinact --version`）。
2. tag 参照の workflow を一時作成 → `git add` → `git commit` で SHA に書き換わってコミットに含まれる。
3. 既存 workflow を変更（コメント追加等）して commit → pinact が不要な書き換えを加えず、コミットが通る。
4. `lefthook run pre-commit --all-files` で全体スキャンが回り、差分 0 で終わる。

## 成功条件

- 上記 3 ファイルの変更がコミットされている。
- 手動検証 1〜4 がすべて期待通りに動く。
- 既存の lint チェック（actionlint / yamllint / golangci-lint 等）に回帰がない。

## 実装時の確認事項

- **Homebrew 参照形式**: `brew "pinact"`（homebrew-core 直指定）で入るか、`tap "suzuki-shunsuke/pinact"` + `brew "suzuki-shunsuke/pinact/pinact"` のような tap 指定が必要かを `brew info pinact` で確認してから Brewfile 記述を確定する。
- **`.pinact.yaml` の `version` キー値**: pinact の現行メジャー（v3 系）で `version: 3` が正しいか公式ドキュメント / `pinact init` の出力で確認する。もしスキーマが変わっていれば最新の推奨形式に合わせる。
