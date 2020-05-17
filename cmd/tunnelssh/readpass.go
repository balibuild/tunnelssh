package main

import "golang.org/x/crypto/ssh"

// GIT_ASKPASS
// SSH_ASKPASS
// GIT_TERMINAL_PROMPT
// git-gui--askpass

// // ReadPassphrase todo
// func ReadPassphrase(prompt string, flags int) string {

// 	return ""
// }

func askIsHostTrusted(host string, key ssh.PublicKey) bool {

	return false
}

func sshPasswordPrompt() (string, error) {
	return "", nil
}
