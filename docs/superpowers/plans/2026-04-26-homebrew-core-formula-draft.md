# Homebrew-core Formula Draft Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** homebrew-core 提出用の Formula 雛形 (`Formula/ccw.rb`) と draft 宣言の `Formula/README.md` を本リポにコミットする。実提出は notability 達成後の別タスク。

**Architecture:** ソースビルド型 Formula。`url` は GitHub source tarball、`depends_on "go" => :build`、ldflags で version 注入、test ブロックは PR #57 と同等の `assert_match` + `system -h`。`Formula/README.md` で「これは draft であり tap ではない」旨を明記し誤用を防ぐ。

**Tech Stack:** Homebrew Formula DSL (Ruby) / Go 1.25 build via `std_go_args(ldflags:)` / `ruby -c` 構文検証。

**Spec:** `docs/superpowers/specs/2026-04-26-homebrew-core-formula-draft-design.md`

---

## ファイル構成

| ファイル | 変更 | 責務 |
|---|---|---|
| `Formula/ccw.rb` | 新規 | homebrew-core 提出用の Formula draft |
| `Formula/README.md` | 新規 | draft 宣言 + 更新手順メモ |
| `docs/superpowers/specs/2026-04-26-homebrew-core-formula-draft-design.md` | 新規 | spec |
| `docs/superpowers/plans/2026-04-26-homebrew-core-formula-draft.md` | 新規 | このファイル |

`.goreleaser.yaml` / `tqer39/homebrew-tap` には触らない。

---

## Task 1: `Formula/ccw.rb` を作成

**Files:**

- Create: `Formula/ccw.rb`

- [ ] **Step 1: ファイル作成**

`Formula/ccw.rb` を以下の内容で新規作成:

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

`sha256` は v0.20.0 source tarball を `curl -sL .../v0.20.0.tar.gz | shasum -a 256` で算出済 (`52b315ed2fffc1e4c15fe68851b67475c1149650ba18df03a0b22dd82cc6e5a7`)。

- [ ] **Step 2: Ruby 構文チェック**

Run: `ruby -c Formula/ccw.rb`
Expected: `Syntax OK`

- [ ] **Step 3: brew audit（任意 / 環境があれば）**

Run: `brew audit --new --formula ./Formula/ccw.rb 2>&1` （`--strict` は notability で必ず落ちるので外す）
Expected: ライセンス検出、desc 形式、url 到達性、sha256 一致 等の検査が PASS。落ちる項目があれば修正。

`brew` がない / 動かない場合はスキップして良い（提出時に GitHub Actions で audit が走る）。

---

## Task 2: `Formula/README.md` を作成

**Files:**

- Create: `Formula/README.md`

- [ ] **Step 1: ファイル作成**

`Formula/README.md` を以下の内容で新規作成:

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

- [ ] **Step 2: markdownlint チェック**

Run: `npx markdownlint-cli2 Formula/README.md` または `pre-commit run --files Formula/README.md`
Expected: PASS。引っかかる場合は `.markdownlint-cli2.jsonc` の設定と diff を見比べる。

このリポは `markdownlint-cli2` を使用しているため、警告があれば対処。

---

## Task 3: コミット & PR 作成

**Files:** なし（git 操作のみ）

- [ ] **Step 1: 差分確認**

Run: `git status && git diff --stat`
Expected: 4 ファイル新規追加（`Formula/ccw.rb`, `Formula/README.md`, spec, plan）。

- [ ] **Step 2: Commit**

```bash
git add Formula/ccw.rb Formula/README.md \
        docs/superpowers/specs/2026-04-26-homebrew-core-formula-draft-design.md \
        docs/superpowers/plans/2026-04-26-homebrew-core-formula-draft.md
git commit -m "$(cat <<'EOF'
docs(brew): homebrew-core 提出用の Formula 雛形を追加

ソースビルド型の Formula draft を `Formula/ccw.rb` に置き、tap として
誤用されないよう `Formula/README.md` で draft 宣言する。実提出は
notability 達成後の別タスク。

Spec: docs/superpowers/specs/2026-04-26-homebrew-core-formula-draft-design.md
Plan: docs/superpowers/plans/2026-04-26-homebrew-core-formula-draft.md
EOF
)"
```

- [ ] **Step 3: Push**

このセッションは PR #57 用ブランチ `worktree-ccw-tqer39-ccw-cli-260426-104642` 上にいる。**A は B と独立した変更なので別ブランチに切るのが望ましい**:

```bash
git switch -c chore/homebrew-core-formula-draft HEAD
# 直前の commit はそのまま新ブランチに乗る（B のコミット履歴も含むので注意）
```

…が、より素直なのは PR #57 にまとめず、main から新ブランチを切って A の commit だけ載せる方式:

```bash
# B の PR が merge される前提が崩れた場合に備え、main から切る
git fetch origin
git switch -c chore/homebrew-core-formula-draft origin/main
# Formula/ccw.rb, Formula/README.md, spec, plan のみ別途作成 → コミット
```

→ **実装は「main からブランチを切り直し、A のファイルだけコミットする」方針**を採る。具体的には:

1. このセッションの commit は B 用の PR #57 で merge される
2. A 用の作業は worktree 切り直しが手間なので、**「A のファイルは現在のブランチ（B のブランチ）には含めず、別途 main から作業する」のが本来の理想**だが、worktree sandbox の都合上、現実的には:
   - 同じブランチに A のコミットを積み、PR #57 と同居させる（簡便）
   - またはこのセッションでは Task 1〜2 まで（ファイル作成のみ）を行い、commit / PR は手動で別ブランチで実施

→ **このセッションでは Task 1〜2 まで実行（ファイル作成・構文検証）し、commit / push / PR は判断を仰ぐ**方式にする。Step 4 は条件分岐で扱う。

- [ ] **Step 4: Commit / PR の進め方をユーザーに確認**

選択肢:

- a) このセッションのブランチ（B 用）に A のコミットを積み、PR #57 を「B + A 統合」に拡張する
- b) このセッションでは A のファイル作成のみ完了（commit せず）、別ブランチでの commit & PR は次セッションで実施
- c) 別ブランチを切って A 単体の PR を出す（git worktree の二重切り替え）

推奨: **a)** が最も摩擦が少ない。B と A は領域が異なるが、いずれも brew 関連 packaging 整備でスコープが近く、同 PR で扱っても review コストは低い。

ユーザーが選んだ方式で commit / push / PR を行う。

---

## 完了基準

- `Formula/ccw.rb` が存在し、`ruby -c` で構文 OK
- `Formula/README.md` が存在し、markdownlint で警告なし
- spec / plan ファイルがリポにコミットされている（または合意した方針で commit 済み）
- `tqer39/homebrew-tap` / `.goreleaser.yaml` / Go コードには一切触っていない
- 実提出 PR は出していない（notability 待ち）

## 完了後（このセッション外）

- notability（30 stars / forks / watchers 等）達成を待つ
- 達成後、最新 release のバージョン / sha256 に bump → `brew search ccw` で名前衝突確認 → homebrew-core への PR 作成
