package main

import (
	"net"
	"net/url"
	"os/user"
	"time"

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
