# ccw

Claude Code worktree launcher — isolates each session in its own git worktree with opinionated defaults.

## Features

- Launches Claude Code (`claude`) in a fresh git worktree each time (no state leakage)
- Auto permission mode by default (no approval prompts)
- Optional superpowers workflow injection (`brainstorming → writing-plans → executing-plans`) via `-s`
- Interactive picker for leftover worktrees: resume, delete, or start new
- Built-in self-update (`--update`) and uninstall (`--uninstall`)
- Version display (`-v`)
- Pass-through of native `claude` arguments after `--`

## Requirements

- `git`
- `bash` 3.2+ (macOS の `/bin/bash` でも動作)
- `tput` (ncurses) — ほぼ常に標準装備
- Claude Code CLI (`claude`) — 起動時に不在なら npm / brew から install するかを対話で選択
- (Optional) superpowers plugin — `-s` 使用時に自動チェック

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
  -s, --superpowers    superpowers プリアンブルを注入して起動
      --update         ccw を最新版に更新
      --uninstall      ccw 自身をアンインストール（symlink のみ）
  -v, --version        バージョン情報を表示
  -h, --help           このヘルプを表示

Arguments after `--` are forwarded to `claude` verbatim.
```

### Examples

```bash
ccw                            # 既存 worktree があれば選択、なければ新規起動
ccw -n                         # 新規 worktree で起動 (picker スキップ)
ccw -s                         # 新規 + superpowers プリアンブル
ccw -- --model claude-opus-4-7 # claude へ引数パススルー
ccw --update                   # 最新版へ更新
ccw --uninstall                # アンインストール
```

## Worktree picker

`ccw` をオプションなしで起動すると、`.claude/worktrees/` 配下に残っている worktree を検出して選択 UI を表示します。

```text
> 🟢 kahan        shimmying-frolicking-kahan  (pushed, clean)
  🟡 pirate       playful-swashbuckling-ai    (local-only)
  🔴 nebula       twinkling-starry-nebula     (dirty)
  ➕ [new]        Start new worktree
  🚪 [quit]       Cancel
```

| アイコン | 状態 |
|---|---|
| 🟢 | `pushed`: upstream 追従、ahead 0、clean |
| 🟡 | `local-only`: upstream なし or ahead あり |
| 🔴 | `dirty`: working tree に未コミット変更あり |

選択後、`resume` / `delete` / `back` を選ぶサブメニューに遷移します。

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

## Update

```bash
ccw --update
```

## Uninstall

```bash
ccw --uninstall
```

symlink のみ削除されます。clone したリポジトリは `rm -rf <repo>` で手動削除してください。

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
- `gofmt` / `golangci-lint`: Go ファイル対象（Phase 1 以降で有効化）
- `shellcheck` / `shfmt`: bash スクリプト対象（`bin/ccw` / `tests/*.bats`）
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
