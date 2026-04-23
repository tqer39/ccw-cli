# 2026-04-24 — README 特長セクション書き換え

## 背景

タグライン改稿 (commit 55e70c5) と同じ問題意識。現状の README「✨ Features / ✨ 特長」の 5 項目はジャーゴンや機構の説明になっていて、ユーザーの嬉しさが伝わらない：

- `🌳 Isolated sessions — each claude run gets its own git worktree`
  - 「isolated sessions」「分離」では嬉しさが不明。しかも **並行作業は `claude --worktree` 標準機能で、ccw 固有の価値ではない**。このままでは読者に誤認させる
- `🦸 Superpowers preamble — -s injects the ... workflow`
  - 「プリアンブル」はジャーゴン。`-s` で何が起きるかが伝わらない

## ゴール

特長項目を ccw 固有の価値にフォーカスして書き直す。タグライン改稿の方針（「仕組み」ではなく「何が起きる / 何が嬉しいか」を先頭に）を踏襲。

また **「ccw は橋渡しに徹するツールだ」というスタンスを前面に出す**：worktree 選択 → claude 起動、それ以降 ccw は手を引く。tmux / zellij / 常駐プロセスではない。セッション中に介入しない、という安心感を最初に伝える。

## 非ゴール

- Features 以外のセクション改稿
- 新機能の実装（Part B / D で別扱い）

## 変更内容

### 1. Features / 特長の行を差し替え

**先頭に「橋渡しに徹する」を新設**し、その下に既存価値を 5 つ並べる構成：

**EN (`README.md`)**

```md
- 🤝 **Hand-off and step aside** — pick (or create) a worktree, launch `claude` in it, then ccw exits. No daemon, no wrapper process, no coupling to tmux/zellij — just the bridge.
- 🧭 **Works from anywhere in the repo** — run `ccw` inside a worktree or subdirectory; ccw resolves the main repo automatically
- 🎯 **Worktree state at a glance** — pushed / ahead / behind / dirty, plus PR info, all in one picker
- 🧹 **Bulk cleanup** — `[clean pushed]` or `ccw --clean-all` sweeps the worktrees you're done with
- 🦸 **"Design first" startup** — `-s` tells claude to follow the brainstorming → writing-plans → executing-plans flow (prompts to install the superpowers plugin if missing)
- ➡️ **claude flags pass through** — anything after `--` goes to claude untouched, so `--model` and friends still work
```

**JA (`docs/README.ja.md`)**

```md
- 🤝 **橋渡しまでが仕事** — worktree を選ぶ（or 新規作成）→ その中で `claude` を起動 → ccw は終了。常駐プロセスもラッパーもなく、tmux/zellij にも噛まない。あとは claude の世界
- 🧭 **リポジトリ内のどこからでも起動** — worktree 内やサブディレクトリからでも `ccw` が動く（main repo を自動解決）
- 🎯 **worktree の状態が一目でわかる** — push 済 / ahead・behind / dirty、PR 番号を picker にまとめて表示
- 🧹 **溜まった worktree を一括掃除** — `[clean pushed]` / `ccw --clean-all` で push 済をまとめて削除
- 🦸 **"設計してから書く" 流儀で起動** — `-s` で brainstorming → writing-plans → executing-plans の手順を claude に指示（plugin 未導入なら入れるか確認）
- ➡️ **claude のオプションはそのまま届く** — `--` 以降の引数は素通しするので `--model` などが使える
```

項目順の意図: 橋渡しスタンス → 起動の自由度 → 観察 → 掃除 → 派生的利用 → 透過性。

### 2. Quick Start 直下のブロッククォート削除

現状：

> `ccw` also works from inside a worktree — it resolves the main repo via `git rev-parse --git-common-dir` and operates there, so you don't need to `cd` back to the project root first.

→ Features の `🧭` 項目と内容が重複するので削除。JA 側も同じ扱い。

## 設計原則への準拠

- Features 間は互いに独立した価値命題 → 並列関係のまま
- 機構（`--worktree` の仕組み）ではなくユーザー視点の結果を先頭に
- 既存の絵文字アイコンは視認性維持のため継承、並び順のみ「起動 → 観察 → 掃除 → 派生的利用 → 透過性」と意味順にそろえる

## 実装

- `README.md` / `docs/README.ja.md` の該当行と直下のブロッククォートを差し替え
- `readme-sync` skill で EN / JA の整合確認

## テスト

- 目視: ローカルで両 README をレンダリング確認
- `lefthook` の markdown lint / cspell / textlint が通ること

## リスク / 影響

- PR-B（auto-install）とは独立。この PR 単独でも整合。
- 既存の機能説明がなくなる箇所はないため、ユーザーの機能理解が毀損されない。

## PR スコープ

この spec は **PR-A** 単独用。Part B / C / D は別 spec / 別 PR。
