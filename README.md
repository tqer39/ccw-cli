# ccw

![ccw-cli — Claude Code x worktree](docs/assets/header.png)

Claude Code worktree launcher — isolates each session in its own git worktree with opinionated defaults.

## Features

- Launches Claude Code (`claude`) in a fresh git worktree each time (no state leakage)
- Auto permission mode by default (no approval prompts)
- Optional superpowers workflow injection (`brainstorming → writing-plans → executing-plans`) via `-s`
- Interactive picker for leftover worktrees: resume, delete, bulk delete, or start new
- Colored status badges, `↑N ↓M ✎N` indicators, and PR number / title (when `gh` is installed)
- Bulk delete via picker menu or the `--clean-all` CLI flag
- Version display (`-v`)
- Pass-through of native `claude` arguments after `--`

## Requirements

- `git`
- Claude Code CLI (`claude`) — if missing at launch, ccw offers to install it via npm / brew interactively
- (Optional) superpowers plugin — auto-checked when `-s` is used
- (Optional) `gh` CLI — enables PR number / title in the picker (picker still works without it)

## Install

### Homebrew (recommended)

```bash
brew install tqer39/tap/ccw
```

### From source

```bash
git clone https://github.com/tqer39/ccw-cli ~/workspace/tqer39/ccw-cli
cd ~/workspace/tqer39/ccw-cli
go build -o ~/.local/bin/ccw ./cmd/ccw

# Make sure ~/.local/bin is on PATH; if not, add it to your shell rc:
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
```

## Usage

```text
Usage: ccw [options] [-- <claude-args>...]

Options:
  -n, --new            Always start a new worktree (skip the picker)
  -s, --superpowers    Inject the superpowers preamble (implies -n)
  -v, --version        Show version info
  -h, --help           Show this help

Bulk delete:
      --clean-all        Bulk delete mode
      --status=<filter>  all | pushed | local-only | dirty (default: all)
      --force            Allow --force removal of dirty worktrees
      --dry-run          Print targets and exit without deleting
  -y, --yes              Skip the confirmation prompt

Arguments after `--` are forwarded to `claude` verbatim.
```

### Examples

```bash
ccw                                       # Pick an existing worktree, or start a new one
ccw -n                                    # Start a new worktree (skip picker)
ccw -s                                    # New worktree + superpowers preamble
ccw -- --model claude-opus-4-7            # Pass-through args to claude
ccw --clean-all --status=pushed --dry-run # Preview what would be deleted
ccw --clean-all --status=all --force -y   # Delete every worktree (incl. dirty) without prompting
```

## Worktree picker

Running `ccw` with no arguments scans `.claude/worktrees/` for leftover worktrees and shows the picker:

![ccw picker demo](docs/assets/picker-demo.gif)

```text
> [PUSHED] feat/login              ↑0 ↓0       #42 open "feat: add login"
    ~/repo/.claude/worktrees/feat-login
  [LOCAL]  feat/picker              ↑3 ↓1       (no PR)
    ~/repo/.claude/worktrees/feat-picker
  [DIRTY]  chore/cleanup            ↑0 ↓2 ✎5   #43 draft "chore: cleanup"
    ~/repo/.claude/worktrees/chore-cleanup
  🗑️  [delete all]
  🧹  [clean pushed]
  ☑️  [custom select]
  ➕  [new]
  🚪  [quit]
```

| Badge | Meaning |
|---|---|
| `[PUSHED]` (green) | Clean, upstream tracked, 0 commits ahead |
| `[LOCAL]` (yellow) | No upstream, or ahead of upstream |
| `[DIRTY]` (red) | Working tree has uncommitted changes |

Indicators:

- `↑N ↓M` — commits ahead / behind upstream
- `✎N` — number of dirty files (only shown for `dirty`)
- `#N state "title"` — PR number / state / title (requires `gh`)

Selecting a worktree opens a `resume` / `delete` / `back` submenu. `[delete all]` / `[clean pushed]` / `[custom select]` are shortcuts for removing multiple worktrees at once. When dirty worktrees are included, the confirmation dialog offers `y` (force), `s` (skip dirty), or `N`.

If `gh` is not installed, the PR column is hidden and the footer shows the hint `💡 gh があったら PR 名も出せます`. If `gh` is installed but the call fails (rate limit, network error, etc.), the PR column is hidden silently without any hint.

## Environment variables

| Variable | Effect |
|---|---|
| `NO_COLOR=1` | Disable colored output |
| `CCW_DEBUG=1` | Verbose debug logging |

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | User error / cancellation (missing dependency, outside a git repo, user declined, etc.) |
| other | Forwarded verbatim from the `claude` command |

## Development

### Prerequisites

Install the pre-commit toolchain used by lefthook:

```bash
brew install lefthook yamllint actionlint
lefthook install
```

`markdownlint-cli2` and `renovate-config-validator` are fetched on demand via `npm exec`, so no upfront install is required (Node.js / npm must be available).

### Hooks

- `check-added-large-files`: block files over 512KB
- `detect-private-key`: catch leaked private keys
- `gofmt` / `golangci-lint`: Go files
- `yamllint` / `actionlint`: YAML and GitHub Actions workflows
- `markdownlint-cli2`: Markdown
- `renovate-config-validator`: `renovate.json5` only

Hooks run in parallel, so the total time is the slowest hook, not the sum.

### Build & test

```bash
go build ./cmd/ccw
go test ./...
go vet ./...
```

## Future work

- Shell completion (bash / zsh)
- Windows support

## License

MIT
