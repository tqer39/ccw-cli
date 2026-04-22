# Phase 1: Go Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Go 版 `ccw` の最小スケルトン（`-h` / `-v` のみ動作、他は "Phase 1 未実装" で exit 1）を構築する。bash 版 `bin/ccw` は温存し、以降のフェーズでロジックを Go 側に移す土台を作る。

**Architecture:** `cmd/ccw/main.go` を薄いエントリポイントとし、`internal/cli` で pflag を使った引数パース、`internal/ui` でカラーメッセージ・ツール存在チェック、`internal/version` で ldflags 注入された版情報を提供する。TDD で各 internal パッケージの小さなユニットを先に固める。

**Tech Stack:** Go 1.23 (`go.mod` / `mise.toml`), `github.com/spf13/pflag`, `golang.org/x/term`, `golangci-lint` v2, GitHub Actions, Makefile。

**関連 spec:** `docs/superpowers/specs/2026-04-23-phase-1-skeleton-design.md`

---

## File Structure

本プランで作成するファイル:

- Create: `mise.toml`
- Create: `go.mod`
- Create: `Makefile`
- Create: `.golangci.yml`
- Create: `cmd/ccw/main.go`
- Create: `internal/version/version.go`
- Create: `internal/version/version_test.go`
- Create: `internal/ui/ui.go`
- Create: `internal/ui/ui_test.go`
- Create: `internal/cli/parse.go`
- Create: `internal/cli/parse_test.go`
- Create: `internal/cli/help.go`
- Create: `.github/workflows/ci.yml`

`go.sum` は `go mod tidy` 実行で自動生成される。

触らないファイル: `bin/ccw`, `tests/*.bats`, 既存 `.github/workflows/lint.yml` / `auto-assign.yml`, `lefthook.yml`, `renovate.json5`, `.gitignore`, `README.md`。

---

## Task 1: Go ツールチェーン基盤（mise.toml / go.mod / Makefile / .golangci.yml）

**Files:**

- Create: `mise.toml`
- Create: `go.mod`
- Create: `Makefile`
- Create: `.golangci.yml`

### Step 1: `mise.toml` を作成

- [ ] **Step 1: `mise.toml` を作成**

作成内容（完全体）:

```toml
[tools]
go = "1.23"
```

ローカル開発者は `mise install` で Go 1.23 が入る。`golangci-lint` は Homebrew 管理（`brew install golangci-lint`）とし mise には含めない（Phase 1 での運用簡素化）。

- [ ] **Step 2: `mise install` で Go が揃うことを確認**

Run:

```bash
mise install
go version
```

期待: `go version go1.23.x darwin/<arch>` または `linux/<arch>`。

mise が入っていない環境では `brew install mise` で事前導入。既に Go が別経路で入っている場合はそのまま使ってよい。

- [ ] **Step 3: `go.mod` を作成**

Run:

```bash
go mod init github.com/tqer39/ccw-cli
```

生成される `go.mod`:

```text
module github.com/tqer39/ccw-cli

go 1.23
```

- [ ] **Step 4: pflag / x/term 依存を追加**

Run:

```bash
go get github.com/spf13/pflag@v1.0.5
go get golang.org/x/term@latest
go mod tidy
```

`go.sum` が生成される。`go.mod` 末尾に `require` ブロックが入る。

- [ ] **Step 5: `Makefile` を作成**

作成内容（完全体、インデントは **必ずタブ文字**）:

```make
.PHONY: build test lint tidy run clean

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
```

- [ ] **Step 6: `.golangci.yml` を作成**

作成内容（完全体、v2 形式）:

```yaml
version: "2"

linters:
  default: none
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - gocritic
    - revive
    - errorlint
    - wrapcheck
    - gocyclo
  settings:
    gocyclo:
      min-complexity: 15
    wrapcheck:
      ignore-sigs:
        - .Errorf(
        - errors.New(
        - errors.Unwrap(
        - errors.Join(
        - .Wrap(
        - .Wrapf(
        - .WithMessage(
        - .WithMessagef(
        - .WithStack(
        - os/exec.Command
  exclusions:
    rules:
      - path: _test\.go
        linters:
          - wrapcheck
          - errcheck

formatters:
  enable:
    - gofmt
    - goimports
```

`gosimple` は v2 で `staticcheck` に統合されたため enable から除外。その他は spec §.golangci.yml 準拠。テストファイルでは `wrapcheck` / `errcheck` を緩める（テスト内で `fmt.Fprintln` の戻り値無視などが頻出するため）。

- [ ] **Step 7: ローカルで golangci-lint が動くか確認**

Run:

```bash
command -v golangci-lint || brew install golangci-lint
golangci-lint --version
golangci-lint config verify
```

期待: `golangci-lint has version 2.x.x built ...` と `valid configuration`。

Phase 1 時点では Go ソースが無いので `golangci-lint run` は実行しない（no Go files で warning を出す可能性がある）。次タスクから。

- [ ] **Step 8: コミット**

Run:

```bash
git add mise.toml go.mod go.sum Makefile .golangci.yml
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
🔧 Phase 1: Go ツールチェーン基盤を追加

- mise.toml: Go 1.23 を宣言（開発者は mise install で揃う）
- go.mod: module github.com/tqer39/ccw-cli / go 1.23
- deps: spf13/pflag, x/term
- Makefile: build / test / lint / tidy / run / clean
- .golangci.yml: v2 形式、厳し目構成（default + goimports /
  misspell / gocritic / revive / errorlint / wrapcheck / gocyclo）

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: `internal/version` パッケージ（TDD）

**Files:**

- Create: `internal/version/version.go`
- Test: `internal/version/version_test.go`

- [ ] **Step 1: 失敗するテストを書く**

Create `internal/version/version_test.go`:

```go
package version

import (
 "strings"
 "testing"
)

func TestString_Default(t *testing.T) {
 got := String()
 want := "ccw dev (commit: none, built: unknown)"
 if got != want {
  t.Fatalf("String() = %q, want %q", got, want)
 }
}

func TestString_WithInjectedValues(t *testing.T) {
 origV, origC, origD := Version, Commit, Date
 t.Cleanup(func() {
  Version, Commit, Date = origV, origC, origD
 })
 Version = "v1.2.3"
 Commit = "abc1234"
 Date = "2026-04-23T00:00:00Z"

 got := String()
 for _, sub := range []string{"v1.2.3", "abc1234", "2026-04-23T00:00:00Z"} {
  if !strings.Contains(got, sub) {
   t.Errorf("String() = %q, want substring %q", got, sub)
  }
 }
}
```

- [ ] **Step 2: テスト失敗を確認**

Run:

```bash
go test ./internal/version/ -v
```

期待: `internal/version/version.go` が存在しないため `package version` が見つからずビルドエラー。

- [ ] **Step 3: 最小実装**

Create `internal/version/version.go`:

```go
// Package version exposes build-time version metadata injected via -ldflags.
package version

import "fmt"

var (
 // Version is the release tag (e.g. "v0.1.0"). Set via -ldflags.
 Version = "dev"
 // Commit is the git commit SHA. Set via -ldflags.
 Commit = "none"
 // Date is the build timestamp. Set via -ldflags.
 Date = "unknown"
)

// String returns a human-readable one-line description of the build.
func String() string {
 return fmt.Sprintf("ccw %s (commit: %s, built: %s)", Version, Commit, Date)
}
```

- [ ] **Step 4: テスト成功を確認**

Run:

```bash
go test ./internal/version/ -v -race
```

期待: `PASS` / `TestString_Default` / `TestString_WithInjectedValues` が通る。

- [ ] **Step 5: lint 実行**

Run:

```bash
golangci-lint run ./internal/version/...
```

期待: 出力なし（警告ゼロ）。

- [ ] **Step 6: コミット**

Run:

```bash
git add internal/version/
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
✨ internal/version: ldflags 注入対応の version パッケージを追加

var Version/Commit/Date と String() を提供。-ldflags で
"-X github.com/tqer39/ccw-cli/internal/version.Version=..."
を指定することで goreleaser / release workflow から値を埋め込める。
デフォルトは dev / none / unknown。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: `internal/ui` パッケージ（TDD）

**Files:**

- Create: `internal/ui/ui.go`
- Test: `internal/ui/ui_test.go`

### 設計メモ

- `SetWriter(out, err io.Writer)` が呼ばれると **color 強制 off**（テスト時の出力比較を安定させるため）
- `InitColor()` は：`NO_COLOR` env が非空 → off / それ以外は stderr が TTY なら on / 非 TTY なら off
- `Info` のみ stdout、他（`Warn` / `Error` / `Success` / `Debug`）は stderr
- prefix: Info=なし / Warn=`⚠` / Error=`✖` / Success=`✓` / Debug=`[debug]`
- 色: Warn=黄(33) / Error=赤(31) / Success=緑(32) / Debug=灰(90)
- `EnsureTool` は `os.Exit` が絡むのでテスト対象外（Phase 1 範囲）

- [ ] **Step 1: 失敗するテストを書く**

Create `internal/ui/ui_test.go`:

```go
package ui

import (
 "bytes"
 "strings"
 "testing"
)

func TestInfoGoesToStdout(t *testing.T) {
 var out, err bytes.Buffer
 SetWriter(&out, &err)

 Info("hello %s", "world")

 if got := out.String(); got != "hello world\n" {
  t.Errorf("stdout = %q, want %q", got, "hello world\n")
 }
 if got := err.String(); got != "" {
  t.Errorf("stderr = %q, want empty", got)
 }
}

func TestWarnGoesToStderrWithPrefix(t *testing.T) {
 var out, err bytes.Buffer
 SetWriter(&out, &err)

 Warn("careful")

 if got := err.String(); !strings.HasPrefix(got, "⚠ careful") {
  t.Errorf("stderr = %q, want prefix %q", got, "⚠ careful")
 }
 if got := out.String(); got != "" {
  t.Errorf("stdout = %q, want empty", got)
 }
}

func TestErrorGoesToStderrWithPrefix(t *testing.T) {
 var out, err bytes.Buffer
 SetWriter(&out, &err)

 Error("bad: %d", 42)

 if got := err.String(); !strings.HasPrefix(got, "✖ bad: 42") {
  t.Errorf("stderr = %q, want prefix %q", got, "✖ bad: 42")
 }
}

func TestSuccessGoesToStderrWithPrefix(t *testing.T) {
 var out, err bytes.Buffer
 SetWriter(&out, &err)

 Success("done")

 if got := err.String(); !strings.HasPrefix(got, "✓ done") {
  t.Errorf("stderr = %q, want prefix %q", got, "✓ done")
 }
}

func TestDebugSilentWhenEnvUnset(t *testing.T) {
 var out, err bytes.Buffer
 SetWriter(&out, &err)
 t.Setenv("CCW_DEBUG", "")

 Debug("diag %s", "x")

 if got := err.String(); got != "" {
  t.Errorf("stderr = %q, want empty when CCW_DEBUG unset", got)
 }
}

func TestDebugVisibleWhenEnvOne(t *testing.T) {
 var out, err bytes.Buffer
 SetWriter(&out, &err)
 t.Setenv("CCW_DEBUG", "1")

 Debug("diag %s", "x")

 if got := err.String(); !strings.HasPrefix(got, "[debug] diag x") {
  t.Errorf("stderr = %q, want prefix %q", got, "[debug] diag x")
 }
}

func TestSetWriterDisablesColor(t *testing.T) {
 var out, err bytes.Buffer
 SetWriter(&out, &err)

 Error("plain")

 // ANSI escape \x1b が含まれていないこと（SetWriter で色 off）
 if got := err.String(); strings.Contains(got, "\x1b[") {
  t.Errorf("stderr = %q, should not contain ANSI escapes", got)
 }
}
```

- [ ] **Step 2: テスト失敗を確認**

Run:

```bash
go test ./internal/ui/ -v
```

期待: `internal/ui/ui.go` が存在せずビルドエラー。

- [ ] **Step 3: 最小実装**

Create `internal/ui/ui.go`:

```go
// Package ui provides colored CLI output helpers and tool-presence checks.
package ui

import (
 "fmt"
 "io"
 "os"
 "os/exec"

 "golang.org/x/term"
)

var (
 stdout       io.Writer = os.Stdout
 stderr       io.Writer = os.Stderr
 colorEnabled bool
)

// InitColor evaluates NO_COLOR and the stderr TTY state once. Call from main
// before any Info/Warn/Error/Success/Debug.
func InitColor() {
 if os.Getenv("NO_COLOR") != "" {
  colorEnabled = false
  return
 }
 colorEnabled = term.IsTerminal(int(os.Stderr.Fd()))
}

// SetWriter redirects stdout/stderr output (tests) and disables color.
func SetWriter(out, err io.Writer) {
 stdout = out
 stderr = err
 colorEnabled = false
}

// Info writes to stdout without prefix or color.
func Info(format string, args ...any) {
 fmt.Fprintf(stdout, format+"\n", args...)
}

// Warn writes a yellow-prefixed message to stderr.
func Warn(format string, args ...any) {
 write(stderr, "⚠ ", 33, format, args...)
}

// Error writes a red-prefixed message to stderr.
func Error(format string, args ...any) {
 write(stderr, "✖ ", 31, format, args...)
}

// Success writes a green-prefixed message to stderr.
func Success(format string, args ...any) {
 write(stderr, "✓ ", 32, format, args...)
}

// Debug writes a gray-prefixed message to stderr only when CCW_DEBUG=1.
func Debug(format string, args ...any) {
 if os.Getenv("CCW_DEBUG") != "1" {
  return
 }
 write(stderr, "[debug] ", 90, format, args...)
}

// EnsureTool aborts with exit 1 if `name` is not found in PATH.
func EnsureTool(name, installHint string) {
 if _, err := exec.LookPath(name); err != nil {
  Error("required tool not found: %s. %s", name, installHint)
  os.Exit(1)
 }
}

func write(w io.Writer, prefix string, ansi int, format string, args ...any) {
 msg := fmt.Sprintf(format, args...)
 if colorEnabled {
  fmt.Fprintf(w, "\x1b[%dm%s%s\x1b[0m\n", ansi, prefix, msg)
  return
 }
 fmt.Fprintf(w, "%s%s\n", prefix, msg)
}
```

- [ ] **Step 4: テスト成功を確認**

Run:

```bash
go test ./internal/ui/ -v -race
```

期待: 全 7 テストが PASS。

- [ ] **Step 5: lint 実行**

Run:

```bash
golangci-lint run ./internal/ui/...
```

期待: 警告ゼロ。`fmt.Fprintf` の戻り値を無視している箇所で `errcheck` が鳴くようなら `//nolint:errcheck` ではなく `_, _ = fmt.Fprintf(...)` で受ける形に書き換える。

- [ ] **Step 6: lint 警告が出た場合の修正**

もし `errcheck` で `fmt.Fprintf` の戻り値無視が検出されたら、`write` / `Info` / `Debug` 内の `fmt.Fprintf` を次のように変更:

```go
_, _ = fmt.Fprintf(w, ...)
```

再実行して警告ゼロを確認:

```bash
golangci-lint run ./internal/ui/...
go test ./internal/ui/ -v -race
```

- [ ] **Step 7: コミット**

Run:

```bash
git add internal/ui/
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
✨ internal/ui: カラーメッセージとツール存在チェックを追加

Info (stdout) / Warn / Error / Success / Debug (stderr) を
提供。NO_COLOR env + stderr TTY 判定で色出力を制御し、
SetWriter(out, err) でテスト時に差し替え可能。
EnsureTool(name, hint) は PATH 不在時に exit 1。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: `internal/cli` パッケージ（TDD）

**Files:**

- Create: `internal/cli/parse.go`
- Create: `internal/cli/help.go`
- Test: `internal/cli/parse_test.go`

- [ ] **Step 1: 失敗するテストを書く**

Create `internal/cli/parse_test.go`:

```go
package cli

import (
 "reflect"
 "testing"
)

func TestParse_Table(t *testing.T) {
 cases := []struct {
  name    string
  argv    []string
  want    Flags
  wantErr bool
 }{
  {
   name: "empty",
   argv: []string{},
   want: Flags{},
  },
  {
   name: "short help",
   argv: []string{"-h"},
   want: Flags{Help: true},
  },
  {
   name: "long help",
   argv: []string{"--help"},
   want: Flags{Help: true},
  },
  {
   name: "short version",
   argv: []string{"-v"},
   want: Flags{Version: true},
  },
  {
   name: "long version",
   argv: []string{"--version"},
   want: Flags{Version: true},
  },
  {
   name: "short new",
   argv: []string{"-n"},
   want: Flags{NewWorktree: true},
  },
  {
   name: "long new",
   argv: []string{"--new"},
   want: Flags{NewWorktree: true},
  },
  {
   name: "short superpowers implies new",
   argv: []string{"-s"},
   want: Flags{NewWorktree: true, Superpowers: true},
  },
  {
   name: "long superpowers implies new",
   argv: []string{"--superpowers"},
   want: Flags{NewWorktree: true, Superpowers: true},
  },
  {
   name: "new and superpowers combined",
   argv: []string{"-n", "-s"},
   want: Flags{NewWorktree: true, Superpowers: true},
  },
  {
   name: "double dash only",
   argv: []string{"--"},
   want: Flags{Passthrough: []string{}},
  },
  {
   name: "double dash with args",
   argv: []string{"--", "foo", "bar"},
   want: Flags{Passthrough: []string{"foo", "bar"}},
  },
  {
   name: "new with passthrough",
   argv: []string{"-n", "--", "--model", "claude-opus-4-7"},
   want: Flags{NewWorktree: true, Passthrough: []string{"--model", "claude-opus-4-7"}},
  },
  {
   name: "superpowers with passthrough",
   argv: []string{"-s", "--", "--resume"},
   want: Flags{NewWorktree: true, Superpowers: true, Passthrough: []string{"--resume"}},
  },
  {
   name:    "unknown long flag",
   argv:    []string{"--unknown"},
   wantErr: true,
  },
  {
   name:    "removed update flag",
   argv:    []string{"--update"},
   wantErr: true,
  },
  {
   name:    "removed uninstall flag",
   argv:    []string{"--uninstall"},
   wantErr: true,
  },
  {
   name:    "positional arg rejected",
   argv:    []string{"positional"},
   wantErr: true,
  },
 }

 for _, tc := range cases {
  t.Run(tc.name, func(t *testing.T) {
   got, err := Parse(tc.argv)
   if tc.wantErr {
    if err == nil {
     t.Fatalf("Parse(%v) expected error, got nil", tc.argv)
    }
    return
   }
   if err != nil {
    t.Fatalf("Parse(%v) unexpected error: %v", tc.argv, err)
   }
   if !reflect.DeepEqual(got, tc.want) {
    t.Errorf("Parse(%v):\n got  = %+v\n want = %+v", tc.argv, got, tc.want)
   }
  })
 }
}
```

- [ ] **Step 2: テスト失敗を確認**

Run:

```bash
go test ./internal/cli/ -v
```

期待: `internal/cli/parse.go` 未作成によるビルドエラー。

- [ ] **Step 3: `internal/cli/parse.go` を実装**

Create `internal/cli/parse.go`:

```go
// Package cli defines ccw's command-line argument surface.
package cli

import (
 "fmt"
 "io"

 "github.com/spf13/pflag"
)

// Flags is the parsed representation of ccw's command-line arguments.
type Flags struct {
 Help        bool
 Version     bool
 NewWorktree bool
 Superpowers bool
 Passthrough []string
}

// Parse interprets argv (without the program name) and returns Flags.
// Unknown flags, positional args, and removed flags (--update / --uninstall)
// return a non-nil error.
func Parse(argv []string) (Flags, error) {
 pre, post := splitAtDoubleDash(argv)

 fs := pflag.NewFlagSet("ccw", pflag.ContinueOnError)
 fs.SetOutput(io.Discard)

 var f Flags
 fs.BoolVarP(&f.Help, "help", "h", false, "show help")
 fs.BoolVarP(&f.Version, "version", "v", false, "show version")
 fs.BoolVarP(&f.NewWorktree, "new", "n", false, "always start a new worktree")
 fs.BoolVarP(&f.Superpowers, "superpowers", "s", false, "inject superpowers preamble (implies --new)")

 if err := fs.Parse(pre); err != nil {
  return Flags{}, fmt.Errorf("parse flags: %w", err)
 }
 if args := fs.Args(); len(args) > 0 {
  return Flags{}, fmt.Errorf("unexpected positional arguments: %v (use -- to pass args to claude)", args)
 }
 if f.Superpowers {
  f.NewWorktree = true
 }
 f.Passthrough = post
 return f, nil
}

// splitAtDoubleDash returns (before, after) around the first bare "--" token.
// If "--" is absent, after is nil and before is argv. If "--" is present,
// after is a non-nil (possibly empty) slice to let callers distinguish it.
func splitAtDoubleDash(argv []string) (before, after []string) {
 for i, a := range argv {
  if a == "--" {
   return argv[:i], append([]string{}, argv[i+1:]...)
  }
 }
 return argv, nil
}
```

- [ ] **Step 4: テスト成功を確認**

Run:

```bash
go test ./internal/cli/ -v -race
```

期待: 全 18 サブテストが PASS。

- [ ] **Step 5: `internal/cli/help.go` を実装**

Create `internal/cli/help.go`:

```go
package cli

import (
 "fmt"
 "io"
)

const usage = `Usage: ccw [options] [-- <claude-args>...]

Options:
  -n, --new            常に新規 worktree で起動（既存 worktree の選択をスキップ）
  -s, --superpowers    superpowers プリアンブルを注入して起動（暗黙に -n）
  -v, --version        バージョン情報を表示
  -h, --help           このヘルプを表示

Arguments after ` + "`--`" + ` are forwarded to ` + "`claude`" + ` verbatim.

Environment:
  NO_COLOR=1           カラー出力を無効化
  CCW_DEBUG=1          詳細ログ出力

Exit codes:
  0  success
  1  user error / cancellation
  *  passthrough from ` + "`claude`" + `
`

// PrintHelp writes the usage string to w.
func PrintHelp(w io.Writer) {
 _, _ = fmt.Fprint(w, usage)
}
```

- [ ] **Step 6: lint + test 実行**

Run:

```bash
golangci-lint run ./internal/cli/...
go test ./internal/cli/ -v -race
```

期待: 警告ゼロ + 全テスト PASS。`wrapcheck` でエラー wrap に警告が出たら、`.golangci.yml` の `ignore-sigs` に追加するか、該当エラー wrapping を維持（既に `fmt.Errorf` で wrap しているため通るはず）。

- [ ] **Step 7: コミット**

Run:

```bash
git add internal/cli/
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
✨ internal/cli: pflag ベースの引数パーサと Help を追加

Parse(argv) (Flags, error) で -h / -v / -n / -s と "--"
以降の passthrough を処理。--update / --uninstall は
登録せず unknown flag として reject（親 spec §CLI 表面）。
-s は暗黙に NewWorktree=true を強制。位置引数は拒否。
PrintHelp(w) で bash 版 README と同じ Usage を出力。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: `cmd/ccw/main.go` とスケルトン動作確認

**Files:**

- Create: `cmd/ccw/main.go`

- [ ] **Step 1: `cmd/ccw/main.go` を作成**

Create `cmd/ccw/main.go`:

```go
// Command ccw launches Claude Code in an isolated git worktree.
//
// Phase 1 status: only -h / -v are functional. Other flag combinations
// print a "not implemented" message and exit 1. Continue to use the
// bash implementation at bin/ccw for day-to-day work.
package main

import (
 "fmt"
 "os"

 "github.com/tqer39/ccw-cli/internal/cli"
 "github.com/tqer39/ccw-cli/internal/ui"
 "github.com/tqer39/ccw-cli/internal/version"
)

func main() {
 ui.InitColor()

 flags, err := cli.Parse(os.Args[1:])
 if err != nil {
  ui.Error("%v", err)
  cli.PrintHelp(os.Stderr)
  os.Exit(2)
 }

 if flags.Help {
  cli.PrintHelp(os.Stdout)
  return
 }
 if flags.Version {
  fmt.Println(version.String())
  return
 }

 ui.Error("Phase 1 スケルトンのため、-n / -s / picker は未実装です。bash 版 bin/ccw を使用してください。")
 os.Exit(1)
}
```

- [ ] **Step 2: ビルド確認**

Run:

```bash
go build -o ccw ./cmd/ccw
ls -la ccw
```

期待: `ccw` バイナリが生成される。

- [ ] **Step 3: `-h` 動作確認**

Run:

```bash
./ccw -h
echo "exit=$?"
./ccw --help
echo "exit=$?"
```

期待: Usage が stdout に表示され、`exit=0`。

- [ ] **Step 4: `-v` 動作確認**

Run:

```bash
./ccw -v
echo "exit=$?"
./ccw --version
echo "exit=$?"
```

期待: `ccw dev (commit: none, built: unknown)` が stdout、`exit=0`。

- [ ] **Step 5: 未実装フラグ経路の動作確認**

Run:

```bash
./ccw -n; echo "exit=$?"
./ccw -s; echo "exit=$?"
./ccw;    echo "exit=$?"
```

期待: いずれも stderr に「Phase 1 スケルトンのため…」が表示され、`exit=1`。

- [ ] **Step 6: 不明フラグ経路の動作確認**

Run:

```bash
./ccw --update;    echo "exit=$?"
./ccw --uninstall; echo "exit=$?"
./ccw --bogus;     echo "exit=$?"
./ccw positional;  echo "exit=$?"
```

期待: いずれも stderr にエラーメッセージ + Usage が出力され、`exit=2`。

- [ ] **Step 7: ldflags 注入確認**

Run:

```bash
go build -ldflags "-X github.com/tqer39/ccw-cli/internal/version.Version=v0.1.0-test -X github.com/tqer39/ccw-cli/internal/version.Commit=abc1234 -X github.com/tqer39/ccw-cli/internal/version.Date=2026-04-23" -o ccw ./cmd/ccw
./ccw -v
rm -f ccw
```

期待: `ccw v0.1.0-test (commit: abc1234, built: 2026-04-23)`。

- [ ] **Step 8: lint + 全 test 実行**

Run:

```bash
golangci-lint run
go test ./... -race -coverprofile=coverage.out
```

期待: 警告ゼロ + 全パッケージのテスト PASS。

- [ ] **Step 9: clean して成果物を残さない**

Run:

```bash
make clean
ls ccw coverage.out 2>&1 | head
```

期待: `No such file or directory` が 2 行。

- [ ] **Step 10: コミット**

Run:

```bash
git add cmd/ccw/main.go
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
✨ cmd/ccw: Go 版 main スケルトンを追加

-h / -v のみ動作。-n / -s / picker は "Phase 1 スケルトン
のため未実装" メッセージを stderr に出して exit 1。不明フラグ
/ 位置引数は usage 表示 + exit 2。bash 版 bin/ccw は引き続き
利用可能。Phase 2 以降で picker / gitx / claude / superpowers を
順次 Go 側に移植する。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: `.github/workflows/ci.yml` を新設

**Files:**

- Create: `.github/workflows/ci.yml`

### 前提

- 既存 `.github/workflows/lint.yml`（shellcheck / shfmt / bats）と `auto-assign.yml` は併存させる。ci.yml は Go 専用。
- SHA ピンは「実装時点の最新 stable を採用 → 以降 Renovate が追従」。下記は spec 時点（2026-04-23）の推奨値だが、実装時に `gh api repos/<owner>/<repo>/commits/v<version>` で確認すること。

### SHA ピン確認コマンド

```bash
gh api repos/actions/checkout/commits/v4 -q '.sha'
gh api repos/actions/setup-go/commits/v5 -q '.sha'
gh api repos/golangci/golangci-lint-action/commits/v6 -q '.sha'
gh api repos/actions/upload-artifact/commits/v4 -q '.sha'
```

以下の `<SHA>` プレースホルダは確認した値に置換する。

- [ ] **Step 1: `.github/workflows/ci.yml` を作成**

Create `.github/workflows/ci.yml`:

```yaml
name: ci

on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

jobs:
  go-lint:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
      - uses: actions/checkout@<CHECKOUT_SHA> # v4.x.x
      - uses: actions/setup-go@<SETUP_GO_SHA> # v5.x.x
        with:
          go-version-file: go.mod
          cache: true
      - uses: golangci/golangci-lint-action@<GOLANGCI_SHA> # v6.x.x
        with:
          version: latest

  go-test:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
      - uses: actions/checkout@<CHECKOUT_SHA> # v4.x.x
      - uses: actions/setup-go@<SETUP_GO_SHA> # v5.x.x
        with:
          go-version-file: go.mod
          cache: true
      - name: Run tests
        run: go test ./... -race -coverprofile=coverage.out
      - uses: actions/upload-artifact@<UPLOAD_SHA> # v4.x.x
        with:
          name: coverage
          path: coverage.out
          if-no-files-found: error
          retention-days: 7

  go-build:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
      - uses: actions/checkout@<CHECKOUT_SHA> # v4.x.x
      - uses: actions/setup-go@<SETUP_GO_SHA> # v5.x.x
        with:
          go-version-file: go.mod
          cache: true
      - name: Build
        run: go build ./cmd/ccw
```

実装者は `<*_SHA>` を上記 `gh api` コマンドで取得した実際の SHA に置換。コメントは `# v<MAJOR>.<MINOR>.<PATCH>` 形式で Renovate が読めるように保つ。

- [ ] **Step 2: actionlint で構文検証**

Run:

```bash
actionlint .github/workflows/ci.yml
```

期待: 出力なし。`actionlint` 未インストールなら `brew install actionlint`。

- [ ] **Step 3: yamllint で検証**

Run:

```bash
yamllint --no-warnings .github/workflows/ci.yml
```

期待: 出力なし。

- [ ] **Step 4: lefthook の pre-commit をドライラン**

Run:

```bash
git add .github/workflows/ci.yml
lefthook run pre-commit
```

期待: `yamllint` / `actionlint` / `markdownlint` 等の関連 hook が PASS。

- [ ] **Step 5: コミット**

Run:

```bash
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
👷 CI: Go 向け workflow (ci.yml) を新設

go-lint / go-test / go-build の 3 job を追加。既存
lint.yml (shellcheck / shfmt / bats) と auto-assign.yml は
併存。setup-go は go-version-file: go.mod を参照し、
mise.toml との単一ソース運用を担保。SHA ピンは Renovate
が追従する。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Final Verification

- [ ] **Step 1: リポジトリ全体で lint + test + build**

Run:

```bash
make clean
golangci-lint run
go test ./... -race -coverprofile=coverage.out
go build -o ccw ./cmd/ccw
./ccw -h | head -5
./ccw -v
./ccw -n 2>&1; echo "exit=$?"
make clean
```

期待:

- `golangci-lint run` 警告ゼロ
- `go test ./...` 全 PASS
- `./ccw -h` が Usage を出力
- `./ccw -v` が `ccw dev (...)` を出力
- `./ccw -n` が `exit=1`
- `make clean` 後 `ccw` / `coverage.out` が消える

- [ ] **Step 2: lefthook で全ファイル対象に pre-commit を走らせる**

Run:

```bash
lefthook run pre-commit --all-files
```

期待: 全 hook（`gofmt`, `golangci-lint`, `shellcheck`, `shfmt`, `yamllint`, `actionlint`, `markdownlint-cli2`, `renovate-config-validator`, safety 系）が PASS。

- [ ] **Step 3: コミット履歴を確認**

Run:

```bash
git log --oneline -8
```

期待（新しい順）:

```text
<sha> 👷 CI: Go 向け workflow (ci.yml) を新設
<sha> ✨ cmd/ccw: Go 版 main スケルトンを追加
<sha> ✨ internal/cli: pflag ベースの引数パーサと Help を追加
<sha> ✨ internal/ui: カラーメッセージとツール存在チェックを追加
<sha> ✨ internal/version: ldflags 注入対応の version パッケージを追加
<sha> 🔧 Phase 1: Go ツールチェーン基盤を追加
<sha> 📝 Phase 1 spec: Go スケルトン設計を追加
<sha> (Phase 0 最終コミット)
```

- [ ] **Step 4: PR 作成（任意）**

Run:

```bash
git push -u origin HEAD
gh pr create --title "Phase 1: Go skeleton (-h / -v only)" --body "$(cat <<'EOF'
## Summary
- Go ツールチェーン基盤を追加 (mise.toml / go.mod / Makefile / .golangci.yml)
- `internal/version` / `internal/ui` / `internal/cli` を TDD で実装
- `cmd/ccw/main.go` スケルトン（`-h` / `-v` のみ動作）
- `.github/workflows/ci.yml` を新設 (go-lint / go-test / go-build)

bash 版 `bin/ccw` は温存し、Phase 2 以降で picker / gitx / claude /
superpowers を順次 Go 側に移植する。

## Test plan
- [ ] `make build && ./ccw -h` が Usage を表示
- [ ] `./ccw -v` が `ccw dev (commit: none, built: unknown)` を表示
- [ ] `./ccw -n` / `./ccw -s` / `./ccw` が "Phase 1 未実装" で exit 1
- [ ] `./ccw --update` / `./ccw --unknown` が exit 2
- [ ] `go test ./... -race` が PASS
- [ ] `golangci-lint run` が警告ゼロ
- [ ] `lefthook run pre-commit --all-files` が PASS
- [ ] CI (ci.yml) の 3 job が PR で PASS

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## 次のプラン

本プラン完了後、以下を順次作成予定（本プランの範囲外）:

1. **Phase 2 plan**: `internal/gitx` + `internal/worktree` + `internal/claude` + `internal/superpowers`（bash 版パリティ）
2. **Phase 3 plan**: `internal/picker`（bubbletea）+ teatest
3. **Phase 4 plan**: `.goreleaser.yaml` + release workflow + `tqer39/homebrew-tap` 初期化
4. **Phase 5 plan**: `v0.1.0` タグ → `brew install` 検証 → README 更新
