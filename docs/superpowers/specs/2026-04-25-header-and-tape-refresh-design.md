# Header 画像差し替え + tape の RESUME/NEW 反映

- 日付: 2026-04-25
- スコープ: ドキュメント / デモアセットのみ。Go コードや CLI 仕様の変更なし。

## 背景

直近の機能追加（PR #43: worktree-resume integration）で、picker に **RESUME / NEW** セッションバッジが追加された。判定は `~/.claude/projects/<encoded-abs-path>/*.jsonl` の有無に依存する（`internal/worktree/has_session.go`）。

現状の `docs/assets/picker-demo.tape` / `picker-demo-setup.sh` は session log を作らないため、すべて NEW で表示される（= 新機能が GIF で見えない）。同時にユーザーから新しいヘッダー画像が提供された（`~/Downloads/header.png`）。

このスペックは以下 2 点をまとめて扱う:

1. `docs/assets/header.png` を新規ファイルで上書き
2. demo tape を最小変更で RESUME / NEW 混在を演出する

GIF を分割する案（picker / cleanup を別 tape に）も検討したが、**現状路線（tape 1 本）** で進めることに決定（README の構成を変えずに済むのと、tape 尺は据え置きで足りるため）。

## 1. ヘッダー画像差し替え

| 項目 | 旧 | 新 |
|---|---|---|
| パス | `docs/assets/header.png` | 同上（上書き） |
| 寸法 | 1778×592 | 2172×724 |
| アスペクト比 | ≈ 3.0:1 | ≈ 3.0:1 |
| サイズ | 430 KB | 約 1023 KB |

- 単純な `cp ~/Downloads/header.png docs/assets/header.png`。
- README の `![ccw-cli — Claude Code x worktree](docs/assets/header.png)` は width 指定なしのため変更不要。
- `docs/README.ja.md` も同じ相対パスを参照しているため、自動で新画像を拾う。

## 2. demo setup script の更新 (`picker-demo-setup.sh`)

最後に「fake HOME と fake `~/.claude/projects/` を作る」ステップ（5）を追加する。

```bash
# 5. fake HOME so the picker can detect RESUME / NEW deterministically
rm -rf /tmp/ccw-demo-home
PROJECTS=/tmp/ccw-demo-home/.claude/projects
mkdir -p "$PROJECTS"
for wt in feat-login feat-dashboard; do
  enc=$(printf '%s' "/tmp/ccw-demo/.claude/worktrees/$wt" | tr '/.' '--')
  mkdir -p "$PROJECTS/$enc"
  printf '{}\n' >"$PROJECTS/$enc/dummy.jsonl"
done
```

エンコード規則は `internal/worktree/has_session.go` の `EncodeProjectPath`（`/` と `.` を `-` に置換）を `tr` 一発で再現する。

### worktree とバッジの対応

| worktree | worktree badge | PR badge | session badge |
|---|---|---|---|
| `feat/login` | 🟢 PUSHED | 🟩 OPEN | 💬 RESUME |
| `feat/dashboard` | 🟢 PUSHED | 🟪 MERGED | 💬 RESUME |
| `feat/picker` | 🟡 LOCAL | ⬛ DRAFT | ⚡ NEW |
| `chore/cleanup` | 🔴 DIRTY | 🟥 CLOSED | ⚡ NEW |

→ picker に RESUME × 2 / NEW × 2 が並ぶ。`Enter` でサブメニューを開く対象は既存の PUSHED+MERGED 行（feat/dashboard = RESUME）なので、「RESUME → `[r] run` で `claude --continue` 復帰」のストーリーが既存 tape のまま成立する。

スクリプト末尾の `echo "ready..."` 行はそのまま。

## 3. tape の更新 (`picker-demo.tape`)

Hide ブロック内の export を 1 行差し替えるのみ:

```text
Type "cd /tmp/ccw-demo && export HOME=/tmp/ccw-demo-home PATH=/tmp/ccw-demo-bin:/tmp/fake-gh:$PATH && clear"
```

- walkthrough（Down × 3 / Up × 2 / Enter / `b` / Down × 4 / Enter / `N` / `q`）は変更しない。
- 尺・寸法（1440×640、~60s）据え置き。
- font / theme / typing speed 据え置き。

## 4. 検証

- `bash docs/assets/picker-demo-setup.sh` 完走（既存 4 worktree + fake HOME）
- `vhs docs/assets/picker-demo.tape` を実行 → `/tmp/ccw-demo.gif` 生成 → `docs/assets/picker-demo.gif` に上書き（既存ワークフロー通り）
- 目視確認:
  - picker 4 行のうち 2 行に `💬 RESUME`、残り 2 行に `⚡ NEW` が見える
  - badge カラム整列が崩れていない
- README ローカルプレビューでヘッダー画像が新しいものに差し替わっている

## 5. 非スコープ

- `internal/picker/*` のコード変更
- README 文言変更（既に RESUME/NEW について記載済み）
- ja README の文言変更
- 別 GIF への分割

## 6. リスク

- HOME 切り替えがユーザーの実 `~/.claude/` を汚さないか: tape は `Hide` 内で export するのみで `vhs` のサブシェルに閉じる。setup script も `/tmp/ccw-demo-home` のみ操作。実 HOME 配下は触らない。
- `EncodeProjectPath` 仕様変更時に setup の `tr '/.' '--'` がずれる: コメントで関連付けを記す。
