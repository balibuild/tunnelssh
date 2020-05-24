package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/balibuild/tunnelssh/pty"
)

// GIT_ASKPASS
// SSH_ASKPASS
// GIT_TERMINAL_PROMPT
// git-gui--askpass

// ReadInput todo

// Ask flags
const (
	AskNone = 0
	AskEcho = 1
)

// AskPrompt todo
func AskPrompt(prompt string) (string, error) {
	if pty.IsTerminal(os.Stdin) {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		respond, err := pty.ReadInputEx(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(respond), nil
	}
	return readAskPass(prompt, AskNone)
}

func readAskPass(prompt string, flags int) (string, error) {
	askpass := os.Getenv("SSH_ASKPASS")
	if len(askpass) == 0 {
		return "", errors.New("SSH_ASKPASS not set")
	}
	cmd := exec.Command(askpass, prompt)
	cmd.Stderr = os.Stderr //bind stderr
	in, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer in.Close()
	br := bufio.NewReader(in)
	cmd.Start()
	ln, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	ln = strings.TrimSuffix(ln, "\r")
	cmd.Wait()
	return ln, nil
}

// AskPassword todo
func (sc *SSHClient) AskPassword() (string, error) {
	if pty.IsTerminal(os.Stdin) {
		return pty.ReadPassword("Password")
	}
	return readAskPass("Password", 0)
}
