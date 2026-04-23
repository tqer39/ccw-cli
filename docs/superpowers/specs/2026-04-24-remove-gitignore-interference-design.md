# 2026-04-24 — `.gitignore` 干渉の削除

## 背景

現状 `cmd/ccw/main.go::maybeSuperpowers` は `-s` 使用時に：

1. `superpowers.EnsureInstalled` でプラグイン導入確認
2. `superpowers.EnsureGitignore` で **プロジェクトの `.gitignore` に介入**

`EnsureGitignore` (`internal/superpowers/gitignore.go`) の現在の挙動：

- `git check-ignore` で `docs/superpowers/` が ignore 対象か確認
- 対象でなく interactive なら `Add to .gitignore?` を表示 → y なら

  ```text
  # superpowers workflow artifacts
  docs/superpowers/
  ```

  を `.gitignore` に追記
- non-interactive なら何もせず続行

### 何が問題か

- **`docs/superpowers/` は commit したい成果物**（design / plan / review のような workflow ドキュメント）。ツール側から ignore を促すのは方針と逆
- ユーザーが誤って y を押すと workflow 成果物が git から見えなくなる
- そもそも worktree ランチャーの責務として `.gitignore` 編集は筋違い（YAGNI 違反）
- bash 実装 (`bin/ccw`) からの移植物で、現代的な運用（commit 派）と合っていない

## ゴール

ccw がユーザーの `.gitignore` に触れない状態にする。

## 非ゴール

- `docs/superpowers/` を ignore したいユーザー向けの opt-in 追加（必要になってから検討）
- superpowers 以外のパスのハンドリング

## 変更内容

### 削除

- `internal/superpowers/gitignore.go` 全削除
- `internal/superpowers/gitignore_test.go` 全削除
- `cmd/ccw/main.go::maybeSuperpowers` から `superpowers.EnsureGitignore(...)` 呼び出しを除去（関連の error wrap もクリーンアップ）

### 残す

- `EnsureInstalled` は無関係なので変更なし
- `Preamble()` は変更なし

### 既存 `.gitignore` に追記済みの箇所は？

過去の ccw 起動で `# superpowers workflow artifacts\ndocs/superpowers/` が既に書き込まれたリポジトリがある可能性。ccw 側から消す（upgrade migration）は過剰。ユーザーが気付いたら手で消すで十分（追加の実装はしない）。

## 実装手順

1. `cmd/ccw/main.go` から `EnsureGitignore` 呼び出しを消す（テストが落ちない範囲で）
2. `gitignore.go` / `gitignore_test.go` を削除
3. `go build ./... && go test ./...` が通ることを確認
4. 手動確認: `ccw -s` を走らせて `.gitignore` が変更されないこと

## リスク / 影響

- 既に ignore 追記を入れているリポでも、ccw が追記しないだけなので破壊的影響はない
- `bin/ccw` (bash 実装) はこの挙動を残しているが、README で「transitional fallback」扱いなので放置で可（別 PR で同期してもよい）

## テスト

- 既存 `internal/superpowers/detect_test.go` / `preamble_test.go` は影響なし
- 削除したファイルのテストは消える（カバレッジ減少は許容）
- `lefthook` / `go vet` / `go test ./...`

## PR スコープ

この spec は **PR-D** 単独用。`bin/ccw` の旧 bash 実装同期は別 PR 候補。
