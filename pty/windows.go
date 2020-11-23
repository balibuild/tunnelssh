// +build windows

package pty

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

// GetWinSize get console size or cygwin terminal size
func GetWinSize() (w int, h int, err error) {
	if isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		cmd := exec.Command("stty", "size")
		cmd.Stdin = os.Stdin
		buf, err := cmd.CombinedOutput()
		if err != nil {
			return 0, 0, err
		}
		bvv := strings.Split(strings.TrimSpace(string(buf)), " ")
		if len(bvv) < 2 {
			return 0, 0, errors.New("invaild stty result:'" + string(buf) + "'")
		}
		y, err := strconv.Atoi(bvv[0])
		if err != nil {
			return 0, 0, errors.New("invaild stty rows: " + bvv[0])
		}
		x, err := strconv.Atoi(bvv[1])
		if err != nil {
			return 0, 0, errors.New("invaild stty columns: " + bvv[1])
		}
		return x, y, nil
	}
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(os.Stderr.Fd()), &info); err != nil {
		return 0, 0, err
	}
	return int(info.Size.X), int(info.Size.Y), nil
}

// Windows Terminaol WT_SESSION todo
// Mintty TTY

// IsTerminal todo
func IsTerminal(fd *os.File) bool {
	return isatty.IsTerminal(fd.Fd()) || isatty.IsCygwinTerminal(fd.Fd())
}

// ReadPassPhrase todo
// openssh-portable-8.1.0.0\contrib\win32\win32compat\msic.c

// save_state=$(stty -g)
//

// http://man7.org/linux/man-pages/man1/stty.1.html

// \Device\NamedPipe\msys-1888ae32e00d56aa-pty0-echoloop
// Cygwin uses special named pipes to simulate TTY, and Mintty listens to these obvious
// pipes to control the terminal emulator, adjust its size, mode, and so on.
//
// If we parse these pipelines to control Mintty, this is a little more complicated.
// In short, we can use stty to complete this work.
// https://github.com/git/git/blob/master/compat/terminal.c#L90

// golang

type cygwinIoctl struct {
	argv []string
}

// windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT
func (ci *cygwinIoctl) Disable(flags int) bool {
	cmd := exec.Command("stty")
	cmd.Stdin = os.Stdin
	if flags&windows.ENABLE_LINE_INPUT != 0 {
		ci.argv = append(ci.argv, "icanon")
		cmd.Args = append(cmd.Args, "-icanon")
	}
	if flags&windows.ENABLE_ECHO_INPUT != 0 {
		ci.argv = append(ci.argv, "echo")
		cmd.Args = append(cmd.Args, "-echo")
	}
	if flags&windows.ENABLE_PROCESSED_INPUT != 0 {
		ci.argv = append(ci.argv, "-ignbrk", "intr", "^c")
		cmd.Args = append(cmd.Args, "ignbrk", "intr", "")
	}
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func (ci *cygwinIoctl) Restore() {
	cmd := exec.Command("stty", ci.argv...)
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

// ReadPassword todo
func ReadPassword(prompt string) (string, error) {
	if fd := int(os.Stdin.Fd()); term.IsTerminal(fd) {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		pwd, err := term.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		return string(pwd), nil
	}
	if isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		var ci cygwinIoctl
		if !ci.Disable(windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT) {
			return "", errors.New("unable set echo mode")
		}
		defer ci.Restore()
	}
	fmt.Fprintf(os.Stderr, "%s: ", prompt)
	pwd, err := ReadInput(os.Stdin, true)
	if err != nil {
		return "", err
	}
	return string(pwd), nil
}

// ReadInputEx todo
func ReadInputEx(fd *os.File) ([]byte, error) {
	if isatty.IsTerminal(fd.Fd()) {
		return ReadInput(fd, false)
	}
	return ReadInput(fd, true)
}

// MakeRaw todo
func MakeRaw(fd *os.File) (*term.State, error) {
	if isatty.IsTerminal(fd.Fd()) {
		return term.MakeRaw(int(fd.Fd()))
	}
	return nil, errors.New("unsupport raw mode")
}
