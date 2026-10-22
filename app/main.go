package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/chzyer/readline"
	"github.com/google/shlex"
)

func main() {
	os.Exit(NewShell(os.Stdin, os.Stdout, os.Stderr).Run())
}

type Shell struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Exited         bool
}

func NewShell(stdin io.Reader, stdout, stderr io.Writer) *Shell {
	return &Shell{Stdin: stdin, Stdout: stdout, Stderr: stderr}
}

func (s *Shell) Run() int {
	var items []readline.PrefixCompleterInterface
	for _, name := range slices.Sorted(maps.Keys(builtins)) {
		items = append(items, readline.PcItem(name))
	}
	completer := readline.NewPrefixCompleter(items...)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		AutoComplete:    completer,
		Stdin:           io.NopCloser(s.Stdin),
		Stdout:          s.Stdout,
		Stderr:          s.Stderr,
		InterruptPrompt: "^C",
	})
	if err != nil {
		fmt.Fprintf(s.Stderr, "readline error: %v\n", err)
		return 1
	}
	defer rl.Close()

	for !s.Exited {
		line, err := rl.Readline()
		if err != nil {
			return 0
		}
		s.runLine(line)
	}
	return 0
}

type env struct {
	sh             *Shell
	stdin          io.Reader
	stdout, stderr io.Writer
}

type builtinFunc func(e *env, args []string)

var builtins map[string]builtinFunc

func init() {
	builtins = map[string]builtinFunc{
		"cd":   cd,
		"echo": echo,
		"exit": exit,
		"pwd":  pwd,
		"type": typeCmd,
	}
}

const (
	truncFlags  = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	appendFlags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
)

type redirect struct {
	toStderr bool
	flags    int
}

var redirects = map[string]redirect{
	">":   {false, truncFlags},
	"1>":  {false, truncFlags},
	"2>":  {true, truncFlags},
	">>":  {false, appendFlags},
	"1>>": {false, appendFlags},
	"2>>": {true, appendFlags},
}

func (s *Shell) runLine(line string) {
	fields, err := shlex.Split(line)
	if err != nil {
		fmt.Fprintf(s.Stderr, "parse error: %v\n", err)
		return
	}
	if len(fields) == 0 {
		return
	}
	name, args := fields[0], fields[1:]

	e := &env{sh: s, stdin: s.Stdin, stdout: s.Stdout, stderr: s.Stderr}

	if len(args) >= 2 {
		op := args[len(args)-2]
		if r, ok := redirects[op]; ok {
			f, err := os.OpenFile(args[len(args)-1], r.flags, 0o644)
			if err != nil {
				fmt.Fprintf(s.Stderr, "cannot open output file: %v\n", err)
				return
			}
			defer f.Close()
			if r.toStderr {
				e.stderr = f
			} else {
				e.stdout = f
			}
			args = args[:len(args)-2]
		}
	}

	if fn, ok := builtins[name]; ok {
		fn(e, args)
		return
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = e.stdout, e.stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintf(e.stderr, "%s: command not found\n", name)
		} else if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			fmt.Fprintf(e.stderr, "%s: %v\n", name, err)
		}
	}
}

func cd(e *env, args []string) {
	dir := "~"
	if len(args) > 0 {
		dir = args[0]
	}
	if dir == "~" {
		dir, _ = os.UserHomeDir()
	}
	if err := os.Chdir(dir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(e.stderr, "cd: %s: No such file or directory\n", dir)
		} else {
			fmt.Fprintf(e.stderr, "cd: %v\n", err)
		}
	}
}

func echo(e *env, args []string) {
	fmt.Fprintln(e.stdout, strings.Join(args, " "))
}

func exit(e *env, _ []string) {
	e.sh.Exited = true
}

func pwd(e *env, _ []string) {
	path, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(e.stderr, "pwd: %v\n", err)
	} else {
		fmt.Fprintln(e.stdout, path)
	}
}

func typeCmd(e *env, args []string) {
	for _, name := range args {
		if _, ok := builtins[name]; ok {
			fmt.Fprintf(e.stdout, "%s is a shell builtin\n", name)
		} else if path, err := exec.LookPath(name); err != nil {
			fmt.Fprintf(e.stdout, "%s: not found\n", name)
		} else {
			fmt.Fprintf(e.stdout, "%s is %s\n", name, path)
		}
	}
}
