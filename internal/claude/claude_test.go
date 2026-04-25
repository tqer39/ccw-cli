package claude

import (
	"reflect"
	"testing"
)

func TestBuildNewArgs_NameOnly(t *testing.T) {
	got := BuildNewArgs("foo", "", nil)
	want := []string{"--permission-mode", "auto", "--worktree", "foo", "-n", "foo"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildNewArgs:\n got  = %v\n want = %v", got, want)
	}
}

func TestBuildNewArgs_WithExtra(t *testing.T) {
	got := BuildNewArgs("foo", "", []string{"--model", "claude-opus-4-7"})
	want := []string{"--permission-mode", "auto", "--worktree", "foo", "-n", "foo", "--model", "claude-opus-4-7"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildNewArgs extra:\n got  = %v\n want = %v", got, want)
	}
}

func TestBuildNewArgs_WithPreamble(t *testing.T) {
	got := BuildNewArgs("foo", "hello", nil)
	want := []string{"--permission-mode", "auto", "--worktree", "foo", "-n", "foo", "--", "hello"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildNewArgs preamble:\n got  = %v\n want = %v", got, want)
	}
}

func TestBuildNewArgs_WithExtraAndPreamble(t *testing.T) {
	got := BuildNewArgs("foo", "hi", []string{"--model", "x"})
	want := []string{"--permission-mode", "auto", "--worktree", "foo", "-n", "foo", "--model", "x", "--", "hi"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildNewArgs both:\n got  = %v\n want = %v", got, want)
	}
}

func TestBuildContinueArgs_Empty(t *testing.T) {
	got := BuildContinueArgs(nil)
	want := []string{"--permission-mode", "auto", "--continue"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildContinueArgs:\n got  = %v\n want = %v", got, want)
	}
}

func TestBuildContinueArgs_WithExtra(t *testing.T) {
	got := BuildContinueArgs([]string{"--model", "x"})
	want := []string{"--permission-mode", "auto", "--continue", "--model", "x"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildContinueArgs extra:\n got  = %v\n want = %v", got, want)
	}
}
