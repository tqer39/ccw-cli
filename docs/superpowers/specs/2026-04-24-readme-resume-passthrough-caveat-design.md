# 2026-04-24 — README `--resume` パススルー注意書き追加

## 背景

ccw は `--` 以降の引数を claude にそのまま渡す（`internal/cli/parse.go::splitAtDoubleDash`）。`claude` の `--resume <id>` を渡すと、argv の組み立て方によって挙動が分岐する：

| 呼び出し | `internal/claude/claude.go` で生成される argv |
|---|---|
| `ccw -- --resume ID`（picker の `[r] run` 経由） | `claude --permission-mode auto --resume ID` |
| `ccw -n -- --resume ID` | `claude --permission-mode auto --worktree --resume ID` |
| `ccw -s -- --resume ID` | `claude --permission-mode auto --worktree --resume ID -- "<preamble>"` |

### 何が問題か

`--worktree` は新しい worktree を作る claude 側フラグ。`--resume` は過去セッション継続。

- **`-n` / `-s` 併用**: 新 worktree を作りつつ、過去セッションの会話履歴中のファイル参照はもとの worktree を指している → 実ファイルと履歴がズレる
- **picker 再入場 + `--resume`**: 選んだ worktree と、resume 対象セッションがもともと動いていた worktree が違うと同様のズレ
- **`-s` 併用**: 末尾の `-- "<preamble>"` は初回メッセージ前提。resume に対して appendix メッセージとして飛ぶのは意図と違う

claude 側の CLI パーサーが併用を拒否するかどうかは未検証（claude 側の実装変動を踏まえると、ドキュメントで運用面から警告するのが妥当）。

## ゴール

README（EN / JA）に `--resume` パススルーの注意書きを追加。ユーザーが踏む前に気づけるようにする。

## 非ゴール

- コード側での検出 / 警告実装（別 PR 候補。今回はドキュメントのみ）
- `--resume` 機能自体のサポート追加

## 変更内容

### 配置

README の picker 説明（`### Worktree picker` 節）の直後、`PR display requires gh` の段落の前後どちらか。JA 側も同じ位置。

### 本文（EN）

```md
> ⚠️ **Passing `--resume` through `--` is unsupported.**
> `ccw -n -- --resume ID` and `ccw -s -- --resume ID` combine `claude --worktree` (new worktree) with `--resume` (continue a prior session); the resumed transcript's file references won't match the freshly-created worktree. Even the picker's re-entry path suffers the same mismatch if the selected worktree differs from the session's original. If a resumed session is what you want, run `claude --resume ID` directly — bypass ccw.
```

### 本文（JA）

```md
> ⚠️ **`-- --resume ID` のパススルーは非推奨です。**
> `ccw -n -- --resume ID` や `ccw -s -- --resume ID` は `claude --worktree`（新 worktree 作成）と `--resume`（過去セッション継続）を同時に使うことになり、resume された会話中のファイル参照が新 worktree の実体と合いません。picker 経由で既存 worktree に再入場する場合も、選んだ worktree と session 元の worktree が違えば同様のズレが出ます。過去セッションを resume したいときは ccw を介さず直接 `claude --resume ID` を呼んでください。
```

## 実装

- `README.md` の該当位置に blockquote 追加
- `docs/README.ja.md` の対応位置にも追加
- `readme-sync` skill で整合性確認

## テスト

- 目視で両 README レンダリング確認
- markdown lint / cspell / textlint

## PR スコープ

この spec は **PR-C** 単独用。
