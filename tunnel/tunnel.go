package tunnel

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/balibuild/tunnelssh/cli"
)

// Proxy library

// BoringMachine todo
type BoringMachine struct {
	Setting *ProxySettings // proxy url
	Debug   func(msg string)
}

// DebugPrint todo
func (bm *BoringMachine) DebugPrint(format string, a ...interface{}) {
	if bm.Debug != nil {
		ss := fmt.Sprintf(format, a...)
		bm.Debug(ss)
	}
}

// Initialize todo
func (bm *BoringMachine) Initialize() error {
	p, err := ResolveProxy()
	if err != nil {
		return nil
	}
	bm.Setting = p
	bm.DebugPrint("Use proxy %s", bm.Setting.ProxyServer)
	return nil
}

// DialTimeout auto dial
func (bm *BoringMachine) DialTimeout(network string, address string, timeout time.Duration) (net.Conn, error) {
	if bm.Setting == nil || !bm.Setting.UseProxy(address) {
		conn, err := net.DialTimeout(network, address, timeout)
		if err != nil {
			return nil, err
		}
		bm.DebugPrint("Establish direct connection %s", address)
		return conn, nil
	}
	proxyurl := bm.Setting.ProxyServer
	if strings.Index(proxyurl, "://") == -1 {
		proxyurl = "http://" + proxyurl // avoid proxy url parse failed
	}
	u, err := url.Parse(proxyurl)
	if err != nil {
		return nil, cli.ErrorCat("invalid proxy url: ", proxyurl)
	}
	proxyaddress := urlMakeAddress(u)
	if proxyaddress == address {
		return nil, cli.ErrorCat("The proxy server address and the target address are the same ", proxyurl, "==", address)
	}
	switch u.Scheme {
	case "https", "http":
		return bm.DialTunnelHTTP(u, proxyaddress, address, timeout)
	case "socks5", "socks5h":
		return bm.DialTunnelSocks5(u, proxyaddress, address, timeout)
	case "ssh":
		return bm.DialTunnelSSH(u, proxyaddress, address, timeout)
	default:
	}
	return nil, cli.ErrorCat("not support current scheme", u.Scheme)
}

// Dial todo
func (bm *BoringMachine) Dial(network string, address string) (net.Conn, error) {
	return bm.DialTimeout(network, address, 30*time.Second)
}
