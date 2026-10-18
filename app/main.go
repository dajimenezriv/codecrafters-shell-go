package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")

		command, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		command = strings.TrimSpace(command)

		if command == "exit" {
			break
		}

		if strings.HasPrefix(command, "echo ") {
			fmt.Println(command[5:])
			continue
		}

		if strings.HasPrefix(command, "type ") {
			v := command[5:]
			if v == "exit" || v == "echo" || v == "type" {
				fmt.Printf("%s is a shell builtin\n", v)
			} else {
				path, ok := searchExecutableCommand(v)
				if ok {
					fmt.Printf("%s is %s\n", v, path)
				} else {
					fmt.Printf("%s not found\n", v)
				}
			}
		} else {
			fmt.Printf("%s: command not found\n", command)
		}
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
