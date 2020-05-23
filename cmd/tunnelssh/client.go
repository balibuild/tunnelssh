package main

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	ssh_config "github.com/balibuild/tunnelssh/external/sshconfig"
	"github.com/balibuild/tunnelssh/pty"
	"golang.org/x/crypto/ssh"
)

type client struct {
	sshconfig           *ssh_config.Config
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
		DebugPrint("SetEnv %s=%s", k, v)
		c.sess.Setenv(k, v)
	}
	return nil
}

func (c *client) Shell() error {
	if c.mode == TerminalModeForce {
		x, y, err := pty.GetWinSize()
		if err != nil {
			return err
		}
		modes := ssh.TerminalModes{ssh.ECHO: 0, ssh.IGNCR: 1}
		if err := c.sess.RequestPty("xterm", y, x, modes); err != nil {
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
	_ = c.SendEnv()
	c.sess.Stdout = os.Stdout
	c.sess.Stderr = os.Stderr
	c.sess.Stdin = os.Stdin
	if len(c.argv) == 0 {
		DebugPrint("ssh mode %s", c.host)
		return c.Shell()
	}
	// git escape argv done
	args := strings.Join(c.argv, " ")
	DebugPrint("cmd: %s", args)
	if err := c.sess.Run(args); err != nil {
		return err
	}
	return c.sess.Wait()
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
	return nil
}

func (c *client) Close() error {
	if c.sess != nil {
		c.sess.Close()
	}
	if c.ka != nil {
		c.ka.Close()
	}
	if c.ssh != nil {
		return c.ssh.Close()
	}
	return nil
}
