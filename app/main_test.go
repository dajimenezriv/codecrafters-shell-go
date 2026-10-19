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

	tests := []struct {
		name, input, wantOutput, wantStdout, wantStderr string
	}{
		{"success stdout >", "echo 'hello world' > " + path, "hello world\n", "$ $ ", ""},
		{"success stdout 1>", "echo 'hello world' 1> " + path, "hello world\n", "$ $ ", ""},
		{"success stderr 2>", "echo 'hello world' 2> " + path, "", "$ hello world\n$ ", ""},
		{"error stdout >", "cd nonexisting > " + path, "", "$ $ ", "cd: nonexisting: No such file or directory\n"},
		{"error stdout 1>", "cd nonexisting > " + path, "", "$ $ ", "cd: nonexisting: No such file or directory\n"},
		{"error stderr 2>", "cd nonexisting 2> " + path, "cd: nonexisting: No such file or directory\n", "$ $ ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			stdout, stderr := runShell(t, tt.input)

			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read file = %v", err)
			}
			if got := string(b); got != tt.wantOutput {
				t.Fatalf("cmd stdout = %q, want %q", got, tt.wantOutput)
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
