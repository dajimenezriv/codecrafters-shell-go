package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type builtin func(args string)

var commands map[string]builtin

func init() {
	commands = map[string]builtin{
		"cd":   cd,
		"echo": echo,
		"pwd":  pwd,
		"type": typeCMD,
	}
}

func main() {
	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")

		input, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		input = strings.TrimSpace(input)
		inputList := strings.Split(input, " ")

		command := inputList[0]
		args := strings.Join(inputList[1:], " ")

		if command == "exit" {
			break
		}

		if fn, ok := commands[command]; ok {
			fn(args)
			continue
		}

		path, ok := searchExecutableCommand(command)
		if !ok {
			fmt.Printf("%s: command not found\n", command)
			continue
		}

		cmd := exec.Cmd{
			Path:   path,
			Args:   inputList,
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
	}
}

func cd(args string) {
	if args == "~" {
		args, _ = os.UserHomeDir()
	}
	err := os.Chdir(args)
	if err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", args)
	}
}

func echo(args string) {
	fmt.Println(args)
}

func pwd(_ string) {
	path, _ := os.Getwd()
	fmt.Println(path)
}

func typeCMD(args string) {
	_, ok := commands[args]
	if args == "exit" || ok {
		fmt.Printf("%s is a shell builtin\n", args)
		return
	}

	path, ok := searchExecutableCommand(args)
	if ok {
		fmt.Printf("%s is %s\n", args, path)
	} else {
		fmt.Printf("%s not found\n", args)
	}
}

func searchExecutableCommand(c string) (string, bool) {
	dirs := filepath.SplitList(os.Getenv("PATH"))

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() && entry.Name() == c {
				path := dir + "/" + entry.Name()
				if _, err := exec.LookPath(path); err == nil {
					return path, true
				}
			}
		}
	}

	return "", false
}
