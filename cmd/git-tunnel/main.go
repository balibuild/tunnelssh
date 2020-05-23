package main

import (
	"fmt"
	"os"
	"os/exec"
)

// GIT_SSH=xxx

// git-tunel
func main() {
	if err := InitializeEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "\x1b[31minitialize env error: %s\x1b[0m\n", err)
		os.Exit(1)
	}
	cmd := exec.Command("git", os.Args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {

	}
	if err := cmd.Wait(); err != nil {
	}
}
