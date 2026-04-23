# ccw

Claude Code worktree launcher — isolates each session in its own git worktree with opinionated defaults.

## Features

- Launches Claude Code (`claude`) in a fresh git worktree each time (no state leakage)
- Auto permission mode by default (no approval prompts)
- Optional superpowers workflow injection (`brainstorming → writing-plans → executing-plans`) via `-s`
- Interactive picker for leftover worktrees: resume, delete, bulk delete, or start new
- カラーバッジ + `↑N ↓M ✎N` インジケータ + PR 番号/タイトル (`gh` 導入時のみ)
- 一括削除: picker メニュー or `--clean-all` CLI フラグ
- Version display (`-v`)
- Pass-through of native `claude` arguments after `--`

## Requirements

- `git`
- Claude Code CLI (`claude`) — 起動時に不在なら npm / brew から install するかを対話で選択
- (Optional) superpowers plugin — `-s` 使用時に自動チェック
- (Optional) `gh` CLI — picker に PR 番号/タイトルを表示（未導入でも picker は動作）

## Install

```bash
git clone https://github.com/tqer39/ccw-cli ~/workspace/tqer39/ccw-cli
mkdir -p ~/.local/bin
ln -s ~/workspace/tqer39/ccw-cli/bin/ccw ~/.local/bin/ccw

# ~/.local/bin が PATH にあるか確認。なければ .zshrc / .bashrc に追加:
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
```

## Usage

```text
Usage: ccw [options] [-- <claude-args>...]

Options:
  -n, --new            常に新規 worktree で起動（既存 worktree の選択をスキップ）
  -s, --superpowers    superpowers プリアンブルを注入して起動（暗黙に -n）
  -v, --version        バージョン情報を表示
  -h, --help           このヘルプを表示

Bulk delete:
      --clean-all        一括削除モード
      --status=<filter>  all | pushed | local-only | dirty (default: all)
      --force            dirty を --force で削除
      --dry-run          対象だけ表示して終了
  -y, --yes              確認プロンプトをスキップ

Arguments after `--` are forwarded to `claude` verbatim.
```

### Examples

```bash
ccw                                       # 既存 worktree があれば選択、なければ新規起動
ccw -n                                    # 新規 worktree で起動 (picker スキップ)
ccw -s                                    # 新規 + superpowers プリアンブル
ccw -- --model claude-opus-4-7            # claude へ引数パススルー
ccw --clean-all --status=pushed --dry-run # 削除対象のプレビュー
ccw --clean-all --status=all --force -y   # dirty 含む全 worktree を確認なしで削除
```

## Worktree picker

`ccw` をオプションなしで起動すると、`.claude/worktrees/` 配下に残っている worktree を検出して選択 UI を表示します。

```text
> [PUSHED] feat/login              ↑0 ↓0       #42 open "feat: add login"
    ~/repo/.claude/worktrees/feat-login
  [LOCAL]  feat/picker              ↑3 ↓1       (no PR)
    ~/repo/.claude/worktrees/feat-picker
  [DIRTY]  chore/cleanup            ↑0 ↓2 ✎5   #43 draft "chore: cleanup"
    ~/repo/.claude/worktrees/chore-cleanup
  🗑️  [delete all]
  🧹  [clean pushed]
  ☑️  [custom select]
  ➕  [new]
  🚪  [quit]
```

| バッジ | 状態 |
|---|---|
| `[PUSHED]` (緑) | upstream 追従、ahead 0、clean |
| `[LOCAL]` (黄) | upstream なし or ahead あり |
| `[DIRTY]` (赤) | working tree に未コミット変更あり |

インジケータ:

- `↑N ↓M` — upstream からの ahead / behind コミット数
- `✎N` — dirty ファイル数（`dirty` のみ）
- `#N state "title"` — PR 番号 / 状態 / タイトル（`gh` 導入時のみ）

worktree を選択すると `resume` / `delete` / `back` のサブメニューに遷移。`[delete all]` / `[clean pushed]` / `[custom select]` は複数 worktree をまとめて削除するショートカット。dirty を含む場合は確認ダイアログで `y` (force) / `s` (skip dirty) / `N` を選べます。

`gh` が未導入の場合は PR 列が消え、footer に `💡 gh があったら PR 名も出せます` のヒントが出ます。導入済みで gh の呼び出しが失敗した場合（rate limit / ネットワークエラー 等）は PR 列のみ非表示にして静かにフォールバックします。

## Environment variables

| 変数 | 効果 |
|---|---|
| `NO_COLOR=1` | カラー出力を無効化 |
| `CCW_DEBUG=1` | 詳細ログ出力 (`set -x` 相当) |

## Exit codes

| コード | 意味 |
|---|---|
| `0` | 成功 |
| `1` | ユーザーエラー / キャンセル（依存欠落、git repo 外、ユーザー拒否 等） |
| その他 | `claude` コマンドの終了コードをそのまま透過 |

## Development

### Prerequisites

ローカルで lefthook pre-commit フックを有効化するため、以下を事前に用意してください。

```bash
brew install lefthook shellcheck shfmt yamllint actionlint
lefthook install
```

`markdownlint-cli2` / `renovate-config-validator` は `npm exec` で都度取得されるため事前インストール不要（Node.js / npm は必要）。

### Hooks

- `check-added-large-files`: 512KB 超のファイルをブロック
- `detect-private-key`: 秘密鍵の混入を検出
- `gofmt` / `golangci-lint`: Go ファイル対象
- `yamllint` / `actionlint`: YAML / GitHub Actions ワークフロー対象
- `markdownlint-cli2`: Markdown 対象
- `renovate-config-validator`: `renovate.json5` のみ対象

pre-commit は parallel 実行されるため、各 hook の所要時間は合算ではなく最長値。

## Future work

- シェル補完 (bash / zsh)
- Windows サポート
- Homebrew 配布（必要になれば）

## License

MIT
