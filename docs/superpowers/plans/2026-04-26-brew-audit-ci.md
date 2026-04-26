# Brew Audit CI Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `tqer39/homebrew-tap` で公開している `ccw` formula を継続的に検証する macOS CI を追加し、`brew audit` / `brew install --build-from-source` / `brew test` のいずれかが壊れた瞬間に検知できるようにする。

**Architecture:** `.github/workflows/brew-audit.yml` を新規追加。`macos-latest` runner（arm64）で `tqer39/tap` を tap して `ccw` formula を audit + install + test する単一ジョブ構成。pull_request / push to main / weekly schedule / workflow_dispatch の 4 トリガーで起動し、いずれの失敗もジョブ失敗として扱う（必須化は別 PR で `workflow-result` aggregator を更新）。

**Tech Stack:** GitHub Actions, `macos-latest` runner, Homebrew 5.x（runner にプリインストール）

**Spec:** N/A（CI 設定のみ。設計文書は本プランに内包）

---

## File Structure

| Path | 役割 | 新規 / 変更 |
|---|---|---|
| `.github/workflows/brew-audit.yml` | tap formula を継続検証する macOS workflow | 新規 |
| `README.md` | バッジ追加（任意） | 変更（任意） |

---

## Task 1: workflow ファイルの新規作成

**Files:**

- Create: `.github/workflows/brew-audit.yml`

- [ ] **Step 1: ファイル作成**

以下の内容で `.github/workflows/brew-audit.yml` を作成する。

```yaml
name: brew-audit

on:
  push:
    branches: [main]
  pull_request:
  schedule:
    # 毎週月曜 09:00 JST = 00:00 UTC
    - cron: '0 0 * * 1'
  workflow_dispatch:

permissions:
  contents: read

jobs:
  audit:
    runs-on: macos-latest
    timeout-minutes: 20
    steps:
      - name: Show Homebrew version
        run: brew --version

      - name: Tap tqer39/homebrew-tap
        run: brew tap tqer39/tap

      - name: Audit formula (strict + online)
        run: brew audit --strict --online tqer39/tap/ccw

      - name: Install from source
        run: brew install --build-from-source tqer39/tap/ccw

      - name: Run formula test block
        run: brew test tqer39/tap/ccw

      - name: Show installed binary version
        run: ccw -v
```

設計上の判断:

- `--strict` で homebrew-core 規約相当の lint を有効化
- `--online` で URL 到達性 / sha256 検証を含む（実 release artifact が生きていることの担保）
- `--build-from-source` で「bottle が無い場合のソースビルド経路」を確認する。ccw は goreleaser が用意した tarball を DL する formula なので **本リポにとっては artifact DL 経路の確認** に相当（ソースビルドではないが、homebrew-core 移行時を見据えて慣例フラグを残す）
- `brew test` で formula 内 `test do system "#{bin}/ccw", "-v"` を実行
- 最後の `ccw -v` は冗長だが「PATH に乗っている」ことの確認として残す
- `arm64` のみ。Intel macOS の検証は homebrew-core 申請時に追加（`macos-13` runner）。今は不要

- [ ] **Step 2: yamllint で構文確認**

Run: `yamllint .github/workflows/brew-audit.yml`
期待: 警告なし（`.yamllint.yml` のルールに従う）。失敗したらインデント / 行長を調整する。

- [ ] **Step 3: actionlint で workflow 構文確認**

Run: `actionlint .github/workflows/brew-audit.yml`（手元に無ければ `brew install actionlint`）
期待: エラーなし。

- [ ] **Step 4: コミット**

```bash
git add .github/workflows/brew-audit.yml
git commit -m "ci(brew): add macOS audit workflow for tqer39/tap/ccw

- brew tap → audit --strict --online → install --build-from-source → test
- triggers: PR / push main / weekly schedule / workflow_dispatch
- catches formula regressions independent of release pipeline"
```

---

## Task 2: 動作確認（PR ドラフト経由）

**Files:** なし（動作検証のみ）

- [ ] **Step 1: ブランチを push して draft PR を作成**

```bash
git push -u origin worktree-ccw-tqer39-ccw-cli-5c15e3
gh pr create --draft \
  --title "ci(brew): add macOS audit workflow for tqer39/tap/ccw" \
  --body "Adds .github/workflows/brew-audit.yml. See docs/superpowers/plans/2026-04-26-brew-audit-ci.md."
```

- [ ] **Step 2: PR の brew-audit ジョブが green になることを確認**

期待:

- `audit` job が成功
- ログに `brew audit --strict --online tqer39/tap/ccw` の出力（warnings/errors なし）
- `ccw -v` が現行 release バージョン（例 `v0.18.0`）を出力

失敗時の対応:

| 症状 | 原因候補 | 対処 |
|---|---|---|
| `audit` で warnings | formula の `desc` が長い / 末尾ピリオド / 冠詞始まり | tap 側 `.goreleaser.yaml` の `brews.description` を修正して再リリース |
| `install --build-from-source` 失敗 | release tarball の sha256 不一致 / URL 404 | release を再生成、または goreleaser 出力を確認 |
| `brew test` 失敗 | `ccw -v` が non-zero exit | `internal/version` パッケージの実装変更が原因の可能性。`go test ./...` で再現 |

- [ ] **Step 3: schedule トリガーの動作確認は skip**

cron は 1 週間後にしか発火しないので PR では検証しない。`workflow_dispatch` で代替確認:

```bash
gh workflow run brew-audit.yml --ref worktree-ccw-tqer39-ccw-cli-5c15e3
gh run watch
```

期待: 手動実行が green になる。

- [ ] **Step 4: PR を ready for review にして merge**

```bash
gh pr ready
gh pr merge --squash --auto
```

---

## Task 3: README にバッジ追加（任意）

**Files:**

- Modify: `README.md`
- Modify: `docs/README.ja.md`

> このタスクは optional。バッジは見栄え用で機能には無関係。スキップ可。

- [ ] **Step 1: バッジ追加**

`README.md` の既存バッジ行（Go / Release / License / Homebrew）の末尾に追加:

```markdown
[![brew-audit](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml/badge.svg)](https://github.com/tqer39/ccw-cli/actions/workflows/brew-audit.yml)
```

`docs/README.ja.md` の対応する行にも同じバッジを追加。

- [ ] **Step 2: README 整合確認**

readme-sync スキルの規約に従い、英語 README と日本語 README で同じバッジ並びになっていることを目視確認。

- [ ] **Step 3: コミット**

```bash
git add README.md docs/README.ja.md
git commit -m "docs(readme): add brew-audit badge"
```

---

## Out of Scope（本プランでは対応しない）

- **homebrew-core 用ソースビルド形式 Formula のリポ内コミット**: notable 条件（star/fork）が満たされてから別プランで対応
- **Intel macOS (`macos-13`) runner の追加**: arm64 で十分。homebrew-core 申請時に追加
- **`workflow-result` ジョブへの依存追加**: 既存 `ci.yml` の集約ジョブとは別 workflow にしているので、必須化判断は本 PR の運用結果を見てから別 PR で実施
- **bottle のビルド / 配布**: tap では goreleaser の tarball で十分。bottle は homebrew-core マージ後に CI で自動生成される
- **release.yml への audit 組み込み**: 「リリース成功 = audit pass」を強制するなら release.yml の goreleaser ステップ後に audit を足す手もあるが、release pipeline が長くなるため今回は独立 workflow で運用

---

## Success Criteria

- [ ] `.github/workflows/brew-audit.yml` が main にマージされている
- [ ] PR 上で workflow が green になっている
- [ ] `gh workflow run brew-audit.yml` で手動実行できる
- [ ] schedule（毎週月曜 00:00 UTC）の cron が GitHub Actions UI に表示されている
