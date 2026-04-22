#!/usr/bin/env bats

CCW="${BATS_TEST_DIRNAME}/../bin/ccw"

setup() {
  TMP_REPO="$(cd "$(mktemp -d)" && pwd -P)"
  (cd "$TMP_REPO" && git init -q -b main && git commit -q --allow-empty -m init)

  FAKE_BIN="$(cd "$(mktemp -d)" && pwd -P)"
  cat > "$FAKE_BIN/claude" <<'SH'
#!/usr/bin/env bash
printf 'fake-claude: %s\n' "$*"
SH
  chmod +x "$FAKE_BIN/claude"

  # superpowers plugin を仮配置
  FAKE_HOME="$(cd "$(mktemp -d)" && pwd -P)"
  mkdir -p "$FAKE_HOME/.claude/plugins/cache/fake/superpowers"
}

teardown() {
  rm -rf "$TMP_REPO" "$FAKE_BIN" "$FAKE_HOME"
}

@test "gitignore with docs/superpowers/ already → no prompt" {
  printf 'docs/superpowers/\n' > "$TMP_REPO/.gitignore"
  run bash -c "cd '$TMP_REPO' && PATH=\"$FAKE_BIN:\$PATH\" HOME='$FAKE_HOME' '$CCW' -s </dev/null"
  [[ "$output" != *"is not ignored by git"* ]]
}

@test "gitignore without entry and non-interactive → skip without modifying" {
  run bash -c "cd '$TMP_REPO' && PATH=\"$FAKE_BIN:\$PATH\" HOME='$FAKE_HOME' '$CCW' -s </dev/null"
  [[ "$output" == *"is not ignored by git"* ]]
  [ ! -f "$TMP_REPO/.gitignore" ] || ! grep -q "docs/superpowers" "$TMP_REPO/.gitignore"
}
