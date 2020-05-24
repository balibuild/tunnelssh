package tunnel

import (
	"net"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

// DialTunnelSocks5 todo
func (bm *BoringMachine) DialTunnelSocks5(u *url.URL, paddr, addr string, timeout time.Duration) (net.Conn, error) {
	var auth *proxy.Auth
	if u.User != nil {
		auth = new(proxy.Auth)
		auth.User = u.User.Username()
		if p, ok := u.User.Password(); ok {
			auth.Password = p
		}
	}
	// see https://github.com/golang/go/issues/37549
	dialer, err := proxy.SOCKS5("tcp", addr, auth, nil)
	if err != nil {
		return nil, err
	}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	bm.DebugPrint("Establish connection to proxy(%s): %s", u.Scheme, paddr)
	return conn, nil
}
