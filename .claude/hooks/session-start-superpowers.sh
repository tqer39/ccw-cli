#!/usr/bin/env bash
# SessionStart hook for new worktree sessions in ccw-cli.
# Outputs additionalContext that nudges Claude to follow the superpowers
# brainstorming -> writing-plans -> executing-plans flow. Only fires on
# new sessions; resume/clear are routed to different matchers in settings.json.
set -euo pipefail
cat <<'JSON'
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "このセッションは Claude Code の --worktree sandbox 内です。\nsuperpowers:brainstorming → superpowers:writing-plans → superpowers:executing-plans\nの順で進めてください。トピックはこれから相談します。"
  }
}
JSON
