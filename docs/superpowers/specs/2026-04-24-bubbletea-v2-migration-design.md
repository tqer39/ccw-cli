# bubbletea/bubbles/lipgloss v2 移行

- ブランチ: `renovate/github.com-charmbracelet-bubbletea-2.x`
- 関連 PR: [#33](https://github.com/tqer39/ccw-cli/pull/33)（本 PR）、[#25](https://github.com/tqer39/ccw-cli/pull/25)（bubbles v2、本対応でクローズ予定）

## 背景と問題

Renovate が生成した PR #33 は `github.com/charmbracelet/bubbletea` を v1 → v2 に上げるが、`go-lint` / `go-test` / `go-build` がすべて失敗している。

原因は 2 つ:

1. **コードの import パスが未変更**: `internal/picker/` 配下 9 ファイルは `github.com/charmbracelet/bubbletea`（v1 パス）を import したまま。v2 では `.../v2` サフィックスが必要。
2. **推移的依存の不整合**: `bubbles@v1.0.0` が `bubbletea/v2` を参照するため、`bubbles` も v2 に上げないと `go.sum` が整合しない。

したがって bubbletea 単独の昇格では済まず、関連する `charmbracelet/*` 群を **同時に v2 に揃え、コード側の API 書き換えも同 PR で行う** 必要がある。

## スコープ（本 PR）

### 含む

1. `go.mod` / `go.sum` で以下を v2 系に統一
   - `github.com/charmbracelet/bubbletea/v2`（既に v2.0.2）
   - `github.com/charmbracelet/bubbles` → `bubbles/v2`
   - `github.com/charmbracelet/lipgloss` → `lipgloss/v2`（現状 indirect だが直接利用、direct 化も可）
   - `github.com/charmbracelet/x/exp/teatest` → v2 対応の最新バージョン
   - `github.com/charmbracelet/x/exp/golden` → 必要に応じて更新
2. `internal/picker/` のコード・テスト書き換え
   - import パスを v2 系へ
   - bubbletea v2 の Key イベント / メッセージ API 変更に追随
   - `bubbles/v2/list` の API 差分に追随
   - `lipgloss/v2` のカラー/スタイル API 差分に追随
3. PR #25（bubbles v2）のクローズ（本 PR が包含するため）

### 含まない（別対応）

- Renovate 設定の `groupName` 追加（`charmbracelet/*` を束ねる設定）は別 PR。
- picker の機能追加・UI 変更（純粋移行のみ）。
- 他ライブラリ（`spf13/pflag` 等）のアップデート。

## 対象ファイル

`internal/picker/` 配下（合計 1,558 行）:

| ファイル | 行数 | 主な v2 影響点 |
|---------|-----|---------------|
| `bulk.go` / `bulk_test.go` | 36 / 51 | - |
| `delegate.go` / `delegate_test.go` | 98 / 133 | `list.ItemDelegate` API |
| `model.go` / `model_test.go` | 229 / 348 | `tea.Model` / `tea.Cmd` / Key 生成 |
| `run.go` / `run_test.go` | 82 / 141 | `tea.NewProgram` / `tea.WithAltScreen` / `teatest.Send` |
| `style.go` / `style_test.go` | 103 / 52 | `lipgloss` API |
| `update.go` | 194 | `tea.KeyMsg` の Type/Runes 参照、`tea.Quit`、`tea.WindowSizeMsg` |
| `view.go` | 91 | `tea.Model.View` / lipgloss 描画 |

## アーキテクチャ方針

本対応は **API 移行のみ** で、`internal/picker/` のアーキテクチャ（責務分割、ファイル境界）は現状維持する。

- 既存のファイル分割（`model` / `update` / `view` / `delegate` / `style` / `run` / `bulk`）はそのまま。
- 内部で直接 `tea.KeyMsg.Type == tea.KeyRunes` のように型定数を使っている箇所は、v2 の推奨パターン（`msg.String()` ベースまたは新しい `KeyPressMsg` / `key.Matches`）に置き換える。置換時に 1 箇所のヘルパーに集約する必要はない（YAGNI）。

## テスト戦略

`*_test.go` の `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}` や `tea.KeyMsg{Type: tea.KeyEnter}` などの v1 リテラル生成が 30 箇所以上ある。v2 の新 API に書き換え、**既存テストケースそのものは保持**（移行で挙動が変わらないことを保証するゴールデンとして機能）。

`run_test.go` の `tm.Send(tea.KeyMsg{...})` も同様に書き換え。`teatest` が v2 対応されているかは実装段階で確認し、未対応なら方針を相談（最悪、該当テストを一時的に skip し、手動検証で代替）。

## 成功基準

- PR #33 の CI で以下がすべて green:
  - `go-build`, `go-lint`, `go-test`, `shellcheck`, `shfmt`, `bats`, `workflow-result`
- `go test ./... -race -coverprofile=coverage.out` がローカルでも pass
- picker を手動起動（`ccw` / `ccw pr` 相当）して、既存キー操作（enter / q / r / d / y / n / b / s / ctrl+c / 矢印キー）が v1 と同等に動作
- `go.mod` に v1 の charmbracelet パッケージが残っていない（`bubbletea`（無印）、`bubbles`（無印）、`lipgloss`（無印）がゼロ）

## リスクと緩和

| リスク | 緩和 |
|-------|-----|
| v2 の breaking change（特に Key 判定と lipgloss）で挙動が微妙に変わる | 実装プラン段階で v2 Release Notes を精読。テストは既存ケースを極力保持し、差分が出たら都度検討 |
| `x/exp/teatest` の v2 対応版が存在しない / 不安定 | 取得できない場合は該当テストを一時 skip、手動検証で代替。方針は発見時に相談 |
| 推移依存の連鎖で想定外のパッケージ更新が発生 | `go mod tidy` の差分を都度確認し、明らかに無関係な更新は別 PR に切り出す |
| v2 で `list.Model` の描画仕様が変わり、現在の `tape` 出力 (`picker-pr-viz-and-tape-resize`) が崩れる | 手動確認でスナップショット感のある視認。崩れがあれば同 PR で最小修正 |

## ロールバック戦略

問題が広範に発生した場合は、PR ブランチを破棄し以下の代替を検討:

- Renovate が再度同じ PR を作るので、その前に Renovate 側で bubbletea メジャーを一時的に `allowedVersions: "<2.0.0"` で固定。
- v2 移行は後日、別ブランチで時間をかけて実施。

ただしこれは **最終手段**。まずは本プランで進める。

## 次ステップ

本仕様の承認後、`superpowers:writing-plans` スキルで詳細実装プランを作成する。
