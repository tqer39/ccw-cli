# 2026-04-26 — homebrew-core 提出用 Formula 雛形

## 背景

`ccw` は現状 `tqer39/homebrew-tap` 配下で `brew install tqer39/tap/ccw` 経由で配布している。homebrew-core (`brew install ccw`) への提出を将来検討するため、**事前に提出可能な Formula 雛形を本リポにコミット**しておきたい。

ただし homebrew-core 提出には notability 要件（30 stars / forks / watchers 相当）があり、現時点ではまだ達していない。**本タスクは雛形のコミットまで**で、実際の提出 PR は notability 達成後に別タスクで行う。

## 制約 — homebrew-core と tap の差分

| 項目 | tap（現状） | homebrew-core（要件） |
|---|---|---|
| ビルド方法 | pre-built tarball （goreleaser 生成） | ソースビルド必須（`depends_on "go" => :build`） |
| desc | 「Launch Claude Code in an isolated git worktree」(47 字) | 冠詞 NG・formula 名禁止・80 字以内 — 現状で OK |
| test | `system bin/"ccw", "-v"`（PR #57 で `assert_match` に拡張中） | `assert_match` 等の意味のあるテスト推奨 |
| 名前衝突 | tap 内自由 | homebrew-core 全体で一意 — `ccw` は要事前確認（提出時） |
| HEAD support | なし | 慣習的に入れる |

→ tap の Formula はそのまま流用できない。**ソースビルド版の新規 Formula** を起こす必要がある。

## ゴール

本リポ（`tqer39/ccw-cli`）に次を追加する:

1. `Formula/ccw.rb` — homebrew-core 提出用の Formula draft（ソースビルド・assert_match テスト・HEAD support 込み）
2. `Formula/README.md` — 「これは draft、tap として使うものではない」旨の宣言

## 非ゴール

- homebrew-core への実 PR 提出（notability 待ち、別タスク）
- `tqer39/homebrew-tap` の Formula 変更（既に goreleaser 自動生成）
- bottle 設定（homebrew-core 側で CI 経由で付与される）
- Linux 用 packaging の追加対応（同 Go ソースで両対応されている）
- 提出時のバージョン bump 自動化（提出時に手動 + 該当 release の sha256 再計算）

## Formula draft 内容

`Formula/ccw.rb`:

```ruby
class Ccw < Formula
  desc "Launch Claude Code in an isolated git worktree"
  homepage "https://github.com/tqer39/ccw-cli"
  url "https://github.com/tqer39/ccw-cli/archive/refs/tags/v0.20.0.tar.gz"
  sha256 "52b315ed2fffc1e4c15fe68851b67475c1149650ba18df03a0b22dd82cc6e5a7"
  license "MIT"
  head "https://github.com/tqer39/ccw-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X github.com/tqer39/ccw-cli/internal/version.Version=#{version}
      -X github.com/tqer39/ccw-cli/internal/version.Commit=brew
      -X github.com/tqer39/ccw-cli/internal/version.Date=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags:), "./cmd/ccw"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/ccw -v")
    system bin/"ccw", "-h"
  end
end
```

**設計ポイント:**

- `url` は GitHub の source tarball（`/archive/refs/tags/v0.20.0.tar.gz`）。homebrew-core はバイナリ tarball を許さない
- `sha256` は `52b315ed2fffc1e4c15fe68851b67475c1149650ba18df03a0b22dd82cc6e5a7`（v0.20.0 source tarball を `shasum -a 256` で算出）
- `head "..." branch: "main"` で `brew install --HEAD ccw` も可能に
- `ldflags` は `.goreleaser.yaml` と同等（`Commit` は `brew` 固定、`Date` はビルド時刻）
- `std_go_args(ldflags:)` は最近の Homebrew の慣習（出力先 `bin/ccw`、`-trimpath` 自動付与）
- `test` ブロックは PR #57 と同じ（`assert_match` + ヘルプ smoke）

## `Formula/README.md` 内容

````markdown
# Formula/ — homebrew-core 提出用 draft

このディレクトリの `ccw.rb` は **homebrew-core 提出を想定した draft** です。

- これは tap ではありません。`brew install ./Formula/ccw.rb` のような直接指定はサポートしていません
- 通常のインストールは `brew install tqer39/tap/ccw`（自動生成された pre-built バイナリ formula）を使ってください
- 提出 PR は notability 要件（stars / forks / watchers ≥ 30 など）達成後に別タスクで実施します
- 提出時には version / sha256 を最新 release に合わせて bump してから PR を出します

## 雛形を更新する手順（参考）

1. 最新 release タグの source tarball sha256 を取得:

   ```bash
   curl -sL https://github.com/tqer39/ccw-cli/archive/refs/tags/v<X.Y.Z>.tar.gz | shasum -a 256
   ```

2. `ccw.rb` の `url` / `sha256` を書き換える
3. `brew audit --new --formula ./Formula/ccw.rb`（ローカル Homebrew 環境がある場合）でドライ検証
````

## テスト戦略

ローカルで実行可能なものに限定（homebrew-core CI 相当のフル audit はローカルで再現困難）:

1. **Ruby 構文チェック**: `ruby -c Formula/ccw.rb`
2. **brew audit（任意）**: 環境がある場合のみ。`brew audit --strict --new --formula ./Formula/ccw.rb`
   - `--strict` で notability エラーが出るのは想定通り（雛形段階では PASS する必要なし）
3. **brew install ローカル検証（任意）**: `brew install --build-from-source ./Formula/ccw.rb` で実際にビルド + test が走るか
   - tap として認識されないため `Error: Calling Installation of ccw from a GitHub commit URL is disabled!` 系エラーが出る場合あり。スキップして OK

## リスク / 影響

- **リポに `Formula/` を置くことで tap として誤検出される可能性**: `Formula/README.md` で明示的に否定する
- **version / sha256 が古くなる**: 雛形なので許容。提出時に bump 必須なのは README に明記
- **homebrew-core での名前衝突**: 提出時に `brew search ccw` で再確認。被ったら `claude-code-worktree` などへの rename 検討（雛形側はそのまま、提出時に判断）
- **既存 CI への影響**: なし。`Formula/` は Go ビルド・test・lint 対象外

## PR スコープ

単一 PR。`Formula/ccw.rb` と `Formula/README.md` の追加 + spec / plan のコミット。
