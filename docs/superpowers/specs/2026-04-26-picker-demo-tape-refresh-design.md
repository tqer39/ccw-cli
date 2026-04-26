# picker-demo.tape refresh for the new picker TUI

## 背景

PR #71 で picker の行レイアウトが変更された。

- Height: 4 → 3（一覧から `path:` 行を削除）
- header に 🌲 prefix と worktree 名
- 右端 4 セルのマージンを確保（`width > 4` なら `effectiveWidth = width - 4`）

`docs/assets/picker-demo.tape` と `docs/assets/picker-demo.gif` は旧 TUI 時点で作られたままで、コメントの記述（path 行への言及など）が現在の表示と齟齬がある。GIF を新 TUI で再生成しつつ、tape 側のコメントを最新表示に揃える。

## 目的

- 最新 picker の表示で `docs/assets/picker-demo.gif` を再生成できる tape を維持する
- tape のコメントを新 TUI 文言（`🌲 <name>`、`[STATUS]`、`RESUME/NEW`、PR セル）に揃え、将来の保守者が混乱しないようにする

## 非目標

- デモのストーリー（4 worktree 巡回 → submenu → bulk → cancel → quit）の再構成
- 複数 tape への分割（用途別 GIF）
- `Output` を `docs/assets/picker-demo.gif` に直接書き出す運用変更
- picker 自体の挙動変更

## 影響範囲

- `docs/assets/picker-demo.tape`
- `docs/assets/picker-demo-setup.sh`（最低限の整合性チェックのみ）
- `docs/assets/picker-demo.gif`（再生成）

## tape の変更点

### コメント

- 旧 TUI 由来の表現（`path: 行` 等）を削除
- walk-through コメントを新 TUI 表現に揃える
  - `[OPEN]` / `[DRAFT]` / `[MERGED]` / `[CLOSED]` の PR セル
  - `[PUSHED]` / `[LOCAL]` / `[DIRTY]` の worktree status
  - `RESUME` / `NEW` のセッションバッジ
  - `🌲 <name>` の worktree 名 prefix
- submenu / delete confirm 画面で「フルパスが見える」点に簡潔に触れる

### Sleep の調整

- 初期 picker 表示の `Sleep 7000ms` は維持（PR fetch と読み込みの双方を待たせる）
- 各 `Down` 後の hover Sleep を `3500ms` → `3000ms` に統一して GIF サイズを微減
- submenu / bulk confirm の Sleep は現行値を踏襲

### terminal サイズ

- `Width 1440 / Height 640 / FontSize 24` を維持
- 右端マージン（width-4）はこのサイズで問題なく収まる

## setup.sh の変更点

- 4 worktree 構成・mock gh・PROJECTS encoding は変更しない
- `chore/cleanup` を `DIRTY` に確実に保つよう、ワークツリー作成直後に
  `git -C /tmp/ccw-demo/.claude/worktrees/chore-cleanup status --porcelain` の結果が空でないことを軽くアサート（`prunable` 行も含まれないことを併せて確認）

## 再生成手順

```bash
bash docs/assets/picker-demo-setup.sh
vhs docs/assets/picker-demo.tape
cp /tmp/ccw-demo.gif docs/assets/picker-demo.gif
```

## 検証

GIF を目視で次の点を確認する。

1. 各行が 3 行構成（header / branch / pr）で表示される
2. header に `🌲 <name>` が表示される
3. 右端で `↑0 ↓0` 等が見切れない
4. 4 worktree の組み合わせが順に hover される
   - `feat/login`: `[PUSHED]` + `[OPEN] #42` + `RESUME`
   - `feat/dashboard`: `[PUSHED]` + `[MERGED] #44` + `RESUME`
   - `feat/picker`: `[LOCAL]` + `[DRAFT] #43` + `NEW`
   - `chore/cleanup`: `[DIRTY]` + `[CLOSED] #45` + `NEW`
5. submenu (`[r] run / [d] delete / [b] back`) でフルパスが表示される
6. `[clean pushed]` 経由の bulk confirm がプレビュー表示される
7. `N` で cancel、`q` で picker 終了

## ロールバック

tape / setup.sh / GIF の変更前バージョンに戻すだけで完結する。picker 本体には触れない。
