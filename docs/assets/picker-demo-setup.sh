#!/bin/bash
# picker-demo-setup.sh — reproduce the demo environment for picker-demo.tape.
# Run once before `vhs docs/assets/picker-demo.tape` to regenerate the GIF.
#
# Requires: git, vhs, a Go toolchain. The mock gh is a local stand-in used
# only to populate the demo GIF with deterministic PR rows.
set -euo pipefail

# 1. build ccw into a dedicated bin dir so the tape can invoke it as `ccw`
rm -rf /tmp/ccw-demo-bin
mkdir -p /tmp/ccw-demo-bin
go build -o /tmp/ccw-demo-bin/ccw ./cmd/ccw

# 2. throwaway demo repo with a bare origin
rm -rf /tmp/ccw-demo /tmp/ccw-demo-origin
mkdir -p /tmp/ccw-demo
git -C /tmp/ccw-demo init -q -b main
git -C /tmp/ccw-demo commit -q --allow-empty -m init
git init -q --bare /tmp/ccw-demo-origin
git -C /tmp/ccw-demo remote add origin /tmp/ccw-demo-origin
git -C /tmp/ccw-demo push -q origin main

# 3. three worktrees with distinct statuses
git -C /tmp/ccw-demo worktree add -q -b feat/login .claude/worktrees/feat-login
(cd /tmp/ccw-demo/.claude/worktrees/feat-login && echo login >README.md && git add README.md && git commit -q -m "feat: add login page" && git push -q -u origin feat/login)

git -C /tmp/ccw-demo worktree add -q -b feat/picker .claude/worktrees/feat-picker
(cd /tmp/ccw-demo/.claude/worktrees/feat-picker && echo picker >README.md && git add README.md && git commit -q -m "feat: picker redesign")

git -C /tmp/ccw-demo worktree add -q -b chore/cleanup .claude/worktrees/chore-cleanup
(cd /tmp/ccw-demo/.claude/worktrees/chore-cleanup && echo untracked >stray.txt && echo more >extra.md)

# 4. mock gh that returns canned PR rows (feat/login -> #42 open, chore/cleanup -> #43 draft)
mkdir -p /tmp/fake-gh
cat >/tmp/fake-gh/gh <<'GH'
#!/bin/bash
case "$1" in
  auth) exit 0 ;;
  pr)
    cat <<'JSON'
[
  {"number": 42, "title": "feat: add login page",       "state": "OPEN",  "headRefName": "feat/login"},
  {"number": 43, "title": "chore: cleanup stray files", "state": "DRAFT", "headRefName": "chore/cleanup"}
]
JSON
    exit 0 ;;
esac
exit 1
GH
chmod +x /tmp/fake-gh/gh

echo "ready. now run: vhs docs/assets/picker-demo.tape"
