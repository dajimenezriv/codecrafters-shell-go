package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/google/shlex"
)

func main() {
	os.Exit(NewShell(os.Stdin, os.Stdout, os.Stderr).Run())
}

type builtin func(args string)

type Shell struct {
	Builtins       map[string]builtin
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Exited         bool

	cmdStdout, cmdStderr io.Writer
}

func NewShell(stdin io.Reader, stdout, stderr io.Writer) *Shell {
	s := &Shell{Stdin: stdin, Stdout: stdout, Stderr: stderr}
	s.Builtins = map[string]builtin{
		"cd":   s.cd,
		"echo": s.echo,
		"exit": s.exit,
		"pwd":  s.pwd,
		"type": s.typeBuiltin,
	}
	return s
}

func (s *Shell) Run() int {
	sc := bufio.NewScanner(s.Stdin)

	for !s.Exited {
		s.prompt(sc)
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintf(s.Stderr, "error reading input: %v\n", err)
	}

	return 0
}

func (s *Shell) prompt(sc *bufio.Scanner) {
	fmt.Fprint(s.Stdout, "$ ")
	if !sc.Scan() {
		s.Exited = true
		return
	}

	fields, err := shlex.Split(sc.Text())
	if err != nil {
		fmt.Fprintf(s.Stderr, "error parsing args: %v\n", err)
		return
	}
	if len(fields) == 0 {
		return
	}
	name, args := fields[0], fields[1:]

	s.cmdStdout, s.cmdStderr = s.Stdout, s.Stderr

	if len(args) >= 2 {
		char := args[len(args)-2]
		var f *os.File
		switch char {
		case ">", "1>":
			f = s.getOutputFile(args[len(args)-1], os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
			s.cmdStdout = f
		case "2>":
			f = s.getOutputFile(args[len(args)-1], os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
			s.cmdStderr = f
		case ">>", "1>>":
			f = s.getOutputFile(args[len(args)-1], os.O_APPEND|os.O_CREATE|os.O_WRONLY)
			s.cmdStdout = f
		case "2>>":
			f = s.getOutputFile(args[len(args)-1], os.O_APPEND|os.O_CREATE|os.O_WRONLY)
			s.cmdStderr = f
		}

		if f != nil {
			defer f.Close()
			args = args[:len(args)-2]
		}
	}

	if fn, ok := s.Builtins[name]; ok {
		fn(strings.Join(args, " "))
		return
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = s.cmdStdout, s.cmdStderr
	if err := cmd.Run(); err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); ok {
		} else if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintf(s.cmdStderr, "%s: command not found\n", name)
		} else {
			fmt.Fprintf(s.cmdStderr, "%s: %v\n", name, err)
		}
	}
}

func (s *Shell) getOutputFile(path string, flags int) *os.File {
	f, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		fmt.Fprintf(s.Stderr, "cannot get output file: %v\n", err)
		return nil
	}
	return f
}

func (s *Shell) cd(args string) {
	if args == "~" {
		args, _ = os.UserHomeDir()
	}
	if err := os.Chdir(args); err != nil {
		fmt.Fprintf(s.cmdStderr, "cd: %s: No such file or directory\n", args)
	}
}

func (s *Shell) echo(args string) { fmt.Fprintln(s.cmdStdout, args) }

func (s *Shell) exit(_ string) { s.Exited = true }

func (s *Shell) pwd(_ string) {
	path, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(s.cmdStderr, "pwd: %v\n", err)
	} else {
		fmt.Fprintln(s.cmdStdout, path)
	}
}

func (s *Shell) typeBuiltin(args string) {
	if _, ok := s.Builtins[args]; ok {
		fmt.Fprintf(s.cmdStdout, "%s is a shell builtin\n", args)
	} else if path, err := exec.LookPath(args); err != nil {
		fmt.Fprintf(s.cmdStdout, "%s: not found\n", args)
	} else {
		fmt.Fprintf(s.cmdStdout, "%s is %s\n", args, path)
	}
}
