package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/balibuild/tunnelssh/pty"
	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/windows"
)

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
		fmt.Fprintf(os.Stderr, "error %v", err)
		return false
	}
	return true
}

func (ci *cygwinIoctl) Restore() {
	cmd := exec.Command("stty", ci.argv...)
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "restore %v", err)
	}
}

// ReadInput todo
func ReadInput(reader io.Reader, unix bool) ([]byte, error) {
	var buf [1]byte
	var ret []byte

	for {
		n, err := reader.Read(buf[:])
		if n > 0 {
			switch buf[0] {
			case '\b':
				if len(ret) > 0 {
					ret = ret[:len(ret)-1]
				}
			case '\n':
				if unix {
					return ret, nil
				}
				// otherwise ignore \n
			case '\r':
				if !unix {
					return ret, nil
				}
				// otherwise ignore \r
			default:
				ret = append(ret, buf[0])
			}
			continue
		}
		if err != nil {
			if err == io.EOF && len(ret) > 0 {
				return ret, nil
			}
			return ret, err
		}
	}
}

// AskPrompt todo
func AskPrompt(prompt string) (string, error) {
	if fd := int(os.Stdin.Fd()); terminal.IsTerminal(fd) {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		respond, err := ReadInput(os.Stdin, false)
		if err != nil {
			return "", err
		}
		return string(respond), nil
	}
	if isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		respond, err := ReadInput(os.Stdin, true)
		if err != nil {
			return "", err
		}
		return string(respond), nil
	}
	return "", nil
}

func askPassword() {
	if isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		var ci cygwinIoctl
		if !ci.Disable(windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT) {
			fmt.Fprintf(os.Stderr, "unable create echo mode")
			os.Exit(1)
		}
		defer ci.Restore()
		fmt.Fprintf(os.Stderr, "Please input password: ")
		passwd, err := ReadInput(os.Stdin, true)
		if err == nil {
			fmt.Fprintf(os.Stderr, "\npassword: %s\n", passwd)
		}

	} else {
		fmt.Fprintf(os.Stderr, "Please input password: ")
		passwd, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err == nil {
			fmt.Fprintf(os.Stderr, "\npassword: %s\n", passwd)
		}
	}
}

func main() {
	fd := int(os.Stdin.Fd())
	xterm := os.Getenv("TERM")
	fmt.Fprintf(os.Stderr, "IsTerminal: %v %s\n", terminal.IsTerminal(fd), xterm)
	askPassword()
	yesno, err := AskPrompt("Please input yes/no")
	if err == nil {
		fmt.Fprintf(os.Stderr, "input: [%s]\n", yesno)
	}
	x, y, err := pty.GetWinSize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetWinSize: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "WinSize: %d %d\n", x, y)
}
