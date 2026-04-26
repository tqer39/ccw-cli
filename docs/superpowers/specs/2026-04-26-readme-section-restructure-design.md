# README セクション再構成 design

- 日付: 2026-04-26
- 対象: `README.md`（英）と `docs/README.ja.md`（日）
- 動機: Usage セクションが肥大化しており、コマンド例・picker のバッジ表・命名規約が H3 として混在。粒度が揃わず目次からのリンク性も悪い。

## ゴール

`Usage` 配下にあった `Worktree picker` と `Naming convention` を H2 に昇格させ、トップレベル目次で並列に並べる。Usage はコマンド例のみに絞る。英日の構成は完全同期を維持する。

## 非ゴール（今回スコープ外）

- Features と Usage の役割分担の見直し
- SessionStart hook セクション（`🪝 Auto-prompt on new worktree sessions`）の位置・粒度の変更
- バッジ表・絵文字アイコンの中身変更
- 文章の加筆・削除（移動のみ）
- README 以外のドキュメント
- `readme-sync` スキルそのものの改修

## 新しいセクション構成

| # | セクション (英) | セクション (日) | 由来 |
|---|---|---|---|
| 1 | `## ⚡ Quick Start` | `## ⚡ Quick Start` | 維持 |
| 2 | `## ✨ Features` | `## ✨ 特長` | 維持 |
| 3 | `## 🎬 Demo` | `## 🎬 デモ` | 維持 |
| 4 | `## 📖 Usage` | `## 📖 使い方` | コマンド例のみに縮小 |
| 5 | `## 🎯 Picker reference` | `## 🎯 Picker リファレンス` | **新 H2**（旧 `### Worktree picker`） |
| 6 | `## 🏷️ Naming` | `## 🏷️ 命名規約` | **新 H2**（旧 `### Naming convention`） |
| 7 | `## 📦 Installation` | `## 📦 インストール` | 維持 |
| 8 | `## 🪝 Auto-prompt on new worktree sessions` | `## 🪝 新規 worktree セッションでの自動プロンプト注入` | 維持 |
| 9 | `## ⚙️ Environment` | `## ⚙️ 環境変数` | 維持 |
| 10 | `## 🛠️ Development` | `## 🛠️ 開発` | 維持 |
| 11 | `## 🤖 Built With` | `## 🤖 作成ツール` | 維持 |
| 12 | `## 📄 License` | `## 📄 ライセンス` | 維持 |

## 移動の詳細

### Usage に残す内容

- 冒頭のコマンド例コードブロック（`ccw` から `ccw --clean-all --force -y` までの 8 行）
- 末尾の "Run `ccw --help` for the full flag reference." / 「全オプションは `ccw --help` で確認できます。」

### Picker reference に切り出す内容（旧 `### Worktree picker` 全体）

- "Worktree status badge:" / 「Worktree 状態バッジ:」見出し + 表
- "PR state badge..." / 「PR 状態バッジ...」見出し + 表
- "Session badge:" / 「セッションバッジ:」見出し + 表
- 直下の動作説明 2 段落（`Selecting a worktree opens...` と `Without gh, the picker stays...`）

新セクションでは旧 H4 相当（"Worktree status badge:" など）はそのまま太字 inline 見出し（コロン付き）として残す。H3 を増やさない。

### Naming に切り出す内容（旧 `### Naming convention` 全体）

- 導入文 + bullet 3 つ（Directory / Branch / Session name）
- 命名規則の段落（`<name>` の生成ロジック説明）

## 影響範囲

- `README.md`（英）：H3 2 つを H2 に昇格、Usage を縮小
- `docs/README.ja.md`（日）：同じ変更を対訳で適用
- 既存の anchor link は `Usage > Worktree picker` を直接参照していないため、外部リンクへの影響はない（仮にあっても anchor は `#worktree-picker` → `#picker-reference` に変わる程度）
- スクリーンショット・GIF・コード例の出力は変更なし

## 受け入れ基準

1. 英 / 日両方の README で H2 セクションが上記表の順序で並ぶ
2. Usage 直下の本文が「コマンド例コードブロック + `ccw --help` の一文」だけになる
3. 旧 `### Worktree picker` の本文が `## 🎯 Picker reference` 配下にそのまま移動している（文言変更なし）
4. 旧 `### Naming convention` の本文が `## 🏷️ Naming` / `## 🏷️ 命名規約` 配下にそのまま移動している（文言変更なし）
5. `readme-sync` スキルが想定する 1:1 構造（H2 数・順序）が両ファイルで一致する
6. markdownlint / cspell が通る
