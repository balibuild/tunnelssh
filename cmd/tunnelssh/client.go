package main

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/balibuild/tunnelssh/cli"
	"github.com/balibuild/tunnelssh/pty"
	"github.com/balibuild/tunnelssh/tunnel"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// SSHClient client
type SSHClient struct {
	ssh                 *ssh.Client
	config              *ssh.ClientConfig
	sess                *ssh.Session
	home                string
	IdentityFile        string
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
func (sc *SSHClient) SendEnv() error {
	if len(sc.env) == 0 {
		return nil
	}
	for k, v := range sc.env {
		DebugPrint("SetEnv %s=%s", k, v)
		sc.sess.Setenv(k, v)
	}
	return nil
}

// Shell to open a shell
func (sc *SSHClient) Shell() error {
	DebugPrint("ssh shell mode. host: %s", sc.host)
	sc.sess.Stdout = os.Stdout
	sc.sess.Stderr = os.Stderr
	sc.sess.Stdin = os.Stdin
	if sc.mode == TerminalModeForce {
		x, y, err := pty.GetWinSize()
		if err != nil {
			return err
		}
		if termstate, err := terminal.MakeRaw(int(os.Stdin.Fd())); err == nil {
			defer terminal.Restore(int(os.Stdin.Fd()), termstate)
		}
		modes := ssh.TerminalModes{
			ssh.ECHO:          0,
			ssh.IGNCR:         1,
			ssh.TTY_OP_ISPEED: 115200, // baud in
			ssh.TTY_OP_OSPEED: 115200, // baud out
		}
		if err := sc.sess.RequestPty("xterm", y, x, modes); err != nil {
			return err
		}
	}
	if err := sc.sess.Shell(); err != nil {
		return err
	}
	return sc.sess.Wait()
}

// Loop todo
func (sc *SSHClient) Loop() error {
	_ = sc.SendEnv()
	if len(sc.argv) == 0 {
		return sc.Shell()
	}
	sc.sess.Stdout = os.Stdout
	sc.sess.Stderr = os.Stderr
	sc.sess.Stdin = os.Stdin
	// git escape argv done
	args := strings.Join(sc.argv, " ")
	DebugPrint("cmd: %s", args)
	return sc.sess.Run(args)
}

// DialTunnel todo
func DialTunnel(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	var bm tunnel.BoringMachine
	if IsDebugMode {
		bm.Debug = func(msg string) {
			_, _ = os.Stderr.WriteString(cli.StrCat("debug3: \x1b[33m", msg, "\x1b[0m\n"))
		}
	}
	_ = bm.Initialize()
	conn, err := bm.DialTimeout(network, addr, config.Timeout)
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
func (sc *SSHClient) Dial() error {
	if sc.connectTimeout != 0 {
		sc.config.Timeout = time.Duration(sc.connectTimeout) * time.Second
	} else {
		sc.config.Timeout = 5 * time.Second
	}
	addr := net.JoinHostPort(sc.host, strconv.Itoa(sc.port))
	conn, err := DialTunnel("tcp", addr, sc.config)
	if err != nil {
		return err
	}
	sc.ssh = conn
	sess, err := sc.ssh.NewSession()
	if err != nil {
		return err
	}
	sc.sess = sess
	return nil
}

// Close client
func (sc *SSHClient) Close() error {
	if sc.sess != nil {
		sc.sess.Close()
	}
	if sc.ka != nil {
		sc.ka.Close()
	}
	if sc.ssh != nil {
		return sc.ssh.Close()
	}
	return nil
}
