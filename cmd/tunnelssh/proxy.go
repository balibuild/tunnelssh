package main

import (
	"encoding/base64"
	"errors"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/balibuild/tunnelssh/cli"
)

// SSH_PROXY

// error
var (
	ErrProxyNotConfigured = errors.New("Proxy is not configured correctly")
)

// ProxySettings todo
type ProxySettings struct {
	ProxyServer   string
	ProxyOverride string // aka no proxy
	ipMatchers    []matcher

	// domainMatchers represent all values in the NoProxy that are a domain
	// name or hostname & domain name
	domainMatchers []matcher
	initialized    bool
	sep            string
}

// UseProxy todo
func (ps *ProxySettings) UseProxy(addr string) bool {
	if !ps.initialized {
		if err := ps.Initialize(); err != nil {
			return true
		}
	}
	if len(addr) == 0 {
		return true
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return false
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() {
			return false
		}
	}

	addr = strings.ToLower(strings.TrimSpace(host))

	if ip != nil {
		for _, m := range ps.ipMatchers {
			if m.match(addr, port, ip) {
				return false
			}
		}
	}
	for _, m := range ps.domainMatchers {
		if m.match(addr, port, ip) {
			return false
		}
	}
	return true
}

func getEnvAny(names ...string) string {
	for _, n := range names {
		if val := os.Getenv(n); val != "" {
			return val
		}
	}
	return ""
}

func schemePort(scheme string) string {
	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	case "socks5", "socks5h":
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
