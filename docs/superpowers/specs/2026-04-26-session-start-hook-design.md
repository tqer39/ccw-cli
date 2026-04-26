# 2026-04-26 — 新規 worktree セッションでの SessionStart hook による初期プロンプト注入

## 背景

ccw で新規 worktree から `claude` を起動するたびに、毎回手で

> このセッションは Claude Code の --worktree sandbox 内です。
> superpowers:brainstorming → superpowers:writing-plans → superpowers:executing-plans
> の順で進めてください。トピックはこれから相談します。

と打ち込んでいる。これを自動化したい。ただし以下の制約がある:

- 適用は **新規セッションのみ**。`--continue` での resume や `/clear` 後では発火させない。
- ccw は「a thin launcher for Claude Code's `--worktree`」を標榜しており、CLI 本体にプロンプト本文の知識を持ち込むのは前 PR (#53) の方針に逆行する。
- 他プロジェクトでも同じパターンを使えるよう、再利用可能な手順として残したい。

## ゴール

- ccw-cli プロジェクト内で `claude --worktree` 新規セッション開始時のみ、上記プロンプトが自動で初期コンテキストに注入される状態にする。
- `--continue` / `/clear` では発火しない。
- ccw 本体（Go コード）は一切変更しない。
- README に「他プロジェクトに転用するための手順」を記載し、ccw-cli 自身を生きたサンプルとして機能させる（dogfooding）。

## 非ゴール

- ccw への新サブコマンドや新フラグ追加（C1 方針: ccw は無関与）。
- プロンプト本文の動的テンプレート化や設定ファイル化（YAGNI、固定文字列で十分）。
- 他プロジェクトへの自動展開機構（例: `ccw init-hooks` 的なもの）。各プロジェクトは README を見て手で設定する。
- 既存 `enabledPlugins` 設定の改変。

## 決定事項

### D1. Claude Code 標準の `SessionStart` hook を使う

`SessionStart` hook は新規セッション開始時のイベントで、`matcher` により以下を分離できる:

- `startup` — 新規セッション開始時のみ発火
- `resume` — `claude --continue` / `--resume` での再開時
- `clear` — `/clear` 後

`matcher: "startup"` を指定することで「resume では効かない」要件を Claude Code 側のセマンティクスに委譲できる。ccw 側で startup と resume を区別する必要はなくなる。

### D2. hook 出力は `hookSpecificOutput.additionalContext` 形式

hook コマンドが stdout に以下の JSON を出力すると、`additionalContext` の値が初期コンテキストに連結される（公式形式）:

```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "..."
  }
}
```

stdout を裸テキストで返す方法もあるが、複数 hook の合成や将来の拡張性を考えると JSON 形式が無難。

### D3. hook 本体は `.claude/hooks/session-start-superpowers.sh` に切り出す

`.claude/settings.json` に長文 JSON のエスケープを埋め込むと改行・引用符の地獄になる。短いシェルスクリプトに切り出し、settings.json からは相対パスで参照する。

スクリプトの内容:

```sh
#!/usr/bin/env bash
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

`#!/usr/bin/env bash` + 実行権限 (`chmod +x`) を付与してコミットする。

### D4. `.claude/settings.json` に hook 定義を追加

既存の `enabledPlugins` を保持したまま、`hooks.SessionStart` を追加:

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

`command` はリポジトリルートからの相対パスで指定する。Claude Code は cwd をプロジェクトルートに置くため相対パスで動作する。

### D5. README に転用手順を追記

`README.md`（英語、ソース）に新規節 "Auto-prompt on new worktree sessions" 相当を追加し、以下を含める:

- なぜこれが便利か（毎回手打ちが不要になる、resume では発火しない）
- 仕組み（`SessionStart` hook + `matcher: "startup"`）
- 最小コピペ手順（`.claude/settings.json` の差分と `.claude/hooks/session-start-superpowers.sh` の中身）
- ccw-cli 自身がこれを使っている旨（`See .claude/settings.json in this repo for a working example.`）

`readme-sync` skill で `docs/README.ja.md` を同期する。

## アーキテクチャへの影響

- ccw の Go コード（`cmd/ccw/`、`internal/`）には影響なし。`internal/claude/claude.go` の `LaunchNew` / `LaunchInWorktree` / `Continue` のいずれも変更不要。
- 影響範囲はリポジトリ直下の `.claude/` 配下と README のみ。
- 既存テストの追加・修正なし。Go の lint / test には影響しない。

## 検証

手動スモークテスト（CI では検証しづらいので手動）:

1. このブランチをチェックアウトし、`ccw -n` を実行して新規 worktree を作成、claude が起動することを確認。
2. 新規セッションの初期メッセージに "このセッションは Claude Code の --worktree sandbox 内です..." が反映されていることを目視確認。
3. 同じ worktree を picker から resume → 上記プロンプトが **注入されないこと** を確認（resume matcher と分離されている確認）。
4. claude セッション内で `/clear` → 上記プロンプトが **再注入されないこと** を確認。
5. hook スクリプトを単体実行 (`bash .claude/hooks/session-start-superpowers.sh`) して JSON が valid であることを確認（`jq .` で parse できる）。

自動チェック:

- `lefthook` の既存 markdownlint / yamllint に影響しないこと（settings.json は JSON なので別系統）。
- shellcheck が通ること（既存の lefthook 設定に応じて、必要なら除外設定）。

## 移行ノート

- 破壊的変更ではない。既存ユーザー（このリポジトリで開発する開発者）にとっては「セッション開始時に毎回手で打っていたプロンプトが自動で出るようになる」という副作用のみ。
- 個人的にこの hook を無効化したい場合は `.claude/settings.local.json` で上書き可能（`.gitignore` 済み）。
- 他プロジェクトに展開したい場合は README の手順に従う。
