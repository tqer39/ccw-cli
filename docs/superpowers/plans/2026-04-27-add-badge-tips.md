# Add badge specification tips Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Picker フッターの tip ローテーションに Status badge / PR badge の意味説明を 2 件追加し、ランダムプールの母数を 5 → 7 にする。

**Architecture:** i18n キー定数を 2 つ追加し、対応する文字列を ja/en 両 locale に登録、`internal/tips/tips.go` の `keys` 配列にも追加する。既存の `TestCatalogParity_AllKeysPresentInBothLocales` を TDD のドライバに利用する(キー追加で fail → locale 追加で pass)。

**Tech Stack:** Go 1.22+, YAML(`gopkg.in/yaml.v3` 系), 標準 testing パッケージ

**Spec:** `docs/superpowers/specs/2026-04-27-add-badge-tips-design.md`

**File map:**

- Modify: `internal/i18n/keys.go` — `KeyTipStatusBadge`, `KeyTipPRBadge` 定数を追加し `AllKeys()` の tip 群に含める
- Modify: `internal/i18n/locales/ja.yaml` — `tip.statusBadge`, `tip.prBadge` を追加
- Modify: `internal/i18n/locales/en.yaml` — `tip.statusBadge`, `tip.prBadge` を追加
- Modify: `internal/tips/tips.go` — `keys` 配列に新 2 キーを追加(母数 5 → 7)

テストコードは既存のものを変更しない(spec で明言)。文言を assert する新テストも追加しない。

---

## Task 1: i18n キー定数を追加し parity test の失敗を確認

**Files:**

- Modify: `internal/i18n/keys.go`(tip 定数群と `AllKeys()` の tip 行)
- Test: `internal/i18n/i18n_test.go::TestCatalogParity_AllKeysPresentInBothLocales`(既存)

- [ ] **Step 1: parity test が現状 pass することを確認(ベースライン)**

Run: `go test ./internal/i18n/ -run TestCatalogParity_AllKeysPresentInBothLocales -v`
Expected: PASS

- [ ] **Step 2: `internal/i18n/keys.go` の tip 定数群に 2 行追加**

`KeyTipResumeBadge` の直下に追加:

```go
const (
 KeyTipRename       Key = "tip.rename"
 KeyTipFromPR       Key = "tip.fromPR"
 KeyTipCleanAll     Key = "tip.cleanAll"
 KeyTipPassthrough  Key = "tip.passthrough"
 KeyTipResumeBadge  Key = "tip.resumeBadge"
 KeyTipStatusBadge  Key = "tip.statusBadge"
 KeyTipPRBadge      Key = "tip.prBadge"

 KeyHelpUsage Key = "help.usage"
```

- [ ] **Step 3: `AllKeys()` の tip 行を更新**

`internal/i18n/keys.go` の `AllKeys()` 内、tip キー行を以下に置き換え:

```go
  KeyTipRename, KeyTipFromPR, KeyTipCleanAll, KeyTipPassthrough, KeyTipResumeBadge,
  KeyTipStatusBadge, KeyTipPRBadge,
```

- [ ] **Step 4: parity test が fail することを確認(TDD red)**

Run: `go test ./internal/i18n/ -run TestCatalogParity_AllKeysPresentInBothLocales -v`
Expected: FAIL — ja/en の両 catalog で `tip.statusBadge` / `tip.prBadge` が欠落していると報告される

このタスクではコミットしない(red 状態のまま次のタスクで green にする)。

---

## Task 2: ja locale に文言を追加

**Files:**

- Modify: `internal/i18n/locales/ja.yaml`(`tip:` セクション)

- [ ] **Step 1: `internal/i18n/locales/ja.yaml` の `tip:` セクションに 2 行追加**

`resumeBadge` の直下に追加:

```yaml
tip:
  rename: "worktree 名 = session 名。/rename で改名しても ccw は追跡しません"
  fromPR: "claude --from-pr <番号> で PR と紐づいた session を直接 resume できます"
  cleanAll: "--clean-all で push 済み worktree を一括削除できます"
  passthrough: "ccw -- --model <id> のように -- 以降は claude にそのまま渡されます"
  resumeBadge: "RESUME バッジは ~/.claude/projects/ から判定しています"
  statusBadge: "Status: PUSHED=push 済 / LOCAL=未 push / DIRTY=未コミット"
  prBadge: "PR: OPEN / DRAFT / MERGED / CLOSED を gh pr list から色分け"
```

- [ ] **Step 2: parity test の状況を確認(en だけ未対応 → 依然 fail)**

Run: `go test ./internal/i18n/ -run TestCatalogParity_AllKeysPresentInBothLocales -v`
Expected: FAIL — `en` catalog で `tip.statusBadge` / `tip.prBadge` が欠落と報告される(ja は OK)

---

## Task 3: en locale に文言を追加して parity test を pass させる

**Files:**

- Modify: `internal/i18n/locales/en.yaml`(`tip:` セクション)

- [ ] **Step 1: `internal/i18n/locales/en.yaml` の `tip:` セクションに 2 行追加**

`resumeBadge` の直下に追加:

```yaml
tip:
  rename: "Worktree name = session name; renaming with /rename is fine, ccw doesn't track it"
  fromPR: "claude --from-pr <number> resumes a PR-linked session directly"
  cleanAll: "--clean-all sweeps pushed worktrees in bulk"
  passthrough: "ccw -- --model <id> passes flags through to claude"
  resumeBadge: "The RESUME badge is derived from ~/.claude/projects/"
  statusBadge: "Status: PUSHED=pushed / LOCAL=unpushed / DIRTY=uncommitted"
  prBadge: "PR: OPEN / DRAFT / MERGED / CLOSED colored by gh pr list state"
```

- [ ] **Step 2: parity test が pass することを確認(TDD green)**

Run: `go test ./internal/i18n/ -run TestCatalogParity -v`
Expected: PASS(`TestCatalogParity_AllKeysPresentInBothLocales` と `TestCatalogParity_FormatVerbsMatch` の双方)

- [ ] **Step 3: i18n パッケージ全体のテストが pass することを確認**

Run: `go test ./internal/i18n/ -v`
Expected: PASS(全テスト)

- [ ] **Step 4: コミット(i18n キー + ロケール文言一式)**

```bash
git add internal/i18n/keys.go internal/i18n/locales/ja.yaml internal/i18n/locales/en.yaml
git commit -m "$(cat <<'EOF'
feat(i18n): badge 仕様説明 tip キーと文言を追加

Status badge / PR badge の意味を説明する tip キー
(tip.statusBadge / tip.prBadge) を i18n に追加し
ja/en 両 locale に文言を登録する。

EOF
)"
```

---

## Task 4: tips プールに新 2 キーを追加

**Files:**

- Modify: `internal/tips/tips.go`(`keys` 配列)
- Test: `internal/tips/tips_test.go`(既存、無改修)

- [ ] **Step 1: tips の既存テストが pass することを確認(ベースライン)**

Run: `go test ./internal/tips/ -v`
Expected: PASS(`TestPickRandom_FromDefaultSet`, `TestPickRandom_Deterministic`, `TestPickFrom_Empty`, `TestDefaults_NonEmpty` の 4 件)

- [ ] **Step 2: `internal/tips/tips.go` の `keys` 配列を 7 要素に拡張**

ファイル全体は以下のように差し替える(変更行は `keys` 配列のみ):

```go
// Package tips provides short rotating tip strings shown in the picker footer.
package tips

import (
 "math/rand/v2"

 "github.com/tqer39/ccw-cli/internal/i18n"
)

var keys = []i18n.Key{
 i18n.KeyTipRename,
 i18n.KeyTipFromPR,
 i18n.KeyTipCleanAll,
 i18n.KeyTipPassthrough,
 i18n.KeyTipResumeBadge,
 i18n.KeyTipStatusBadge,
 i18n.KeyTipPRBadge,
}

// Defaults returns the current language's tip strings.
func Defaults() []string {
 out := make([]string, len(keys))
 for i, k := range keys {
  out[i] = i18n.T(k)
 }
 return out
}

// PickRandom returns a single tip selected deterministically from seed.
func PickRandom(seed uint64) string {
 return pickFrom(Defaults(), seed)
}

func pickFrom(set []string, seed uint64) string {
 if len(set) == 0 {
  return ""
 }
 r := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
 return set[r.IntN(len(set))]
}
```

- [ ] **Step 3: tips のテストが引き続き pass することを確認**

Run: `go test ./internal/tips/ -v`
Expected: PASS(母数増加でも `Defaults()` の各要素が非空かつ `PickRandom` の決定性が維持されている)

- [ ] **Step 4: コミット**

```bash
git add internal/tips/tips.go
git commit -m "$(cat <<'EOF'
feat(tips): badge 仕様説明 tip 2 件をローテーションに追加

internal/tips/tips.go の keys 配列に
KeyTipStatusBadge と KeyTipPRBadge を追加し
ランダム選択母数を 5 → 7 に拡張する。

EOF
)"
```

---

## Task 5: 全体テストと lint を実行

**Files:** なし(検証のみ)

- [ ] **Step 1: 全パッケージのテストが pass することを確認**

Run: `go test ./...`
Expected: PASS(既存テスト + 今回触った i18n / tips パッケージのテストが緑)

- [ ] **Step 2: `go vet` を実行**

Run: `go vet ./...`
Expected: 出力なし(問題なし)

- [ ] **Step 3: `golangci-lint` を実行**

Run: `golangci-lint run ./...`
Expected: 出力なし(問題なし)。`golangci-lint` がローカルに無い場合は `mise exec golangci-lint -- run ./...` でも可。

- [ ] **Step 4: `lefthook` の pre-commit 相当をドライ実行(cspell 等)**

Run: `lefthook run pre-commit` または `lefthook run pre-push`
Expected: cspell / markdownlint / yamllint が pass。`lefthook` が無い場合はスキップ可。

このタスクはコミット不要。

---

## Task 6: ローカル動作確認(目視)

**Files:** なし(検証のみ)

- [ ] **Step 1: ccw を実機ビルド**

Run: `go build -o /tmp/ccw-badge-test ./cmd/ccw`
Expected: ビルド成功(エラーなし)。

- [ ] **Step 2: 各 tip が seed を変えれば必ず引けることを確認するため、tips パッケージを直接呼び出して全 7 件を網羅**

`internal/tips/tips_check.go` のような検証用ファイルは作らず、Go の REPL 代わりに `go run` で短い検証スクリプトを書くか、もしくは以下のワンライナーで母数 = 7 を確認:

Run: `go test ./internal/tips/ -run TestDefaults_NonEmpty -v -count=1` の出力を見るのではなく、目視確認のために以下を実行:

```bash
cat <<'EOF' > /tmp/tips_dump.go
package main

import (
 "fmt"

 "github.com/tqer39/ccw-cli/internal/i18n"
 "github.com/tqer39/ccw-cli/internal/tips"
)

func main() {
 if err := i18n.Init("ja"); err != nil {
  panic(err)
 }
 for _, t := range tips.Defaults() {
  fmt.Println(t)
 }
 fmt.Println("---")
 if err := i18n.Init("en"); err != nil {
  panic(err)
 }
 for _, t := range tips.Defaults() {
  fmt.Println(t)
 }
}
EOF
go run /tmp/tips_dump.go
rm /tmp/tips_dump.go
```

Expected: ja 7 行 + 区切り + en 7 行が出力され、それぞれの末尾 2 行が badge 仕様説明 tip(`Status: ...`, `PR: ...`)になっている。各行が 80 列以内で 1 行に収まっていることを目視確認。

- [ ] **Step 3: 実機で picker を起動して新 tip が時々表示されることを確認**

Run: `cd /tmp/ccw-cli-demo-repo && /tmp/ccw-badge-test`(または既存リポジトリで `/tmp/ccw-badge-test`)

- 起動して picker のフッターに `💡 Tip: ...` が表示されることを確認
- 何度か起動し直して `Status: ...` または `PR: ...` の tip が引けることを確認(seed は時刻ベースなのでタイミングを変えれば異なる tip が出る)
- `q` で終了

期待: 新 tip が picker フッターで折返さず 1 行で表示される。

このタスクはコミット不要。

---

## 完了条件(Acceptance criteria)

spec の Acceptance criteria を再掲:

- [ ] `internal/i18n/keys.go` に `KeyTipStatusBadge` / `KeyTipPRBadge` が定義され、`AllKeys()` にも含まれる(Task 1, 3)
- [ ] `internal/i18n/locales/ja.yaml` / `en.yaml` の `tip:` に `statusBadge` / `prBadge` の 2 文字列が追加されている(Task 2, 3)
- [ ] `internal/tips/tips.go` の `keys` が 7 要素になっている(Task 4)
- [ ] `go test ./...` が pass する(Task 5)
- [ ] `go vet ./...` / `golangci-lint run` 相当の lint が clean(Task 5)
- [ ] picker をローカル実行して、新 tip が時々表示されることを目視確認(Task 6)
- [ ] 各 tip が 80 列端末で 1 行に収まる(折返さない)(Task 6 Step 2 で目視)
