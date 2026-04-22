# ccw-cli: Go 書き換え + Homebrew 配布 + リポジトリ品質基盤整備

- 作成日: 2026-04-23
- 作成者: tqer39 / Claude Code (brainstorming session)
- ステータス: draft (user review 待ち)

## 目的

現状 bash 約 840 行で実装されている `bin/ccw` を Go に書き換え、Homebrew の独自 tap 経由で配布する。合わせて、リポジトリの品質基盤（Renovate / pre-commit フック / auto-assign）を整備し、Go プロジェクトとして自走できる状態にする。

## 背景

- 現 `bin/ccw` は bash + `tput` + `read -rsn1` で TUI を自作しており、依存は git / bash / tput / claude のみで動くが、macOS 以外での可搬性や機能拡張の容易さに限界がある
- 手動 symlink install しか想定していないため、アップデート体験（`ccw --update`）や配布の敷居が高い
- 単一バイナリ化で macOS/Linux × amd64/arm64 を跨ぎ、`brew install` で 1 コマンド化したい

## 非目標

- Windows ネイティブサポート（PowerShell / cmd）。`README` の Future work に残す
- bash 版の完全削除。段階移行として温存
- homebrew-core への収録（将来検討。今回は独自 tap のみ）
- `--update` / `--uninstall` の Go 版実装（Homebrew に委ねる）

## 採用方針（サマリ）

| 項目 | 決定 |
|---|---|
| 既存 bash 版との関係 | 段階移行（B）: `bin/ccw` は温存し、Go 版を並行開発 |
| 配布経路 | 独自 tap `tqer39/homebrew-tap`（A） |
| `--update` / `--uninstall` | Go 版から削除（B）。`brew upgrade ccw` / `brew uninstall ccw` を案内 |
| TUI ライブラリ | `charmbracelet/bubbletea` + `lipgloss`（A） |
| git 操作 | `os/exec` で `git` コマンド呼び出し（A） |
| CLI パーサ | `spf13/pflag`（short/long alias、サブコマンド不要） |
| リリース自動化 | `goreleaser` + GitHub Actions（tag push トリガ） |

## アーキテクチャ

### モジュール構成

```
github.com/tqer39/ccw-cli

cmd/
  ccw/
    main.go                 # エントリポイント（version 埋め込みのみ）
internal/
  cli/
    parse.go                # pflag 定義 + flag struct
    run.go                  # main ルータ（help/version/main flow の分岐）
  gitx/
    git.go                  # exec ラッパ (run, runOutput)
    worktree.go             # worktree list/add/remove/status
    repo.go                 # resolve_main_repo / require_git_repo
  worktree/
    list.go                 # .claude/worktrees/ 配下のみ抽出 + WorktreeInfo
    flags.go                # pushed / local-only / dirty 判定
  picker/
    model.go                # bubbletea model（list + submenu の state machine）
    view.go                 # lipgloss styling
    update.go               # key handling
  claude/
    launch.go               # claude --permission-mode auto [--worktree] [--] <preamble>
    install.go              # claude 不在時の npm/brew 誘導
  superpowers/
    detect.go               # ~/.claude/plugins/cache/*/superpowers 検出
    preamble.go             # 埋め込み preamble 文字列
    gitignore.go            # docs/superpowers/ 追記フロー
  ui/
    color.go                # NO_COLOR 対応 + info/warn/error/success
    debug.go                # CCW_DEBUG=1 で詳細ログ
  version/
    version.go              # var Version/Commit/Date (ldflags で注入)
.github/
  workflows/
    ci.yml                  # lint + test
    release.yml             # tag push → goreleaser
    auto-assign.yml
  auto_assign.yml
.goreleaser.yaml
.golangci.yml
lefthook.yml
renovate.json5
.gitignore                  # Go 向けに更新
go.mod / go.sum
```

### データフロー

1. `main` → `cli.Parse(os.Args)` → `Flags` struct
2. `help` / `version` は即 return
3. `require_git_repo` → `resolve_main_repo`
4. `check_base_tool_or_exit("git")` / `claude.EnsureInstalled()`
5. `cd main_repo` 相当（Go では `os.Chdir`）+ `git remote set-head origin -a`（失敗無視）
6. `-s` なら `superpowers.EnsureInstalled()` → `superpowers.EnsureGitignore()` → `preamble := superpowers.Preamble()`
7. `-n` or `-s` なら `claude.LaunchNew(preamble, passthroughArgs)` → `exit`
8. それ以外は `picker.Run(mainRepo)`:
   - ccw 配下 worktree を列挙（空なら直接 new）
   - bubbletea で矢印選択 → resume/delete/back サブメニュー
   - resume: `cd wt && claude --permission-mode auto [passthrough]`
   - delete: `git worktree remove [--force]`

### CLI 表面

bash 版と同一（`--update` / `--uninstall` のみ削除）:

```
Usage: ccw [options] [-- <claude-args>...]

Options:
  -n, --new            新規 worktree で起動（picker スキップ）
  -s, --superpowers    superpowers preamble を注入して起動（暗黙に -n）
  -v, --version        バージョン情報を表示
  -h, --help           ヘルプを表示

Environment:
  NO_COLOR=1           カラー出力を無効化
  CCW_DEBUG=1          詳細ログ出力

Exit codes:
  0  success
  1  user error / cancellation
  *  passthrough from `claude`
```

`--update` / `--uninstall` は pflag にも登録しない（完全削除）。結果として渡すと「unknown flag」として exit 2 で即終了する。README とリリースノートで `brew upgrade ccw` / `brew uninstall ccw` への移行を明記する。

### エラーハンドリング方針

- 「ユーザー操作ミス / 依存欠落 / キャンセル」→ `exit 1` + 日本語メッセージ（stderr）
- 「git / claude の exec 失敗」→ そのコマンドの exit code を透過
- パニックは使わない。`fmt.Errorf` で wrap して top-level で整形表示

## 機能パリティ表

| bash 関数 | Go 移行先 | 備考 |
|---|---|---|
| `init_color` | `ui.InitColor()` | `os.Stdout` が terminal か + `NO_COLOR` 参照 |
| `print_help` / `print_version` | `cli.PrintHelp()` / `cli.PrintVersion()` | version は ldflags |
| `run_update` | **削除** | README で `brew upgrade ccw` 案内 |
| `run_uninstall` | **削除** | README で `brew uninstall ccw` 案内 |
| `resolve_main_repo` | `gitx.ResolveMainRepo()` | `git rev-parse --git-common-dir` |
| `require_git_repo` | `gitx.RequireRepo()` | |
| `check_base_tool_or_exit` | `ui.EnsureTool(name, url)` | `exec.LookPath` |
| `launch_claude_new` | `claude.LaunchNew(preamble, extra)` | |
| `check_superpowers` | `superpowers.EnsureInstalled()` | `filepath.Glob` |
| `ensure_gitignore` | `superpowers.EnsureGitignore(root)` | 末尾改行保証ロジックも移植 |
| `list_ccw_worktrees` | `worktree.List(root)` → `[]WorktreeInfo` | porcelain パース |
| `worktree_flags` | `worktree.Status(path)` → enum | `pushed` / `localOnly` / `dirty` |
| `worktree_icon` | `picker` 内で lipgloss アイコンへ写像 | |
| `tui_select` / `run_picker_flow` | `picker.Run(root)` | bubbletea model |
| `resume_worktree` | `picker.resume` + `claude.Resume(wt, extra)` | |
| `delete_worktree` | `picker.delete` + `gitx.RemoveWorktree(path, force)` | |
| `superpowers_preamble` | `superpowers.Preamble()` | `embed.FS` で同梱 |
| `check_claude` | `claude.EnsureInstalled()` | npm/brew 選択 |
| `parse_args` | `cli.Parse(os.Args)` | pflag |

## リリース & 配布

### goreleaser

`.goreleaser.yaml`（要点）:

- `builds`: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
- `ldflags`: `-s -w -X github.com/tqer39/ccw-cli/internal/version.Version={{.Version}} -X .Commit={{.Commit}} -X .Date={{.Date}}`
- `archives`: `tar.gz`（binary + LICENSE + README）
- `brews`: `tqer39/homebrew-tap` へ formula を自動 push
- `changelog`: `use: github` で PR / commit の差分生成

### Homebrew tap

- 新規リポ `tqer39/homebrew-tap` を作成（空の `Formula/` ディレクトリのみで初期化）
- goreleaser が tag release ごとに `Formula/ccw.rb` を更新
- ユーザーインストール: `brew install tqer39/tap/ccw`
- ユーザー更新: `brew upgrade ccw`
- ユーザー削除: `brew uninstall ccw`

### リリースワークフロー

`.github/workflows/release.yml`:

- trigger: `push: tags: ['v*']`
- job: ubuntu-latest, `goreleaser-action` で `--clean` 実行
- secrets: `HOMEBREW_TAP_GITHUB_TOKEN`（tap リポへの push 用 PAT）

### バージョニング

- SemVer（`v0.1.0` から開始予定）
- 最初の Go 版リリースは `v1.0.0` ではなく `v0.1.0`。`v1.0.0` は bash 版と機能同等になり安定運用が確認できた時点で打つ
- bash 版はタグ打ちしていないため、旧版との比較は commit hash 参照

## リポジトリ品質基盤

### .gitignore 更新

`/gitignore` スキルを使って Go + macOS + editors 相当を追記。追加項目の最低ライン:

- `ccw`（ビルド成果物）
- `dist/`（goreleaser 出力）
- `coverage.out` / `*.out`
- `*.test` / `*.prof`
- 既存の OS / editors / logs は維持

### Renovate

`renovate.json5`（参考: `terraform-github/renovate.json5`）:

```json5
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": [
    "github>tqer39/renovate-config"
  ]
}
```

対応対象:

- `go.mod`（Go モジュール）
- GitHub Actions（`.github/workflows/*.yml` の `uses:` ピン）
- Homebrew tap formula（`tqer39/homebrew-tap` 側にも同じ設定を置く想定。今回のリポの spec 外）

### pre-commit: lefthook

`lefthook.yml`（参考: `terraform-github/lefthook.yml` を Go プロジェクトに合わせて調整）:

採用コマンド（`pre-commit: parallel: true`）:

- **セーフティ**: `check-added-large-files`, `detect-private-key`, `betterleaks`, `end-of-file-fixer`, `mixed-line-ending`, `trailing-whitespace`
- **Go**: `gofmt -l -d`（diff があれば fail）, `golangci-lint run`（差分ファイル対象）
- **Shell**: `shellcheck`（`bin/ccw` + `scripts/*.sh`）, `shfmt -d -i 2 -ci -bn`
- **YAML / Actions**: `yamllint`（`.github/workflows/*.yml` 含む）, `actionlint`
- **Markdown**: `markdownlint-cli2 --fix`, `textlint`, `cspell lint`
- **Renovate**: `renovate-config-validator`（`renovate.json5` のみ）

`stage_fixed: true` は末尾改行 / 改行統一 / 末尾空白 / markdownlint に付与し、自動修正をそのまま staging に残す。

除外方針:

- `terraform-fmt` / `biome-format`: ccw-cli は Terraform / JSON を生成しないため除外
- `check-json`: 設定ファイルは `renovate.json5`（JSON5）のみ。必要時に追加
- `detect-aws-credentials`: ccw-cli スコープでは過剰。除外

### auto-assign

`.github/auto_assign.yml`（参考そのまま）:

```yaml
# see: https://github.com/kentaro-m/auto-assign-action
addAssignees: author
```

`.github/workflows/auto-assign.yml`（参考そのまま。SHA ピンは Renovate に任せる）:

- trigger: `pull_request: types: [opened, ready_for_review]`
- concurrency group で二重起動防止
- permissions: `contents: read`, `pull-requests: write`
- assignee 空の場合のみ assign

### CI (`.github/workflows/ci.yml`)

Job 構成:

1. `lint`:
   - `actions/checkout@<sha>`（Renovate 管理）
   - `actions/setup-go@<sha>` (`.tool-versions` or `go.mod` を参照)
   - `golangci-lint-action`
   - 既存の shellcheck / shfmt / bats は bash 版向けに継続（別 job または別 workflow として残す）
2. `test`:
   - `go test ./... -race -coverprofile=coverage.out`
   - TUI テストは `teatest`（bubbletea 公式テストヘルパ）で矢印・サブメニューをシミュレート

## 段階移行プラン

1. **Phase 0（本 spec 対応）**: Go 化は未着手。リポジトリ品質基盤（gitignore / renovate / lefthook / auto-assign）先行導入
2. **Phase 1**: `go.mod` 作成 + `cmd/ccw/main.go` スケルトン + `cli` + `version` + `ui`
3. **Phase 2**: `gitx` + `worktree` + `claude` + `superpowers`（bash 版とパリティ確保）
4. **Phase 3**: `picker`（bubbletea）導入 + teatest テスト
5. **Phase 4**: `.goreleaser.yaml` + release workflow + `tqer39/homebrew-tap` 初期化
6. **Phase 5**: `v0.1.0` タグ → 独自 tap へ formula 自動反映 → 実機 `brew install tqer39/tap/ccw` で動作確認
7. **Phase 6**: README 更新（bash 版からの移行手順を明記）。bash 版は `bin/ccw` のまま温存し、新規推奨は brew 版とする

bash 版削除のタイミングは別 spec で扱う（本 spec の非目標）。

## テスト戦略

- **unit**: `gitx` / `worktree` / `superpowers` / `ui` はテーブルテスト。`exec` 呼び出しは `gitx` 内で interface 経由にし、fake runner で境界を切る
- **integration**: `t.TempDir()` 内で `git init` → 本物の git を呼び出してシナリオ検証（CI 上の `git` で動く）
- **TUI**: `teatest`（bubbletea 公式）で矢印送信 → 画面 snapshot 一致確認。resume / delete / back / new / quit の各遷移を網羅
- **CLI**: `cli.Parse` のテーブルテスト（`-n`, `-s`, `--`, 不明フラグ、競合組み合わせ）
- **smoke (CI)**: tag 直前に `go run ./cmd/ccw -h` と `-v` を実行し出力差分検査

## オープン項目（実装時に確定）

- `goreleaser` の `archives` 命名規則（`{{ .Binary }}-{{ .Version }}-{{ .Os }}-{{ .Arch }}.tar.gz`）
- `golangci.yml` の有効リンタ集合（最初は `default` + `errcheck` / `gosimple` / `staticcheck` / `unused` / `ineffassign` から）
- `lefthook` の Go 系コマンドで `go test` を含めるか（パフォーマンス次第。初期は除外）
- Homebrew formula の `depends_on "git"`（macOS は Xcode CLT で事足りるため不要の可能性）

## 成功条件

1. `brew install tqer39/tap/ccw` で macOS (Intel/Apple Silicon) と Linux (amd64/arm64) にインストール可能
2. `ccw -h` / `ccw -v` / `ccw -n` / `ccw -s` / `ccw` (picker) が bash 版と同等の挙動
3. `brew upgrade ccw` でバージョン更新できる
4. pre-commit (`lefthook`) が PR 作成前に Go / shell / markdown / yaml / workflow を全部ブロックする
5. Renovate が Go 依存と GitHub Actions を自動 PR 化する
6. auto-assign で PR 作成時に author が assignee に設定される
7. `go test ./...` が CI で通る（-race 付き）
