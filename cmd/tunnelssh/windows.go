// +build windows

package main

import (
	"fmt"
	"os"

	"github.com/balibuild/tunnelssh/cli"
	"github.com/balibuild/winio"
	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/windows/registry"
)

// HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings

// ResolveRegistryProxy todo
func ResolveRegistryProxy() (*ProxySettings, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()
	ps := &ProxySettings{sep: ";"}
	if d, _, err := k.GetIntegerValue("ProxyEnable"); err == nil && d == 1 {
		if s, _, err := k.GetStringValue("ProxyServer"); err == nil && len(s) > 0 {
			ps.ProxyServer = s
		}
	} else {
		if s, _, err := k.GetStringValue("AutoConfigURL"); err == nil && len(s) > 0 {
			ps.ProxyServer = s
		}
	}
	if s, _, err := k.GetStringValue("ProxyOverride"); err == nil && len(s) > 0 {
		ps.ProxyOverride = s
	}
	if ps.ProxyServer != "" {
		return ps, nil
	}
	return nil, ErrProxyNotConfigured
}

// feature read proxy from registry

// ResolveProxy todo
func ResolveProxy() (*ProxySettings, error) {
	if s, err := ResolveRegistryProxy(); err == nil {
		return s, nil
	}
	ps := &ProxySettings{sep: ","}
	ps.ProxyOverride = os.Getenv("NO_PROXY")
	if ps.ProxyServer = os.Getenv("SSH_PROXY"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	if ps.ProxyServer = os.Getenv("HTTPS_PROXY"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	if ps.ProxyServer = os.Getenv("HTTP_PROXY"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	if ps.ProxyServer = os.Getenv("ALL_PROXY"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	return nil, ErrProxyNotConfigured
}

// MakeAgent make agent
// Windows use pipe now
// https://github.com/PowerShell/openssh-portable/blob/latestw_all/contrib/win32/win32compat/ssh-agent/agent.c#L40
func (ka *KeyAgent) MakeAgent() error {
	if len(os.Getenv("SSH_AUTH_SOCK")) == 0 {
		return cli.ErrorCat("ssh agent not initialized")
	}
	// \\\\.\\pipe\\openssh-ssh-agent
	conn, err := winio.DialPipe("\\\\.\\pipe\\openssh-ssh-agent", nil)
	if err != nil {
		return err
	}
	ka.conn = conn
	return nil
}

// Windows Terminaol WT_SESSION todo
// Mintty TTY

// IsTerminal todo
func IsTerminal(fd *os.File) bool {
	return isatty.IsTerminal(fd.Fd()) || isatty.IsCygwinTerminal(fd.Fd())
}

// ReadPassPhrase todo
// openssh-portable-8.1.0.0\contrib\win32\win32compat\msic.c
func ReadPassPhrase(prompt string, flags int) (string, error) {

	return "", nil
}

//Askpass ask password
func Askpass(prompt string, verify bool) (string, error) {
	if fd := int(os.Stdin.Fd()); terminal.IsTerminal(fd) {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		pwd, err := terminal.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		return string(pwd), nil
	}
	if isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		pwd, err := readPasswordLine(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(pwd), nil
	}
	return "", nil
}
