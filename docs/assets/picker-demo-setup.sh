#!/bin/bash
# picker-demo-setup.sh — reproduce the demo environment for picker-demo.tape.
# Run once before `vhs docs/assets/picker-demo.tape` to regenerate the GIF.
#
# Requires: git, vhs, a Go toolchain. The mock gh is a local stand-in used
# only to populate the demo GIF with deterministic PR rows.
#
# Produces four worktrees covering every worktree badge (PUSHED / LOCAL /
# DIRTY) and every PR badge (OPEN / DRAFT / MERGED / CLOSED):
#   - feat/login       → PUSHED + [OPEN]   #42
#   - feat/dashboard   → PUSHED + [MERGED] #44
#   - feat/picker      → LOCAL  + [DRAFT]  #43
#   - chore/cleanup    → DIRTY  + [CLOSED] #45
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

# 3. four worktrees covering all badge combinations
#    feat/login: PUSHED (clean, upstream tracked) + OPEN PR #42
git -C /tmp/ccw-demo worktree add -q -b feat/login .claude/worktrees/feat-login
(cd /tmp/ccw-demo/.claude/worktrees/feat-login && echo login >README.md && git add README.md && git commit -q -m "feat: add login page" && git push -q -u origin feat/login)

#    feat/dashboard: PUSHED (clean, upstream tracked) + MERGED PR #44
git -C /tmp/ccw-demo worktree add -q -b feat/dashboard .claude/worktrees/feat-dashboard
(cd /tmp/ccw-demo/.claude/worktrees/feat-dashboard && echo dashboard >README.md && git add README.md && git commit -q -m "feat: analytics dashboard" && git push -q -u origin feat/dashboard)

#    feat/picker: LOCAL (no upstream) + DRAFT PR #43
git -C /tmp/ccw-demo worktree add -q -b feat/picker .claude/worktrees/feat-picker
(cd /tmp/ccw-demo/.claude/worktrees/feat-picker && echo picker >README.md && git add README.md && git commit -q -m "feat: picker redesign")

#    chore/cleanup: DIRTY (untracked files) + CLOSED PR #45
git -C /tmp/ccw-demo worktree add -q -b chore/cleanup .claude/worktrees/chore-cleanup
(cd /tmp/ccw-demo/.claude/worktrees/chore-cleanup && echo untracked >stray.txt && echo more >extra.md)

# 4. mock gh that returns canned PR rows covering every PR state
mkdir -p /tmp/fake-gh
cat >/tmp/fake-gh/gh <<'GH'
#!/bin/bash
case "$1" in
  auth) exit 0 ;;
  pr)
    cat <<'JSON'
[
  {"number": 42, "title": "feat: add login page",        "state": "OPEN",   "headRefName": "feat/login"},
  {"number": 43, "title": "feat: picker redesign",       "state": "DRAFT",  "headRefName": "feat/picker"},
  {"number": 44, "title": "feat: analytics dashboard",   "state": "MERGED", "headRefName": "feat/dashboard"},
  {"number": 45, "title": "chore: cleanup stray files",  "state": "CLOSED", "headRefName": "chore/cleanup"}
]
JSON
    exit 0 ;;
esac
exit 1
GH
chmod +x /tmp/fake-gh/gh

# 5. fake HOME so the picker can detect RESUME / NEW deterministically.
# Encoding mirrors internal/worktree/has_session.go:EncodeProjectPath
# (replaces '/' and '.' with '-'). Use the realpath of each worktree —
# on macOS /tmp is a symlink to /private/tmp, and the picker uses the
# resolved absolute path as the lookup key.
rm -rf /tmp/ccw-demo-home
PROJECTS=/tmp/ccw-demo-home/.claude/projects
mkdir -p "$PROJECTS"
for wt in feat-login feat-dashboard; do
  abs=$(cd "/tmp/ccw-demo/.claude/worktrees/$wt" && pwd -P)
  enc=$(printf '%s' "$abs" | tr '/.' '--')
  mkdir -p "$PROJECTS/$enc"
  printf '{}\n' >"$PROJECTS/$enc/dummy.jsonl"
done

echo "ready. now run: vhs docs/assets/picker-demo.tape"
