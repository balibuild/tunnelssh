package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
)

// GIT_ASKPASS
// SSH_ASKPASS
// GIT_TERMINAL_PROMPT
// git-gui--askpass

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

// Ask flags
const (
	AskNone = 0
	AskEcho = 1
)

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
	DebugPrint("Fingerprint %s", fingerprint)
	return false
}

func sshPasswordPrompt() (string, error) {
	return AskPassword("Please input password")
}
