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

func TestPromptYN_YesVariants(t *testing.T) {
	for _, in := range []string{"y\n", "Y\n", "yes\n", "YES\n"} {
		t.Run("input="+in, func(t *testing.T) {
			var out bytes.Buffer
			ok, err := PromptYN(strings.NewReader(in), &out, "Go?")
			if err != nil {
				t.Fatalf("PromptYN: %v", err)
			}
			if !ok {
				t.Errorf("PromptYN(%q) = false, want true", in)
			}
			if !strings.Contains(out.String(), "Go?") {
				t.Errorf("prompt not written to out: %q", out.String())
			}
		})
	}
}

func TestPromptYN_NoAndEmpty(t *testing.T) {
	for _, in := range []string{"", "n\n", "\n", "maybe\n"} {
		t.Run("input="+in, func(t *testing.T) {
			var out bytes.Buffer
			ok, err := PromptYN(strings.NewReader(in), &out, "Go?")
			if err != nil {
				t.Fatalf("PromptYN: %v", err)
			}
			if ok {
				t.Errorf("PromptYN(%q) = true, want false", in)
			}
		})
	}
}

func TestPromptChoice_Valid(t *testing.T) {
	var out bytes.Buffer
	got, err := PromptChoice(strings.NewReader("2\n"), &out, "Pick:", []rune{'1', '2', 'q'})
	if err != nil {
		t.Fatalf("PromptChoice: %v", err)
	}
	if got != '2' {
		t.Errorf("PromptChoice = %q, want '2'", got)
	}
}

func TestPromptChoice_Invalid(t *testing.T) {
	var out bytes.Buffer
	_, err := PromptChoice(strings.NewReader("x\n"), &out, "Pick:", []rune{'1', '2', 'q'})
	if err == nil {
		t.Fatal("PromptChoice invalid: want error, got nil")
	}
}

func TestPromptChoice_EmptyInput(t *testing.T) {
	var out bytes.Buffer
	_, err := PromptChoice(strings.NewReader(""), &out, "Pick:", []rune{'1'})
	if err == nil {
		t.Fatal("PromptChoice empty: want error, got nil")
	}
}
