# ccw -s plugin preload Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `ccw -s` 起動直後の最初のターンから superpowers skill が使える状態にする。`--plugin-dir` で superpowers プラグインを明示注入し、preamble にも自己修復行を追加する。

**Architecture:** `internal/superpowers/plugindir.go` に `ResolvePluginDir()` を追加 (a)→(b)→(c) のカスケード解決。`cmd/ccw/main.go::run()` で `flags.Superpowers` 時に解決結果を `--plugin-dir <path>` として `flags.Passthrough` の先頭に prepend する。失敗時は i18n 警告を出して preamble だけで続行。preamble にもフォールバック文を 1 行追記。

**Tech Stack:** Go 1.23, `os.UserHomeDir`, `filepath.Glob`, `encoding/json`, `embed`, 既存の `internal/i18n` / `internal/ui` / `internal/cli` / `internal/claude` / `internal/superpowers`。

**Spec:** [docs/superpowers/specs/2026-04-26-ccw-s-plugin-preload-design.md](../specs/2026-04-26-ccw-s-plugin-preload-design.md)

---

## File Structure

- **Create:**
  - `internal/superpowers/plugindir.go` — `ResolvePluginDir()` 本体
  - `internal/superpowers/plugindir_test.go` — カスケード解決のテスト
- **Modify:**
  - `internal/superpowers/preamble_ja.txt` — 自己修復文 1 行追記
  - `internal/superpowers/preamble_en.txt` — 自己修復文 1 行追記
  - `internal/superpowers/preamble_test.go` — 新文言を検証するアサート追加
  - `internal/i18n/keys.go` — `KeySuperpowersPluginDirNotFound` 定数 + `AllKeys()` 追加
  - `internal/i18n/locales/en.yaml` — `superpowers.warn.pluginDirNotFound` 追加
  - `internal/i18n/locales/ja.yaml` — `superpowers.warn.pluginDirNotFound` 追加
  - `cmd/ccw/main.go` — `run()` で `flags.Superpowers` 時に `--plugin-dir` を passthrough 先頭に prepend、失敗時は警告
  - `cmd/ccw/main_test.go` — 既存 `TestMaybePreamble_*` は変更不要（preamble 文字列変更で `superpowers:brainstorming` の含有チェックは引き続き pass）

`internal/claude/` のシグネチャは変更しない（`extra []string` に `--plugin-dir <path>` が含まれるだけ）。

---

## Task 1: i18n 警告メッセージのキーとロケール追加

**Files:**

- Modify: `internal/i18n/keys.go`
- Modify: `internal/i18n/locales/en.yaml`
- Modify: `internal/i18n/locales/ja.yaml`

- [ ] **Step 1: `keys.go` に定数と `AllKeys` エントリを追加**

`internal/i18n/keys.go` の `const` ブロック末尾、`KeyFallbackPrompt Key = "fallback.prompt"` の直下に追加:

```go
KeySuperpowersPluginDirNotFound Key = "superpowers.warn.pluginDirNotFound"
```

`AllKeys()` の return slice 末尾、`KeyFallbackQuit, KeyFallbackPrompt,` の直後に同行を追加:

```go
KeySuperpowersPluginDirNotFound,
```

- [ ] **Step 2: `en.yaml` にエントリ追加**

`internal/i18n/locales/en.yaml` 末尾に以下のブロックを追記（既存 `fallback:` ブロックの後）:

```yaml
superpowers:
  warn:
    pluginDirNotFound: "Could not resolve the superpowers plugin path. If skills are not yet loaded, run /reload-plugins inside Claude Code."
```

- [ ] **Step 3: `ja.yaml` に同等のエントリ追加**

`internal/i18n/locales/ja.yaml` 末尾に追記:

```yaml
superpowers:
  warn:
    pluginDirNotFound: "superpowers プラグインのパスを解決できませんでした。skill が読み込まれていない場合は Claude Code 内で /reload-plugins を実行してください。"
```

- [ ] **Step 4: i18n テストで両ロケールが新キーを持つことを確認**

Run: `go test ./internal/i18n/...`
Expected: PASS（既存 parity test が新キーを両言語に求めて両方追加済みなので通る）

- [ ] **Step 5: Commit**

```bash
git add internal/i18n/keys.go internal/i18n/locales/en.yaml internal/i18n/locales/ja.yaml
git commit -m "feat(i18n): add superpowers.warn.pluginDirNotFound key"
```

---

## Task 2: preamble に自己修復行を追記

**Files:**

- Modify: `internal/superpowers/preamble_ja.txt`
- Modify: `internal/superpowers/preamble_en.txt`
- Modify: `internal/superpowers/preamble_test.go`

- [ ] **Step 1: `preamble_ja.txt` 末尾に自己修復行を追加**

`internal/superpowers/preamble_ja.txt` の現状:

```text
このセッションは Claude Code の --worktree sandbox 内です。
superpowers:brainstorming → superpowers:writing-plans → superpowers:executing-plans
の順で進めてください。トピックはこれから相談します。
```

末尾に空行 1 行と次の 1 行を追加（最終的なファイル）:

```text
このセッションは Claude Code の --worktree sandbox 内です。
superpowers:brainstorming → superpowers:writing-plans → superpowers:executing-plans
の順で進めてください。トピックはこれから相談します。

もし superpowers の skill がまだ読み込まれていない場合は、まず /reload-plugins を実行してから brainstorming を始めてください。
```

- [ ] **Step 2: `preamble_en.txt` 末尾に自己修復行を追加**

最終形:

```text
You are inside a Claude Code --worktree sandbox.
Proceed in this order: superpowers:brainstorming → superpowers:writing-plans → superpowers:executing-plans.
The topic will follow.

If the superpowers skills are not yet loaded, run /reload-plugins first, then begin brainstorming.
```

- [ ] **Step 3: `preamble_test.go` に新文言の検証を追加**

`internal/superpowers/preamble_test.go` の `TestPreamble_JA` 末尾に追加:

```go
 if !strings.Contains(got, "/reload-plugins") {
  t.Errorf("ja preamble missing /reload-plugins fallback: %q", got)
 }
```

`TestPreamble_EN` 末尾に追加:

```go
 if !strings.Contains(got, "/reload-plugins") {
  t.Errorf("en preamble missing /reload-plugins fallback: %q", got)
 }
```

- [ ] **Step 4: テスト実行**

Run: `go test ./internal/superpowers/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/superpowers/preamble_ja.txt internal/superpowers/preamble_en.txt internal/superpowers/preamble_test.go
git commit -m "feat(superpowers): add /reload-plugins self-heal line to preamble"
```

---

## Task 3: `ResolvePluginDir()` の実装と単体テスト (TDD)

**Files:**

- Create: `internal/superpowers/plugindir.go`
- Create: `internal/superpowers/plugindir_test.go`

`ResolvePluginDir()` のシグネチャは `(path string, ok bool)`。テストのために内部処理を `home string` 引数で抽象化したヘルパー `resolvePluginDirIn(home string)` を持ち、`ResolvePluginDir()` は `os.UserHomeDir()` の結果でそれを呼び出すラッパーにする。

- [ ] **Step 1: 失敗するテストを書く: 全部失敗 → `(_, false)` を返す**

`internal/superpowers/plugindir_test.go` を以下で新規作成:

```go
package superpowers

import (
 "os"
 "path/filepath"
 "testing"
)

func TestResolvePluginDir_AllMiss(t *testing.T) {
 home := t.TempDir()
 got, ok := resolvePluginDirIn(home)
 if ok {
  t.Fatalf("expected miss, got %q", got)
 }
}

func TestResolvePluginDir_WellKnownHit(t *testing.T) {
 home := t.TempDir()
 dir := filepath.Join(home, ".claude", "plugins", "cache", "claude-plugins-official", "superpowers", "latest", ".claude-plugin")
 if err := os.MkdirAll(dir, 0o755); err != nil {
  t.Fatal(err)
 }
 if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte("{}"), 0o644); err != nil {
  t.Fatal(err)
 }
 got, ok := resolvePluginDirIn(home)
 if !ok {
  t.Fatal("expected hit, got miss")
 }
 want := filepath.Join(home, ".claude", "plugins", "cache", "claude-plugins-official", "superpowers", "latest")
 if got != want {
  t.Errorf("got %q, want %q", got, want)
 }
}

func TestResolvePluginDir_GlobHit(t *testing.T) {
 home := t.TempDir()
 dir := filepath.Join(home, ".claude", "plugins", "cache", "third-party", "superpowers", "latest", ".claude-plugin")
 if err := os.MkdirAll(dir, 0o755); err != nil {
  t.Fatal(err)
 }
 if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte("{}"), 0o644); err != nil {
  t.Fatal(err)
 }
 got, ok := resolvePluginDirIn(home)
 if !ok {
  t.Fatal("expected hit, got miss")
 }
 want := filepath.Join(home, ".claude", "plugins", "cache", "third-party", "superpowers", "latest")
 if got != want {
  t.Errorf("got %q, want %q", got, want)
 }
}

func TestResolvePluginDir_InstalledJSONHit(t *testing.T) {
 home := t.TempDir()
 // Build the cache dir for a custom marketplace + version.
 cacheDir := filepath.Join(home, ".claude", "plugins", "cache", "my-marketplace", "superpowers", "v2", ".claude-plugin")
 if err := os.MkdirAll(cacheDir, 0o755); err != nil {
  t.Fatal(err)
 }
 if err := os.WriteFile(filepath.Join(cacheDir, "plugin.json"), []byte("{}"), 0o644); err != nil {
  t.Fatal(err)
 }
 // Write installed_plugins.json that points at the custom version.
 pluginsDir := filepath.Join(home, ".claude", "plugins")
 jsonBody := `{
  "version": 2,
  "plugins": {
   "superpowers@my-marketplace": [
    {"scope": "user", "installPath": "", "version": "v2"}
   ]
  }
 }`
 if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(jsonBody), 0o644); err != nil {
  t.Fatal(err)
 }
 got, ok := resolvePluginDirIn(home)
 if !ok {
  t.Fatal("expected hit, got miss")
 }
 want := filepath.Join(home, ".claude", "plugins", "cache", "my-marketplace", "superpowers", "v2")
 if got != want {
  t.Errorf("got %q, want %q", got, want)
 }
}

func TestResolvePluginDir_InstalledJSONExplicitInstallPath(t *testing.T) {
 home := t.TempDir()
 customDir := filepath.Join(home, "custom", "abs", "superpowers-checkout", ".claude-plugin")
 if err := os.MkdirAll(customDir, 0o755); err != nil {
  t.Fatal(err)
 }
 if err := os.WriteFile(filepath.Join(customDir, "plugin.json"), []byte("{}"), 0o644); err != nil {
  t.Fatal(err)
 }
 pluginsDir := filepath.Join(home, ".claude", "plugins")
 if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
  t.Fatal(err)
 }
 parent := filepath.Dir(customDir)
 jsonBody := `{
  "version": 2,
  "plugins": {
   "superpowers@local": [
    {"scope": "project", "installPath": "` + parent + `", "version": "dev"}
   ]
  }
 }`
 if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(jsonBody), 0o644); err != nil {
  t.Fatal(err)
 }
 got, ok := resolvePluginDirIn(home)
 if !ok {
  t.Fatal("expected hit, got miss")
 }
 if got != parent {
  t.Errorf("got %q, want %q", got, parent)
 }
}
```

- [ ] **Step 2: テストが失敗することを確認 (実装ファイルがないため)**

Run: `go test ./internal/superpowers/...`
Expected: FAIL with "undefined: resolvePluginDirIn"

- [ ] **Step 3: 最小実装を `plugindir.go` に書く**

`internal/superpowers/plugindir.go` を新規作成:

```go
package superpowers

import (
 "encoding/json"
 "os"
 "path/filepath"
 "sort"
 "strings"
)

// ResolvePluginDir returns the on-disk directory of the superpowers plugin
// suitable for `claude --plugin-dir <path>`. Returns ("", false) if no
// candidate location can be confirmed.
//
// Resolution cascade:
//
//  1. Read ~/.claude/plugins/installed_plugins.json, find the first key
//     matching `superpowers@<marketplace>`, and use its explicit installPath
//     when set, otherwise derive
//     ~/.claude/plugins/cache/<marketplace>/superpowers/<version>.
//  2. Try the well-known path
//     ~/.claude/plugins/cache/claude-plugins-official/superpowers/latest.
//  3. Glob ~/.claude/plugins/cache/*/superpowers/latest/.claude-plugin/plugin.json
//     and return the alphabetically first hit.
//
// Each candidate is validated by checking that
// <candidate>/.claude-plugin/plugin.json exists.
func ResolvePluginDir() (string, bool) {
 home, err := os.UserHomeDir()
 if err != nil || home == "" {
  return "", false
 }
 return resolvePluginDirIn(home)
}

func resolvePluginDirIn(home string) (string, bool) {
 if p, ok := lookupFromInstalledJSON(home); ok {
  return p, true
 }
 if p, ok := lookupWellKnown(home); ok {
  return p, true
 }
 if p, ok := lookupViaGlob(home); ok {
  return p, true
 }
 return "", false
}

func validateCandidate(dir string) bool {
 if dir == "" {
  return false
 }
 manifest := filepath.Join(dir, ".claude-plugin", "plugin.json")
 st, err := os.Stat(manifest)
 if err != nil || st.IsDir() {
  return false
 }
 return true
}

type installedPluginEntry struct {
 Scope       string `json:"scope"`
 InstallPath string `json:"installPath"`
 Version     string `json:"version"`
}

type installedPluginsFile struct {
 Plugins map[string][]installedPluginEntry `json:"plugins"`
}

func lookupFromInstalledJSON(home string) (string, bool) {
 path := filepath.Join(home, ".claude", "plugins", "installed_plugins.json")
 data, err := os.ReadFile(path)
 if err != nil {
  return "", false
 }
 var doc installedPluginsFile
 if err := json.Unmarshal(data, &doc); err != nil {
  return "", false
 }
 for key, entries := range doc.Plugins {
  marketplace, ok := superpowersMarketplace(key)
  if !ok {
   continue
  }
  if len(entries) == 0 {
   continue
  }
  entry := entries[0]
  if entry.InstallPath != "" {
   if validateCandidate(entry.InstallPath) {
    return entry.InstallPath, true
   }
   continue
  }
  version := entry.Version
  if version == "" {
   version = "latest"
  }
  dir := filepath.Join(home, ".claude", "plugins", "cache", marketplace, "superpowers", version)
  if validateCandidate(dir) {
   return dir, true
  }
 }
 return "", false
}

func superpowersMarketplace(key string) (string, bool) {
 const prefix = "superpowers@"
 if !strings.HasPrefix(key, prefix) {
  return "", false
 }
 mp := strings.TrimPrefix(key, prefix)
 if mp == "" {
  return "", false
 }
 return mp, true
}

func lookupWellKnown(home string) (string, bool) {
 dir := filepath.Join(home, ".claude", "plugins", "cache", "claude-plugins-official", "superpowers", "latest")
 if validateCandidate(dir) {
  return dir, true
 }
 return "", false
}

func lookupViaGlob(home string) (string, bool) {
 pattern := filepath.Join(home, ".claude", "plugins", "cache", "*", "superpowers", "latest", ".claude-plugin", "plugin.json")
 matches, err := filepath.Glob(pattern)
 if err != nil || len(matches) == 0 {
  return "", false
 }
 sort.Strings(matches)
 for _, m := range matches {
  dir := filepath.Dir(filepath.Dir(m))
  if validateCandidate(dir) {
   return dir, true
  }
 }
 return "", false
}
```

- [ ] **Step 4: テストが通ることを確認**

Run: `go test ./internal/superpowers/... -run ResolvePluginDir -v`
Expected: PASS — 5 件全部 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/superpowers/plugindir.go internal/superpowers/plugindir_test.go
git commit -m "feat(superpowers): add ResolvePluginDir for --plugin-dir injection"
```

---

## Task 4: `cmd/ccw/main.go` で `--plugin-dir` を注入

**Files:**

- Modify: `cmd/ccw/main.go`

`run()` 内、`preamble := maybePreamble(flags.Superpowers)` の直後に passthrough 構築ロジックを挿入する。`flags.NewWorktree` ブランチ内の `claude.LaunchNew` 呼び出しがそれを使う。

- [ ] **Step 1: `cmd/ccw/main.go` の import に `"github.com/tqer39/ccw-cli/internal/i18n"` が既にあることを確認**

既存 import に `i18n` は含まれている (line 17)。追加 import 不要。

- [ ] **Step 2: `run()` 内で passthrough を組み立てるヘルパー関数を追加**

`maybePreamble` 関数の直下（`func runPicker(...)` の前）に追加:

```go
// withPluginDir prepends `--plugin-dir <path>` to passthrough when -s was
// passed and the superpowers plugin can be resolved on disk. When resolution
// fails it emits a warning and returns passthrough unchanged so the preamble
// still reaches Claude.
func withPluginDir(enabled bool, passthrough []string) []string {
 if !enabled {
  return passthrough
 }
 dir, ok := superpowers.ResolvePluginDir()
 if !ok {
  ui.Warn("%s", i18n.T(i18n.KeySuperpowersPluginDirNotFound))
  return passthrough
 }
 out := make([]string, 0, len(passthrough)+2)
 out = append(out, "--plugin-dir", dir)
 out = append(out, passthrough...)
 return out
}
```

- [ ] **Step 3: `run()` で `flags.Passthrough` の代わりに `withPluginDir` 結果を使う**

`cmd/ccw/main.go::run()` の `flags.NewWorktree` ブロックを更新（line 84-96 周辺）:

変更前:

```go
 preamble := maybePreamble(flags.Superpowers)

 if flags.NewWorktree {
  name, err := namegen.Generate(mainRepo)
  if err != nil {
   ui.Error("generate worktree name: %v", err)
   return 1
  }
  code, err := claude.LaunchNew(mainRepo, name, preamble, flags.Passthrough)
```

変更後:

```go
 preamble := maybePreamble(flags.Superpowers)
 passthrough := withPluginDir(flags.Superpowers, flags.Passthrough)

 if flags.NewWorktree {
  name, err := namegen.Generate(mainRepo)
  if err != nil {
   ui.Error("generate worktree name: %v", err)
   return 1
  }
  code, err := claude.LaunchNew(mainRepo, name, preamble, passthrough)
```

下段の `runPicker(mainRepo, flags.Passthrough, interactive)` は `-s` ブランチに到達しない（`flags.Superpowers` が立てば `flags.NewWorktree=true` なので picker には入らない）ので変更不要。

- [ ] **Step 4: ビルドと既存テスト**

Run:

```bash
go build ./...
go test ./...
```

Expected: 全 PASS。

- [ ] **Step 5: 動作確認 (`-s` を実 worktree で起動するのは重いので pflag 解析と passthrough 構築を直接呼べるようなテストを追加)**

`cmd/ccw/main_test.go` 末尾に追加:

```go
func TestWithPluginDir_Disabled(t *testing.T) {
 in := []string{"--model", "opus"}
 got := withPluginDir(false, in)
 if len(got) != len(in) {
  t.Errorf("disabled should not modify passthrough, got %v", got)
 }
}
```

`-s` 有効ブランチは `superpowers.ResolvePluginDir()` の実環境依存でユニットテスト困難。Task 3 でロジックは網羅済みなので Task 4 では disabled パスのみ検証して可。

- [ ] **Step 6: テスト実行**

Run: `go test ./cmd/ccw/...`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add cmd/ccw/main.go cmd/ccw/main_test.go
git commit -m "feat(ccw): inject --plugin-dir for -s to preload superpowers"
```

---

## Task 5: 統合検証

**Files:** なし（既存ファイルでの動作確認）

- [ ] **Step 1: 全テスト + ビルド + lint**

Run:

```bash
go build ./...
go test ./...
```

Expected: 全 PASS。

- [ ] **Step 2: ローカルバイナリビルドして手動スモーク**

Run:

```bash
go build -o /tmp/ccw-dev ./cmd/ccw
/tmp/ccw-dev -h | head -30
```

Expected: ヘルプが表示され、`-s, --superpowers` 行が見える。

- [ ] **Step 3: 解決パスが妥当に出ることを確認するワンショット**

Run:

```bash
go run ./cmd/internal-debug-not-checked-in 2>/dev/null || cat <<'EOF' >/tmp/check_resolve.go
package main

import (
 "fmt"

 "github.com/tqer39/ccw-cli/internal/superpowers"
)

func main() {
 dir, ok := superpowers.ResolvePluginDir()
 fmt.Println("ok:", ok, "dir:", dir)
}
EOF
go run /tmp/check_resolve.go
rm /tmp/check_resolve.go
```

Expected: `ok: true dir: /Users/<you>/.claude/plugins/cache/.../superpowers/latest` のように、`.claude-plugin/plugin.json` を持つディレクトリが返る。

（このステップはローカル開発機で superpowers が導入されている前提。CI では skip。）

- [ ] **Step 4: フォールバック動作の手動確認 (任意)**

`~/.claude/plugins` を一時的にリネームして上記スクリプトを再実行:

```bash
mv ~/.claude/plugins ~/.claude/plugins.bak
go run /tmp/check_resolve.go   # 前ステップで作って消したなら再生成
mv ~/.claude/plugins.bak ~/.claude/plugins
```

Expected: `ok: false dir:` が返ること。

- [ ] **Step 5: 修正コミットがなければ完了**

Run: `git status`
Expected: `nothing to commit, working tree clean`

---

## Self-Review

**1. Spec coverage:**

| Spec 決定事項 | 実装タスク |
|---|---|
| D1. `flags.Superpowers` 時に `--plugin-dir <path>` を prepend | Task 4 |
| D2. パス解決カスケード (a)→(b)→(c) | Task 3 |
| D3. 解決失敗時の `ui.Warn` + 通常 preamble 続行 | Task 4 (`withPluginDir` 内の警告) |
| D4. preamble に `/reload-plugins` 自己修復行を追記 | Task 2 |
| D5. 新規 API は `ResolvePluginDir() (string, bool)` のみ、`internal/claude/` は無変更 | Task 3, Task 4 (passthrough 経由で渡すため `claude` パッケージ未変更) |
| 検証 (a)/(b)/(c)/(全失敗) のテスト | Task 3 Step 1 |
| preamble 文言テスト追従 | Task 2 Step 3 |

ギャップなし。

**2. Placeholder scan:** "TBD" / "TODO" / "Similar to" / 未定型なし。全コードブロックは実コード。

**3. Type consistency:**

- `ResolvePluginDir() (string, bool)` の名前と引数: Task 3 で定義、Task 4 で同じシグネチャを呼んでいる ✓
- `withPluginDir(enabled bool, passthrough []string) []string`: Task 4 内で 1 回宣言・1 回呼出、整合 ✓
- i18n キー名 `KeySuperpowersPluginDirNotFound` / dotted path `superpowers.warn.pluginDirNotFound`: Task 1 で 3 箇所すべて同じ綴り ✓
- `lookupFromInstalledJSON` / `lookupWellKnown` / `lookupViaGlob` / `validateCandidate` / `superpowersMarketplace`: Task 3 内のみ参照、命名一貫 ✓

問題なし。
