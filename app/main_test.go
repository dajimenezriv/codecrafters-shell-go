package main

import (
	"fmt"
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
		{"echo", "echo hello   world\n", "hello world\n", ""},
		{"empty line", "\n\n", "", ""},
		{"type builtin", "type echo\n", "echo is a shell builtin\n", ""},
		{"type missing", "type nosuchcmd\n", "nosuchcmd: not found\n", ""},
		{"unknown command", "nosuchcmd\n", "", "nosuchcmd: command not found\n"},
		{"exit stops processing", "exit\necho nope\n", "", ""},
		{"single quotes", "echo 'hello    world'\n", "hello    world\n", ""},
		{"double quotes", "echo \"shell's test\"\n", "shell's test\n", ""},
		{"backslash", "echo ignore\\_backslash\n", "ignore_backslash\n", ""},
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
	const errMsg = "cd: nonexisting: No such file or directory\n"

	success := func(op string) func(path string) string {
		return func(path string) string {
			return fmt.Sprintf("echo 'hello world' %s %s", op, path)
		}
	}
	invalid := func(op string) func(path string) string {
		return func(path string) string {
			return fmt.Sprintf("cd nonexisting %s %s", op, path)
		}
	}
	twice := func(op string) func(path string) string {
		return func(path string) string {
			line := fmt.Sprintf("echo 'hello world' %s %s\n", op, path)
			return line + line
		}
	}

	tests := []struct {
		name       string
		input      func(path string) string
		wantFile   string
		wantStdout string
		wantStderr string
	}{
		{"success stdout >", success(">"), "hello world\n", "", ""},
		{"success stdout 1>", success("1>"), "hello world\n", "", ""},
		{"success stderr 2>", success("2>"), "", "hello world\n", ""},
		{"error stdout >", invalid(">"), "", "", errMsg},
		{"error stdout 1>", invalid("1>"), "", "", errMsg},
		{"error stderr 2>", invalid("2>"), errMsg, "", ""},
		{"append stdout >>", twice(">>"), "hello world\nhello world\n", "", ""},
		{"append stdout 1>>", twice("1>>"), "hello world\nhello world\n", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "output")

			stdout, stderr := runShell(t, tt.input(path))

			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read file: %v", err)
			}
			if got := string(b); got != tt.wantFile {
				t.Errorf("file contents = %q, want %q", got, tt.wantFile)
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Errorf("stdout = %q, want %q", got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Errorf("stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}
