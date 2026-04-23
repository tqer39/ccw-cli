# picker の PR 視覚強化と tape サイズ縮小

- Date: 2026-04-24
- Status: Draft

## 背景

現状の picker 行は `internal/picker/delegate.go:renderRow` が描画しており、
上段に `[BADGE] branch  ↑N ↓M  #42 open "title"` を、下段に worktree path を
出力する。`[BADGE]` はステータス色（pushed=緑 / local=黄 / dirty=赤）が付くが、
PR 部は単なる文字列連結で背景色や動線表現がなく、worktree と PR の対応関係が
視覚的に弱い。

またデモ用の `docs/assets/picker-demo.tape` は `1280x780 / FontSize 28 / Padding 28`
で録画しており、README に埋め込んだ時に横幅が大きすぎる。もう少しコンパクトに
録画し直したい。

## ゴール

1. picker 行の PR 部分に 3 つの視覚強化を入れる
   - **A. 状態バッジ**: `[OPEN] / [DRAFT] / [MERGED] / [CLOSED]` を GitHub 風配色で
   - **B. PR ブロック強調**: `[STATE] #N "title"` 全体を状態色の薄い背景で囲む
   - **C. 動線グリフ**: worktree メタ情報と PR ブロックの間に `→` を挟む
2. `picker-demo.tape` を `1024x640 / FontSize 24 / Padding 20` で録画し直す

## 非ゴール

- PR タイトル最大長（現状 30 文字）の変更
- bulk 画面・menu 画面の配色変更
- `gh` PR 取得ロジックの変更
- picker 2 行構成（top = status/branch/indicators/PR, bottom = path）の変更

## レンダリング仕様

### 現状

```text
  [PUSHED] feat/login           ↑1 ↓0  #42 open "add login page"
  /path/to/worktree
```

### 変更後

```text
  [PUSHED] feat/login           ↑1 ↓0  →  [OPEN] #42 "add login page"
  /path/to/worktree
                                          ^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                          状態色の薄背景で PR セル全体を包む
```

- `→` は worktree メタ情報と PR セルの間にだけ挟む（デフォルト色 / 前後に空白）
- 状態バッジ（濃色背景）と PR セル全体（薄色背景）で 2 段階の色強調
- 状態バッジ文字列から `state` の小文字表記を削除し、`[OPEN]` のみで状態を表現

### 状態別配色（8-bit ANSI）

| 状態 | 濃背景 (バッジ) | バッジ文字色 | 薄背景 (PR セル) | PR セル文字色 |
|---|---|---|---|---|
| OPEN | `2` (green) | `0` | `22` (dark green) | デフォルト |
| DRAFT | `8` (dark gray) | `15` | `237` (very dark gray) | デフォルト |
| MERGED | `5` (magenta) | `15` | `53` (dark purple) | デフォルト |
| CLOSED | `1` (red) | `15` | `52` (dark red) | デフォルト |

状態バッジ（濃背景）は PR セル（薄背景）の上に重ねて描画する。`lipgloss` は
ネストされたスタイルの背景色を内側優先で描画するので、バッジ部分は濃背景、
その他の `#N "title"` 部分は薄背景になる。

### PR が取得できない場合

| ケース | 描画 |
|---|---|
| `prUnavailable=true` | PR 列省略（`→` も出さない） |
| PR が `nil`（該当ブランチに PR 無し） | `→ (no PR)`（`(no PR)` はグレー `240`） |

### NO_COLOR=1 フォールバック

- 状態バッジ: `[open] / [draft] / [merged] / [closed]`（小文字プレーン）
- PR セル背景なし
- 動線グリフは `->`（半角 2 文字、一定幅）
- `(no PR)` は色なし

## 実装影響範囲

| ファイル | 変更内容 |
|---|---|
| `internal/picker/style.go` | PR 状態 → 配色マッピングを追加（`prBadgeStyle(state)` / `prCellBackground(state)` / `PRBadge(state)` / `PRCellStyle(state)` 的な関数）。`NO_COLOR=1` 時のフォールバックも同ファイルで扱う |
| `internal/picker/delegate.go` | `renderRow` で PR 列を `renderPRCell(pr, selected)` として切り出し、動線グリフ `→` を間に挟む。`lipgloss` の背景色が行末まで伸びるのを防ぐため PR セルの幅を明示的に絞る |
| `internal/picker/delegate_test.go` | 新レンダリングに追従。`NO_COLOR=1` / `prUnavailable` / PR nil の 3 系統テストは維持し、各 PR 状態（OPEN / DRAFT / MERGED / CLOSED）で状態色とグリフが出ることを検証するテストを追加 |
| `docs/assets/picker-demo.tape` | `Set FontSize 24` / `Set Width 1024` / `Set Height 640` / `Set Padding 20` に変更。`TypingSpeed` / `PlaybackSpeed` / `Sleep` 値は据え置き |
| `docs/assets/picker-demo.gif` | `picker-demo-setup.sh` → `vhs docs/assets/picker-demo.tape` で再生成 |

## テスト計画

- `go test ./internal/picker/...` をグリーンに
- `renderRow` のテストで以下を網羅
  - 各 PR 状態（OPEN / DRAFT / MERGED / CLOSED）でバッジ文字列 `[OPEN]` 等が含まれる
  - 動線グリフ `→`（NO_COLOR では `->`）が worktree 列と PR 列の間に入る
  - `prUnavailable=true` で PR 列と `→` の両方が省略される
  - PR が `nil` で `(no PR)` が出る
  - `NO_COLOR=1` で背景色が含まれない
- `go vet ./...` をグリーンに
- 手元で `vhs docs/assets/picker-demo.tape` を走らせ、GIF サイズ・見た目を目視確認

## リスクと緩和

- **`lipgloss.Background` が行末まで延びる**: PR セルに `Width()` で明示幅を指定して抑える。
  テストで背景色 ANSI エスケープの出現位置を検証するのはノイズが高いので、
  目視確認 + 既存の「文字列に特定トークンが含まれる」系のアサーションで代替する。
- **タイトル末端で薄背景が途切れる見た目**: タイトル最大 30 文字 + 前後の `"` で閉じるので、
  セル幅は「バッジ幅 + PR 番号幅 + タイトル 30 + 記号」を基準に固定的に見積もる。
- **GIF ファイルサイズ変動**: 解像度が下がる方向なので悪化リスクは低い。

## 作業順序（実装時）

1. `style.go` に PR 状態配色 / NO_COLOR フォールバックを追加
2. `delegate.go` の `renderRow` を更新（PR セル切り出し + グリフ挿入）
3. テスト追加・修正 → `go test ./internal/picker/...` グリーン
4. `picker-demo.tape` の数値を差し替え
5. `picker-demo-setup.sh` → `vhs` で GIF 再生成し README に差分を確認
6. コミットは論理単位で分ける（picker の色変更 / tape リサイズ / GIF 再生成）
