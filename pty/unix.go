// +build !windows

package pty

import (
	"errors"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// GetWinSize todo
func GetWinSize() (w int, h int, err error) {
	wsz, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return -1, -1, err
	}
	return int(wsz.Col), int(wsz.Row), nil
}

// IsTerminal todo
func IsTerminal(fd *os.File) bool {
	return isatty.IsTerminal(fd.Fd())
}

// ReadPassword todo
func ReadPassword(prompt string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", prompt)
	pwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	return string(pwd), nil
}

// MakeRaw todo
func MakeRaw(fd *os.File) (*term.State, error) {
	if IsTerminal(fd) {
		return term.MakeRaw(int(fd.Fd()))
	}
	return nil, errors.New("not terminal")
}

// ReadInputEx todo
func ReadInputEx(fd *os.File) ([]byte, error) {
	return ReadInput(fd, true)
}
