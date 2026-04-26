package listmode

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// RenderText writes out as a column-aligned ASCII table with no ANSI escapes.
// Columns: NAME STATUS AHEAD/BEHIND PR SESSION BRANCH.
func RenderText(out *Output, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tSTATUS\tAHEAD/BEHIND\tPR\tSESSION\tBRANCH"); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for _, e := range out.Worktrees {
		ahead := "-"
		if e.Status != "prunable" {
			ahead = fmt.Sprintf("%d/%d", e.Ahead, e.Behind)
		}
		pr := "-"
		if e.PR != nil {
			pr = fmt.Sprintf("#%d %s", e.PR.Number, e.PR.State)
		}
		session := "NEW"
		if e.Session.Exists {
			session = "RESUME"
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			e.Name, e.Status, ahead, pr, session, e.Branch); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("flush tabwriter: %w", err)
	}
	return nil
}
