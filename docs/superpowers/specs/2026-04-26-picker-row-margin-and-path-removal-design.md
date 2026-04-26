# Picker 行の右端マージン確保と `path:` 行の削除

## 背景

`ccw` をオプションなしで起動すると `internal/picker` の TUI が `ccw worktrees` 一覧を表示する。各行は 4 行構成で、左端に resume バッジ、`·` 区切りで worktree 名、右端に status バッジと indicators (`↑0 ↓0` 等) を右寄せ表示する。

二つの問題が発生している。

1. **右端の見切れ**
   - Cursor / cmux などの IDE 内蔵ターミナルでは PTY が報告する `WindowSizeMsg.Width` と実際の可視幅にズレがあり、右端の `↑0 ↓0` や `path:` が数文字分切り取られる。
   - スタンドアロンの Terminal.app 80×24 でも右端ピッタリで `↑0 ↓0` が描画され、マージンが無く窮屈。
2. **情報の冗長さ**
   - `path:` 行はフルパスを表示するが、ccw 管理 worktree は仕様上 `<repo>/.claude/worktrees/<name>` 配下に固定されており (`internal/worktree/worktree.go:115` `ccwPathMarker`)、全行で prefix が一致する。
   - フルパスは選択後の menu / delete confirm 画面で改めて表示されるため、一覧画面で並べる価値が薄い。
   - header に表示している worktree 名 (path basename) はラベルが無く `·` で繋がっているだけで、初見では「これが worktree 名」と読み取りにくい。

## ゴール

- 右端見切れを解消する。狭い (80 桁) 環境でも IDE 内蔵ターミナルでも、右端の status / indicators が確実に可視範囲内に収まる。
- 一覧の情報密度を上げる。冗長な行を削り、worktree 名であることが視覚的に分かるようにする。

## 非ゴール

- 一覧の情報を増やすこと (PR 状態の高度な可視化、メタデータ追加など)。
- bubbles list / lipgloss のレンダリング層への手入れ。
- フォールバック (gh 不在時) のテキスト出力レイアウト変更。
- 選択後の menu / delete confirm 画面のレイアウト変更。

## 設計

### レイアウト変更後

```text
> [⚡ NEW] 🌲 ccw-tqer39-ccw-cli-260426-155328              [LOCAL] ↑0 ↓0
    branch:  worktree-ccw-tqer39-ccw-cli-260426-155328
    pr:      (no PR)
```

(`branch:` 後 2 スペース、`pr:` 後 6 スペースで値開始位置を揃える既存実装に従う。)

Height 4 → 3 行。worktree 名の前に 🌲 (evergreen tree) を置き、worktree 識別子であることを示唆する。`branch:` `pr:` 行は従来通り。`path:` 行は削除する。

### 変更点

1. **右端マージン: 固定 4 文字**
   - `internal/picker/delegate.go` の `renderRow` に渡される `width` から 4 を引いた値を、`padBetween` と `truncateToWidth` の両方に渡す。
   - Cursor / cmux の幅報告ズレ (経験的に 1〜2 セル) を吸収しつつ、80 桁でも `↑0 ↓0` の表示は維持できる。
   - `width <= 4` の極端なケースは元の `width` を使う (フォールバック)。

2. **`path:` 行の削除**
   - `renderRow` 内で `pathLine` の組み立て・truncate・連結を削除する。
   - フルパスが必要な場面 (実体 path のコピーや確認) は menu / delete confirm 画面で従来通り表示されるため、機能的な後退は無い。

3. **header の worktree 名前に 🌲 を追加**
   - `header := fmt.Sprintf("%s%s · 🌲 %s", prefix, resume, name)` のように 1 文字 + space を追加する。
   - 既存の `padBetween` がセル幅で右寄せ位置を計算するので、絵文字の幅増 (2 セル) はそのまま吸収される。

4. **`rowDelegate.Height()` を 3 に変更**
   - 4 → 3。`Spacing(): 1` は据え置き。

5. **NO_COLOR / fallback パスの整合**
   - `noColor()` 経路でもアイコンと右端マージンが同様に効くことを確認する。
   - フォールバック出力 (`internal/picker/run.go:49`) は今回触らない (TUI ではない別経路)。

6. **README に RESUME 名非表示の注意書きを追加**
   - `README.md` / `docs/README.ja.md` の `## 🎬 Demo` (デモ) セクション直下に短い注意書きを追加する。文意:
     - 「`💬 RESUME` バッジは『この worktree に紐づく最新 session が存在する』ことだけを示す」
     - 「session のタイトルや最初のプロンプトを picker で表示することは現状していない (再開は `claude --continue` 任せで、claude code 側に委ねているため)」
   - 既存の `## 🎯 Picker reference` 内の Session badge 表とは役割を分け、Demo セクションでは GIF を見たユーザが期待を誤らないよう一文添える程度に留める。
   - 英文 (README.md) と日本語 (docs/README.ja.md) の両方を `readme-sync` skill の整合ルールに従って同時に更新する。

### 影響範囲

- `internal/picker/delegate.go`
  - `rowDelegate.Height` を 4 → 3
  - `renderRow` から `pathLine` 関連を削除
  - `renderRow` 内で右端マージン用の `effectiveWidth = width - 4` (条件付き) を導入
  - header に `🌲` prefix を追加
- `internal/picker/delegate_test.go`
  - `Height` の期待値 4 → 3
  - `renderRow` 系のテストで path 行を assert している箇所を更新
  - 必要なら header の `🌲` を assert
  - 右端マージンの境界テストを 1 つ追加 (狭い width で見切れないこと)
- `internal/picker/view_test.go`
  - path を assert している箇所があれば調整 (削除 or 期待値修正)
- `internal/picker/update.go` は **触らない** (`list.SetSize` の引数を縮めると bubbles list 全体の左寄せが起き、width 不一致が別経路で出る懸念があるため、行レベルで縮める方針)
- `README.md`
  - `## 🎬 Demo` 直下に RESUME 名非表示の注意書きを 1〜2 文追加
- `docs/README.ja.md`
  - `## 🎬 デモ` 直下に同等の注意書きを日本語で追加

### 検証

- 手元で `ccw` を起動し、Terminal.app 80×24 と Cursor / cmux 内蔵ターミナルで右端の見切れが無いこと、レイアウトが Height 3 で密に並ぶことを目視確認。
- `go test ./internal/picker/...` が通る。
- フォールバック (`gh` 不在) パスが従来通り (TUI 起動側に影響無いこと)。

## 受け入れ基準

- 一覧画面で各行が `header / branch / pr` の 3 行構成になっている。
- `path:` 行が一覧画面に表示されない。
- header の worktree 名の前に `🌲` が付いている。
- 80×24 の Terminal.app 上で、`↑0 ↓0` の右にスペース 4 文字以上の余白がある。
- Cursor / cmux 内蔵ターミナルで右端の `↑0 ↓0` が見切れない。
- 選択後の menu / delete confirm 画面では従来通りフルパスが見える (機能的後退なし)。
- `README.md` と `docs/README.ja.md` の Demo セクション直下に「RESUME バッジは最新 session の有無のみを示し、session 名や内容のプレビューは表示しない (`claude --continue` 任せ)」の注意書きが入っている。
- `go test ./...` が通る。
- markdownlint / cspell / 既存 lefthook hook を pre-commit で通過する。

## 今回外 (別 spec で扱う)

- **RESUME 名の表示**: 現在 `[💬 RESUME]` バッジのみ表示しているが、`~/.claude/projects/<encoded>/<最新mtime>.jsonl` の最初のユーザープロンプトを要約して `[💬 RESUME] "最初のプロンプト..."` のように出すと、`claude --continue` (`internal/claude/claude.go:42`) で実際に再開される session の中身が事前に分かる。本 PR とは独立に新 spec で扱う。jsonl format が public contract でない点 (`internal/worktree/has_session.go:11` のコメント) と、複数 jsonl 存在時の選択ルール、要約の長さ・多言語切り詰めが論点になる。
