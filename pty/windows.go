// +build windows

package pty

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"golang.org/x/sys/windows"
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
