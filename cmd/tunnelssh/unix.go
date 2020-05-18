// +build !windows

package main

import (
	"net"
	"os"

	"github.com/balibuild/tunnelssh/cli"
	"github.com/mattn/go-isatty"
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
