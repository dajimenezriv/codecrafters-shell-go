package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	os.Exit(NewShell(os.Stdin, os.Stdout, os.Stderr).Run())
}

type builtin func(args string)

type Shell struct {
	builtins map[string]builtin
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	exited   bool
}

func NewShell(stdin io.Reader, stdout, stderr io.Writer) *Shell {
	s := &Shell{stdin: stdin, stdout: stdout, stderr: stderr}
	s.builtins = map[string]builtin{
		"cd":   s.cd,
		"echo": s.echo,
		"exit": s.exit,
		"pwd":  s.pwd,
		"type": s.typeBuiltin,
	}
	return s
}

func (s *Shell) Run() int {
	sc := bufio.NewScanner(s.stdin)

	for !s.exited {
		fmt.Fprint(s.stdout, "$ ")
		if !sc.Scan() {
			break
		}

		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		name, args := fields[0], fields[1:]

		if fn, ok := s.builtins[name]; ok {
			fn(strings.Join(args, " "))
			continue
		}

		cmd := exec.Command(name, args...)
		cmd.Stdout, cmd.Stderr = s.stdout, s.stderr
		if err := cmd.Run(); err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				fmt.Fprintf(s.stderr, "%s: command not found\n", name)
			}
		}
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintf(s.stderr, "error reading input: %v\n", err)
	}
	return 0
}

func (s *Shell) cd(args string) {
	if args == "~" {
		args, _ = os.UserHomeDir()
	}
	if err := os.Chdir(args); err != nil {
		fmt.Fprintf(os.Stderr, "cd: %s: No such file or directory\n", args)
	}
}

func (s *Shell) echo(args string) { fmt.Fprintln(s.stdout, args) }

func (s *Shell) exit(_ string) { s.exited = true }

func (s *Shell) pwd(_ string) {
	path, _ := os.Getwd()
	fmt.Fprintln(s.stdout, path)
}

func (s *Shell) typeBuiltin(args string) {
	if _, ok := s.builtins[args]; ok {
		fmt.Fprintf(s.stdout, "%s is a shell builtin\n", args)
	} else if path, err := exec.LookPath(args); err != nil {
		fmt.Fprintf(s.stdout, "%s: not found\n", args)
	} else {
		fmt.Fprintf(s.stdout, "%s is %s\n", args, path)
	}
}
