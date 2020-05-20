package main

import (
	"net"
	"net/url"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

// DialTunnelSock5 todo
func DialTunnelSock5(u *url.URL, paddr, addr string, config *ssh.ClientConfig) (net.Conn, error) {
	var auth *proxy.Auth
	if u.User != nil {
		auth = new(proxy.Auth)
		auth.User = u.User.Username()
		if p, ok := u.User.Password(); ok {
			auth.Password = p
		}
	}
	dialer, err := proxy.SOCKS5("tcp", addr, auth, nil)
	if err != nil {
		return nil, err
	}
	return dialer.Dial("tcp", addr)
}
