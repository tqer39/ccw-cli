# picker デモ GIF の横幅拡大 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** README 掲載デモ GIF のターミナル横幅を広げ、picker の meta 行が truncate されずに PR セルまで表示されるようにする。

**Architecture:** `docs/assets/picker-demo.tape` の `Set Width` を `1024` → `1440` に変更し（他パラメータは据え置き）、`picker-demo-setup.sh` + `vhs` で `docs/assets/picker-demo.gif` を再生成する。picker 側コード（`internal/picker/*`）は変更しない。

**Tech Stack:** vhs (tape ファイル), bash, Go 1.25（setup スクリプト内で `go build` するため）

**Spec:** `docs/superpowers/specs/2026-04-24-picker-demo-tape-widen-design.md`

---

## File Structure

| File | Role | Action |
|---|---|---|
| `docs/assets/picker-demo.tape` | VHS tape 定義（サイズ設定） | Modify |
| `docs/assets/picker-demo.gif` | README に貼られているデモ | Regenerate |

picker のソース (`internal/picker/delegate.go` 等) と `picker-demo-setup.sh` は変更しない。

---

## Task 1: tape ファイルの Width を 1440 に変更

**Files:**

- Modify: `docs/assets/picker-demo.tape:5`

- [ ] **Step 1: 現在の tape 定数を確認**

Run: `sed -n '1,11p' docs/assets/picker-demo.tape`
Expected:

```text
Output "/tmp/ccw-demo.gif"

Set Shell "bash"
Set FontSize 24
Set Width 1024
Set Height 640
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 110ms
Set PlaybackSpeed 0.92
```

- [ ] **Step 2: `Set Width 1024` を `Set Width 1440` に書き換える**

`docs/assets/picker-demo.tape` の 5 行目を次のように変更:

```text
Set Width 1440
```

他の行（`FontSize 24` / `Height 640` / `Padding 20` / テーマ / 再生速度）は触らない。

- [ ] **Step 3: 差分を確認**

Run: `git diff docs/assets/picker-demo.tape`
Expected:

```diff
-Set Width 1024
+Set Width 1440
```

変更は 1 行のみ。他の差分が出ていたら取り消す。

- [ ] **Step 4: コミット**

```bash
git add docs/assets/picker-demo.tape
git commit -m "docs: widen picker demo tape to 1440 cols"
```

---

## Task 2: デモ GIF を再生成

**Files:**

- Regenerate: `docs/assets/picker-demo.gif`

**背景:** tape の Width を広げただけでは README に貼られている GIF は変わらない。`vhs` で実際にレンダリングしてファイルを差し替える。

- [ ] **Step 1: 前提ツールの存在を確認**

Run: `which vhs && which go && which gh`
Expected: いずれもパスが返る。
失敗時: `brew install vhs` / `brew install go` / `brew install gh` を案内し、ユーザー対応を待つ。

- [ ] **Step 2: デモ環境をセットアップ**

Run: `bash docs/assets/picker-demo-setup.sh`
Expected: 末尾に `ready. now run: vhs docs/assets/picker-demo.tape` が出力され、`/tmp/ccw-demo`, `/tmp/ccw-demo-bin`, `/tmp/fake-gh` が作成される（スクリプトは冪等）。

- [ ] **Step 3: vhs で GIF を生成**

Run: `vhs docs/assets/picker-demo.tape`
Expected: `/tmp/ccw-demo.gif` が生成される。各フレームの進捗が出る。

- [ ] **Step 4: 生成物を docs/assets に配置**

Run: `cp /tmp/ccw-demo.gif docs/assets/picker-demo.gif`
Expected: なし（成功時は無出力）。

- [ ] **Step 5: 目視確認 — truncate が解消しているか**

Run: `open docs/assets/picker-demo.gif`（macOS）または任意のビューアで開く。

確認ポイント:

- meta 行の PR セル `[STATE] #NNN "title"` が末尾 `"` まで表示されている（内側の 30 文字切り詰めによる `…` は許容）
- branch 名（`%-24s`）とインジケータ `↑N ↓N [✎N]` が欠けていない
- bulk confirm 画面などで表示崩れが出ていない

truncate が残る場合: tape の `Set Width` をさらに +128（1568）で再生成し Step 3〜5 を繰り返す。

- [ ] **Step 6: ファイルサイズ・寸法が想定内か確認**

Run: `file docs/assets/picker-demo.gif && ls -lh docs/assets/picker-demo.gif`
Expected: `GIF image data, version 89a, 1440 x 640` と表示される。サイズは元と同程度〜1.5 倍程度（横幅拡大による増加）。

- [ ] **Step 7: コミット**

```bash
git add docs/assets/picker-demo.gif
git commit -m "docs: regenerate picker demo GIF at 1440x640"
```

---

## Task 3: README での見え方を最終確認

**Files:**

- Verify: `README.md`, `docs/README.ja.md`

このタスクは確認のみ。ファイル編集はしない（README 側の相対パス参照は変わらない）。

- [ ] **Step 1: README が GIF を相対パス参照していることを確認**

Run: `grep -n 'picker-demo.gif' README.md docs/README.ja.md`
Expected:

```text
README.md:41:![picker demo](docs/assets/picker-demo.gif)
docs/README.ja.md:41:![picker demo](assets/picker-demo.gif)
```

（リンクテキスト内の別行参照は許容）

- [ ] **Step 2: GitHub プレビューで実寸を見る**

ブランチをプッシュ済みなら GitHub 上の README プレビューで、そうでなければローカルの Markdown プレビューで GIF が正しく再生され、横長になった分が自動で縮小表示されていることを確認する。

確認ポイント:

- README 本文幅内に収まっている（はみ出していない）
- 縮小表示でもバッジ・ブランチ名が読み取れる

問題があれば Task 2 の Step 5 の「truncate が残る場合」と同様に tape を調整して再生成する。

---

## Final Verification

- [ ] `docs/assets/picker-demo.tape` の `Set Width` が `1440` になっている
- [ ] `docs/assets/picker-demo.gif` が 1440x640 で再生成されている
- [ ] 新 GIF の meta 行で PR セルが truncate されずに表示されている
- [ ] README / `docs/README.ja.md` のリンクはそのままで GIF が表示できる
- [ ] picker のソース (`internal/picker/`) に変更が入っていない

---

## Commit Log (想定)

1. `docs: widen picker demo tape to 1440 cols`
2. `docs: regenerate picker demo GIF at 1440x640`
