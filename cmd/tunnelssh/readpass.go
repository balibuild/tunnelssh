package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/balibuild/tunnelssh/cli"
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

// AttachConsole

func readAskPass(prompt, user string, passwd bool) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	askpass := filepath.Join(filepath.Dir(exe), "ssh-askpass")
	cmd := exec.Command(askpass)
	if passwd {
		cmd.Args = append(cmd.Args, "-p", prompt, "-u", user)
	} else {
		cmd.Args = append(cmd.Args, prompt)
	}
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
	if err := cmd.Wait(); err != nil {
		return "", err
	}
	return ln, nil
}

// AskPassword todo
func (sc *SSHClient) AskPassword() (string, error) {
	if pty.IsTerminal(os.Stdin) {
		return pty.ReadPassword("Password")
	}
	return readAskPass(cli.StrCat("TunnelSSH connect to ", sc.config.User, "@", sc.host, " password: "), sc.config.User, true)
}
