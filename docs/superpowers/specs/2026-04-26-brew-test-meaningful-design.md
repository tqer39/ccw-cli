# 2026-04-26 — `brew test` を実質化（goreleaser brews テンプレート）

## 背景

現状 `.goreleaser.yaml` の `brews[0].test` ブロック:

```ruby
test: |
  system bin/"ccw", "-v"
```

`system` は exit code 0 のみを検証する。バイナリが起動して終了することは確認できるが、

- 期待した version 情報が出力されているか（ldflags 注入の検証）
- そもそも標準出力に何かが出ているか

を見ていないため、release pipeline の ldflags 設定ミス（version=`dev` のまま注入されない等）を `brew test` では捕捉できない。

## ゴール

`brew test ccw` で以下を実質的に検証する:

1. `ccw -v` の出力に Formula の `version` 文字列が含まれる（ldflags 注入の正常性）
2. `ccw -h` がクラッシュせず exit 0 で終わる（i18n.Init / cli.PrintHelp パスの smoke test）

## 非ゴール

- homebrew-core 用 Formula 雛形の作成（別タスク A）
- `--strict` の再有効化（次 release 後の別タスク）
- Linux 側の挙動差異対応（同一バイナリ・同一テストで OK）
- `--lang=en` を強制してロケール依存文字列まで検証する（過剰、A の homebrew-core 提出時に検討）

## 変更内容

### `.goreleaser.yaml` の `test:` ブロック

変更前:

```ruby
test: |
  system bin/"ccw", "-v"
```

変更後:

```ruby
test: |
  assert_match version.to_s, shell_output("#{bin}/ccw -v")
  system bin/"ccw", "-h"
```

**ポイント:**

- `version.to_s` は Formula の `version` 属性で、goreleaser の `{{.Version}}`（`v` プレフィックスなし、例 `0.18.0`）と一致する。
- `ccw -v` の出力形式は `ccw 0.18.0 (commit: abc1234, built: 2026-04-26T...)`（`internal/version/version.go::String`）。`0.18.0` が含まれるので `assert_match` は成功する。
- `ccw -h` はロケール依存（日本語デフォルト / `--lang=en` 指定で英語）。アサート文字列を入れず exit code のみ検証することでロケール非依存にする。
- 2 行とも `bin/"ccw"` は Homebrew が提供する `bin` プレフィックスを参照する。

### Homebrew-core との差異

homebrew-core の audit (`brew audit --new --strict`) は `assert_match` を使った非自明テストを推奨している。本変更はその方向へ寄せる第一歩で、A の homebrew-core 提出時の追加コストを下げる。ただし audit クリアそのものはここではゴールにしない。

## テスト

### ローカルでの検証

`brew test` 実体は release 後の Formula 経由でしか走らないため、ここでは以下で間接的に検証する:

1. **YAML 構文チェック**: `goreleaser check`
2. **Ruby 構文チェック**: `test:` ブロックを抜き出して `ruby -c` で確認（ヒアドキュメント形式なので `ruby -c -e "..."` で評価できる）
3. **次回 release 時の `brew test` 実行ログ**: GitHub Actions の `brew-audit.yml`（PR #54 で導入済み）が `brew test ccw` を回す。次のタグ push（v0.19.0 以降）で初めて実環境検証される。

### 既存 CI への影響

- `release.yml`（goreleaser 実行）: 影響なし。`.goreleaser.yaml` の構文は崩れない。
- `brew-audit.yml`: 影響なし。`brew test` は失敗時のみエラーになる。次回 release 時に新 test ブロックで動作する。

## リスク

- `version.to_s` が `ccw -v` 出力と一致しないケース: goreleaser の `{{.Version}}` 仕様（`v` を除く）と内部 `version.Version` の代入が一致している前提。これは現状の release で実証済み（345ad56 の release が成功している）。
- `ccw -h` が将来クラッシュする変更を入れた場合: `brew test` で検出できる（メリット）。
- `homebrew-tap` 側の Formula を上書きする merge_method: goreleaser brews は毎リリースで Formula を再生成 + commit する仕組みなので、test ブロックの変更は次 release 時に自動反映される。

## 影響範囲

- `tqer39/homebrew-tap` リポジトリの `Formula/ccw.rb` が次回 release 時に test ブロック更新される
- 既存ユーザーへの影響なし（`brew install` の挙動は変わらない）
- 次回 release（v0.19.0 想定）で初めて新 test ブロックが effective になる

## PR スコープ

単一 PR。`.goreleaser.yaml` のみ変更。
