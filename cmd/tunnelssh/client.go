package main

import (
	"golang.org/x/crypto/ssh"
)

//

type client struct {
	ssh        *ssh.Client
	config     *ssh.ClientConfig
	sess       *ssh.Session
	ka         *KeyAgent
	argv       []string // unresolved command argv
	env        map[string]string
	host       string
	port       int
	forcetty   bool
	forcenotty bool
}

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

// SendEnv todo
func (c *client) SendEnv() error {
	if len(c.env) == 0 {
		return nil
	}
	// sess, err := c.ssh.NewSession()
	// if err != nil {
	// 	return err
	// }
	// for k, v := range c.env {
	// 	sess.Setenv(k, v)
	// }
	return nil
}

func (c *client) Shell() error {
	if c.forcetty {
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
