package main

import (
	"golang.org/x/crypto/ssh"
)

//

type client struct {
	ssh  *ssh.Client
	argv []string
}

// DialTunnel todo
func DialTunnel(p, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := DailTunnelInternal(p, addr, config)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

// Dial todo
func Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	if ps, err := ResolveProxy(); err == nil {
		DebugPrint("resolve proxy config: %s", ps.ProxyServer)
		return DialTunnel(ps.ProxyServer, network, addr, config)
	}
	DebugPrint("no proxy env found direct dail: %s", addr)
	return ssh.Dial(network, addr, config)
}
