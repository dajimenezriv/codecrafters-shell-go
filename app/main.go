package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"github.com/google/shlex"
)

func main() {
	os.Exit(NewShell(os.Stdin, os.Stdout, os.Stderr).Run())
}

type builtinFunc func(args []string)

type Shell struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Exited         bool
}

func NewShell(stdin io.Reader, stdout, stderr io.Writer) *Shell {
	return &Shell{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
}

func (s *Shell) Run() int {
	sc := bufio.NewScanner(s.Stdin)

	for !s.Exited {
		fmt.Fprint(s.Stdout, "$ ")
		if !sc.Scan() {
			break
		}
		s.prompt(sc.Text())
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintf(s.Stderr, "error reading input: %v\n", err)
		return 1
	}
	return 0
}

type command struct {
	s              *Shell
	builtins       map[string]builtinFunc
	stdout, stderr io.Writer
}

func newCommand(s *Shell, stdout, stderr io.Writer) *command {
	c := &command{
		s:      s,
		stdout: stdout,
		stderr: stderr,
	}
	c.builtins = map[string]builtinFunc{
		"cd":   c.cd,
		"echo": c.echo,
		"exit": c.exit,
		"pwd":  c.pwd,
		"type": c.typeCmd,
	}
	return c
}

type redirect struct {
	fd    int // 1 for stdout, 2 for stderr
	flags int
}

var redirects = map[string]redirect{
	">":   {1, os.O_WRONLY | os.O_CREATE | os.O_TRUNC},
	"1>":  {1, os.O_WRONLY | os.O_CREATE | os.O_TRUNC},
	"2>":  {2, os.O_WRONLY | os.O_CREATE | os.O_TRUNC},
	">>":  {1, os.O_APPEND | os.O_CREATE | os.O_WRONLY},
	"1>>": {1, os.O_APPEND | os.O_CREATE | os.O_WRONLY},
	"2>>": {2, os.O_APPEND | os.O_CREATE | os.O_WRONLY},
}

func (s *Shell) prompt(line string) {
	fields, err := shlex.Split(line)
	if err != nil {
		fmt.Fprintf(s.Stderr, "parse error: %v", err)
		return
	}
	if len(fields) == 0 {
		return
	}
	name, args := fields[0], fields[1:]

	stdout, stderr := s.Stdout, s.Stderr

	if len(args) >= 2 {
		char := args[len(args)-2]
		if r, ok := redirects[char]; ok {
			f, err := os.OpenFile(args[len(args)-1], r.flags, 0o644)
			if err != nil {
				fmt.Fprintf(s.Stderr, "cannot open output file: %v", err)
				return
			}
			defer f.Close()
			switch r.fd {
			case 1:
				stdout = f
			case 2:
				stderr = f
			}
			args = args[:len(args)-2]
		}
	}

	c := newCommand(s, stdout, stderr)
	if fn, ok := c.builtins[name]; ok {
		fn(args)
		return
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintf(stderr, "%s: command not found\n", name)
		} else if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			fmt.Fprintf(stderr, "%s: %v\n", name, err)
		}
	}
}

func (c *command) cd(args []string) {
	dir := "~"
	if len(args) > 0 {
		dir = args[0]
	}
	if dir == "~" {
		dir, _ = os.UserHomeDir()
	}
	if err := os.Chdir(dir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(c.stderr, "cd: %s: No such file or directory\n", dir)
		} else {
			fmt.Fprintf(c.stderr, "cd: %v", err)
		}
	}
}

func (c *command) echo(args []string) {
	fmt.Fprintln(c.stdout, strings.Join(args, " "))
}

func (c *command) exit(_ []string) {
	c.s.Exited = true
}

func (c *command) pwd(_ []string) {
	path, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(c.stderr, "pwd: %v\n", err)
	} else {
		fmt.Fprintf(c.stdout, "%s\n", path)
	}
}

func (c *command) typeCmd(args []string) {
	for _, name := range args {
		if _, ok := c.builtins[name]; ok {
			fmt.Fprintf(c.stdout, "%s is a shell builtin\n", name)
		} else if path, err := exec.LookPath(name); err != nil {
			fmt.Fprintf(c.stdout, "%s: not found\n", name)
		} else {
			fmt.Fprintf(c.stdout, "%s is %s\n", name, path)
		}
	}
}
