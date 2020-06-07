package tunnel

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"syscall"
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

// DialTunnel todo
func (bm *BoringMachine) DialTunnel(network string, address string, timeout time.Duration) (net.Conn, error) {
	proxyurl := bm.Setting.ProxyServer
	if !strings.Contains(proxyurl, "://") {
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

func isProxyOffline(err error) bool {
	netErr, ok := err.(net.Error)
	if !ok {
		return false
	}
	if netErr.Timeout() {
		return true
	}
	opErr, ok := netErr.(*net.OpError)
	if !ok {
		return false
	}
	switch t := opErr.Err.(type) {
	case *net.DNSError:
		return false
	case *os.SyscallError:
		if errno, ok := t.Err.(syscall.Errno); ok {
			switch errno {
			case syscall.ECONNREFUSED:
				return true
			case syscall.ETIMEDOUT:
				return true
			}
		}
	}
	return false
}

// DialDirect todo
func (bm *BoringMachine) DialDirect(network string, address string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	bm.DebugPrint("Establish direct connection %s", conn.RemoteAddr().String())
	return conn, nil
}

// DialTimeout auto dial
func (bm *BoringMachine) DialTimeout(network string, address string, timeout time.Duration) (net.Conn, error) {
	if bm.Setting == nil || !bm.Setting.UseProxy(address) {
		return bm.DialDirect(network, address, timeout)
	}
	conn, err := bm.DialTunnel(network, address, timeout)
	if err == nil && !isProxyOffline(err) {
		return conn, err
	}
	bm.DebugPrint("Tunnel cannot establish, try connect direct %s", address)
	return bm.DialDirect(network, address, timeout)
}

// Dial todo
func (bm *BoringMachine) Dial(network string, address string) (net.Conn, error) {
	return bm.DialTimeout(network, address, 30*time.Second)
}
