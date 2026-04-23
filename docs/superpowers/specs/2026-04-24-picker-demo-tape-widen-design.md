# picker デモ GIF の横幅拡大 Design

## 背景

README（`README.md` / `docs/README.ja.md`）に掲載している `docs/assets/picker-demo.gif` は `docs/assets/picker-demo.tape` で生成している。現行設定は `Width 1024 / Height 640 / FontSize 24 / Padding 20` で、実効ターミナル幅は約 70 カラム。

picker の meta 行は次の構造で描画される（`internal/picker/delegate.go:44` `renderRow`）:

```text
> [badge]  [%-24s branch]  ↑N ↓N [✎N]  →  [PR badge] #NNN "<=30-char title…"
```

最大消費幅はおおよそ **100〜105 カラム**。現在は `truncateToWidth(top, width)` によって PR セルの末尾（タイトルや `"`）が切られて見える。

## 目的

デモ GIF のターミナル幅を拡張し、PR セルを含む meta 行全体を truncate せずに表示する。picker 側のコード挙動は変えない（読み手が「何が省略されず表示されるのか」を一目で把握できる状態にする）。

## 非目的

- picker のレンダリングロジック変更
- フォント・テーマ・再生速度の変更
- README の文面変更

## 変更点

### 1. tape パラメータ

対象ファイル: `docs/assets/picker-demo.tape`

| 項目 | 現行 | 新 |
|---|---|---|
| `Set Width` | `1024` | `1440` |
| `Set Height` | `640` | `640`（据え置き） |
| `Set FontSize` | `24` | `24`（据え置き） |
| `Set Padding` | `20` | `20`（据え置き） |

`Width` のみ 1440 に拡大する。想定カラム数はおよそ 100 カラム（`(1440 − 40) / ~14px ≒ 100`）で、`%-24s` の branch 幅 + 30 文字上限 PR タイトル + 固定メタ（badge / indicators / arrow / PR badge / `#NNN`）の合計を truncate なしで収容できる。

### 2. GIF 再生成

対象ファイル: `docs/assets/picker-demo.gif`

再生成手順は既存どおり:

```bash
bash docs/assets/picker-demo-setup.sh
vhs docs/assets/picker-demo.tape
cp /tmp/ccw-demo.gif docs/assets/picker-demo.gif
```

`picker-demo-setup.sh` 側のスクリプト変更は不要。

### 3. 検証

- 新 GIF の 1 周再生中、選択中行の meta に `"…"` による途中省略が発生していないこと（PR タイトル側の `…` は既存 30 文字切り詰めによるものなので許容）
- branch 列（最大 24 文字）とインジケータ（`↑N ↓N ✎N`）が切れていないこと
- bulk confirm 画面など別ビューでも表示崩れが出ていないこと

## 採用しなかった代替案

- **FontSize 22 + Width 1280** … 縦は同じでもフォントを小さくする必要があり、README サムネイルでの可読性が下がる
- **Width 1280 のみ**（中間案） … 約 88 カラムで、PR タイトル末尾がぎりぎり削られるケースが残る可能性があるため不採用

## リスク・懸念

- GIF の横:縦比が 1440:640 ≒ 2.25:1 と横長になる。GitHub README は画像を本文幅に合わせて縮小表示するため、原寸でのピクセル表示はされない。可読性が極端に落ちない範囲（1 行 100 カラム表示）に収めているため許容。
- VHS の実効カラム数はフォントメトリクスに依存するため、生成後に truncate が残っていたら tape を微調整（Width を +128 程度）する可能性がある。

## 参考

- `internal/picker/delegate.go` — `renderRow`, `renderPRCell`, `truncateToWidth`
- `docs/assets/picker-demo.tape` — VHS tape 定義
- `docs/assets/picker-demo-setup.sh` — 前準備スクリプト
- `docs/superpowers/specs/2026-04-24-picker-pr-viz-and-tape-resize-design.md` — 前回の tape リサイズ（縮小）設計。本設計はその方向を 1 部巻き戻す形になる
