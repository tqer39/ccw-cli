<div align="center">

![ccw-cli — Claude Code x worktree](docs/assets/header.png)

**A thin launcher for [Claude Code](https://docs.claude.com/claude-code)'s built-in `--worktree` — run `ccw` anywhere in the repo to pick an existing worktree (PR info attached) or start fresh. Plain CLI, stays out of your tmux/zellij setup.**

[![Go](https://img.shields.io/badge/go-1.25-00ADD8?logo=go&logoColor=white)](go.mod)
[![Release](https://img.shields.io/github/v/release/tqer39/ccw-cli?logo=github)](https://github.com/tqer39/ccw-cli/releases)
[![License](https://img.shields.io/github/license/tqer39/ccw-cli)](LICENSE)
[![Homebrew](https://img.shields.io/badge/brew-tqer39%2Ftap%2Fccw-FBB040?logo=homebrew&logoColor=white)](https://github.com/tqer39/homebrew-tap)
[![brew-audit](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml/badge.svg)](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml)
[![codecov](https://codecov.io/gh/tqer39/ccw-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/tqer39/ccw-cli)

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
- 🎯 **Worktree state at a glance** — pushed / ahead / behind / dirty, plus PR info, all in one picker
- 🧹 **Bulk cleanup** — `[clean pushed]` or `ccw --clean-all` sweeps the worktrees you're done with
- 📋 **Machine-readable list** — `ccw -L --json` aggregates worktree × git × PR × session info in one shot, ideal for scripts and Claude Code agent use

## 🎬 Demo

![picker demo](docs/assets/picker-demo.gif)

> **Note:** the `💬 RESUME` badge only signals that a session log exists for the worktree. The session title or first prompt is **not** previewed in the picker — `ccw` simply runs `claude --continue` and lets the Claude Code CLI pick the most recent session.

## 📖 Usage

```bash
ccw                                       # pick an existing worktree, or start fresh
ccw -n                                    # new worktree, skip picker
ccw -s                                    # new worktree + inject the localized superpowers preamble as first prompt
ccw -- --model <model-id>                 # pass-through: any flags after `--` go to claude verbatim
ccw -L                                    # list ccw worktrees (text table)
ccw -L --json                             # same, JSON for scripts / agents
ccw -L -d ~/repo --no-pr --no-session     # target a specific repo, skip gh and session lookup
ccw --clean-all --status=pushed --dry-run # preview bulk delete targets
ccw --clean-all --force -y                # nuke everything without prompt
```

Run `ccw --help` for the full flag reference.

## 🎯 Picker reference

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

Session badge:

| Badge | Meaning |
|---|---|
| 💬 `RESUME` | Past session log exists — `run` restores the conversation |
| ⚡ `NEW`    | No session log — `run` starts fresh |

Selecting a worktree opens `[r] run` / `[d] delete` / `[b] back`. `run` calls `claude --continue` to restore the past conversation when a session log exists, or `claude -n <worktree>` to start fresh otherwise. Bulk shortcuts (`[delete all]`, `[clean pushed]`, `[custom select]`) remove many at once; dirty items require either `--force` or a three-choice confirm (`y` force · `s` skip dirty · `N` cancel).

Without `gh`, the picker stays functional and shows a hint; rate-limit / network failures hide the PR column silently.

## 🏷️ Naming

When ccw creates a new worktree, the worktree directory and the Claude Code session name are kept 1:1:

- Directory: `<repo>/.claude/worktrees/<name>/`
- Session name: `<name>` (set via `claude -n <name>`)

`<name>` is generated as `ccw-<owner>-<repo>-<yymmdd>-<hhmmss>` (e.g. `ccw-tqer39-ccw-cli-260426-143055`). `<owner>` / `<repo>` come from the `origin` remote URL; the timestamp is the worktree creation time in your local timezone. When `origin` is unset, `<owner>` becomes `local` and `<repo>` is the directory basename. Duplicate names (e.g. two worktrees created within the same second) are disambiguated with `-2`, `-3`, … Renaming the session manually with `/rename` is fine — ccw does not track it, and `--continue` keys off the working directory so conversation restore is unaffected.

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
- [Claude Code](https://docs.claude.com/claude-code) `>= 2.1.76` — ccw uses `--worktree <name>` (added in 2.1.49) together with `-n <name>` (added in 2.1.76). ccw offers to install `claude` via npm / brew if missing.
- *(optional)* [`gh`](https://cli.github.com/) — enables PR info in the picker
- *(optional)* [superpowers](https://github.com/obra/superpowers) plugin — declared in [`.claude/settings.json`](./.claude/settings.json) so Claude Code prompts to install it on first launch in this repo

## ⚙️ Environment

| Variable | Effect |
|---|---|
| `NO_COLOR=1` | Disable colored output |
| `CCW_DEBUG=1` | Verbose debug logging |
| `CCW_LANG=en\|ja` | Force output language. Overridden by `--lang`. Falls back to system locale (`LC_ALL` / `LC_MESSAGES` / `LANG`), then English. |

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
