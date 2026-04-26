package listmode

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderJSON_RoundTrip(t *testing.T) {
	out := &Output{
		Version:   1,
		Repo:      RepoInfo{Owner: "o", Name: "r", DefaultBranch: "main", MainPath: "/p"},
		Worktrees: []WorktreeEntry{{Name: "x", Path: "/p/.claude/worktrees/x", Status: "pushed"}},
	}
	var buf bytes.Buffer
	if err := RenderJSON(out, &buf); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	var got Output
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Version != 1 || got.Worktrees[0].Name != "x" {
		t.Errorf("round trip failed: %+v", got)
	}
}

func TestRenderJSON_Indented(t *testing.T) {
	out := &Output{Version: 1, Worktrees: []WorktreeEntry{}}
	var buf bytes.Buffer
	if err := RenderJSON(out, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\n") {
		t.Errorf("output not indented: %s", buf.String())
	}
}

func TestRenderJSON_TrailingNewline(t *testing.T) {
	var buf bytes.Buffer
	_ = RenderJSON(&Output{Version: 1, Worktrees: []WorktreeEntry{}}, &buf)
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Errorf("missing trailing newline")
	}
}
