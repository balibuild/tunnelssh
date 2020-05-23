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
	"golang.org/x/crypto/ssh"
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
		respond, err := pty.ReadInput(os.Stdin, true)
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

// The authenticity of host 'github.com (140.82.113.4)' can't be established.
// RSA key fingerprint is SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8.
// Are you sure you want to continue connecting (yes/no/[fingerprint])

func askIsHostTrusted(host string, key ssh.PublicKey) bool {
	fintgerprint := ssh.FingerprintSHA256(key)
	DebugPrint("Fingerprint %s", fintgerprint)
	return false
}

// AskPassword todo
func AskPassword(prompt string) (string, error) {
	if pty.IsTerminal(os.Stdin) {
		return pty.ReadPassword(prompt)
	}
	return readAskPass(prompt, 0)
}
