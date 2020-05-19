package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
)

func enableEchoInput() (string, error) {
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = os.Stdin
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	cmd2 := exec.Command("stty", "-echo")
	cmd2.Stdin = os.Stdin
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(string(buf)), nil
}

func restoreInput(state string) error {
	cmd := exec.Command("stty", state)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

func main() {
	fd := int(os.Stdin.Fd())
	xterm := os.Getenv("TERM")
	fmt.Fprintf(os.Stderr, "IsTerminal: %v %s\n", terminal.IsTerminal(fd), xterm)
	if isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		fmt.Fprintf(os.Stderr, "IsCygwinTerminal true\n")
		state, err := enableEchoInput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "sttyState error %v\n", err)
			return
		}
		defer restoreInput(state)
		fmt.Fprintf(os.Stderr, "sttyState  %s\n", state)
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
	yesno, err := AskPrompt("Please input yes/no")
	if err == nil {
		fmt.Fprintf(os.Stderr, "input: [%s]\n", yesno)
	}
}
