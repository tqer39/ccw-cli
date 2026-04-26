# 2026-04-26 — Codecov によるテストカバレッジ可視化（ハイブリッド構成）

## 背景

ccw-cli は Go CLI として `internal/` 配下にユニットテストを持ち、CI (`.github/workflows/ci.yml` の `go-test` ジョブ) で `go test ./... -race -coverprofile=coverage.out` を実行している。`coverage.out` は GitHub Actions の artifact としてアップロードされているが、

- バッジで全体カバレッジ率がパッと見えない
- PR ごとに「このPRで上がった/下がった」「未カバー行」が分からない
- Web 上で行単位のヒートマップが見られない
- ローカルで HTML を開く手順が暗黙知（`go tool cover -html=coverage.out`）

という状態で、カバレッジを「読みに行く」コストが高い。CODECOV_TOKEN は GitHub Secrets に既に設定済み。

## ゴール

- README にカバレッジバッジ（Codecov 提供）を表示し、main の現状値が常時見える
- PR ごとに Codecov が差分コメントを投稿し、未カバー行を line-level で確認できる
- カバレッジ率が直前の base commit より退行したら GitHub Checks が fail する（auto しきい値、絶対値ではなく差分でゲート）
- ローカルで `make coverage-html` 一発で `coverage.html` を生成できる
- Codecov サービスが落ちても CI 全体は通る（耐障害性）

## 非ゴール

- Bash (`bin/ccw`) のカバレッジ取得（kcov 等）。将来検討。
- Codecov 専用 CI ジョブの分離（既存 `go-test` への追加で十分、ジョブ増加を避ける）。
- 固定しきい値（`target: 80%` 等）の導入。auto で退行防止だけ行い、絶対値の運用ルールは導入後に別途検討。
- テストコード (`*_test.go`, `tests/`) のカバレッジ計測対象化（自分自身のテストはノイズなので除外）。
- 既存 artifact upload (`actions/upload-artifact`) の削除。Codecov と二重になるが、CI 内ダウンロード用途で残す。

## 決定事項

### D1. Codecov action を `go-test` ジョブに追加（ジョブ分離なし）

`.github/workflows/ci.yml` の `go-test` ジョブの末尾、既存の `actions/upload-artifact` ステップの後に `codecov/codecov-action` を追加する。

```yaml
- uses: codecov/codecov-action@<SHA> # vX.Y.Z
  with:
    token: ${{ secrets.CODECOV_TOKEN }}
    files: ./coverage.out
    fail_ci_if_error: false
    verbose: true
```

ポイント:

- **SHA pin**: 既存ワークフローは全アクションを SHA で pin している (`actions/checkout@de0fac2e...`, `actions/setup-go@4a3601...`)。Codecov action も同じ規約で SHA pin する。実装時は `codecov/codecov-action@vX` で記述してコミットし、リポジトリの `pinact` 設定により SHA に展開される（既存の運用フローと同じ）。本スペック上の `<SHA>` プレースホルダは実装手順の一部として解決される。
- **`fail_ci_if_error: false`**: Codecov 障害時に CI を落とさないため。テスト結果と coverage アップロードを分離。
- **`token`**: public repo でも 2024 以降の Codecov v4+ では token 必須。CODECOV_TOKEN は設定済み。
- **`files: ./coverage.out`**: 自動検出に任せず明示パス指定。
- **`verbose: true`**: 初期導入のトラブルシュート用。安定したら外すかは別途判断。

`workflow-result` ジョブの `needs` リストには Codecov 由来のステップは含めない（`go-test` の中の最終ステップとして埋め込む）。

### D2. `codecov.yml` をリポジトリルートに作成

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

ポイント:

- **`status.project.target: auto`** + **`threshold: 1%`**: 直前の base 比で 1% 以内の低下は許容（測定揺れ吸収）、それ以上の退行で fail。
- **`status.patch.target: auto`**: PR で追加/変更された行のカバレッジが、追加された行に対して妥当な比率を維持していること。
- **`comment.require_changes: false`**: 変化が小さい PR でもコメントを出す（最初は可視化重視、ノイズが多ければ後で `true` に）。
- **`ignore`**:
  - `tests/**`: bats テストとサポートファイル（Go ファイルではないが念のため）
  - `**/*_test.go`: Go テストコード自身
  - `cmd/ccw/main.go` は **含める**（オーケストレーションロジックがあるため指標から外さない）。
- **yamllint**: `.yamllint.yml` は `extends: default` + `line-length: 160 (warning)` + `truthy: [true, false, on]`。上記設定はこれらに収まる。

### D3. `Makefile` に `coverage` / `coverage-html` ターゲットを追加

```makefile
.PHONY: bootstrap build test lint tidy run clean release-check release-snapshot release-clean coverage coverage-html

# ... 既存ターゲット ...

coverage:
 go test ./... -race -coverprofile=coverage.out
 go tool cover -func=coverage.out | tail -n 1

coverage-html: coverage
 go tool cover -html=coverage.out -o coverage.html
 @echo "open coverage.html"

clean:
 rm -f ccw coverage.out coverage.html
```

ポイント:

- 既存 `test` ターゲットは触らない（`-coverprofile=coverage.out` は既に付いている）。`coverage` は `test` の上位互換ではなく「総合行のサマリ表示」を兼ねた別ターゲット。
- `coverage-html` は `coverage` を依存に呼び、`coverage.html` を生成。`open` は OS 依存なので echo に留め、ユーザーが手で開く（macOS なら `open coverage.html`、Linux なら `xdg-open`）。
- `clean` に `coverage.html` を追記。

### D4. README にバッジを追加（英語・日本語両方）

`README.md`（英語、ソース）と `docs/README.ja.md`（日本語、readme-sync で同期）の両方の badge 行に Codecov バッジを追加する。配置は既存の `brew-audit` バッジの隣。

```markdown
[![codecov](https://codecov.io/gh/tqer39/ccw-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/tqer39/ccw-cli)
```

`readme-sync` skill の規約に従い、英語ソース → 日本語の順で編集する。

### D5. `.gitignore` は変更不要

確認済み: 既存 `.gitignore` の "Go coverage" セクションに `coverage.out` と `coverage.html` が両方含まれている。追加変更なし。

## アーキテクチャへの影響

- Go コード（`cmd/`, `internal/`）には変更なし。
- CI ワークフローは `go-test` ジョブ内に 1 ステップ追加のみ。並列性・所要時間にほぼ影響なし。
- 新規ファイル: `codecov.yml`, `docs/superpowers/specs/2026-04-26-codecov-coverage-visualization-design.md`（本ファイル）。
- 既存テスト・lint への影響なし。`yamllint` が `codecov.yml` を新規対象として拾うが、上記設定はインデント 2 スペース・行長制限内に収まっている前提。

## 検証

ローカル:

1. `make coverage-html` 実行 → `coverage.out` と `coverage.html` が生成され、`coverage.html` をブラウザで開けることを確認
2. `make clean` で両ファイルが削除されることを確認
3. `yamllint codecov.yml` が pass することを確認（lefthook 経由でも可）

CI（PR 上で確認）:

1. PR を作成 → `go-test` ジョブが完了し、Codecov へのアップロードログが出ることを確認
2. PR に Codecov bot のコメントが投稿されることを確認（`reach,diff,flags,files` レイアウト）
3. GitHub Checks に `codecov/project` と `codecov/patch` のステータスが現れることを確認
4. 既存 `coverage` artifact upload が引き続き成功すること
5. main にマージ後、`https://codecov.io/gh/tqer39/ccw-cli` にダッシュボードが現れ、README のバッジが緑/数値を返すこと

退行検知の動作確認（任意）:

- 意図的にカバー済みの行を未カバーにする変更を別 PR で出し、`codecov/project` が fail することを確認 → 確認後 PR は close

## 移行ノート

- 破壊的変更なし。
- 既存開発者への影響: PR ごとに Codecov bot コメントが追加される。コメントが過剰と感じたら `codecov.yml` の `comment.require_changes: true` に切り替える（後追いで PR 1 本）。
- Codecov サービス障害時: `fail_ci_if_error: false` により CI は通る。バッジは古い値を返すが、復旧後の次回 push で更新される。
- Token ローテーション: `CODECOV_TOKEN` を更新する場合は GitHub Secrets で差し替えるのみ。コードに変更不要。
