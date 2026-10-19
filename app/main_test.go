package main

import (
	"strings"
	"testing"
)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr strings.Builder

			code := NewShell(strings.NewReader(tt.input), &stdout, &stderr).Run()

			if code != 0 {
				t.Errorf("Run() = %d, want 0", code)
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Errorf("Run().stdout = %q, want %q", got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Errorf("Run().stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}
