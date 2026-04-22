#!/usr/bin/env bats

CCW="${BATS_TEST_DIRNAME}/../bin/ccw"

setup() {
  TMP_REPO="$(cd "$(mktemp -d)" && pwd -P)"
  (cd "$TMP_REPO" && git init -q -b main && git commit -q --allow-empty -m "init")

  # fake claude that records args and exits 0
  FAKE_BIN="$(cd "$(mktemp -d)" && pwd -P)"
  cat > "$FAKE_BIN/claude" <<'SH'
#!/usr/bin/env bash
printf 'fake-claude: %s\n' "$*"
SH
  chmod +x "$FAKE_BIN/claude"
}

teardown() {
  rm -rf "$TMP_REPO" "$FAKE_BIN"
}

@test "non-git dir triggers warning and exit 1" {
  local outside
  outside="$(mktemp -d)"
  run bash -c "cd '$outside' && '$CCW'"
  [ "$status" -eq 1 ]
  [[ "$output" == *"must be run inside a git repository"* ]]
  rm -rf "$outside"
}

@test "inside main repo forwards to claude with --worktree" {
  run bash -c "cd '$TMP_REPO' && PATH=\"$FAKE_BIN:\$PATH\" '$CCW' -n"
  [[ "$output" == *"fake-claude: --permission-mode auto --worktree"* ]]
}

@test "inside a worktree still resolves to main repo and launches claude" {
  local wt="${TMP_REPO}/.claude/worktrees/test"
  mkdir -p "$(dirname "$wt")"
  (cd "$TMP_REPO" && git worktree add -q "$wt" -b test-branch)
  run bash -c "cd '$wt' && PATH=\"$FAKE_BIN:\$PATH\" '$CCW' -n"
  [[ "$output" == *"fake-claude: --permission-mode auto --worktree"* ]]
}

@test "-- passthrough forwards extra args" {
  run bash -c "cd '$TMP_REPO' && PATH=\"$FAKE_BIN:\$PATH\" '$CCW' -n -- --model foo"
  [[ "$output" == *"--worktree --model foo"* ]]
}
