# Picker ヘッダー行のインライン化 (`|` 区切り)

## 背景

`internal/picker` の TUI worktree 一覧では、各行のヘッダーが「左に名前ブロック、右に status バッジ + indicators を右寄せ」という構成になっている (`internal/picker/delegate.go:35` `renderRow`)。`padBetween` がターミナル幅に応じて間にスペースを詰めることで右寄せを実現している。

このレイアウトは 80〜100 桁程度では機能するが、ユーザーがターミナルを広く開いている (e.g. 200 桁超) と、左の worktree 名と右の status badge の間が大きく離れ、視線移動が辛い。各行で「この worktree の status はどれ？」を読むのに毎回左右往復が発生する。

picker のメニュー行 (`tagNew/tagQuit/tagDeleteAll/tagCleanPushed/tagCustomSelect`) は左寄せの 2 行構成なので影響しない。問題があるのは worktree 行のヘッダー 1 行だけ。

## ゴール

- ターミナル幅に依らず、worktree 名と status バッジ / indicators の視覚的距離を一定に保つ。
- 1 行に意味のあるブロックが何個並んでいるか (resume / 名前 / status / indicators) が一目で分かる。

## 非ゴール

- branch / pr 行のレイアウト変更。
- メニュー行 (delete all / clean pushed など) のレイアウト変更。
- フォールバック (`gh` 不在) のテキスト出力レイアウト変更。
- 選択後の menu / delete confirm 画面のレイアウト変更。
- 区切り文字の動的切替 (env var / 設定) — `|` 固定。

## 設計

### レイアウト変更後

通常の worktree 行:

```text
> [💬 RESUME] | 🌲 feat-login | [pushed] | ↑0 ↓0
    branch:  feat/login
    pr:      [open] #42 "feat: add login page"
```

prunable (実体ディレクトリが消えている worktree):

```text
> [💬 RESUME] | 🌲 stale-feature | (missing on disk)
    branch:  stale/feature
    pr:      (no PR)
```

prunable では status バッジと indicators を省略し、`(missing on disk)` 1 セルに集約する。これらは prunable 状態では意味を持たないため。

### `|` 区切りの仕様

- 文字: ASCII の `|`
- 前後: 半角スペース 1 個ずつ (= `" | "`)
- 色: lipgloss `Color("240")` (dim grey)。bg 透過。
- NO_COLOR (`noColor() == true`) のときは色を付けず素の `" | "`

### 既存 `·` の扱い

現状ヘッダー内で唯一使われている区切り文字は `RESUME · 🌲 name` の `·`。これも今回 `|` に置換する。混在 (`·` と `|` の併存) は読みにくいので統一。

### 右寄せの廃止

- `padBetween` 関数は呼び出し元が消えるため **削除** する。
- ヘッダー行は左から `Separator()` で連結された単一文字列になる。右側に余白が残るが、それは仕様。

### 4-cell 右マージン

- `effectiveWidth = width - 4` の保険は **残す**。`truncateToWidth` のみで使う。
- 理由: Cursor / cmux など IDE 内蔵ターミナルで報告幅と可視幅にズレがあるケースで、ヘッダーが極端に長くなったときに右端で 1〜2 セル切られるのを防ぐ。

### `Separator()` ヘルパ

`internal/picker/style.go` に追加:

```go
// Separator returns the dim-grey vertical bar used to join inline header
// segments (resume, name, status, indicators). Falls back to plain " | "
// when noColor() is true.
func Separator() string {
    if noColor() {
        return " | "
    }
    return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" | ")
}
```

`renderRow` ではローカルに `sep := Separator()` を一度取り出し、必要箇所で連結する。

### `renderRow` の新しい組み立て

```go
sep := Separator()

if wt.Status == worktree.StatusPrunable {
    // resume | tree name | (missing on disk)
    missing := dimMissingOnDisk()  // dim grey "(missing on disk)" or plain
    header := prefix + resume + sep + "🌲 " + name + sep + missing
    ...
} else {
    // resume | tree name | status | indicators
    header := prefix + resume + sep + "🌲 " + name + sep + status + sep + indicators
    ...
}
```

`(missing on disk)` の dim 化ヘルパも `style.go` に置く (`MissingOnDisk()` 程度)。

### 影響範囲

- `internal/picker/delegate.go`
  - `renderRow` のヘッダー組み立てを `Separator()` ベースに書き換え。
  - prunable 分岐を明示化 (現状は `indicators` を `"(missing on disk)"` で上書きしているだけ)。
  - `padBetween` 関数を削除。
  - `effectiveWidth` の計算は維持し、`truncateToWidth` のみで使用。
- `internal/picker/style.go`
  - `Separator()` を追加。
  - `MissingOnDisk()` を追加 (dim 化された `"(missing on disk)"`)。
- `internal/picker/delegate_test.go`
  - 既存テスト (`[RESUME]` / `🌲 foo` / `[dirty]` / `↑2 ↓1 ✎5` / 4-cell 右マージン) は内容に追加・並びの変更があっても Contains で通る想定。実行して落ちる箇所のみ更新。
  - 新規:
    - `TestRenderRow_HeaderUsesPipeSeparator` — NO_COLOR で `[RESUME] | 🌲 foo |` 相当のパターンを Contains。
    - `TestRenderRow_PrunableShowsMissingOnDisk` — prunable 行で `(missing on disk)` を Contains し、`[pushed]` `[dirty]` 等の status バッジを Contains しないこと、indicators (`↑0 ↓0`) を含まないこと。
    - `TestRenderRow_HeaderHasNoLargeRightPadding` — width=200 でヘッダー行の visible width が中身ぴったり (例: 80 セル未満) に収まり、200 に近い右端まで詰められていないこと。
- `internal/picker/style_test.go`
  - `TestSeparator_NoColor` — NO_COLOR で `" | "` を返すことを assert。
  - `TestMissingOnDisk_NoColor` — NO_COLOR で `"(missing on disk)"` を返すことを assert。
- `internal/picker/view_test.go`
  - 既存スナップショット系で `·` を assert している箇所があれば `|` に更新。

### 触らないもの

- `internal/picker/update.go`: `list.SetSize` 等は無関係。
- `internal/picker/run.go`: フォールバック出力は別経路。
- branch / pr 行レンダリング (`branchLine`, `prLine`, `renderPRCell`)。
- `tagNew/tagQuit/tagDeleteAll/tagCleanPushed/tagCustomSelect` の 2 行レンダリング。
- i18n キー (新規文言は `(missing on disk)` 既存のみ流用)。

### vhs tape / GIF

- `docs/picker-demo.tape` のコマンド列は変更不要。
- ただし出力の見た目が変わるため tape 再録 → GIF 差し替えが必要。
- 本 spec のスコープ外として、別 commit / 別 PR で対応する (本 PR の作業順としては最後に tape を再録)。

## 検証

- 手元で `ccw` を起動し、80 桁、120 桁、200 桁のターミナルで以下を目視確認:
  - 全ての worktree 行で resume / 🌲 name / status / indicators が `|` で連結されて連続表示される。
  - 200 桁でも各セル間の距離が一定 (3 セル) で、視線往復が発生しない。
- prunable な worktree (実体ディレクトリ削除) で `(missing on disk)` が status / indicators の代わりに表示される。
- NO_COLOR=1 で起動して `|` が dim 化されず素の `|` で出ること、`(missing on disk)` も同様。
- `go test ./internal/picker/...` が通る。
- `go test ./...` が通る。
- markdownlint / cspell / 既存 lefthook hook を pre-commit で通過する。

## 受け入れ基準

- worktree 行のヘッダーが `prefix + resume | 🌲 name | status | indicators` の左寄せインライン形式になっている。
- 既存の `·` 区切りが `|` に置換されている (混在しない)。
- `|` は color 環境では dim grey (240)、NO_COLOR では素の `|`。
- prunable 行は `prefix + resume | 🌲 name | (missing on disk)` で、status バッジ / indicators が含まれない。
- ターミナル幅 200 桁でも各要素間が `" | "` 固定 (大きな余白を挟まない)。
- 4-cell 右端マージン (`truncateToWidth`) により header / branch / pr 行の visible width は `width - 4` 以下に収まり、IDE 内蔵ターミナル (Cursor / cmux 等) の幅報告ズレでも右端が見切れない。
- `padBetween` が削除されている。
- `internal/picker/style.go` に `Separator()` と `MissingOnDisk()` が追加されている。
- 新規テスト (pipe separator / prunable / 大幅 width で右側余白なし / NO_COLOR Separator / NO_COLOR MissingOnDisk) が追加され通る。
- `go test ./...` が通る。

## 今回外 (別 spec で扱う)

- vhs tape / GIF の差し替え (本 PR 内の最後に handler として組み込むか、後続 PR にするかは plan 段階で決定)。
- README の picker レイアウト説明文の更新 (該当箇所が無ければ不要)。
