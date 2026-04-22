# Phase 1: Go スケルトン設計

- 作成日: 2026-04-23
- 作成者: tqer39 / Claude Code (brainstorming session)
- ステータス: draft (user review 待ち)
- 親 spec: `docs/superpowers/specs/2026-04-23-go-rewrite-and-brew-design.md`

## 目的

親 spec §段階移行プラン の **Phase 1** を実装可能な単位まで詳細化する。Go 版 `ccw` の最小スケルトン（`-h` / `-v` のみ動作）を構築し、後続フェーズでロジックを埋めていく土台を作る。

## スコープ

- **動かす**: `ccw -h` / `ccw -v`
- **動かさない**: `-n` / `-s` / 無引数（picker）は stderr に "Phase 1 スケルトンのため未実装" を出力し `exit 1`
- **併存**: bash 版 `bin/ccw` は温存。ユーザーの推奨は引き続き bash 版

## 非目標

- `internal/gitx` / `worktree` / `claude` / `superpowers` / `picker` の実装（Phase 2 / 3）
- goreleaser / Homebrew formula（Phase 4）
- `--update` / `--uninstall` の削除アナウンス（Phase 6）
- `bin/ccw` (bash 版) への変更

## 採用方針サマリ

| 項目 | 決定 |
|---|---|
| スケルトン動作範囲 | A: `-h` / `-v` のみ。他は "not implemented" で `exit 1` |
| Go バージョン管理 | `mise.toml` + `go.mod` の `go` directive（`.tool-versions` は使わない） |
| CI | `.github/workflows/ci.yml` を新設。既存 `lint.yml`（bash 版向け）は併存 |
| タスクランナー | Makefile |
| `ui` パッケージ | 関数群 + パッケージレベル状態（`SetWriter` でテスト差し替え） |
| `golangci-lint` | 厳し目構成（default + `goimports` / `misspell` / `gocritic` / `revive` / `errorlint` / `wrapcheck` / `gocyclo`） |
| `version` 注入 | ldflags（`-X internal/version.Version=...`） |
| `cli.Parse` 失敗時 | `error` 返却 → main で整形出力 + `os.Exit(2)`（pflag は `ContinueOnError`） |

## 作成するファイル

```text
cmd/ccw/main.go                  # エントリポイント
internal/cli/parse.go            # pflag + Flags struct + Parse(argv) (Flags, error)
internal/cli/parse_test.go       # テーブルテスト
internal/cli/help.go             # PrintHelp / usage 文字列
internal/ui/ui.go                # InitColor, Info/Warn/Error/Success/Debug, EnsureTool, SetWriter
internal/ui/ui_test.go
internal/version/version.go      # var Version/Commit/Date + String()
internal/version/version_test.go
go.mod / go.sum
mise.toml                        # [tools] go = "1.23"
Makefile                         # build / test / lint / run / tidy / clean
.golangci.yml
.github/workflows/ci.yml         # go-lint / go-test / go-build 3 job
```

既存ファイルへの変更:

- `.gitignore`: 既に Go 向けエントリが入っているため追加不要（Phase 0 で完了済み）
- `README.md`: Phase 1 段階では変更しない（Go 版の install 手順は Phase 5 で追加）

## モジュール設計

### `go.mod`

- module: `github.com/tqer39/ccw-cli`
- `go 1.23`
- 依存: `github.com/spf13/pflag` のみ
  - `mattn/go-isatty` は追加しない。TTY 判定は `golang.org/x/term.IsTerminal`（標準ライブラリ準拠）で行う

### `internal/version`

```go
package version

import "fmt"

var (
    Version = "dev"
    Commit  = "none"
    Date    = "unknown"
)

func String() string {
    return fmt.Sprintf("ccw %s (commit: %s, built: %s)", Version, Commit, Date)
}
```

- ldflags フォーマット: `-X github.com/tqer39/ccw-cli/internal/version.Version=... -X ....Commit=... -X ....Date=...`
- Phase 4 の goreleaser 設定とキーが一致する

### `internal/ui`

パッケージレベル状態:

```go
var (
    stdout       io.Writer = os.Stdout
    stderr       io.Writer = os.Stderr
    colorEnabled bool
)
```

公開 API:

- `InitColor()`: `NO_COLOR` env と stderr が TTY か（`golang.org/x/term.IsTerminal(int(os.Stderr.Fd()))`）を一度評価して `colorEnabled` を確定
- `SetWriter(out, err io.Writer)`: テスト用。呼ぶと color は強制 off（TTY でないため自然に off になるが明示）
- `Info(format string, args ...any)` / `Warn` / `Error` / `Success`: prefix + ANSI（色 ON 時） + `fmt.Fprintf(stderr, ...)` ※ `Info` のみ stdout
- `Debug(format string, args ...any)`: `os.Getenv("CCW_DEBUG") == "1"` のときだけ stderr に出す
- `EnsureTool(name, installHint string)`: `exec.LookPath(name)` 失敗なら `Error("required tool not found: %s. %s", name, installHint)` + `os.Exit(1)`

色プレフィックス（bash 版準拠）:

- Info: なし（白）
- Warn: `⚠` 黄
- Error: `✖` 赤
- Success: `✓` 緑
- Debug: `[debug]` 灰

### `internal/cli`

```go
package cli

type Flags struct {
    Help        bool
    Version     bool
    NewWorktree bool     // -n
    Superpowers bool     // -s（Parse 内で暗黙に NewWorktree = true）
    Passthrough []string // `--` 以降
}

func Parse(argv []string) (Flags, error)
func PrintHelp(w io.Writer)
```

Parse の実装方針:

1. `argv` を `--` で手動分割（前半: pflag に渡す / 後半: `Passthrough` にそのまま格納）
2. `pflag.NewFlagSet("ccw", pflag.ContinueOnError)` を作成
3. `fs.SetOutput(io.Discard)` で pflag 自身の stderr 出力を抑止（エラーは呼び出し元で整形）
4. `-h / --help`, `-v / --version`, `-n / --new`, `-s / --superpowers` を定義
5. `--update` / `--uninstall` は **登録しない**（親 spec §CLI 表面 の方針）→ 渡されると "unknown flag" エラー
6. `fs.Parse(pre)` → エラーなら wrap して返す
7. 位置引数は想定外。`fs.Args()` が非空ならエラー（bash 版も同様）
8. `-s` 指定時は `Flags.NewWorktree = true` を強制

`PrintHelp` は `cmd/ccw/main.go` から呼ばれ、bash 版 README の Usage と同じフォーマットを出す。

### `cmd/ccw/main.go`

```go
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

## ビルド / タスク / CI

### `Makefile`

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

### `.golangci.yml`

- v2 形式
- enable: `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`, `gosimple`, `gofmt`, `goimports`, `misspell`, `gocritic`, `revive`, `errorlint`, `wrapcheck`, `gocyclo`
- `settings.gocyclo.min-complexity: 15`
- `settings.wrapcheck.ignoreSigs`: `os/exec.Command` 系を除外（Go 慣例）
- `settings.revive.rules`: default preset のみ（Phase 1 では細かい rule 追加しない）

### `.github/workflows/ci.yml`（新設）

`on: [push to main, pull_request]`。3 job（並列）:

1. **go-lint**: `actions/checkout` → `actions/setup-go`(go-version-file: `go.mod`) → `golangci/golangci-lint-action`
2. **go-test**: `actions/checkout` → `actions/setup-go`(go-version-file: `go.mod`) → `go test ./... -race -coverprofile=coverage.out` → `actions/upload-artifact` で coverage 保存
3. **go-build**: `actions/checkout` → `actions/setup-go`(go-version-file: `go.mod`) → `go build ./cmd/ccw` で smoke

Actions の SHA ピンは参考リポ（`terraform-github` など）と同じ粒度で固定し、以降は Renovate が追従。

### `lefthook.yml` への影響

Phase 0 で `gofmt` / `golangci-lint` は `glob: "*.go"` で既に定義済み。Phase 1 で Go ファイルが現れた時点で自動的に作動する。追加設定不要。

## テスト戦略（Phase 1 範囲）

### `internal/cli/parse_test.go`

テーブルテストで以下を網羅:

- `[]` → 全フラグ false、Passthrough 空
- `["-h"]` → Help = true
- `["--help"]` → Help = true
- `["-v"]` / `["--version"]` → Version = true
- `["-n"]` / `["--new"]` → NewWorktree = true
- `["-s"]` → Superpowers = true, NewWorktree = true
- `["-n", "-s"]` → 両方 true
- `["--"]` → Passthrough = []
- `["--", "foo", "bar"]` → Passthrough = ["foo", "bar"]
- `["-n", "--", "--model", "claude-opus-4-7"]` → NewWorktree = true, Passthrough = ["--model", "claude-opus-4-7"]
- `["-s", "--", "--resume"]` → Superpowers+NewWorktree = true, Passthrough = ["--resume"]
- `["--unknown"]` → error
- `["--update"]` → error（unknown flag）
- `["--uninstall"]` → error
- `["positional"]` → error（位置引数禁止）

### `internal/version/version_test.go`

- `String()` のデフォルト出力が `"ccw dev (commit: none, built: unknown)"` になる
- 変数を書き換え後の文字列が期待通りに差し替わる

### `internal/ui/ui_test.go`

- `SetWriter(out, err)` + `Info` / `Warn` / `Error` / `Success` の出力を buffer で検証
- `NO_COLOR` 設定時・未設定時で ANSI シーケンス有無を確認
- `Debug` は `CCW_DEBUG=1` のときのみ出力されることを検証
- `EnsureTool` は `os.Exit` 経由のため Phase 1 ではテスト対象外

## エラーハンドリング方針（Phase 1 範囲）

- パースエラー: `Error` + help 表示 + `exit 2`
- 未実装フラグ組み合わせ（`-n` / `-s` / 無引数）: `Error` + `exit 1`
- `ui.EnsureTool` 失敗: `exit 1`（親 spec 準拠。Phase 1 では呼ばないが実装のみ入れる）

## 成功条件

1. `go build ./cmd/ccw` が warnings なく成功
2. `./ccw -h` が bash 版と同じ usage を出力して `exit 0`
3. `./ccw -v` が `ccw dev (commit: none, built: unknown)` を出力して `exit 0`
4. `./ccw -n` / `./ccw -s` / `./ccw` は "Phase 1 スケルトンのため…" を出して `exit 1`
5. `go test ./... -race` が通る（cli / version / ui 全テスト PASS）
6. `golangci-lint run` が警告 0 で通る
7. `.github/workflows/ci.yml` の 3 job（go-lint / go-test / go-build）が PR で PASS する
8. `lefthook run pre-commit --all-files` が PASS する

## オープン項目（実装時に確定）

- `actions/setup-go` / `golangci-lint-action` の SHA ピン値（実装時点の最新 stable を採用し Renovate に委ねる）
- `golangci-lint` v2 YAML の細部（`settings.*` の具体キー名が v2 で変わる箇所があれば調整）
- `wrapcheck` の `ignoreSigs` 具体エントリ（`os/exec` 以外にも除外が必要なら追記）
