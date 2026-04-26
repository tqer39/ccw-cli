# Add badge specification tips

**Status**: approved
**Date**: 2026-04-27

## Goal

Picker フッターに表示される `💡 Tip:` のローテーションに、Status badge と PR badge の意味を説明する 2 件を追加する。
既存の `resumeBadge` tip と同じスタイル(1 tip = 1 行 = 1 badge)で、ランダムプールの母数を 5 → 7 に増やす。

## Scope

In-scope:

1. `internal/i18n/keys.go` に tip キーを 2 つ追加(`KeyTipStatusBadge` / `KeyTipPRBadge`)し、`Defaults()` 用のキー集合にも含める
2. `internal/i18n/locales/ja.yaml` と `en.yaml` の `tip:` セクションに `statusBadge` / `prBadge` 文字列を追加
3. `internal/tips/tips.go` の `keys` 配列に新 2 キーを追加(母数 5 → 7)

Out-of-scope:

- Badge の色・絵文字・レンダリング自体の変更
- tip カテゴリ分け / フィルタ機能の追加
- 新規言語(zh/ko 等)の追加
- badge 仕様変更時に tip を自動同期する仕組み
- picker レイアウトロジック・フッター描画ロジックの変更
- `docs/assets/picker-demo.gif` の再生成(GIF はランダム 1 件しか映らないので必須ではない)

## Design

### 1. 文言

ja:

```yaml
tip:
  statusBadge: "Status: PUSHED=push 済 / LOCAL=未 push / DIRTY=未コミット"
  prBadge: "PR: OPEN / DRAFT / MERGED / CLOSED を gh pr list から色分け"
```

en:

```yaml
tip:
  statusBadge: "Status: PUSHED=pushed / LOCAL=unpushed / DIRTY=uncommitted"
  prBadge: "PR: OPEN / DRAFT / MERGED / CLOSED colored by gh pr list state"
```

ラベル列挙は picker 実装の現実と一致させる:

- Status badge ラベル(`internal/picker/style.go:32` 周辺): `PUSHED` / `LOCAL` / `DIRTY`
- PR badge ラベル(`internal/picker/style.go:81` 周辺): `OPEN` / `DRAFT` / `MERGED` / `CLOSED`

### 2. i18n キー追加

`internal/i18n/keys.go` の tip キー定義に 2 行追加:

```go
KeyTipStatusBadge Key = "tip.statusBadge"
KeyTipPRBadge     Key = "tip.prBadge"
```

`Defaults()` の検証用キー集合(`KeyTipRename, KeyTipFromPR, KeyTipCleanAll, KeyTipPassthrough, KeyTipResumeBadge` が並んでいる箇所)にも `KeyTipStatusBadge, KeyTipPRBadge` を追加。

### 3. tips プール拡張

`internal/tips/tips.go` の `keys` 配列に追加:

```go
var keys = []i18n.Key{
    i18n.KeyTipRename,
    i18n.KeyTipFromPR,
    i18n.KeyTipCleanAll,
    i18n.KeyTipPassthrough,
    i18n.KeyTipResumeBadge,
    i18n.KeyTipStatusBadge,
    i18n.KeyTipPRBadge,
}
```

順序は既存末尾に追加。`PickRandom` / `pickFrom` のロジックは無改修。

### 4. テスト影響

`internal/tips/tips_test.go` は文字列内容を assert していない:

- `TestPickRandom_FromDefaultSet` — `Defaults()` のいずれかと一致するかのみ → 母数増加でも pass
- `TestPickRandom_Deterministic` — 同一 seed で同一結果 → 母数増加で結果は変わるが決定性は維持
- `TestPickFrom_Empty` — 空集合ハンドリング → 影響なし
- `TestDefaults_NonEmpty` — 各要素が空白文字のみでないこと → 新 2 件も非空文字列なので pass

i18n 側の locale 完全性検証(全 locale で全 tip キーが定義されている)は新 2 キーも自動的にカバー。

テストコードは触らない方針。文言を assert するテストは追加しない(文言改訂のたびに壊れるため)。

## Acceptance criteria

- [ ] `internal/i18n/keys.go` に `KeyTipStatusBadge` / `KeyTipPRBadge` が定義され、`Defaults()` 用キー集合にも含まれる
- [ ] `internal/i18n/locales/ja.yaml` / `en.yaml` の `tip:` に `statusBadge` / `prBadge` の 2 文字列が追加されている
- [ ] `internal/tips/tips.go` の `keys` が 7 要素になっている
- [ ] `go test ./...` が pass する
- [ ] `go vet ./...` / `golangci-lint run` 相当の lint が clean
- [ ] picker をローカル実行して、新 tip が時々表示されることを目視確認(seed を変えれば必ず引ける)
- [ ] 各 tip が 80 列端末で 1 行に収まる(折返さない)

## Risks / open questions

- **狭い端末での折返し**: ja `prBadge` は表示幅 ≈ 55 列、ja `statusBadge` ≈ 60 列。標準 80 列なら 1 行に収まるが、60 列以下の端末では折返す可能性。既存 tip(`fromPR` ≈ 50 列)と同水準なので新規リスクではないと判断、許容。
- **badge ラベル変更との整合性**: 将来 picker 側で badge ラベルを変更した場合、tip 文言も追従する必要がある。自動同期は YAGNI として見送り、`internal/picker/style.go` 変更時に手動で確認する運用とする。
