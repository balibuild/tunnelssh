package main

import (
	"net"
	"net/url"

	"golang.org/x/net/proxy"
)

// DialTunnelSock5 todo
func DialTunnelSock5(u *url.URL, paddr, addr string) (net.Conn, error) {
	dialer, err := proxy.SOCKS5("tcp", addr, nil, nil)
	if err != nil {
		return nil, err
	}
	return dialer.Dial("tcp", addr)
}
