# Translate picker tips to English

**Status**: approved
**Date**: 2026-04-25

## Goal

Picker フッターに表示される `💡 Tip:` の文言を日本語から英語へ統一する。
英語が main のドキュメント (`README.md`) と同じ言語になり、海外 OSS ユーザにも picker の操作ヒントが伝わるようにする。

## Scope

In-scope:

1. `internal/tips/tips.go` の `defaults` 5 行を英語化
2. `docs/assets/picker-demo.gif` の再生成（picker の `💡 Tip:` フッターが英語で焼き込まれた状態にする）
3. このブランチを `worktree-rapid-lion-5f9c` のまま PR 化せず、`git:create-branch` skill で意味のあるブランチ名を切り直してから push & PR 化

Out-of-scope:

- 他の日本語 UI 文言（picker のメニュー、エラーメッセージなど。今回の指摘対象は tips のみ）
- `docs/README.ja.md` 等の日本語ドキュメント（ja は意図的に日本語のまま維持）
- tips 数の増減・並び替え（純粋に翻訳のみ）
- tip 選択ロジック（`PickRandom`, `pickFrom`）の変更

## Design

### 1. tips の英訳

`internal/tips/tips.go` の `defaults` を以下に置き換える:

```go
var defaults = []string{
    "Worktree name = session name; renaming with /rename is fine, ccw doesn't track it",
    "claude --from-pr <number> resumes a PR-linked session directly",
    "--clean-all sweeps pushed worktrees in bulk",
    "ccw -- --model <id> passes flags through to claude",
    "The RESUME badge is derived from ~/.claude/projects/",
}
```

並び順は元のまま（既存テストは順序非依存だが、git diff のレビュー容易性のため）。

#### テスト影響

`internal/tips/tips_test.go` は文字列内容を assert していない:

- `TestPickRandom_FromDefaultSet` — `Defaults()` のいずれかと一致するかのみ
- `TestPickRandom_Deterministic` — 同一 seed で同一結果
- `TestPickFrom_Empty` — 空集合のハンドリング
- `TestDefaults_NonEmpty` — 各要素が空白文字のみでないこと

すべて翻訳後も pass する想定。テストコードは触らない。

### 2. picker-demo.gif の再生成

`docs/assets/picker-demo.tape` 自体には日本語が含まれないので、tape ファイルの編集は不要。
ただし picker のフッターは録画時に `tips.PickRandom(time.Now().UnixNano())` で 1 つの tip を引いて表示する。現行の `picker-demo.gif` には日本語の tip が焼き付いているので、ビルド済 ccw を差し替えた上で GIF を撮り直す:

```bash
bash docs/assets/picker-demo-setup.sh
vhs docs/assets/picker-demo.tape
cp /tmp/ccw-demo.gif docs/assets/picker-demo.gif
```

`picker-demo-setup.sh` 内で `go build -o /tmp/ccw-demo-bin/ccw ./cmd/ccw` を行うので、翻訳後の `tips.go` がそのまま反映される。

GIF はランダムに 1 件しか映らないため、5 件すべての英語化を視覚で保証することはできない。コードレビュー + ユニットテストで担保し、GIF は「英語が映っている」だけ確認する。

### 3. PR ブランチ名

現在の作業ブランチは ccw が生成した `worktree-rapid-lion-5f9c`。このままでは PR タイトルとブランチ名の対応が読み取りにくい。

`git:create-branch` skill を使い、`main` から `docs/translate-tips-to-english` を切り、そこに今回の変更を載せて push する。worktree 自体の Git ブランチ名（`worktree-rapid-lion-5f9c`）は ccw 内部の命名規約として残し、PR を上げる用のブランチだけ別名で作る方針。

具体的な手順は writing-plans で詰める。

## Acceptance criteria

- [ ] `internal/tips/tips.go` の 5 行が英語化されている
- [ ] `go test ./...` が pass する（既存テストは無改修）
- [ ] `go vet ./...` / `golangci-lint run` 相当の lint が clean
- [ ] `docs/assets/picker-demo.gif` が再生成され、フッターに英語の tip が映っている
- [ ] PR ブランチ名が `worktree-*` ではなく内容を表す名前になっている
- [ ] `docs/README.ja.md` 等の日本語ドキュメントは無変更

## Risks / open questions

- **vhs 環境依存**: GIF 再生成はローカルに `vhs` が必要。実行に失敗した場合は GIF 差し替えだけ別 PR にする選択肢あり。
- **picker-demo.tape の Sleep 値**: 英語の tip は日本語より文字数が多くフッター行が長くなる可能性がある。tape の幅設定は `Width 1440` で十分余裕あり、Sleep 調整は不要見込み。再生成後の GIF を目視確認して問題があれば追修正。
