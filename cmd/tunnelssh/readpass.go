package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func lookupAskPass() (string, error) {
	var suffix string
	if runtime.GOOS == "windows" {
		if askpass, err := exec.LookPath("ssh-askpass-baulk"); err == nil {
			return askpass, nil
		}
		suffix = ".exe"
	}
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "lookup executable %v", err)
		return "", err
	}
	askpassname := "ssh-askpass" + suffix
	askpass := filepath.Join(filepath.Dir(exe), askpassname)
	if _, err := os.Stat(askpass); err != nil {
		return "", err
	}
	return askpass, nil
}

func readAskPass(prompt, user string, passwd bool) (string, error) {
	askpass, err := lookupAskPass()
	if err != nil {
		return "", err
	}
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
	if err = cmd.Start(); err != nil {
		return "", err
	}
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
