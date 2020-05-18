package main

import (
	"errors"
	"os"

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

func readAskPass() error {
	askpass := os.Getenv("SSH_ASKPASS")
	if len(askpass) == 0 {
		return errors.New("SSH_ASKPASS not set")
	}
	return nil
}

func askIsHostTrusted(host string, key ssh.PublicKey) bool {
	if IsTerminal(os.Stdin) {
		// read input
		//
	}
	return false
}

func sshPasswordPrompt() (string, error) {
	if IsTerminal(os.Stdin) {
		// read input
		//
	}
	return "", nil
}
