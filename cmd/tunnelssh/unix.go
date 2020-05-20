// +build !windows

package main

import (
	"fmt"
	"net"
	"os"

	"github.com/balibuild/tunnelssh/cli"
	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
)

// ResolveProxy todo
func ResolveProxy() (*ProxySettings, error) {
	ps := &ProxySettings{sep: ","}
	ps.ProxyOverride = getEnvAny("NO_PROXY", "no_proxy")
	if ps.ProxyServer = getEnvAny("SSH_PROXY", "ssh_proxy"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	if ps.ProxyServer = getEnvAny("HTTPS_PROXY", "https_proxy"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	if ps.ProxyServer = getEnvAny("HTTP_PROXY", "http_proxy"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	if ps.ProxyServer = getEnvAny("ALL_PROXY", "all_proxy"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	return nil, ErrProxyNotConfigured
}

// MakeAgent make agent
func (ka *KeyAgent) MakeAgent() error {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if len(sock) == 0 {
		return cli.ErrorCat("ssh agent not initialized")
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return err
	}
	ka.conn = conn
	return nil
}

// IsTerminal todo
func IsTerminal(fd *os.File) bool {
	return isatty.IsTerminal(fd.Fd())
}

//AskPassword ask password
func AskPassword(prompt string) (string, error) {
	if fd := int(os.Stdin.Fd()); terminal.IsTerminal(fd) {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		pwd, err := terminal.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		return string(pwd), nil
	}
	return readAskPass(prompt, AskEcho)
}

// AskPrompt todo
func AskPrompt(prompt string) (string, error) {
	if fd := int(os.Stdin.Fd()); terminal.IsTerminal(fd) {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		respond, err := ReadInput(os.Stdin, true)
		if err != nil {
			return "", err
		}
		return string(respond), nil
	}
	return readAskPass(prompt, AskNone)
}

func DefaultKnownHosts() string {
	return os.ExpandEnv("$HOME/.ssh/known_hosts")
}

// Initialize todo
func (ks *KeySearcher) Initialize() error {
	ks.home = os.ExpandEnv("$HOME")
	return nil
}
