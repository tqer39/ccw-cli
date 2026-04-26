# SessionStart hook for new worktree sessions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** ccw-cli プロジェクトで `claude` の新規セッション開始時のみ、superpowers のフロー誘導プロンプトを `additionalContext` として自動注入する。

**Architecture:** Claude Code 標準の `SessionStart` hook を `matcher: "startup"` で登録。hook 本体は `.claude/hooks/session-start-superpowers.sh` に切り出し、JSON を stdout に書く。`.claude/settings.json` の既存 `enabledPlugins` を保持しつつ `hooks` キーを追記。ccw 本体（Go コード）は無変更。

**Tech Stack:** bash, jq (動作確認のみ), JSON (settings.json), Claude Code hooks API。

**Spec:** [docs/superpowers/specs/2026-04-26-session-start-hook-design.md](../specs/2026-04-26-session-start-hook-design.md)

---

## File Structure

| ファイル | 種別 | 責務 |
|---|---|---|
| `.claude/hooks/session-start-superpowers.sh` | 新規 | `additionalContext` を含む JSON を stdout に出力する hook 本体 |
| `.claude/settings.json` | 変更 | 既存 `enabledPlugins` に `hooks.SessionStart[matcher=startup]` を追記 |
| `README.md` | 変更 | 「Auto-prompt on new worktree sessions」節を追加 |
| `docs/README.ja.md` | 変更 | 上記の日本語訳（`readme-sync` skill で同期） |

---

## Task 1: hook スクリプトの作成と JSON 出力検証

**Files:**

- Create: `.claude/hooks/session-start-superpowers.sh`

- [ ] **Step 1: スクリプトを書く**

`.claude/hooks/session-start-superpowers.sh` を以下の内容で作成:

```sh
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
```

- [ ] **Step 2: 実行権限を付与**

```bash
chmod +x .claude/hooks/session-start-superpowers.sh
```

- [ ] **Step 3: 出力が valid JSON かを確認（=動作テスト）**

```bash
.claude/hooks/session-start-superpowers.sh | jq .
```

期待: JSON が parse でき、`hookSpecificOutput.additionalContext` フィールドに 3 行のメッセージが入っていること。エラーで終了しないこと。

`jq` が無い環境でもよいよう、フォールバック確認:

```bash
.claude/hooks/session-start-superpowers.sh | python3 -c 'import json,sys; print(json.load(sys.stdin)["hookSpecificOutput"]["additionalContext"])'
```

期待: 3 行のメッセージがそのまま表示されること。

- [ ] **Step 4: shellcheck / shfmt 通過確認**

```bash
shellcheck .claude/hooks/session-start-superpowers.sh
shfmt -d -i 2 -ci -bn .claude/hooks/session-start-superpowers.sh
```

期待: いずれも警告ゼロ・差分ゼロ。差分が出たら `shfmt -w -i 2 -ci -bn .claude/hooks/session-start-superpowers.sh` で適用してから再確認。

- [ ] **Step 5: コミット**

```bash
git add .claude/hooks/session-start-superpowers.sh
git commit -m "feat(hooks): add SessionStart hook script for new worktree sessions"
```

---

## Task 2: settings.json に hook を登録

**Files:**

- Modify: `.claude/settings.json`

- [ ] **Step 1: 現状確認**

```bash
cat .claude/settings.json
```

期待: `{"enabledPlugins": {"superpowers@claude-plugins-official": true}}` のみが入っている。

- [ ] **Step 2: hooks キーを追加**

`.claude/settings.json` を以下に書き換える:

```json
{
  "enabledPlugins": {
    "superpowers@claude-plugins-official": true
  },
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/session-start-superpowers.sh"
          }
        ]
      }
    ]
  }
}
```

- [ ] **Step 3: JSON validity 確認**

```bash
python3 -m json.tool .claude/settings.json > /dev/null && echo OK
```

期待: `OK` が表示される。

- [ ] **Step 4: 手動スモークテスト（新規セッション）**

別ターミナルで:

```bash
ccw -n
```

期待: 新規 worktree が作成され、起動した claude セッションの初期コンテキストに「このセッションは Claude Code の --worktree sandbox 内です。」で始まる 3 行のメッセージが反映されている（Claude が superpowers のフロー手順を踏もうとする / または最初の応答でその旨に言及する）。

確認したら `/exit` で終了。

- [ ] **Step 5: 手動スモークテスト（resume）**

picker から既存 worktree（同じ worktree でも別の既存でも可）を `[r] run` で resume:

```bash
ccw
# 既存の worktree を選んで [r] を押す
```

期待: 復帰後の会話に上記プロンプトが **再注入されない**（matcher が `startup` のため `resume` では発火しない）。

- [ ] **Step 6: 手動スモークテスト（/clear）**

新規セッションを開始したのち claude 内で `/clear` を実行。

期待: `/clear` 後にも上記プロンプトが **再注入されない**（matcher が `startup` のため `clear` では発火しない）。

- [ ] **Step 7: コミット**

```bash
git add .claude/settings.json
git commit -m "feat(claude): wire SessionStart hook in .claude/settings.json"
```

---

## Task 3: README.md に転用手順を追記

**Files:**

- Modify: `README.md`（新規節を `## ⚙️ Environment` の **直前** に挿入。位置は `## 🛠️ Development` よりも前、`## 📦 Installation` の後ろが自然）

実際の挿入位置は `## ⚙️ Environment` の手前（現状 `:122` の前）。

- [ ] **Step 1: 新規節を README.md に追加**

`## 📦 Installation` セクションの末尾（`*(optional)* [superpowers]...` 行の次の空行の後）と `## ⚙️ Environment` の間に、以下を挿入:

````markdown
## 🪝 Auto-prompt on new worktree sessions

This repo ships a `SessionStart` hook that injects a fixed instruction at the start of every **new** Claude Code session — never on `--continue` or `/clear`. Useful for steering each fresh worktree session into a consistent workflow (here: brainstorming → writing-plans → executing-plans).

**How it works**

- [`.claude/hooks/session-start-superpowers.sh`](./.claude/hooks/session-start-superpowers.sh) prints a JSON object whose `hookSpecificOutput.additionalContext` is added to Claude's initial context.
- [`.claude/settings.json`](./.claude/settings.json) wires it as a `SessionStart` hook with `matcher: "startup"`, so resume / clear are excluded.

**Reuse in another project**

1. Drop a script at `.claude/hooks/session-start-<your-name>.sh` that outputs JSON in the form:

   ```json
   {
     "hookSpecificOutput": {
       "hookEventName": "SessionStart",
       "additionalContext": "<your instruction>"
     }
   }
   ```

2. `chmod +x` it.
3. Add to `.claude/settings.json`:

   ```json
   {
     "hooks": {
       "SessionStart": [
         {
           "matcher": "startup",
           "hooks": [
             { "type": "command", "command": ".claude/hooks/session-start-<your-name>.sh" }
           ]
         }
       ]
     }
   }
   ```

See this repo's [`.claude/`](./.claude) for a working example.
````

- [ ] **Step 2: markdownlint を通す**

```bash
npm exec -y --package=markdownlint-cli2 -- markdownlint-cli2 --fix README.md
```

期待: 警告なしで終了。差分が出ていたら採用する。

- [ ] **Step 3: 目視確認**

`README.md` の追加節を読み返して、コードブロックのフェンス揃え・リンク先の妥当性をチェック。

- [ ] **Step 4: コミット**

```bash
git add README.md
git commit -m "docs(readme): document SessionStart hook for new worktree sessions"
```

---

## Task 4: docs/README.ja.md を同期（readme-sync skill）

**Files:**

- Modify: `docs/README.ja.md`

- [ ] **Step 1: readme-sync skill を起動**

Skill ツールで `readme-sync` を呼び出し、引数なし（既定動作）で実行。

期待: `README.md` の追加節と同等の日本語版が `docs/README.ja.md` の対応位置（`## ⚙️ 環境変数` の直前 / `## 📦 インストール` の後）に挿入される。

- [ ] **Step 2: 日本語訳の妥当性を目視確認**

特に確認したい点:

- 節タイトルは「🪝 新規 worktree セッションでの自動プロンプト注入」のような日本語になっているか
- 仕組みの説明（`SessionStart` hook + `matcher: "startup"`）が訳出されているか
- 「Reuse in another project」が「他プロジェクトへの転用」相当に訳されているか
- リンク先（`./.claude/hooks/session-start-superpowers.sh` など）は英語版と同じファイルを指しているか

ニュアンスがずれている場合は手で修正する。

- [ ] **Step 3: markdownlint を通す**

```bash
npm exec -y --package=markdownlint-cli2 -- markdownlint-cli2 --fix docs/README.ja.md
```

期待: 警告なしで終了。

- [ ] **Step 4: コミット**

```bash
git add docs/README.ja.md
git commit -m "docs(readme-ja): sync SessionStart hook section"
```

---

## Task 5: 最終検証とプッシュ準備

**Files:**

- なし（読み取りのみ）

- [ ] **Step 1: 全ファイルの差分確認**

```bash
git log --oneline main..HEAD
git diff --stat main..HEAD
```

期待: 4 コミット（hook script / settings.json / README.md / docs/README.ja.md）が並ぶ。差分は `.claude/` と README 2 本のみで、Go コードの変更が含まれていないこと。

- [ ] **Step 2: Go テストが影響を受けていないことを確認**

```bash
go test ./...
```

期待: 全部 pass（既存テストへの影響なし）。

- [ ] **Step 3: lefthook を手元で再実行（任意）**

```bash
lefthook run pre-commit
```

期待: 全 lint pass。pre-commit で自動 fix が入った場合は対応するファイルを再ステージしてアメンドではなく **新規コミット** を作る。

- [ ] **Step 4: PR 作成（ユーザー指示があれば）**

ユーザーから明示的に PR 作成を依頼された場合のみ:

```bash
gh pr create --title "feat: SessionStart hook for new worktree sessions" --body "$(cat <<'EOF'
## Summary
- 新規セッション開始時のみ superpowers フロー誘導プロンプトを `additionalContext` として注入する `SessionStart` hook を追加
- ccw 本体（Go コード）は無変更。Claude Code 標準の hook 機構のみで完結
- README に他プロジェクトへの転用手順を追加（日英同期済み）

Spec: docs/superpowers/specs/2026-04-26-session-start-hook-design.md

## Test plan
- [x] `.claude/hooks/session-start-superpowers.sh | jq .` で valid JSON を確認
- [x] `ccw -n` で新規セッション → 初期メッセージ反映
- [x] picker からの resume → 注入されない
- [x] `/clear` → 再注入されない
- [x] `go test ./...` pass

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

ユーザーから PR 依頼がない場合はこのステップはスキップ。

---

## Self-Review

**Spec coverage:**

- 仕様 D1 (`SessionStart` + `matcher: "startup"`) → Task 2 Step 2 で実装
- 仕様 D2 (`hookSpecificOutput.additionalContext` 形式) → Task 1 Step 1 でスクリプトに記述、Task 1 Step 3 で出力検証
- 仕様 D3 (hook を `.claude/hooks/session-start-superpowers.sh` に切り出し) → Task 1 Step 1
- 仕様 D4 (settings.json に hooks 追記) → Task 2 Step 2
- 仕様 D5 (README に転用手順) → Task 3、日本語同期は Task 4
- 検証項目（new / resume / clear のスモークテスト） → Task 2 Steps 4–6
- 検証項目（hook 単体実行で JSON valid） → Task 1 Step 3
- 検証項目（lefthook の lint 影響なし） → Task 1 Step 4 / Task 3 Step 2 / Task 5 Step 3
- 非ゴール（ccw Go コード無変更） → Task 5 Step 1 / Step 2 で間接検証

**Placeholder scan:** TBD/TODO/「適切なエラーハンドリング」「テストを書く（コードなし）」等の placeholder は検出されず。すべての code step に actual content あり。

**Type / path consistency:**

- スクリプトのパス `.claude/hooks/session-start-superpowers.sh` が Task 1, Task 2, Task 3, Task 5 で一貫
- settings.json のキー構造 (`hooks.SessionStart[].matcher`, `hooks.SessionStart[].hooks[].type/command`) が Claude Code 公式仕様準拠で Task 2 / Task 3 のサンプル間で一致
- メッセージ本文（3 行）は Task 1 のスクリプトと spec ドキュメントで一致

問題なし。
