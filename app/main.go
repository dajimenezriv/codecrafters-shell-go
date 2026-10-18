package main

import (
	"bufio"
	"fmt"
	"os"
)

var _ = fmt.Print

func main() {
	fmt.Print("$ ")

	r := bufio.NewReader(os.Stdin)

	command, err := r.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		os.Exit(1)
	}
	command = command[:len(command)-1]
	fmt.Printf("%s: command not found\n", command)
}
