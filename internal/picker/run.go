package picker

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

// Run displays the picker against mainRepo and returns the user's decision.
// In non-interactive mode a numbered text fallback is used.
func Run(mainRepo string, interactive bool, in io.Reader, out io.Writer) (Action, Selection, BulkDeletion, error) {
	infos, err := worktree.List(mainRepo)
	if err != nil {
		return ActionCancel, Selection{}, BulkDeletion{}, fmt.Errorf("list worktrees: %w", err)
	}
	if len(infos) == 0 {
		return ActionNew, Selection{}, BulkDeletion{}, nil
	}
	if !interactive {
		a, s, err := runFallback(infos, in, out)
		return a, s, BulkDeletion{}, err
	}
	return runTUI(infos)
}

func runTUI(infos []worktree.Info) (Action, Selection, BulkDeletion, error) {
	p := tea.NewProgram(New(infos), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return ActionCancel, Selection{}, BulkDeletion{}, fmt.Errorf("picker run: %w", err)
	}
	m, ok := final.(Model)
	if !ok {
		return ActionCancel, Selection{}, BulkDeletion{}, fmt.Errorf("picker: unexpected final model type %T", final)
	}
	return m.Action(), m.Selection(), m.Bulk(), nil
}

func runFallback(infos []worktree.Info, in io.Reader, out io.Writer) (Action, Selection, error) {
	_, _ = fmt.Fprintln(out, "Select a worktree to resume:")
	for i, w := range infos {
		_, _ = fmt.Fprintf(out, "  %d) %s  (%s)  %s\n", i+1, w.Branch, w.Status, w.Path)
	}
	_, _ = fmt.Fprintln(out, "  n) new")
	_, _ = fmt.Fprintln(out, "  q) quit")
	_, _ = fmt.Fprint(out, "#? ")

	r := bufio.NewReader(in)
	line, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return ActionCancel, Selection{}, fmt.Errorf("read choice: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	switch answer {
	case "", "q":
		return ActionCancel, Selection{}, nil
	case "n":
		return ActionNew, Selection{}, nil
	}
	n, err := strconv.Atoi(answer)
	if err != nil || n < 1 || n > len(infos) {
		return ActionCancel, Selection{}, fmt.Errorf("invalid choice: %q", answer)
	}
	w := infos[n-1]
	return ActionResume, Selection{Path: w.Path, Branch: w.Branch, Status: w.Status}, nil
}
