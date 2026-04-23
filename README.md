<div align="center">

![ccw-cli — Claude Code x worktree](docs/assets/header.png)

**Launch [Claude Code](https://docs.claude.com/claude-code) in an isolated git worktree — no state leakage, no switching headaches.**

[![Go](https://img.shields.io/badge/go-1.25-00ADD8?logo=go&logoColor=white)](go.mod)
[![Release](https://img.shields.io/github/v/release/tqer39/ccw-cli?logo=github)](https://github.com/tqer39/ccw-cli/releases)
[![License](https://img.shields.io/github/license/tqer39/ccw-cli)](LICENSE)
[![Homebrew](https://img.shields.io/badge/brew-tqer39%2Ftap%2Fccw-FBB040?logo=homebrew&logoColor=white)](https://github.com/tqer39/homebrew-tap)

[🇺🇸 English](README.md) · [🇯🇵 日本語](docs/README.ja.md)

</div>

---

## ⚡ Quick Start

```bash
# 1. install
brew install tqer39/tap/ccw

# 2. run inside any git repo
ccw
```

That's it. `ccw` scans `.claude/worktrees/` and shows the picker, or spins up a fresh worktree if none exist.

> `ccw` also works from inside a worktree — it resolves the main repo via `git rev-parse --git-common-dir` and operates there, so you don't need to `cd` back to the project root first.

## ✨ Features

- 🌳 **Isolated sessions** — each `claude` run gets its own git worktree
- 🎯 **Smart picker** — status badges, `↑N ↓M ✎N` indicators, PR info via `gh`
- 🧹 **Bulk delete** — `[clean pushed]` from the picker or `ccw --clean-all`
- 🦸 **Superpowers preamble** — `-s` injects the `brainstorming → writing-plans → executing-plans` workflow
- ➡️ **Transparent passthrough** — anything after `--` reaches `claude` verbatim

## 🎬 Demo

![picker demo](docs/assets/picker-demo.gif)

## 📖 Usage

```bash
ccw                                       # pick an existing worktree, or start fresh
ccw -n                                    # new worktree, skip picker
ccw -s                                    # new worktree + superpowers preamble
ccw -- --model <model-id>                 # pass-through: any flags after `--` go to claude verbatim
ccw --clean-all --status=pushed --dry-run # preview bulk delete targets
ccw --clean-all --force -y                # nuke everything without prompt
```

Run `ccw --help` for the full flag reference.

### Worktree picker

| Badge | Meaning |
|---|---|
| 🟢 `[PUSHED]` | Clean, upstream tracked, 0 commits ahead |
| 🟡 `[LOCAL]` | No upstream, or ahead of upstream |
| 🔴 `[DIRTY]` | Working tree has uncommitted changes |

Selecting a worktree opens `[r] run` / `[d] delete` / `[b] back`. `run` launches a fresh `claude --permission-mode auto` in that worktree — ccw does **not** reuse Claude Code session IDs (no `--resume` under the hood). Bulk shortcuts (`[delete all]`, `[clean pushed]`, `[custom select]`) remove many at once; dirty items require either `--force` or a three-choice confirm (`y` force · `s` skip dirty · `N` cancel).

PR display requires [`gh`](https://cli.github.com/). Without `gh`, the picker stays functional and shows a hint; rate-limit / network failures hide the PR column silently.

## 📦 Installation

### Homebrew (recommended)

```bash
brew install tqer39/tap/ccw
```

### From source

```bash
git clone https://github.com/tqer39/ccw-cli ~/ccw-cli
go build -o ~/.local/bin/ccw ~/ccw-cli/cmd/ccw
```

Make sure `~/.local/bin` is on your `PATH`.

### Requirements

- `git`
- [Claude Code](https://docs.claude.com/claude-code) `>= 2.1.49` — the `--worktree` flag that ccw relies on was introduced in 2.1.49 (2026-02-19). ccw offers to install `claude` via npm / brew if missing.
- *(optional)* [`gh`](https://cli.github.com/) — enables PR info in the picker
- *(optional)* superpowers plugin — auto-checked when `-s` is used

## ⚙️ Environment

| Variable | Effect |
|---|---|
| `NO_COLOR=1` | Disable colored output |
| `CCW_DEBUG=1` | Verbose debug logging |

Exit codes: `0` success · `1` user error / cancel · anything else is forwarded from `claude`.

## 🛠️ Development

```bash
go test ./...
go vet ./...
go build ./cmd/ccw
```

Pre-commit hooks are managed by [lefthook](https://github.com/evilmartians/lefthook):

```bash
brew install lefthook yamllint actionlint
lefthook install
```

See [`docs/assets/picker-demo-setup.sh`](docs/assets/picker-demo-setup.sh) + [`picker-demo.tape`](docs/assets/picker-demo.tape) to regenerate the demo GIF with [vhs](https://github.com/charmbracelet/vhs).

## 🗺️ Roadmap

- Shell completion (bash / zsh)
- Windows support

## 🤖 Built With

This project was built with [Claude Code](https://docs.claude.com/claude-code) using Claude **Opus 4.7**.

## 📄 License

[MIT](LICENSE)
