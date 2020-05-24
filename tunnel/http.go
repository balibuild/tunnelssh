package tunnel

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/balibuild/tunnelssh/cli"
)

type proxyconn struct {
	conn net.Conn
	br   *bufio.Reader
}

// Read reads data from the connection.
func (pc *proxyconn) Read(b []byte) (int, error) {
	return pc.br.Read(b)
}

// Write writes data
func (pc *proxyconn) Write(b []byte) (int, error) {
	return pc.conn.Write(b)
}

// Close closes the connection.
func (pc *proxyconn) Close() error {
	return pc.conn.Close()
}

// LocalAddr returns the local network address.
func (pc *proxyconn) LocalAddr() net.Addr {
	return pc.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (pc *proxyconn) RemoteAddr() net.Addr {
	return pc.conn.RemoteAddr()
}

// SetDeadline wapper
func (pc *proxyconn) SetDeadline(t time.Time) error {
	return pc.conn.SetDeadline(t)
}

// SetReadDeadline wapper
func (pc *proxyconn) SetReadDeadline(t time.Time) error {
	return pc.conn.SetReadDeadline(t)
}

// SetWriteDeadline wapper
func (pc *proxyconn) SetWriteDeadline(t time.Time) error {
	return pc.conn.SetWriteDeadline(t)
}

// DialTunnelHTTP use http proxy
func (bm *BoringMachine) DialTunnelHTTP(u *url.URL, paddr, addr string, timeout time.Duration) (net.Conn, error) {
	var err error
	var conn net.Conn
	if u.Scheme == "https" {
		d := &net.Dialer{Timeout: timeout}
		conn, err = tls.DialWithDialer(d, "tcp", paddr, nil)
	} else {
		conn, err = net.DialTimeout("tcp", paddr, timeout)
	}
	if err != nil {
		return nil, cli.ErrorCat("Counld't establish connection to proxy: ", err.Error())
	}
	var buf bytes.Buffer
	buf.Grow(512)
	ph, _ := splitHostPort(addr)
	_, _ = buf.WriteString("CONNECT ")
	_, _ = buf.WriteString(addr)
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
	pc := &proxyconn{conn: conn, br: bufio.NewReader(conn)}
	resp, err := http.ReadResponse(pc.br, nil)
	if err != nil {
		pc.Close()
		return nil, cli.ErrorCat("reading HTTP response from CONNECT to ", addr, " via proxy ", paddr, " failed: ", err.Error())
	}
	if resp.StatusCode != 200 {
		pc.Close()
		return nil, cli.ErrorCat("proxy error from ", paddr, " while dialing ", addr, ":", resp.Status)
	}
	bm.DebugPrint("Establish connection to proxy(%s): %s", u.Scheme, paddr)
	return pc, nil
}
