package cli

import (
	"fmt"
	"io"
)

const usage = `Usage: ccw [options] [-- <claude-args>...]

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

Arguments after ` + "`--`" + ` are forwarded to ` + "`claude`" + ` verbatim.

Environment:
  NO_COLOR=1           カラー出力を無効化
  CCW_DEBUG=1          詳細ログ出力

Exit codes:
  0  success
  1  user error / cancellation
  *  passthrough from ` + "`claude`" + `
`

// PrintHelp writes the usage string to w.
func PrintHelp(w io.Writer) {
	_, _ = fmt.Fprint(w, usage)
}
