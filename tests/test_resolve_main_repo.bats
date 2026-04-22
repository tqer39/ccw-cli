#!/usr/bin/env bats

CCW="${BATS_TEST_DIRNAME}/../bin/ccw"

setup() {
  TMP_REPO="$(cd "$(mktemp -d)" && pwd -P)"
  (cd "$TMP_REPO" && git init -q -b main && git commit -q --allow-empty -m "init")
}

teardown() {
  rm -rf "$TMP_REPO"
}

@test "non-git dir triggers warning and exit 1" {
  local outside
  outside="$(mktemp -d)"
  run bash -c "cd '$outside' && '$CCW'"
  [ "$status" -eq 1 ]
  [[ "$output" == *"must be run inside a git repository"* ]]
  rm -rf "$outside"
}

@test "inside main repo returns the repo path via debug output" {
  run bash -c "cd '$TMP_REPO' && '$CCW'"
  [[ "$output" == *"main_repo: $TMP_REPO"* ]]
}

@test "inside a worktree resolves to main repo" {
  local wt="${TMP_REPO}/.claude/worktrees/test"
  mkdir -p "$(dirname "$wt")"
  (cd "$TMP_REPO" && git worktree add -q "$wt" -b test-branch)
  run bash -c "cd '$wt' && '$CCW'"
  [[ "$output" == *"main_repo: $TMP_REPO"* ]]
}
