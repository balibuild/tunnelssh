package main

import (
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

//

// DialTunnel todo
func DialTunnel(p, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := DailTunnelInternal(p, addr, config)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

// Dial todo
func Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	if ps, err := ResolveProxy(); err == nil {
		DebugPrint("resolve proxy config: %s", ps.ProxyServer)
		return DialTunnel(ps.ProxyServer, network, addr, config)
	}
	DebugPrint("no proxy env found direct dail: %s", addr)
	return ssh.Dial(network, addr, config)
}

type client struct {
	ssh                 *ssh.Client
	config              *ssh.ClientConfig
	sess                *ssh.Session
	ka                  *KeyAgent
	argv                []string // unresolved command argv
	env                 map[string]string
	host                string
	port                int
	mode                TerminalMode
	v4                  bool
	v6                  bool
	serverAliveInterval int
	connectTimeout      int
}

// SendEnv todo
func (c *client) SendEnv() error {
	if len(c.env) == 0 {
		return nil
	}
	for k, v := range c.env {
		c.sess.Setenv(k, v)
	}
	return nil
}

func (c *client) Shell() error {
	if c.mode == TerminalModeForce {
		// Set up terminal modes
		// https://net-ssh.github.io/net-ssh/classes/Net/SSH/Connection/Term.html
		// https://www.ietf.org/rfc/rfc4254.txt
		// https://godoc.org/golang.org/x/crypto/ssh
		// THIS IS THE TITLE
		// https://pythonhosted.org/ANSIColors-balises/ANSIColors.html
		modes := ssh.TerminalModes{ssh.ECHO: 0, ssh.IGNCR: 1}
		if err := c.sess.RequestPty("vt100", 90, 30, modes); err != nil {
			return err
		}
	}
	if err := c.sess.Shell(); err != nil {
		return err
	}
	return c.sess.Wait()
}

// Loop todo
func (c *client) Loop() error {
	if len(c.argv) == 0 {
		return c.Shell()
	}
	return nil
}

// Dial todo
func (c *client) Dial() error {
	if c.connectTimeout != 0 {
		c.config.Timeout = time.Duration(c.connectTimeout) * time.Second
	} else {
		c.config.Timeout = 5 * time.Second
	}
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := Dial("tcp", addr, c.config)
	if err != nil {
		return err
	}
	c.ssh = conn
	sess, err := c.ssh.NewSession()
	if err != nil {
		return err
	}
	c.sess = sess
	c.sess.Stdin = os.Stdin
	c.sess.Stderr = os.Stderr
	c.sess.Stdout = os.Stdout
	return nil
}

func (c *client) Close() error {
	if c.sess != nil {
		c.sess.Close()
	}
	if c.ssh != nil {
		return c.ssh.Close()
	}
	return nil
}
