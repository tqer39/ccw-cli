package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestInfoGoesToStdout(t *testing.T) {
	var out, err bytes.Buffer
	SetWriter(&out, &err)

	Info("hello %s", "world")

	if got := out.String(); got != "hello world\n" {
		t.Errorf("stdout = %q, want %q", got, "hello world\n")
	}
	if got := err.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestWarnGoesToStderrWithPrefix(t *testing.T) {
	var out, err bytes.Buffer
	SetWriter(&out, &err)

	Warn("careful")

	if got := err.String(); !strings.HasPrefix(got, "⚠ careful") {
		t.Errorf("stderr = %q, want prefix %q", got, "⚠ careful")
	}
	if got := out.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
}

func TestErrorGoesToStderrWithPrefix(t *testing.T) {
	var out, err bytes.Buffer
	SetWriter(&out, &err)

	Error("bad: %d", 42)

	if got := err.String(); !strings.HasPrefix(got, "✖ bad: 42") {
		t.Errorf("stderr = %q, want prefix %q", got, "✖ bad: 42")
	}
}

func TestSuccessGoesToStderrWithPrefix(t *testing.T) {
	var out, err bytes.Buffer
	SetWriter(&out, &err)

	Success("done")

	if got := err.String(); !strings.HasPrefix(got, "✓ done") {
		t.Errorf("stderr = %q, want prefix %q", got, "✓ done")
	}
}

func TestDebugSilentWhenEnvUnset(t *testing.T) {
	var out, err bytes.Buffer
	SetWriter(&out, &err)
	t.Setenv("CCW_DEBUG", "")

	Debug("diag %s", "x")

	if got := err.String(); got != "" {
		t.Errorf("stderr = %q, want empty when CCW_DEBUG unset", got)
	}
	_ = out
}

func TestDebugVisibleWhenEnvOne(t *testing.T) {
	var out, err bytes.Buffer
	SetWriter(&out, &err)
	t.Setenv("CCW_DEBUG", "1")

	Debug("diag %s", "x")

	if got := err.String(); !strings.HasPrefix(got, "[debug] diag x") {
		t.Errorf("stderr = %q, want prefix %q", got, "[debug] diag x")
	}
	_ = out
}

func TestSetWriterDisablesColor(t *testing.T) {
	var out, err bytes.Buffer
	SetWriter(&out, &err)

	Error("plain")

	if got := err.String(); strings.Contains(got, "\x1b[") {
		t.Errorf("stderr = %q, should not contain ANSI escapes", got)
	}
	_ = out
}
