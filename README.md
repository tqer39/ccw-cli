<div align="center">

![ccw-cli — Claude Code x worktree](docs/assets/header.png)

**A thin launcher for [Claude Code](https://docs.claude.com/claude-code)'s built-in `--worktree` — run `ccw` anywhere in the repo to pick an existing worktree (PR info attached) or start fresh. Plain CLI, stays out of your tmux/zellij setup.**

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

## ✨ Features

- 🤝 **Hand-off and step aside** — pick (or create) a worktree, launch `claude` in it, then ccw exits. No daemon, no wrapper process, no coupling to tmux/zellij — just the bridge.
- 🧭 **Works from anywhere in the repo** — run `ccw` inside a worktree or subdirectory; ccw resolves the main repo automatically
- 🎯 **Worktree state at a glance** — pushed / ahead / behind / dirty, plus PR info, all in one picker
- 🧹 **Bulk cleanup** — `[clean pushed]` or `ccw --clean-all` sweeps the worktrees you're done with
- 🦸 **"Design first" startup** — `-s` tells claude to follow the brainstorming → writing-plans → executing-plans flow (prompts to install the superpowers plugin if missing)
- ➡️ **claude flags pass through** — anything after `--` goes to claude untouched, so `--model` and friends still work

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

Worktree status badge:

| Badge | Meaning |
|---|---|
| 🟢 `[PUSHED]` | Clean, upstream tracked, 0 commits ahead |
| 🟡 `[LOCAL]` | No upstream, or ahead of upstream |
| 🔴 `[DIRTY]` | Working tree has uncommitted changes |

PR state badge (shown only when [`gh`](https://cli.github.com/) is installed and authenticated):

| Badge | Meaning |
|---|---|
| 🟩 `[OPEN]` | PR is open and awaiting review / merge |
| ⬛ `[DRAFT]` | PR is a draft |
| 🟪 `[MERGED]` | PR has been merged |
| 🟥 `[CLOSED]` | PR was closed without merging |

Selecting a worktree opens `[r] run` / `[d] delete` / `[b] back`. `run` launches a fresh `claude --permission-mode auto` in that worktree — ccw does **not** reuse Claude Code session IDs (no `--resume` under the hood). Bulk shortcuts (`[delete all]`, `[clean pushed]`, `[custom select]`) remove many at once; dirty items require either `--force` or a three-choice confirm (`y` force · `s` skip dirty · `N` cancel).

Without `gh`, the picker stays functional and shows a hint; rate-limit / network failures hide the PR column silently.

> ⚠️ **Passing `--resume` through `--` is unsupported.**
> `ccw -n -- --resume ID` and `ccw -s -- --resume ID` combine `claude --worktree` (new worktree) with `--resume` (continue a prior session); the resumed transcript's file references won't match the freshly-created worktree. Even the picker's re-entry path suffers the same mismatch if the selected worktree differs from the session's original. If a resumed session is what you want, run `claude --resume ID` directly — bypass ccw.

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

- [`git`](https://git-scm.com/)
- [Claude Code](https://docs.claude.com/claude-code) `>= 2.1.49` — the `--worktree` flag that ccw relies on was introduced in 2.1.49 (2026-02-19). ccw offers to install `claude` via npm / brew if missing.
- *(optional)* [`gh`](https://cli.github.com/) — enables PR info in the picker
- *(optional)* [superpowers](https://github.com/obra/superpowers) plugin — auto-checked when `-s` is used

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

Set up the full dev environment (Homebrew required) with:

```bash
make bootstrap
```

This installs the Homebrew packages listed in [`Brewfile`](Brewfile), provisions Go / Node via [`mise`](https://mise.jdx.dev/), and enables [lefthook](https://github.com/evilmartians/lefthook) pre-commit hooks.

See [`docs/assets/picker-demo-setup.sh`](docs/assets/picker-demo-setup.sh) + [`picker-demo.tape`](docs/assets/picker-demo.tape) to regenerate the demo GIF with [vhs](https://github.com/charmbracelet/vhs).

## 🤖 Built With

This project was built with [Claude Code](https://docs.claude.com/claude-code) using Claude **Opus 4.7**.

## 📄 License

[MIT](LICENSE)
