# Brew Test 実質化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `.goreleaser.yaml` の `brews[0].test` を `system bin/"ccw", "-v"` から `assert_match version.to_s, shell_output("#{bin}/ccw -v")` + `system bin/"ccw", "-h"` に置き換え、`brew test ccw` で ldflags 注入とヘルプパスを実質的に検証する。

**Architecture:** 単一ファイル (`.goreleaser.yaml`) のテスト用ヒアドキュメントを 2 行に置き換える。Go コードや CI ワークフローには触らない。

**Tech Stack:** goreleaser v2 / Homebrew formula DSL（Ruby）/ goreleaser check による構文検証。

**Spec:** `docs/superpowers/specs/2026-04-26-brew-test-meaningful-design.md`

---

## ファイル構成

| ファイル | 変更 | 責務 |
|---|---|---|
| `.goreleaser.yaml` | 修正 | `brews[0].test` ブロックを 2 行に拡張 |

その他のファイル（README、CI ワークフロー、Go コード）は変更しない。

---

## Task 1: `.goreleaser.yaml` の test ブロックを書き換える

**Files:**

- Modify: `.goreleaser.yaml:78-79`

- [ ] **Step 1: test ブロックを置き換える**

`.goreleaser.yaml` 末尾の `brews:` 配下の `test:` ブロックを次に置換。

変更前（L78-79）:

```yaml
    test: |
      system bin/"ccw", "-v"
```

変更後:

```yaml
    test: |
      assert_match version.to_s, shell_output("#{bin}/ccw -v")
      system bin/"ccw", "-h"
```

インデント（4 スペース + `|` ヒアドキュメント本文 6 スペース）は既存に揃える。

- [ ] **Step 2: goreleaser check で構文確認**

Run: `goreleaser check`
Expected: `config is valid` 等の成功表示。エラーが出たらインデント・キー名を見直す。

`goreleaser` が PATH にない場合はインストール: `brew install goreleaser` あるいは `mise use -g goreleaser`。

- [ ] **Step 3: ヒアドキュメント中の Ruby 構文を確認**

Run:

```bash
ruby -c -e 'assert_match version.to_s, shell_output("#{bin}/ccw -v")
system bin/"ccw", "-h"'
```

Expected: `Syntax OK`。

（`assert_match` / `shell_output` / `bin` / `version` は Homebrew formula のメソッドで Ruby 単体評価では未定義だが、構文チェックは通る。）

- [ ] **Step 4: 既存 lint / format に違反していないか確認**

Run: `pre-commit run --files .goreleaser.yaml` または `yamllint .goreleaser.yaml`
Expected: PASS。yamllint 設定（`.yamllint.yml`）でブロックスタイルを許容しているはずなので、引っかかる場合は変更前の構造と diff を見比べる。

- [ ] **Step 5: Commit**

```bash
git add .goreleaser.yaml docs/superpowers/specs/2026-04-26-brew-test-meaningful-design.md docs/superpowers/plans/2026-04-26-brew-test-meaningful.md
git commit -m "ci(brew): brew test を assert_match + ヘルプ smoke に拡張"
```

---

## Task 2: PR 作成

**Files:** なし（git 操作のみ）

- [ ] **Step 1: ブランチ push**

このセッションの worktree は `worktree-ccw-tqer39-ccw-cli-260426-104642` ブランチ上にある。Push:

```bash
git push -u origin worktree-ccw-tqer39-ccw-cli-260426-104642
```

- [ ] **Step 2: PR 作成**

```bash
gh pr create \
  --title "ci(brew): brew test を assert_match + ヘルプ smoke に拡張" \
  --body "$(cat <<'EOF'
## Summary

- `brew test ccw` が `system bin/"ccw", "-v"`（exit code チェックのみ）だったのを `assert_match version.to_s, shell_output(...)` に変更し、ldflags 注入の正常性を実質検証する
- `system bin/"ccw", "-h"` を追加して i18n.Init / cli.PrintHelp パスの smoke test も兼ねる
- 次の release 時から `tqer39/homebrew-tap` の Formula に反映される
- homebrew-core 提出（別タスク A）に向けた前準備でもある

## Spec

`docs/superpowers/specs/2026-04-26-brew-test-meaningful-design.md`

## Test plan

- [x] `goreleaser check` で `.goreleaser.yaml` の構文を検証
- [x] ヒアドキュメント本文を `ruby -c` で構文検証
- [ ] 次回 release（v0.19.0 想定）後に `brew-audit.yml` ワークフローが走り、`brew test ccw` が新ブロックで PASS することを確認
EOF
)"
```

- [ ] **Step 3: PR URL を控える**

`gh pr view --web` または `gh pr view --json url -q .url` で確認できる。

---

## 完了基準

- `.goreleaser.yaml` の `brews[0].test` ブロックが新 2 行に置き換わっている
- `goreleaser check` が PASS する
- spec / plan ファイルがリポにコミットされている
- PR が作成されている
- Go コード / 他 CI ワークフロー / README には一切触っていない

## 完了後（このセッション外）

- 次の release v0.19.0 タグを切ったあと、`brew-audit.yml` の `brew test ccw` ステップが新ブロックで PASS することをログ確認
- 次タスク（A: homebrew-core 用 Formula 雛形）の brainstorming に進む
