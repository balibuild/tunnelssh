package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
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

func readPasswordLine(reader io.Reader) ([]byte, error) {
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
				if runtime.GOOS != "windows" {
					return ret, nil
				}
				// otherwise ignore \n
			case '\r':
				if runtime.GOOS == "windows" {
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
	if err != nil {
		return "", err
	}
	ln = strings.TrimSuffix(ln, "\r")
	cmd.Wait()
	return ln, nil
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
		//terminal.ReadPassword(os.Stdin.Fd())
	}
	DebugPrint("stdin is not a tty")
	return "", nil
}
