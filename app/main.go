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
	NewShell(os.Stdin, os.Stdout, os.Stderr).Run()
}

type builtin func(args string)

type Shell struct {
	builtins map[string]builtin
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
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

func (s *Shell) Run() {
	sc := bufio.NewScanner(s.stdin)

	for {
		fmt.Print("$ ")

		if !sc.Scan() {
			break
		}

		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		name, argsList := fields[0], fields[1:]
		args := strings.Join(argsList, " ")

		if fn, ok := s.builtins[name]; ok {
			fn(args)
			continue
		}

		cmd := exec.Command(name, argsList...)
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
}

func (s *Shell) cd(args string) {
	if args == "~" {
		args, _ = os.UserHomeDir()
	}
	if err := os.Chdir(args); err != nil {
		fmt.Fprintf(os.Stderr, "cd: %s: No such file or directory\n", args)
	}
}

func (s *Shell) echo(args string) {
	fmt.Println(args)
}

func (s *Shell) exit(_ string) {
	os.Exit(0)
}

func (s *Shell) pwd(_ string) {
	path, _ := os.Getwd()
	fmt.Println(path)
}

func (s *Shell) typeBuiltin(args string) {
	if _, ok := s.builtins[args]; ok {
		fmt.Printf("%s is a shell builtin\n", args)
	} else if path, err := exec.LookPath(args); err != nil {
		fmt.Printf("%s is %s\n", args, path)
	} else {
		fmt.Printf("%s not found\n", args)
	}
}
