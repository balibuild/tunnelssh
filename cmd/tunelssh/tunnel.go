package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"os/user"
	"time"

	"github.com/balibuild/tunnelssh/cli"
	"golang.org/x/crypto/ssh"
)

type sshconn struct {
	client *ssh.Client
	chcon  net.Conn
}

// Read reads data from the connection.
func (conn *sshconn) Read(b []byte) (int, error) {
	return conn.chcon.Read(b)
}

// Write writes data
func (conn *sshconn) Write(b []byte) (int, error) {
	return conn.chcon.Write(b)
}

// Close closes the connection.
func (conn *sshconn) Close() error {
	if conn.chcon != nil {
		_ = conn.chcon.Close()
	}
	return conn.client.Close()
}

// LocalAddr returns the local network address.
func (conn *sshconn) LocalAddr() net.Addr {
	return conn.client.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (conn *sshconn) RemoteAddr() net.Addr {
	return conn.client.RemoteAddr()
}

// SetDeadline wapper
func (conn *sshconn) SetDeadline(t time.Time) error {
	return conn.chcon.SetDeadline(t)
}

// SetReadDeadline wapper
func (conn *sshconn) SetReadDeadline(t time.Time) error {
	return conn.chcon.SetReadDeadline(t)
}

// SetWriteDeadline wapper
func (conn *sshconn) SetWriteDeadline(t time.Time) error {
	return conn.chcon.SetWriteDeadline(t)
}

// DialTunnelSSH todo
func DialTunnelSSH(u *url.URL, paddr, addr string, config *ssh.ClientConfig) (net.Conn, error) {
	conn := &sshconn{}
	var err error
	configcopy := *config
	if u.User != nil {
		configcopy.User = u.User.Username()
	} else {
		current, err := user.Current()
		if err != nil {
			return nil, err
		}
		configcopy.User = current.Name
	}
	if conn.client, err = ssh.Dial("tcp", paddr, &configcopy); err != nil {
		return nil, err
	}
	if conn.chcon, err = conn.client.Dial("tcp", addr); err != nil {
		return nil, err
	}
	return conn, nil
}

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

// DialTunnelHTTP todo
func DialTunnelHTTP(u *url.URL, paddr, addr string) (net.Conn, error) {
	var err error
	var conn net.Conn
	if u.Scheme == "https" {
		conn, err = tls.Dial("tcp", paddr, nil)
	} else {
		conn, err = net.DialTimeout("tcp", paddr, 10*time.Second)
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
	return pc, nil
}
