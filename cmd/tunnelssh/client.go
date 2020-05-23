package main

import (
	"bytes"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	ssh_config "github.com/balibuild/tunnelssh/external/sshconfig"
	"golang.org/x/crypto/ssh"
)

const (
	specialChars      = "\\'\"`${[|&;<>()*?!"
	extraSpecialChars = " \t\n"
	prefixChars       = "~"
)

func quote(word string, buf *bytes.Buffer) {
	// We want to try to produce a "nice" output. As such, we will
	// backslash-escape most characters, but if we encounter a space, or if we
	// encounter an extra-special char (which doesn't work with
	// backslash-escaping) we switch over to quoting the whole word. We do this
	// with a space because it's typically easier for people to read multi-word
	// arguments when quoted with a space rather than with ugly backslashes
	// everywhere.
	origLen := buf.Len()

	if len(word) == 0 {
		// oops, no content
		buf.WriteString("''")
		return
	}

	cur, prev := word, word
	atStart := true
	for len(cur) > 0 {
		c, l := utf8.DecodeRuneInString(cur)
		cur = cur[l:]
		if strings.ContainsRune(specialChars, c) || (atStart && strings.ContainsRune(prefixChars, c)) {
			// copy the non-special chars up to this point
			if len(cur) < len(prev) {
				buf.WriteString(prev[0 : len(prev)-len(cur)-l])
			}
			buf.WriteByte('\\')
			buf.WriteRune(c)
			prev = cur
		} else if strings.ContainsRune(extraSpecialChars, c) {
			// start over in quote mode
			buf.Truncate(origLen)
			goto quote
		}
		atStart = false
	}
	if len(prev) > 0 {
		buf.WriteString(prev)
	}
	return

quote:
	// quote mode
	// Use single-quotes, but if we find a single-quote in the word, we need
	// to terminate the string, emit an escaped quote, and start the string up
	// again
	inQuote := false
	for len(word) > 0 {
		i := strings.IndexRune(word, '\'')
		if i == -1 {
			break
		}
		if i > 0 {
			if !inQuote {
				buf.WriteByte('\'')
				inQuote = true
			}
			buf.WriteString(word[0:i])
		}
		word = word[i+1:]
		if inQuote {
			buf.WriteByte('\'')
			inQuote = false
		}
		buf.WriteString("\\'")
	}
	if len(word) > 0 {
		if !inQuote {
			buf.WriteByte('\'')
		}
		buf.WriteString(word)
		buf.WriteByte('\'')
	}
}

// QuoteArgs todo
func QuoteArgs(args []string) string {
	var buf bytes.Buffer
	for i := 0; i < len(args); i++ {
		if i != 0 {
			_ = buf.WriteByte(' ')
		}
		quote(args[i], &buf)
	}
	return buf.String()
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

	c.sess.Stdout = os.Stdout
	c.sess.Stderr = os.Stderr
	if len(c.argv) == 0 {
		stdin, err := c.sess.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}
		defer stdin.Close()
		DebugPrint("ssh mode %s", c.host)
		return c.Shell()
	}
	arg := QuoteArgs(c.argv)
	//os.Stdin = c.sess.Stdin
	if err := c.sess.Run(arg); err != nil {
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
