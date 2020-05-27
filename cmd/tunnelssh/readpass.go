package main

import (
	"bufio"
	"fmt"
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
		fmt.Fprintf(os.Stderr, "%v", err)
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
	out, err := cmd.StdoutPipe()
	if err != nil {
		DebugPrint("read askpass StdoutPipe %v", err)
		return "", err
	}
	defer out.Close()
	br := bufio.NewReader(out)
	cmd.Start()
	ln, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		DebugPrint("read askpass ReadString %v", err)
		return "", err
	}
	ln = strings.TrimSpace(ln)
	if err := cmd.Wait(); err != nil {
		DebugPrint("read askpass Wait %v", err)
		return "", err
	}
	return ln, nil
}

// AskPassword todo
func (sc *SSHClient) AskPassword() (string, error) {
	if pty.IsTerminal(os.Stdin) {
		return pty.ReadPassword("Password")
	}
	prompt := cli.StrCat("Enter your credentials for ", sc.config.User, "@", sc.host)
	return readAskPass(prompt, sc.config.User, true)
}
