package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runShell(t *testing.T, input string) (strings.Builder, strings.Builder) {
	var stdout, stderr strings.Builder
	code := NewShell(strings.NewReader(input), &stdout, &stderr).Run()
	if code != 0 {
		t.Errorf("Run() = %d, want 0", code)
	}
	return stdout, stderr
}

func TestBuiltins(t *testing.T) {
	tests := []struct {
		name, input, wantStdout, wantStderr string
	}{
		{"echo", "echo hello   world\n", "$ hello world\n$ ", ""},
		{"empty line", "\n\n", "$ $ $ ", ""},
		{"type builtin", "type echo\n", "$ echo is a shell builtin\n$ ", ""},
		{"type missing", "type nosuchcmd\n", "$ nosuchcmd: not found\n$ ", ""},
		{"unknown command", "nosuchcmd\n", "$ $ ", "nosuchcmd: command not found\n"},
		{"exit stops processing", "exit\necho nope\n", "$ ", ""},
		{"single quotes", "echo 'hello    world'\n", "$ hello    world\n$ ", ""},
		{"double quotes", "echo \"shell's test\"\n", "$ shell's test\n$ ", ""},
		{"backslash", "echo ignore\\_backslash\n", "$ ignore_backslash\n$ ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr := runShell(t, tt.input)

			if got := stdout.String(); got != tt.wantStdout {
				t.Errorf("Run().stdout = %q, want %q", got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Errorf("Run().stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}

func TestRedirectOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output")

	stdout, stderr := runShell(t, "echo 'hello world' > "+path)

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file = %v", err)
	}
	if got, want := string(b), "hello world\n"; got != want {
		t.Fatalf("cmd stdout = %q, want %q", got, want)
	}
	if got, want := stdout.String(), "$ $ "; got != want {
		t.Errorf("Run().stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("Run().stderr = %q, want empty", got)
	}
}
