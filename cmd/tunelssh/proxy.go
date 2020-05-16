package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net"
	"net/url"
	"os/user"
	"strings"
	"time"

	"github.com/balibuild/tunnelssh/cli"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

// SSH_PROXY

// error
var (
	ErrProxyNotConfigured = errors.New("Proxy is not configured correctly")
)

func schemePort(scheme string) string {
	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	case "socks5":
		return "1080"
	case "ssh":
		return "22"
	}
	return "80"
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func splitHostPort(hostport string) (host, port string) {
	host = hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

func urlMakeAddress(u *url.URL) string {
	host := u.Host
	port := u.Port()
	if len(port) != 0 {
		return host
	}

	if strings.IndexByte(host, ':') != -1 {
		cli.StrCat("[", host, "]:", schemePort(u.Scheme))
	}
	return cli.StrCat(host, ":", schemePort(u.Scheme))
}

func basicAuth(ui *url.Userinfo) string {
	passwd, _ := ui.Password()
	auth := ui.Username() + ":" + passwd
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// DialTunnelSock5 todo
func DialTunnelSock5(u *url.URL, addr string) (net.Conn, error) {
	//proxy.SOCKS5()
	//proxy.SOCKS5()
	proxy.SOCKS5("tcp", addr, nil, nil)
	return nil, nil
}

// DialTunnelSSH over ssh
func DialTunnelSSH(u *url.URL, addr string, config *ssh.ClientConfig) (net.Conn, error) {
	//conn,err:=ssh.Dial()
	var name string
	if u.User != nil {
		name = u.User.Username()
	} else {
		current, err := user.Current()
		if err != nil {
			return nil, err
		}
		name = current.Name
	}
	DebugPrint("user %s", name)
	return nil, nil
}

// DailTunnelInternal todo
func DailTunnelInternal(pu, addr string, config *ssh.ClientConfig) (net.Conn, error) {
	if strings.Index(pu, "://") == -1 {
		pu = "http://" + pu // avoid proxy url parse failed
	}
	u, err := url.Parse(pu)
	if err != nil {
		return nil, err
	}
	ph, _ := splitHostPort(addr)
	address := urlMakeAddress(u)
	var conn net.Conn
	switch u.Scheme {
	case "https":
		conn, err = tls.Dial("tcp", address, nil)
	case "http":
		conn, err = net.DialTimeout("tcp", address, 10*time.Second)
	case "socks5":
		return DialTunnelSock5(u, addr)
	case "ssh":
		return DialTunnelSSH(u, addr, config)
	default:
		return nil, cli.ErrorCat("not support current scheme", u.Scheme)
	}
	if err != nil {
		return nil, cli.ErrorCat("Counld't establish connection to proxy: ", err.Error())
	}
	var buf bytes.Buffer
	buf.Grow(512)
	_, _ = buf.WriteString("CONNECT ")
	_, _ = buf.WriteString(address)
	_, _ = buf.WriteString(" HTTP/1.1\nHost: ")
	_, _ = buf.WriteString(ph) // Host information
	_, _ = buf.WriteString("\nProxy-Connection: Keep-Alive\nContent-Length: 0\nUser-Agent:SSH/9.0\n")
	if u.User != nil {
		_, _ = buf.WriteString("\nProxy-Authorization: Basic ")
		_, _ = buf.WriteString(basicAuth(u.User))
	}
	_, _ = buf.WriteString("\r\n\r\n")
	if _, err := conn.Write(buf.Bytes()); err != nil {
		conn.Close()
		return nil, cli.ErrorCat("Counld't send CONNECT request to proxy: ", err.Error())
	}
	res, err := StateMachineCONNECT(conn)
	if err != nil {
		return nil, cli.ErrorCat("reading HTTP response from CONNECT to ", addr, " via proxy ", pu, " failed: ", err.Error())
	}
	// HTTP/1.1 200 Connection Established
	// HTTP/1.1 407 Unauthorized
	if res.StatusCode != 200 {
		return nil, cli.ErrorCat("proxy error from ", pu, " while dialing ", addr, ":", res.Status)
	}
	return conn, nil
}
