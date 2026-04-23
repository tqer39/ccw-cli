# Superpowers Auto-Install (PR-B) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `ccw -s` を `-y` 指定時 / 非対話時でも使えるように、`EnsureInstalled` に auto-install 分岐を追加する。

**Architecture:** `internal/superpowers/detect.go::EnsureInstalled` に `assumeYes bool` を追加。`!interactive || assumeYes` のとき事前メッセージ → `installRunner()` → 成功メッセージの順で自動インストールする。既存のインタラクティブ（`-y` なし）プロンプト分岐は変更しない。呼び出し側 `cmd/ccw/main.go::maybeSuperpowers` で `flags.AssumeYes` を引き回す。

**Tech Stack:** Go 1.25, `testing` + in-process `installRunner` 差し替え、`spf13/pflag`（`-y`）。

**Spec:** `docs/superpowers/specs/2026-04-24-superpowers-auto-install-design.md`

---

## ファイル構成

| ファイル | 変更 | 責務 |
|---|---|---|
| `internal/superpowers/detect.go` | 修正 | `EnsureInstalled` にシグネチャ `assumeYes bool` 追加 + auto-install 分岐 |
| `internal/superpowers/detect_test.go` | 修正 | 既存テストのシグネチャ更新 + auto-install ケース追加 |
| `cmd/ccw/main.go` | 修正 | `maybeSuperpowers` が `flags.AssumeYes` を受けて `EnsureInstalled` に渡す |
| `internal/cli/parse.go` | 修正 | `-y` のヘルプ文言を `skip confirmation prompts (--clean-all, -s plugin install)` に変更 |

`EnsureGitignore` はスコープ外（spec 非ゴール）。`cli/parse_test.go` は help 文字列を assert していないので修正不要。

---

## Task 1: `EnsureInstalled` シグネチャ拡張（機械的リファクタ）

このタスクは挙動を変えない。`assumeYes` を受け取るようシグネチャを広げ、既存呼び出し側・テストを追従させる。`assumeYes=false` を渡せば従来と完全に等価。

**Files:**

- Modify: `internal/superpowers/detect.go:45`
- Modify: `internal/superpowers/detect_test.go:59,70,82,92,108`
- Modify: `cmd/ccw/main.go:248`

- [ ] **Step 1: `EnsureInstalled` のシグネチャを変更（挙動は据え置き）**

`internal/superpowers/detect.go` の `EnsureInstalled` を次の形に置換。ロジックは `interactive` パラメータのみを見て従来通り動く（`assumeYes` は未使用＝unused 警告防止のため `_ = assumeYes` を入れる）。

```go
// EnsureInstalled returns nil if superpowers is detected under home; otherwise
// it prompts in interactive mode, or errors in non-interactive mode.
// assumeYes is accepted for forward-compatibility (wired up in Task 3).
func EnsureInstalled(in io.Reader, out io.Writer, home string, interactive, assumeYes bool) error {
 _ = assumeYes // wired up in Task 3
 ok, err := DetectInstalled(home)
 if err != nil {
  return err
 }
 if ok {
  return nil
 }

 _, _ = fmt.Fprintln(out, "⚠ missing dependency: superpowers plugin (required for -s)")
 _, _ = fmt.Fprintln(out, "The following command will install it:")
 _, _ = fmt.Fprintln(out, "  claude plugin install claude-plugins-official/superpowers")
 _, _ = fmt.Fprintln(out, "(reference: https://docs.claude.com/en/docs/claude-code/plugins )")

 if !interactive {
  return errors.New("superpowers plugin not installed (non-interactive)")
 }

 yes, err := ui.PromptYN(in, out, "Run now?")
 if err != nil {
  return fmt.Errorf("prompt: %w", err)
 }
 if !yes {
  return errors.New("cancelled by user")
 }
 if err := installRunner(); err != nil {
  return fmt.Errorf("plugin install failed: %w", err)
 }
 return nil
}
```

- [ ] **Step 2: 既存テストのシグネチャを更新**

`internal/superpowers/detect_test.go` 内の 5 箇所の `EnsureInstalled` 呼び出しに `false` を追加。

置換対象（すべて第 4 引数の後ろに `, false` を足す）:

```go
// TestEnsureInstalled_Present (64 行目付近)
if err := EnsureInstalled(strings.NewReader(""), &out, home, true, false); err != nil {

// TestEnsureInstalled_MissingNonInteractive (72 行目付近)
err := EnsureInstalled(strings.NewReader(""), &out, home, false, false)

// TestEnsureInstalled_UserCancels (80 行目付近)
err := EnsureInstalled(strings.NewReader("n\n"), &out, home, true, false)

// TestEnsureInstalled_InstallSucceeds (92 行目付近)
if err := EnsureInstalled(strings.NewReader("y\n"), &out, home, true, false); err != nil {

// TestEnsureInstalled_InstallFails (107 行目付近)
err := EnsureInstalled(strings.NewReader("y\n"), &out, home, true, false)
```

- [ ] **Step 3: `cmd/ccw/main.go` の呼び出し側を更新**

`cmd/ccw/main.go:248` を次に変更:

```go
 if err := superpowers.EnsureInstalled(os.Stdin, os.Stderr, home, interactive, false); err != nil {
```

まだ `flags.AssumeYes` を渡さない。Task 3 で繋ぐ。

- [ ] **Step 4: ビルド + 既存テスト通過を確認**

Run: `go build ./... && go test ./internal/superpowers/...`
Expected: すべて PASS（挙動変更なし）。

- [ ] **Step 5: Commit**

```bash
git add internal/superpowers/detect.go internal/superpowers/detect_test.go cmd/ccw/main.go
git commit -m "refactor(superpowers): EnsureInstalled に assumeYes を受ける余白を追加"
```

---

## Task 2: auto-install の失敗テストを書く（TDD red）

Task 3 で実装する新 path を赤テストで先に定義する。3 ケース追加:

- interactive=true + assumeYes=true + 未導入 → プロンプトせず `installRunner()` が呼ばれる
- interactive=false + 未導入 → プロンプトせず `installRunner()` が呼ばれる
- 成功時に事前メッセージと成功メッセージが出力に含まれる

**Files:**

- Modify: `internal/superpowers/detect_test.go` (末尾に追加)

- [ ] **Step 1: 3 つのテストを追加**

`internal/superpowers/detect_test.go` の末尾、`fakeErr` 型定義の前に以下を追加:

```go
func TestEnsureInstalled_AutoInstallWithAssumeYes(t *testing.T) {
 home := t.TempDir()
 var called bool
 orig := installRunner
 installRunner = func() error { called = true; return nil }
 t.Cleanup(func() { installRunner = orig })

 var out bytes.Buffer
 // interactive=true, assumeYes=true: プロンプトなしで install が走る
 if err := EnsureInstalled(strings.NewReader(""), &out, home, true, true); err != nil {
  t.Fatalf("EnsureInstalled assumeYes: %v", err)
 }
 if !called {
  t.Error("installRunner not called")
 }
 if strings.Contains(out.String(), "Run now?") {
  t.Errorf("prompt should not appear with assumeYes, got %q", out.String())
 }
 if !strings.Contains(out.String(), "Installing superpowers plugin") {
  t.Errorf("missing pre-install message, got %q", out.String())
 }
 if !strings.Contains(out.String(), "Installed superpowers plugin.") {
  t.Errorf("missing success message, got %q", out.String())
 }
}

func TestEnsureInstalled_AutoInstallNonInteractive(t *testing.T) {
 home := t.TempDir()
 var called bool
 orig := installRunner
 installRunner = func() error { called = true; return nil }
 t.Cleanup(func() { installRunner = orig })

 var out bytes.Buffer
 // interactive=false: assumeYes に関係なく install が走る
 if err := EnsureInstalled(strings.NewReader(""), &out, home, false, false); err != nil {
  t.Fatalf("EnsureInstalled non-interactive: %v", err)
 }
 if !called {
  t.Error("installRunner not called")
 }
 if strings.Contains(out.String(), "Run now?") {
  t.Errorf("prompt should not appear when non-interactive, got %q", out.String())
 }
}

func TestEnsureInstalled_AutoInstallFails(t *testing.T) {
 home := t.TempDir()
 orig := installRunner
 installRunner = func() error { return errFake }
 t.Cleanup(func() { installRunner = orig })

 var out bytes.Buffer
 err := EnsureInstalled(strings.NewReader(""), &out, home, false, false)
 if err == nil {
  t.Fatal("EnsureInstalled auto-install fail: want error")
 }
}
```

- [ ] **Step 2: テストを走らせて赤を確認**

Run: `go test ./internal/superpowers/... -run 'AutoInstall' -v`
Expected: 3 ケースとも FAIL。`TestEnsureInstalled_AutoInstallWithAssumeYes` は `Missing Dependency` メッセージは出るが installRunner は呼ばれず（`called` が false）アサーションで失敗。`AutoInstallNonInteractive` は `EnsureInstalled` がエラーを返して fail。`AutoInstallFails` は err == nil なので fail。

（3 ケースのうち 1 つだけ成功しても OK とはしない。想定される失敗理由が仕様と一致していることを必ず目視確認する。）

- [ ] **Step 3: Commit**

```bash
git add internal/superpowers/detect_test.go
git commit -m "test(superpowers): auto-install 分岐の失敗テストを追加"
```

---

## Task 3: auto-install 分岐を実装（TDD green）

**Files:**

- Modify: `internal/superpowers/detect.go::EnsureInstalled`

- [ ] **Step 1: auto-install 分岐を追加**

`internal/superpowers/detect.go` の `EnsureInstalled` を次に置換:

```go
// EnsureInstalled returns nil if superpowers is detected under home. Otherwise:
//   - interactive && !assumeYes: prompts "Run now?" before installing
//   - !interactive || assumeYes: auto-installs (with pre/post messages)
func EnsureInstalled(in io.Reader, out io.Writer, home string, interactive, assumeYes bool) error {
 ok, err := DetectInstalled(home)
 if err != nil {
  return err
 }
 if ok {
  return nil
 }

 _, _ = fmt.Fprintln(out, "⚠ missing dependency: superpowers plugin (required for -s)")
 _, _ = fmt.Fprintln(out, "The following command will install it:")
 _, _ = fmt.Fprintln(out, "  claude plugin install claude-plugins-official/superpowers")
 _, _ = fmt.Fprintln(out, "(reference: https://docs.claude.com/en/docs/claude-code/plugins )")

 if !interactive || assumeYes {
  return autoInstall(out)
 }

 yes, err := ui.PromptYN(in, out, "Run now?")
 if err != nil {
  return fmt.Errorf("prompt: %w", err)
 }
 if !yes {
  return errors.New("cancelled by user")
 }
 if err := installRunner(); err != nil {
  return fmt.Errorf("plugin install failed: %w", err)
 }
 return nil
}

func autoInstall(out io.Writer) error {
 _, _ = fmt.Fprintln(out, "Installing superpowers plugin (claude plugin install claude-plugins-official/superpowers)…")
 if err := installRunner(); err != nil {
  return fmt.Errorf("plugin install failed: %w", err)
 }
 _, _ = fmt.Fprintln(out, "Installed superpowers plugin.")
 return nil
}
```

- [ ] **Step 2: テストを走らせて緑を確認**

Run: `go test ./internal/superpowers/... -v`
Expected: 全テスト PASS（既存 5 + 新規 3 + その他）。

- [ ] **Step 3: Commit**

```bash
git add internal/superpowers/detect.go
git commit -m "feat(superpowers): -y / 非対話時の auto-install 分岐を追加"
```

---

## Task 4: `cmd/ccw/main.go` で `flags.AssumeYes` を引き回す

**Files:**

- Modify: `cmd/ccw/main.go:67,240,248`

- [ ] **Step 1: `maybeSuperpowers` シグネチャに `assumeYes` を追加**

`cmd/ccw/main.go:240-255` 付近を次に変更:

```go
func maybeSuperpowers(enabled bool, mainRepo string, interactive, assumeYes bool) (string, error) {
 if !enabled {
  return "", nil
 }
 home, err := os.UserHomeDir()
 if err != nil {
  return "", fmt.Errorf("resolve HOME: %w", err)
 }
 if err := superpowers.EnsureInstalled(os.Stdin, os.Stderr, home, interactive, assumeYes); err != nil {
  return "", fmt.Errorf("superpowers install: %w", err)
 }
 if err := superpowers.EnsureGitignore(os.Stdin, os.Stderr, mainRepo, interactive); err != nil {
  return "", fmt.Errorf("superpowers gitignore: %w", err)
 }
 return superpowers.Preamble(), nil
}
```

- [ ] **Step 2: 呼び出し側を更新**

`cmd/ccw/main.go:67` を次に変更（現状 `maybeSuperpowers(flags.Superpowers, mainRepo, interactive)`）:

```go
 preamble, err := maybeSuperpowers(flags.Superpowers, mainRepo, interactive, flags.AssumeYes)
```

- [ ] **Step 3: ビルド + 既存 CLI テスト通過を確認**

Run: `go build ./... && go test ./...`
Expected: 全パッケージ PASS。

- [ ] **Step 4: Commit**

```bash
git add cmd/ccw/main.go
git commit -m "feat(cli): maybeSuperpowers に assumeYes を引き回す"
```

---

## Task 5: `-y` ヘルプ文言を更新

**Files:**

- Modify: `internal/cli/parse.go:43`

- [ ] **Step 1: ヘルプ文字列を変更**

`internal/cli/parse.go:43` を次に変更:

```go
 fs.BoolVarP(&f.AssumeYes, "yes", "y", false, "skip confirmation prompts (--clean-all, -s plugin install)")
```

- [ ] **Step 2: 既存テストを確認**

Run: `go test ./internal/cli/...`
Expected: PASS（`parse_test.go` は help 文字列を assert していない）。

- [ ] **Step 3: ヘルプ出力の目視確認**

Run: `go run ./cmd/ccw --help 2>&1 | grep -- '-y,'`
Expected: 次の行が出る。

```text
  -y, --yes                   skip confirmation prompts (--clean-all, -s plugin install)
```

- [ ] **Step 4: Commit**

```bash
git add internal/cli/parse.go
git commit -m "docs(cli): -y のヘルプ文言を auto-install 含むスコープに更新"
```

---

## Task 6: 統合検証

自動テストだけでは確認しづらい 2 ケースを手動で実行。`installRunner` は実際の `claude plugin install` を呼ぶため、安全側で `CCW_DEBUG=1` も流しつつ、プラグインが既に入っている環境ではまず uninstall する（あるいは本手順をスキップして OK）。

- [ ] **Step 1: 非対話モードで auto-install が走ることを確認**

Run: `ccw -s < /dev/null` あるいは `printf '' | go run ./cmd/ccw -s`
Expected:

- プロンプトなし
- 出力に `Installing superpowers plugin …` と `Installed superpowers plugin.` が出る（プラグイン未導入の場合）
- 既に導入済みなら何も出ず worktree 作成に進む

（実環境で副作用が嫌なら `installRunner` を差し替えずに代わりに `go test -v -run AutoInstall ./internal/superpowers/...` の緑を根拠としてよい。）

- [ ] **Step 2: `-y` + TTY でも auto-install が走ることを確認**

Run: `go run ./cmd/ccw -s -y`
Expected: プロンプトなしで auto-install。

- [ ] **Step 3: 全テスト + vet**

Run: `go test ./... && go vet ./...`
Expected: 全 PASS / 警告なし。

- [ ] **Step 4: 最終 commit（必要なら）**

本タスクでコード変更は発生しない想定。もし調整が必要になれば個別 commit を作る。

---

## 完了基準

- 既存テスト全 PASS。
- 追加テスト（`TestEnsureInstalled_AutoInstall*` 3 件）全 PASS。
- `ccw -s -y` / `ccw -s` 非対話のいずれでも auto-install が走り、エラーで止まらない。
- `ccw --help` で `-y` 説明文が新文言になっている。
- README は変更しない（spec 非ゴール）。
- `EnsureGitignore` は触らない（spec 非ゴール）。
