// +build windows

package main

import (
	"os"

	"github.com/balibuild/tunnelssh/cli"
	"github.com/balibuild/winio"
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

// DefaultKnownHosts todo
func DefaultKnownHosts() string {
	return os.ExpandEnv("$USERPROFILE\\.ssh\\known_hosts")
}

// Initialize todo
func (ks *KeySearcher) Initialize() error {
	ks.home = os.ExpandEnv("$USERPROFILE")
	return nil
}
