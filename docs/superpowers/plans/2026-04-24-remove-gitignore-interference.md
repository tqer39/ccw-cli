# `.gitignore` 干渉削除（PR-D）実装計画

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** ccw が `-s` 実行時にユーザーリポの `.gitignore` を編集する `EnsureGitignore` を完全に削除する。

**Architecture:** `internal/superpowers/gitignore.go` および対応するテストを丸ごと削除し、`cmd/ccw/main.go::maybeSuperpowers` から呼び出しと不要となる引数 (`mainRepo`) を取り除く。`EnsureInstalled` と `Preamble()` は変更しない。

**Tech Stack:** Go (標準ライブラリ + `os/exec`)、`go test`, `go vet`, `lefthook`

**Spec:** [`docs/superpowers/specs/2026-04-24-remove-gitignore-interference-design.md`](../specs/2026-04-24-remove-gitignore-interference-design.md)

**Branch:** `worktree-federated-sparking-bear`（main 起点、spec は既に main にコミット済み）

**Working directory:** worktree ルート (`.claude/worktrees/federated-sparking-bear`)

---

## ファイル構成

| ファイル | 操作 | 責務 |
|---|---|---|
| `internal/superpowers/gitignore.go` | 削除 | `EnsureGitignore` 実装本体 |
| `internal/superpowers/gitignore_test.go` | 削除 | `EnsureGitignore` のテスト |
| `cmd/ccw/main.go` | 編集 | `maybeSuperpowers` から `EnsureGitignore` 呼び出しおよび `mainRepo` 引数を削除、呼び出し側 (`run`) も合わせて修正 |
| `internal/superpowers/preamble.go` | 編集 | パッケージ doc コメントから ".gitignore augmentation" の文言を削除 |

---

## Task 1: 失敗するテストで「呼び出されないこと」をロックする前に、現状確認

**Files:**

- なし（read-only）

このタスクは TDD ループに入る前のセーフティチェック。Spec 通りの状態かを確認する。

- [ ] **Step 1: 現状のビルドとテストがグリーンであることを確認**

Run:

```bash
go build ./...
go test ./...
```

Expected:

- ビルド成功
- 全テスト PASS（`internal/superpowers` の既存 `EnsureGitignore_*` 系テストを含む）

ここで失敗する場合は spec が想定する出発状態と差分があるため、原因を調べてから次に進む。

---

## Task 2: `cmd/ccw/main.go` から `EnsureGitignore` 呼び出しを削除

**Files:**

- Modify: `cmd/ccw/main.go:67`（`maybeSuperpowers` 呼び出し側）
- Modify: `cmd/ccw/main.go:240-255`（`maybeSuperpowers` 関数本体）

- [ ] **Step 1: `maybeSuperpowers` の関数定義を編集**

`cmd/ccw/main.go` の `maybeSuperpowers` を以下に置換する。`mainRepo` 引数は不要になるので削除。

```go
func maybeSuperpowers(enabled bool, interactive bool) (string, error) {
 if !enabled {
  return "", nil
 }
 home, err := os.UserHomeDir()
 if err != nil {
  return "", fmt.Errorf("resolve HOME: %w", err)
 }
 if err := superpowers.EnsureInstalled(os.Stdin, os.Stderr, home, interactive); err != nil {
  return "", fmt.Errorf("superpowers install: %w", err)
 }
 return superpowers.Preamble(), nil
}
```

- [ ] **Step 2: 呼び出し側 (`run` 内) を新シグネチャに合わせる**

`cmd/ccw/main.go:67` の呼び出しを以下に変更。

変更前:

```go
preamble, err := maybeSuperpowers(flags.Superpowers, mainRepo, interactive)
```

変更後:

```go
preamble, err := maybeSuperpowers(flags.Superpowers, interactive)
```

- [ ] **Step 3: ビルド確認**

Run:

```bash
go build ./...
```

Expected: 成功（`gitignore.go` 自体はまだ残っているため `EnsureGitignore` シンボル自体は未参照のままビルドは通る）。

- [ ] **Step 4: コミット**

```bash
git add cmd/ccw/main.go
git commit -m "refactor(superpowers): drop EnsureGitignore call from maybeSuperpowers"
```

---

## Task 3: `internal/superpowers/gitignore.go` と対応テストを削除

**Files:**

- Delete: `internal/superpowers/gitignore.go`
- Delete: `internal/superpowers/gitignore_test.go`

- [ ] **Step 1: ファイル削除**

Run:

```bash
git rm internal/superpowers/gitignore.go internal/superpowers/gitignore_test.go
```

Expected: 2 ファイルが unstaged から staged の削除に移行。

- [ ] **Step 2: `go vet` と `go build` で未使用 import / 残存参照がないことを確認**

Run:

```bash
go vet ./...
go build ./...
```

Expected: 両方成功。`internal/superpowers` パッケージ内の `errors` / `io` / `os/exec` / `path/filepath` / `internal/ui` への参照は `gitignore.go` のみだったため、削除と同時に問題なくクリーンになる想定。

> 補足: `internal/ui` は他ファイルから参照されていないため、`internal/superpowers` 配下の他 `.go` で `ui` import が残ることはない。それでも `go vet` / `go build` が通れば OK。

- [ ] **Step 3: 全テスト実行**

Run:

```bash
go test ./...
```

Expected: 全テスト PASS。`internal/superpowers` パッケージは `detect_test.go` と `preamble_test.go` のみが残り、いずれも `EnsureGitignore` に依存しないため落ちないはず。

- [ ] **Step 4: コミット**

```bash
git add -A internal/superpowers/
git commit -m "refactor(superpowers): remove EnsureGitignore implementation and tests"
```

---

## Task 4: パッケージ doc コメントの更新

**Files:**

- Modify: `internal/superpowers/preamble.go:1-3`

`Preamble` のあるファイルがパッケージ doc コメント (`// Package superpowers ...`) を持っており、現状 ".gitignore augmentation" を含んでいる。実体が消えたので文言を整える。

- [ ] **Step 1: パッケージ doc を編集**

`internal/superpowers/preamble.go` の冒頭を以下に置換。

変更前:

```go
// Package superpowers handles the optional superpowers plugin preamble,
// plugin presence detection, and .gitignore augmentation used by `ccw -s`.
package superpowers
```

変更後:

```go
// Package superpowers handles the optional superpowers plugin preamble
// and plugin presence detection used by `ccw -s`.
package superpowers
```

- [ ] **Step 2: 確認 & コミット**

Run:

```bash
go vet ./...
go test ./...
```

Expected: 成功。

```bash
git add internal/superpowers/preamble.go
git commit -m "docs(superpowers): drop .gitignore mention from package doc"
```

---

## Task 5: 手動確認 — `ccw -s` で `.gitignore` が変化しないこと

**Files:**

- なし（実行確認のみ）

spec の「実装手順 4」を満たすための手動確認。サンドボックス用の使い捨てリポを作って観測する。

- [ ] **Step 1: 開発ビルドを生成**

Run:

```bash
go build -o /tmp/ccw-prd ./cmd/ccw
```

Expected: `/tmp/ccw-prd` が作成される。

- [ ] **Step 2: 一時リポで挙動を観察**

Run:

```bash
TMP=$(mktemp -d)
git -C "$TMP" init -q
echo "node_modules/" > "$TMP/.gitignore"
( cd "$TMP" && /tmp/ccw-prd -s --help >/dev/null 2>&1 || true )
diff <(printf 'node_modules/\n') "$TMP/.gitignore"
rm -rf "$TMP"
```

Expected:

- `diff` が差分なし（exit 0）。すなわち `.gitignore` は touch されていない。
- `superpowers workflow artifacts` のコメントブロックが追記されていないこと。

> 注: `--help` でも `maybeSuperpowers` が走らない場合は、代わりに `-s -h` 等、ccw 本体の help / dry-run 系で `EnsureInstalled` を素通りできる経路を使う。それも難しければ、`EnsureGitignore` を呼ぶ唯一のコードパスが消えていることをコードレビューで確認するに留める（spec の「破壊的影響なし」と整合）。

- [ ] **Step 3: クリーンアップ**

Run:

```bash
rm -f /tmp/ccw-prd
```

このタスクには commit なし。

---

## Task 6: lefthook / 最終チェックと PR 作成

**Files:**

- なし（運用）

- [ ] **Step 1: pre-commit / pre-push 相当のローカルフックをドライ実行**

Run:

```bash
lefthook run pre-commit
```

Expected: 全フック PASS。失敗した場合は spec の方針（`.gitignore` を触らない）を逸脱しない範囲で個別に対処。

- [ ] **Step 2: ブランチを push & PR 作成**

ブランチは worktree のまま (`worktree-federated-sparking-bear`)。

```bash
git push -u origin worktree-federated-sparking-bear
gh pr create --title "refactor(superpowers): remove EnsureGitignore (.gitignore interference)" --body "$(cat <<'EOF'
## Summary
- ccw の `-s` 実行時にユーザーリポの `.gitignore` を編集していた `EnsureGitignore` を完全に削除
- `docs/superpowers/` は commit したい成果物であり、ignore を促すのは方針と逆だったため
- 関連の `internal/superpowers/gitignore.go` / テスト / `maybeSuperpowers` の `mainRepo` 引数も合わせて整理

## Spec
- `docs/superpowers/specs/2026-04-24-remove-gitignore-interference-design.md`

## Test plan
- [ ] `go test ./...` PASS
- [ ] `lefthook run pre-commit` PASS
- [ ] 一時リポで `ccw -s` 起動後も `.gitignore` が無変更（Task 5 の手順）

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Expected: PR が作成され URL が返る。

---

## Out of scope（PR-D ではやらない）

- 既に `# superpowers workflow artifacts\ndocs/superpowers/` を追記してしまったリポの自動 migration
- `bin/ccw`（旧 bash 実装）の同期 — 別 PR で扱う
- `docs/superpowers/` を opt-in で ignore する代替フロー
