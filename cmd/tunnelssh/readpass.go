package main

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
)

// GIT_ASKPASS
// SSH_ASKPASS
// GIT_TERMINAL_PROMPT
// git-gui--askpass

// // ReadPassphrase todo
// func ReadPassphrase(prompt string, flags int) string {

// 	return ""
// }

func readAskPass(prompt string, flags int) (string, error) {
	askpass := os.Getenv("SSH_ASKPASS")
	if len(askpass) == 0 {
		return "", errors.New("SSH_ASKPASS not set")
	}
	cmd := exec.Command(askpass, prompt)
	in, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer in.Close()
	br := bufio.NewReader(in)
	cmd.Start()
	for i := 0; i < 3; i++ {
		ln, err := br.ReadString('\n')
		if err != nil {
			break
		}
		ln = strings.TrimSuffix(ln, "\r")
	}
	cmd.Wait()
	return "", nil
}

func askIsHostTrusted(host string, key ssh.PublicKey) bool {
	if IsTerminal(os.Stdin) {
		// read input
		//
	}
	DebugPrint("stdin is not a tty")
	return false
}

func sshPasswordPrompt() (string, error) {
	if IsTerminal(os.Stdin) {
		// read input
		//
	}
	DebugPrint("stdin is not a tty")
	return "", nil
}
