# `-s` superpowers flag 廃止 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `ccw` の `-s` / `--superpowers` flag および `internal/superpowers/` パッケージを削除し、代わりにリポジトリの `.claude/settings.json` で `superpowers@claude-plugins-official` プラグインを宣言する。

**Architecture:** Claude Code 標準の `enabledPlugins` 機構に責務を委譲し、`ccw` を「薄い worktree ランチャー」という本来のスコープに戻す。`-s` 由来のコード（プリアンブル injection、`EnsureInstalled` ロジック、関連 i18n 文字列、README 記述）を一掃する。

**Tech Stack:** Go 1.25, spf13/pflag, ccw 既存 i18n（YAML ベース）。

**Spec:** [`docs/superpowers/specs/2026-04-26-remove-s-flag-design.md`](../specs/2026-04-26-remove-s-flag-design.md)

**Working branch:** `chore/remove-superpowers-flag`（spec commit 済み）

---

## File Structure

| ファイル | 種別 | 役割 |
|---|---|---|
| `.claude/settings.json` | **新規** | プロジェクトレベルで `superpowers@claude-plugins-official` を要求。clone した開発者は `claude` 起動時に導入を促される |
| `.gitignore` | 修正 | `.claude/settings.local.json` を除外（ローカル設定の誤コミット防止） |
| `internal/superpowers/` ディレクトリ | **削除** | `preamble.go`, `preamble.txt`, `detect.go`, それぞれの `_test.go` ごと撤去 |
| `internal/cli/parse.go` | 修正 | `Superpowers` フィールド・`-s` flag 定義・`Superpowers` から `NewWorktree` への伝播・`-y` のヘルプ文字列 |
| `internal/cli/parse_test.go` | 修正 | `-s` 関連 4 ケース削除 |
| `internal/claude/claude.go` | 修正 | `BuildNewArgs` / `BuildInWorktreeArgs` / `LaunchNew` / `LaunchInWorktree` から `preamble` 引数を撤去 |
| `internal/claude/claude_test.go` | 修正 | preamble 関連テスト削除＋既存テストの呼び出しシグネチャ更新 |
| `cmd/ccw/main.go` | 修正 | `superpowers` import・`maybeSuperpowers`・`preamble` ローカル変数・呼び出し時の preamble 引数を削除 |
| `internal/i18n/locales/en.yaml` | 修正 | help テキストから `-s, --superpowers` 行を削除＋ `-y` の説明から `-s plugin install` を削除 |
| `internal/i18n/locales/ja.yaml` | 修正 | 同上（日本語版） |
| `README.md` | 修正 | Features 第 3 項目（line 36 付近）、Quick Start の `ccw -s` 行（line 48 付近）、Optional dependency の `-s` 言及（line 118 付近）を更新 |
| `docs/README.ja.md` | 修正 | 同上（日本語版） |

各ファイルは「責務がひとつ」になるよう既存構造を維持。`internal/superpowers/` は丸ごと不要になるため、ディレクトリごと削除する。

---

## Task 1: プロジェクト設定ファイル

ccw-cli を clone した開発者に superpowers の導入を促すため、リポジトリレベル設定を追加する。コードに依存しない独立タスク。

**Files:**

- Create: `.claude/settings.json`
- Modify: `.gitignore`（末尾に 1 行追加）

- [ ] **Step 1: `.claude/settings.json` を作成する**

  内容:

  ```json
  {
    "enabledPlugins": {
      "superpowers@claude-plugins-official": true
    }
  }
  ```

- [ ] **Step 2: `.gitignore` の末尾に `.claude/settings.local.json` を追加**

  既存の `.claude/worktrees/` 行の直後など、`.claude/` 関連の塊に並べる。最終差分:

  ```diff
   .claude/worktrees/
  +.claude/settings.local.json
  ```

  既存 `.gitignore` の `.claude/worktrees/` の場所を `grep -n '.claude/worktrees/' .gitignore` で確認してから挿入する。

- [ ] **Step 3: `git status` で意図した 2 ファイルだけ変更されていることを確認**

  Run: `git status --porcelain`
  Expected: `?? .claude/settings.json` と `M .gitignore` の 2 行だけ。

- [ ] **Step 4: Commit**

  ```bash
  git add .claude/settings.json .gitignore
  git commit -m "chore(claude): superpowers プラグインを enabledPlugins で宣言"
  ```

---

## Task 2: i18n help テキスト更新

`-s` を help から消す。コード変更前にやっておくと、後続タスクでテスト実行時に help を視認したときに整合する。

**Files:**

- Modify: `internal/i18n/locales/en.yaml:17,27`
- Modify: `internal/i18n/locales/ja.yaml:17,27`

- [ ] **Step 1: `internal/i18n/locales/en.yaml` から `-s` 行と `-y` の括弧書きを修正**

  Before（17 行目周辺）:

  ```yaml
        -n, --new            Always start a new worktree (skip picker)
        -s, --superpowers    Inject superpowers preamble (implies -n)
        -v, --version        Show version
  ```

  After:

  ```yaml
        -n, --new            Always start a new worktree (skip picker)
        -v, --version        Show version
  ```

  Before（27 行目）:

  ```yaml
        -y, --yes              Skip confirmation prompts (--clean-all, -s plugin install)
  ```

  After:

  ```yaml
        -y, --yes              Skip confirmation prompts (--clean-all)
  ```

- [ ] **Step 2: `internal/i18n/locales/ja.yaml` も同様に更新**

  Before（17 行目周辺）:

  ```yaml
        -n, --new            常に新しい worktree で開始 (picker をスキップ)
        -s, --superpowers    superpowers プリアンブルを注入 (-n を含む)
        -v, --version        バージョン表示
  ```

  After:

  ```yaml
        -n, --new            常に新しい worktree で開始 (picker をスキップ)
        -v, --version        バージョン表示
  ```

  Before（27 行目）:

  ```yaml
        -y, --yes              確認プロンプトをスキップ (--clean-all, -s plugin install)
  ```

  After:

  ```yaml
        -y, --yes              確認プロンプトをスキップ (--clean-all)
  ```

- [ ] **Step 3: ビルドが通ることだけ確認（i18n の YAML がロードできるか）**

  Run: `go build ./...`
  Expected: 0 exit、エラーなし。

- [ ] **Step 4: Commit**

  ```bash
  git add internal/i18n/locales/en.yaml internal/i18n/locales/ja.yaml
  git commit -m "i18n(help): -s / --superpowers 行と -y 括弧書きから -s 言及を削除"
  ```

---

## Task 3: cli パッケージから `-s` を撤去（テスト先行）

`Flags.Superpowers` フィールド・`fs.BoolVarP("superpowers", "s", ...)`・`if f.Superpowers { f.NewWorktree = true }` を削除する。同時に `cmd/ccw/main.go` 側で `flags.Superpowers` を参照している箇所と `maybeSuperpowers` 呼び出しも消さないとビルドが落ちるため、同じコミットでまとめる。`internal/superpowers/` パッケージ自体（detect.go の `EnsureInstalled` / preamble.go の `Preamble`）は Task 5 で削除する。今は `cmd/ccw/main.go` の `superpowers` import と `maybeSuperpowers` を消すだけにとどめる。

**Files:**

- Modify: `internal/cli/parse_test.go` — `-s` 関連 4 ケース削除
- Modify: `internal/cli/parse.go` — `Superpowers` フィールド・flag 定義・伝播削除
- Modify: `cmd/ccw/main.go` — `superpowers` import、`maybeSuperpowers` 関数、その呼び出し、`preamble` ローカル変数を削除（preamble 引数は Task 4 で消す。一旦空文字列を渡す）

- [ ] **Step 1: `internal/cli/parse_test.go` から `-s` 関連ケースを削除**

  以下 4 ケースを `cases` スライスから取り除く（前後のコンマ含めて整える）:

  - `"short superpowers implies new"` (`-s`)
  - `"long superpowers implies new"` (`--superpowers`)
  - `"new and superpowers combined"` (`-n`, `-s`)
  - `"superpowers with passthrough"` (`-s`, `--`, `--resume`)

  これらはそれぞれ `Flags{NewWorktree: true, Superpowers: true, ...}` を期待しており、`Superpowers` フィールド削除によりコンパイル不能になるため、まず削除する。

- [ ] **Step 2: テストを走らせて、現状（コードはまだ古い）でも他のケースは pass することを確認**

  Run: `go test ./internal/cli/...`
  Expected: PASS（削除した 4 ケースが消えただけで、残りは通る）。

- [ ] **Step 3: `internal/cli/parse.go` の `Flags` 構造体から `Superpowers` を削除**

  Before（13 行目周辺）:

  ```go
  type Flags struct {
      Help         bool
      Version      bool
      NewWorktree  bool
      Superpowers  bool
      CleanAll     bool
      ...
  }
  ```

  After:

  ```go
  type Flags struct {
      Help         bool
      Version      bool
      NewWorktree  bool
      CleanAll     bool
      ...
  }
  ```

- [ ] **Step 4: `internal/cli/parse.go` から flag 定義 1 行を削除**

  Before（39 行目）:

  ```go
  fs.BoolVarP(&f.Superpowers, "superpowers", "s", false, "inject superpowers preamble (implies --new)")
  ```

  → 該当行を削除する。

- [ ] **Step 5: `internal/cli/parse.go` から `Superpowers` 伝播ブロックを削除**

  Before:

  ```go
  if f.Superpowers {
      f.NewWorktree = true
  }
  ```

  → 該当 3 行を削除する（直前/直後の空行も整える）。

- [ ] **Step 6: `internal/cli/parse.go` の `--yes` 説明文から `-s plugin install` を削除**

  Before（44 行目）:

  ```go
  fs.BoolVarP(&f.AssumeYes, "yes", "y", false, "skip confirmation prompts (--clean-all, -s plugin install)")
  ```

  After:

  ```go
  fs.BoolVarP(&f.AssumeYes, "yes", "y", false, "skip confirmation prompts (--clean-all)")
  ```

- [ ] **Step 7: `cmd/ccw/main.go` の冒頭コメントから `-s` 言及を削除**

  Before（4 行目周辺）:

  ```go
  // Phase 3 status: -h / -v / -n / -s and the picker are at parity with the
  // bash implementation. The bash `bin/ccw` is kept as a transitional fallback.
  ```

  After:

  ```go
  // Phase 3 status: -h / -v / -n and the picker are at parity with the
  // bash implementation. The bash `bin/ccw` is kept as a transitional fallback.
  ```

- [ ] **Step 8: `cmd/ccw/main.go` から `superpowers` import を削除**

  Before:

  ```go
  import (
      ...
      "github.com/tqer39/ccw-cli/internal/picker"
      "github.com/tqer39/ccw-cli/internal/superpowers"
      "github.com/tqer39/ccw-cli/internal/ui"
      ...
  )
  ```

  → `"github.com/tqer39/ccw-cli/internal/superpowers"` 行のみ削除。

- [ ] **Step 9: `cmd/ccw/main.go` から `maybeSuperpowers` 関数を完全削除**

  ファイル末尾付近の以下ブロックを削除:

  ```go
  func maybeSuperpowers(enabled bool, interactive, assumeYes bool) (string, error) {
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
      return superpowers.Preamble(), nil
  }
  ```

- [ ] **Step 10: `cmd/ccw/main.go` の `run` 関数内で `maybeSuperpowers` 呼び出しを削除し、`LaunchNew` の preamble 引数に空文字列を渡す**

  Before（73-86 行目周辺）:

  ```go
      preamble, err := maybeSuperpowers(flags.Superpowers, interactive, flags.AssumeYes)
      if err != nil {
          ui.Error("%v", err)
          return 1
      }

      if flags.NewWorktree {
          name, err := namegen.Generate(mainRepo)
          if err != nil {
              ui.Error("generate worktree name: %v", err)
              return 1
          }
          code, err := claude.LaunchNew(mainRepo, name, preamble, flags.Passthrough)
  ```

  After:

  ```go
      if flags.NewWorktree {
          name, err := namegen.Generate(mainRepo)
          if err != nil {
              ui.Error("generate worktree name: %v", err)
              return 1
          }
          code, err := claude.LaunchNew(mainRepo, name, "", flags.Passthrough)
  ```

  preamble 引数自体は Task 4 で claude 側から消す。

- [ ] **Step 11: ビルドとテストが通ることを確認**

  Run: `go build ./...`
  Expected: 0 exit。

  Run: `go test ./...`
  Expected: 全 pass。

- [ ] **Step 12: Commit**

  ```bash
  git add internal/cli/parse.go internal/cli/parse_test.go cmd/ccw/main.go
  git commit -m "refactor(cli): -s / --superpowers flag を削除し maybeSuperpowers を撤去"
  ```

---

## Task 4: claude パッケージから `preamble` 引数を撤去（テスト先行）

`BuildNewArgs(name, preamble, extra)` などから `preamble` を削除し、preamble が空でないときに `--` を末尾に挿入する分岐も消す。`cmd/ccw/main.go` で `LaunchNew(mainRepo, name, "", flags.Passthrough)` および `LaunchInWorktree(path, name, "", passthrough)` と空文字列を渡しているため、シグネチャ変更と同コミットで呼び出し側も詰める。

**Files:**

- Modify: `internal/claude/claude_test.go` — preamble 含むテストを削除・呼び出しシグネチャを 2 引数化
- Modify: `internal/claude/claude.go` — 4 関数のシグネチャを変更し `preamble` 分岐を削除
- Modify: `cmd/ccw/main.go` — `LaunchNew` / `LaunchInWorktree` 呼び出しの空文字列引数を削除

- [ ] **Step 1: `internal/claude/claude_test.go` を新シグネチャに書き換える**

  以下の最終内容にファイル全体を置き換える（preamble ケース 2 件を削除し、残りを `(name, extra)` シグネチャ向けに更新）:

  ```go
  package claude

  import (
      "reflect"
      "testing"
  )

  func TestBuildNewArgs_NameOnly(t *testing.T) {
      got := BuildNewArgs("foo", nil)
      want := []string{"--permission-mode", "auto", "--worktree", "foo", "-n", "foo"}
      if !reflect.DeepEqual(got, want) {
          t.Errorf("BuildNewArgs:\n got  = %v\n want = %v", got, want)
      }
  }

  func TestBuildNewArgs_WithExtra(t *testing.T) {
      got := BuildNewArgs("foo", []string{"--model", "claude-opus-4-7"})
      want := []string{"--permission-mode", "auto", "--worktree", "foo", "-n", "foo", "--model", "claude-opus-4-7"}
      if !reflect.DeepEqual(got, want) {
          t.Errorf("BuildNewArgs extra:\n got  = %v\n want = %v", got, want)
      }
  }

  func TestBuildInWorktreeArgs_NameOnly(t *testing.T) {
      got := BuildInWorktreeArgs("foo", nil)
      want := []string{"--permission-mode", "auto", "-n", "foo"}
      if !reflect.DeepEqual(got, want) {
          t.Errorf("BuildInWorktreeArgs:\n got  = %v\n want = %v", got, want)
      }
  }

  func TestBuildInWorktreeArgs_WithExtra(t *testing.T) {
      got := BuildInWorktreeArgs("foo", []string{"--model", "x"})
      want := []string{"--permission-mode", "auto", "-n", "foo", "--model", "x"}
      if !reflect.DeepEqual(got, want) {
          t.Errorf("BuildInWorktreeArgs extra:\n got  = %v\n want = %v", got, want)
      }
  }

  func TestBuildContinueArgs_Empty(t *testing.T) {
      got := BuildContinueArgs(nil)
      want := []string{"--permission-mode", "auto", "--continue"}
      if !reflect.DeepEqual(got, want) {
          t.Errorf("BuildContinueArgs:\n got  = %v\n want = %v", got, want)
      }
  }

  func TestBuildContinueArgs_WithExtra(t *testing.T) {
      got := BuildContinueArgs([]string{"--model", "x"})
      want := []string{"--permission-mode", "auto", "--continue", "--model", "x"}
      if !reflect.DeepEqual(got, want) {
          t.Errorf("BuildContinueArgs extra:\n got  = %v\n want = %v", got, want)
      }
  }
  ```

- [ ] **Step 2: テストを走らせて期待通り fail することを確認（実装はまだ古いシグネチャ）**

  Run: `go test ./internal/claude/...`
  Expected: FAIL — 「too many arguments」「too few arguments」など型エラーでコンパイル不能。

- [ ] **Step 3: `internal/claude/claude.go` の 4 関数のシグネチャと本体を更新**

  ファイル全体を以下に置き換える:

  ```go
  // Package claude wraps launching the `claude` CLI in ccw-appropriate ways
  // (new worktree session vs. continue existing worktree).
  package claude

  import (
      "errors"
      "fmt"
      "os"
      "os/exec"
  )

  // BuildNewArgs constructs argv (excluding the program name) for
  // `claude --permission-mode auto --worktree <name> -n <name> [extra...]`.
  func BuildNewArgs(name string, extra []string) []string {
      args := make([]string, 0, 6+len(extra))
      args = append(args, "--permission-mode", "auto", "--worktree", name, "-n", name)
      args = append(args, extra...)
      return args
  }

  // BuildInWorktreeArgs is BuildNewArgs without `--worktree`. Use when cwd is
  // already an existing worktree, since passing `--worktree <name>` from inside
  // a worktree risks a name-collision error against the existing git registration.
  func BuildInWorktreeArgs(name string, extra []string) []string {
      args := make([]string, 0, 4+len(extra))
      args = append(args, "--permission-mode", "auto", "-n", name)
      args = append(args, extra...)
      return args
  }

  // BuildContinueArgs constructs argv for `claude --permission-mode auto --continue [extra...]`.
  func BuildContinueArgs(extra []string) []string {
      args := make([]string, 0, 3+len(extra))
      args = append(args, "--permission-mode", "auto", "--continue")
      return append(args, extra...)
  }

  // LaunchNew execs claude with BuildNewArgs in cwd. Returns claude's exit code
  // (0 on success, the child exit code on non-zero exit, -1 on exec error).
  func LaunchNew(cwd, name string, extra []string) (int, error) {
      return runClaude(cwd, BuildNewArgs(name, extra))
  }

  // LaunchInWorktree execs claude with BuildInWorktreeArgs in cwd (no `--worktree`).
  func LaunchInWorktree(cwd, name string, extra []string) (int, error) {
      return runClaude(cwd, BuildInWorktreeArgs(name, extra))
  }

  // Continue execs claude with BuildContinueArgs in cwd.
  func Continue(cwd string, extra []string) (int, error) {
      return runClaude(cwd, BuildContinueArgs(extra))
  }

  func runClaude(cwd string, args []string) (int, error) {
      cmd := exec.Command("claude", args...)
      cmd.Dir = cwd
      cmd.Stdin = os.Stdin
      cmd.Stdout = os.Stdout
      cmd.Stderr = os.Stderr
      err := cmd.Run()
      if err == nil {
          return 0, nil
      }
      var exitErr *exec.ExitError
      if errors.As(err, &exitErr) {
          return exitErr.ExitCode(), nil
      }
      return -1, fmt.Errorf("run claude: %w", err)
  }
  ```

- [ ] **Step 4: `cmd/ccw/main.go` の呼び出し 3 箇所から空文字列引数を取り除く**

  Search: `grep -n 'claude.LaunchNew\|claude.LaunchInWorktree' cmd/ccw/main.go`

  期待される変更:

  - `code, err := claude.LaunchNew(mainRepo, name, "", flags.Passthrough)` → `code, err := claude.LaunchNew(mainRepo, name, flags.Passthrough)`
  - `code, err := claude.LaunchNew(mainRepo, name, "", passthrough)` → `code, err := claude.LaunchNew(mainRepo, name, passthrough)`（`runPicker` 内 `ActionNew` ケース）
  - `code, err := claude.LaunchInWorktree(path, name, "", passthrough)` → `code, err := claude.LaunchInWorktree(path, name, passthrough)`（`launchInPlace`）

  3 箇所すべて修正したことを再 grep で確認: `grep -n '"",' cmd/ccw/main.go` で `claude.Launch...` 行が 0 件になっていること。

- [ ] **Step 5: ビルドとテストが通ることを確認**

  Run: `go build ./...`
  Expected: 0 exit。

  Run: `go test ./...`
  Expected: 全 pass。

- [ ] **Step 6: Commit**

  ```bash
  git add internal/claude/claude.go internal/claude/claude_test.go cmd/ccw/main.go
  git commit -m "refactor(claude): preamble 引数を 4 関数から撤去"
  ```

---

## Task 5: `internal/superpowers/` パッケージを削除

ここまでで `internal/superpowers/` への参照はすべて消えているため、ディレクトリごと削除して dead code を排除する。

**Files:**

- Delete: `internal/superpowers/preamble.go`
- Delete: `internal/superpowers/preamble.txt`
- Delete: `internal/superpowers/preamble_test.go`
- Delete: `internal/superpowers/detect.go`
- Delete: `internal/superpowers/detect_test.go`

- [ ] **Step 1: 残存参照がないことを確認**

  Run: `grep -rn 'internal/superpowers\|superpowers\.Preamble\|superpowers\.EnsureInstalled' cmd/ internal/ 2>&1 | grep -v '^internal/superpowers/'`
  Expected: 0 matches。何か出たら該当箇所を先に潰す。

- [ ] **Step 2: ディレクトリを削除**

  ```bash
  rm -rf internal/superpowers
  ```

- [ ] **Step 3: ビルドとテストが通ることを確認**

  Run: `go build ./...`
  Expected: 0 exit。

  Run: `go test ./...`
  Expected: 全 pass。

- [ ] **Step 4: Commit**

  ```bash
  git add -A internal/superpowers
  git commit -m "refactor: internal/superpowers パッケージを削除"
  ```

  注: `git add -A internal/superpowers` でディレクトリ削除を staging する。`git status` で `D` 行 5 件が並ぶことを確認してから commit。

---

## Task 6: README 更新（英語 + 日本語）

両 README から `-s` 言及 3 箇所をそれぞれ更新する。

**Files:**

- Modify: `README.md:36, 48, 118`
- Modify: `docs/README.ja.md:36, 48, 118`

- [ ] **Step 1: 現状の該当行を確認**

  Run: `grep -n '\-s\b\|--superpowers\|superpowers' README.md docs/README.ja.md`
  Expected: 各ファイルに 3 行ずつヒット（features bullet・quick start example・optional dependency）。

- [ ] **Step 2: `README.md` の Features 第 3 項目（line 36 付近）を削除**

  Before:

  ```markdown
  - 🦸 **"Design first" startup** — `-s` tells claude to follow the brainstorming → writing-plans → executing-plans flow (prompts to install the superpowers plugin if missing)
  ```

  → 該当 1 行を削除する（前後の改行整理含む）。理由: superpowers の自動発火は plugin 側 (`using-superpowers` skill) が担うため、ccw 固有機能ではなくなる。

- [ ] **Step 3: `README.md` の Quick Start `ccw -s` 行（line 48 付近）を削除**

  Before:

  ```bash
  ccw -s                                    # new worktree + superpowers preamble
  ```

  → 該当 1 行を削除。

- [ ] **Step 4: `README.md` の Optional dependency 記述（line 118 付近）を更新**

  Before:

  ```markdown
  - *(optional)* [superpowers](https://github.com/obra/superpowers) plugin — auto-checked when `-s` is used
  ```

  After:

  ```markdown
  - *(required for development)* [superpowers](https://github.com/obra/superpowers) plugin — declared in `.claude/settings.json`; Claude Code prompts to install it on first launch in this repo
  ```

- [ ] **Step 5: `docs/README.ja.md` を同じ趣旨で更新**

  - line 36 付近の features 項目（`🦸 "設計してから書く" 流儀で起動 — -s で...` を含む 1 行）を削除。
  - line 48 付近の `ccw -s` 行を削除。
  - line 118 付近を以下に書き換える:

    ```markdown
    - *(開発時に必須)* [superpowers](https://github.com/obra/superpowers) プラグイン — `.claude/settings.json` で宣言済み。リポジトリ内での初回 `claude` 起動時に Claude Code がインストールを促す
    ```

- [ ] **Step 6: 残存する `-s` / `--superpowers` 参照がないことを確認**

  Run: `grep -n '\-s\b\|--superpowers' README.md docs/README.ja.md | grep -v '^Binary'`
  Expected: 0 matches。

- [ ] **Step 7: Commit**

  ```bash
  git add README.md docs/README.ja.md
  git commit -m "docs(readme): -s 廃止に伴い両 README を更新"
  ```

---

## Task 7: 最終検証と PR 作成

仕上げ。spec の Verification 節の残項目を消化し、PR を立てる。

**Files:** なし（検証のみ）

- [ ] **Step 1: ビルド・テスト・lint をクリーンに通す**

  Run: `go build ./...`
  Expected: 0 exit。

  Run: `go test ./...`
  Expected: 全 pass。

  Run: `golangci-lint run ./...`（リポジトリの定常 lint。設定済み: `.golangci.yml`）
  Expected: 0 issues。lint が見つけた warning は本タスクの差分に起因するものを修正する。

- [ ] **Step 2: 残存する superpowers / `-s` 参照がコードベース全体に無いことを確認**

  Run: `grep -rn '\-s\b\|--superpowers\|superpowers\.\|internal/superpowers' --include='*.go' --include='*.yaml' --include='*.md' --include='*.json' . 2>&1 | grep -v 'docs/superpowers/\|\.git/\|\.claude/worktrees/' | head -40`

  Expected: ヒットは `.claude/settings.json` の `enabledPlugins` キー、および本 plan / spec / 過去の docs/superpowers/plans 配下のみ。コード・i18n・README には残らない。

- [ ] **Step 3: 手元で `ccw -s` を実行し pflag エラーで終了することを確認**

  Run:

  ```bash
  go build -o /tmp/ccw-removed-s ./cmd/ccw
  cd "$(mktemp -d)" && git init -q
  /tmp/ccw-removed-s -s; echo "exit=$?"
  ```

  Expected: stderr に「unknown shorthand flag: 's'」相当のエラー、exit code が `2`。

- [ ] **Step 4: 手元で `ccw --help` を実行し `-s` が消えていることを目視確認**

  Run: `/tmp/ccw-removed-s --help`
  Expected: ヘルプ出力に `-s` / `--superpowers` の行が無く、`-y` の説明から `-s plugin install` の括弧書きも消えている。

- [ ] **Step 5: ブランチを push して PR を作成**

  ```bash
  git push -u origin chore/remove-superpowers-flag
  gh pr create --title "chore: -s / --superpowers flag を削除し .claude/settings.json で宣言" --body "$(cat <<'EOF'
  ## Summary
  - `-s` / `--superpowers` flag を即時削除し、Claude Code 標準機能と重複していた preamble 注入と auto-install ロジックを撤去
  - `.claude/settings.json` で `superpowers@claude-plugins-official` を `enabledPlugins` 宣言。clone した開発者は `claude` 起動時にインストールを促される
  - `internal/superpowers/` パッケージを丸ごと削除、claude pkg の preamble 引数も整理
  - 両 README と i18n help テキストから `-s` 言及を除去

  ## Spec / Plan
  - Spec: `docs/superpowers/specs/2026-04-26-remove-s-flag-design.md`
  - Plan: `docs/superpowers/plans/2026-04-26-remove-s-flag.md`

  ## Breaking change
  `-s` / `--superpowers` を渡すと pflag の unknown flag エラーで終了します。代替は `claude` 起動時の `enabledPlugins` プロンプト、もしくは `/plugin install superpowers@claude-plugins-official` を手動実行してください。初期プロンプトを渡したい場合は `ccw -n -- "<text>"` をそのまま使えます。

  ## Test plan
  - [ ] `go build ./...`
  - [ ] `go test ./...`
  - [ ] `golangci-lint run ./...`
  - [ ] 手元で `ccw -s` が pflag エラーで exit 2
  - [ ] 手元で `ccw --help` の出力から `-s` 行が消えていることを確認

  🤖 Generated with [Claude Code](https://claude.com/claude-code)
  EOF
  )"
  ```

  Expected: PR URL が出力される。

---

## Self-Review

- **Spec coverage:** D1（`-s` 即時削除）→ Task 3 + Task 7 Step 3。D2（`internal/superpowers/` パッケージ丸ごと削除）→ Task 5。D3（claude.go の preamble 引数撤去）→ Task 4。D4（`.claude/settings.json` コミット）→ Task 1 Step 1。D5（`.gitignore` に `.claude/settings.local.json` 追加）→ Task 1 Step 2。D6（README 3 箇所更新 + readme-sync 規約での同期）→ Task 6。Verification（go test/build, `ccw -s` 動作確認）→ Task 7 Step 1-4。Migration ノート（CHANGELOG / リリースノート）→ PR 説明（Task 7 Step 5）でカバー。CHANGELOG ファイルがリポジトリにあるかは未確認だが、現時点で `find . -maxdepth 2 -name 'CHANGELOG*'` の結果が空であれば、リリースノートは GitHub Release / PR description で代替する（goreleaser が PR/commit から生成する想定）。
- **Placeholder scan:** TBD / TODO / "implement later" などのプレースホルダなし。各 step に実コードまたは具体コマンドを掲載済み。
- **Type consistency:** `BuildNewArgs(name string, extra []string) []string` のシグネチャは Task 4 Step 1（テスト更新）→ Step 3（実装更新）→ Step 4（呼び出し更新）で一貫。`Flags` から `Superpowers` を消す手順は Task 3 Step 1（テスト先行削除）→ Step 3-5（フィールドと flag 定義削除）→ Step 10（`flags.Superpowers` 参照削除）で一貫。

---

## Execution Handoff

**Plan complete and saved to `docs/superpowers/plans/2026-04-26-remove-s-flag.md`. Two execution options:**

1. **Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.
2. **Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

**Which approach?**
