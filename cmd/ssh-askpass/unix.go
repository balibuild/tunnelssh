// +build !windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// connect tty

const (
	ttypath string = "/dev/tty"
)

func AskYes(caption, title string) int {
	ttyin, err := os.OpenFile(ttypath, os.O_RDONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable open tty: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "%s: ", caption)
	defer ttyin.Close()
	b := bufio.NewReader(ttyin)
	for {
		answer, err := b.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read answer: %s", err)
			return 1
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "yes" {
			fmt.Fprintf(os.Stdout, "Yes\n")
			return 0
		}
		if answer == "no" {
			fmt.Fprintf(os.Stdout, "No\n")
			return 0
		}
		fmt.Fprintf(os.Stderr, "Please type 'yes' or 'no': ")
	}
	return 0
}

// AskPassword todo
func AskPassword(caption, user string) int {
	ttyin, err := os.OpenFile(ttypath, os.O_RDONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable open tty: %v\n", err)
		return 1
	}
	defer ttyin.Close()
	state, err := terminal.GetState(int(ttyin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get terminal state: %s", err)
		return 1
	}

	stopC := make(chan struct{})
	defer func() {
		close(stopC)
	}()

	go func() {
		sigC := make(chan os.Signal, 1)
		signal.Notify(sigC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case <-sigC:
			terminal.Restore(int(ttyin.Fd()), state)
			os.Exit(1)
		case <-stopC:
		}
	}()
	if len(user) != 0 {
		fmt.Fprintf(os.Stderr, "Password for %s: ", user)
	} else {
		fmt.Fprintf(os.Stderr, "Password: ")
	}
	b, err := terminal.ReadPassword(int(ttyin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read password: %s", err)
		return 1
	}

	fmt.Fprintf(os.Stdout, "%s\n", string(b))
	return 0
}
