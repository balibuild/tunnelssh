// +build !windows

package pty

import (
	"os"

	"golang.org/x/sys/unix"
)

// GetWinSize2 todo
func GetWinSize2() (w int, h int, err error) {
	wsz, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return -1, -1, err
	}
	return int(wsz.Col), int(wsz.Row), nil
}
