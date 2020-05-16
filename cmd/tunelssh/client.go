package main

import "golang.org/x/crypto/ssh"

//

type client struct {
	ssh *ssh.Client
}

// DialTunnell todo
func DialTunnell(proxy, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {

	return nil, nil
}

// Dial todo
func Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	if p, err := ResolveProxy(); err == nil {
		return DialTunnell(p, network, addr, config)
	}
	return ssh.Dial(network, addr, config)
}
