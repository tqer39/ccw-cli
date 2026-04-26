package listmode

import (
	"encoding/json"
	"fmt"
	"io"
)

// RenderJSON writes out as indented JSON followed by a trailing newline.
func RenderJSON(out *Output, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
