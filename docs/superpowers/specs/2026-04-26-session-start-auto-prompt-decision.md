# 2026-04-26 — SessionStart hook で「能動的な初期ターン」を起こすかの検討（結論: 現状維持）

## 背景

`.claude/hooks/session-start-superpowers.sh` + `.claude/settings.json` の SessionStart hook（matcher: `startup`）は仕様通り発火しており、`additionalContext` として以下の文字列が新規セッション開始時に注入されている:

> このセッションは Claude Code の --worktree sandbox 内です。
> superpowers:brainstorming → superpowers:writing-plans → superpowers:executing-plans
> の順で進めてください。トピックはこれから相談します。

しかし「`ccw -n` 直後に Claude が能動的に最初のターンを切り出してくれない（ユーザー入力を待つ）」という体験ギャップがある。これを埋めるべきかをブレインストーミングで検討した。

## 検討した選択肢

- **A. ccw 本体に責務を置く** — Go コード（`internal/claude/claude.go`）を改修し、`claude` 起動時に位置引数として初期プロンプトを渡す（フラグ or 固定テキスト）。
- **B. ccw + 設定ファイル** — `.claude/initial-prompt.txt` のような汎用ファイルを ccw が読み取って `claude` の位置引数に渡す薄い機構を追加する。
- **C. ユーザー側 alias / shell 関数** — ccw 改修ゼロで `ccw -n -- "<prompt>"` を alias 化する運用に倒す。
- **D. Claude Code の hook 仕様内で能動ターンを起こす** — SessionStart hook の出力スキーマで「prompt として送出」できるかを再調査。

## 決定: 現状維持（A〜D いずれも採用しない）

判断理由:

1. **hook は開発者ごとにカスタマイズする領域** — プロジェクト固有のワークフロー文字列を ccw 本体（Go）に持ち込むのは、ccw の "thin launcher for `claude --worktree`" 方針（PR #53、PR #56 spec D1）に反する。A は却下。
2. **B は ccw に新たな読み取り機構を増やす** — 「個別カスタマイズは Claude Code 側の機能で完結すべき」という整理に反する。汎用ファイル規約を ccw が定義し始めると、ccw が hook ランタイムの一部になってしまう。B は却下。
3. **C は機能としては成立するが、現行 hook と二重管理になる** — シェル alias と `.claude/settings.json` の両方を見ないと挙動が把握できなくなる。C は却下。
4. **D は Claude Code の現行仕様の範囲では不可能** — SessionStart hook の `hookSpecificOutput` は `additionalContext` / `systemMessage` 等の受動的な情報注入のみで、Claude のターンを能動キックする機構はない。仕様変更を待つ立場。
5. **`additionalContext` で十分なケースが多い** — ユーザーが何か発話した時点で hook の文脈が活きるため、UX 上の損失は「最初の一言を自分で書く必要がある」程度。能動ターン化のために本体改修や規約追加を行う費用対効果が見合わない。

## 副次決定

- 既存 SessionStart hook（`startup` matcher）はそのまま維持する。これは設計通り動いており、resume / clear で発火しないという要件も満たしている。
- README の "Auto-prompt on new worktree sessions" セクションも現状のままで、`additionalContext` ベースの説明として正しい。誤解を招く「自動でターンが始まる」のような表現は元々していないため修正不要。
- 将来 Claude Code が SessionStart hook で「初期ユーザーメッセージ注入」を正式サポートしたら、その時点で本決定を見直す。

## 関連

- 設計 spec: `docs/superpowers/specs/2026-04-26-session-start-hook-design.md`
- 実装 plan: `docs/superpowers/plans/2026-04-26-session-start-hook.md`
- 実装コミット: `44d4178` (PR #56)
