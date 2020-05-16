package main

import (
	"net"

	"golang.org/x/crypto/ssh"
)

//

type client struct {
	ssh  *ssh.Client
	conn net.Conn
}

// DialTunnel todo
func DialTunnel(p, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := DailTunnelInternal(p, addr, config)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

// Dial todo
func Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	if p, err := ResolveProxy(); err == nil {
		return DialTunnel(p, network, addr, config)
	}
	return ssh.Dial(network, addr, config)
}
