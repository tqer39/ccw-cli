package claude

import (
	"reflect"
	"testing"
)

func TestBuildNewArgs_NoExtraNoPreamble(t *testing.T) {
	got := BuildNewArgs("", nil)
	want := []string{"--permission-mode", "auto", "--worktree"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildNewArgs:\n got  = %v\n want = %v", got, want)
	}
}

func TestBuildNewArgs_WithExtra(t *testing.T) {
	got := BuildNewArgs("", []string{"--model", "claude-opus-4-7"})
	want := []string{"--permission-mode", "auto", "--worktree", "--model", "claude-opus-4-7"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildNewArgs extra:\n got  = %v\n want = %v", got, want)
	}
}

func TestBuildNewArgs_WithPreamble(t *testing.T) {
	got := BuildNewArgs("hello", nil)
	want := []string{"--permission-mode", "auto", "--worktree", "--", "hello"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildNewArgs preamble:\n got  = %v\n want = %v", got, want)
	}
}

func TestBuildNewArgs_WithExtraAndPreamble(t *testing.T) {
	got := BuildNewArgs("hi", []string{"--resume"})
	want := []string{"--permission-mode", "auto", "--worktree", "--resume", "--", "hi"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildNewArgs both:\n got  = %v\n want = %v", got, want)
	}
}

func TestBuildResumeArgs_Empty(t *testing.T) {
	got := BuildResumeArgs(nil)
	want := []string{"--permission-mode", "auto"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildResumeArgs:\n got  = %v\n want = %v", got, want)
	}
}

func TestBuildResumeArgs_WithExtra(t *testing.T) {
	got := BuildResumeArgs([]string{"--resume"})
	want := []string{"--permission-mode", "auto", "--resume"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildResumeArgs extra:\n got  = %v\n want = %v", got, want)
	}
}
