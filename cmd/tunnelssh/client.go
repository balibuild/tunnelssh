package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
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
	sys                 *sysInfo
	wg                  sync.WaitGroup
}

// error
var (
	ErrGotSignal = errors.New("got signal")
)

func (sc *SSHClient) onFinal(err error) {
	if err == nil {
		DebugPrint("ssh connetion to %s successfully", sc.host)
		return
	}
	if err == ErrGotSignal {
		DebugPrint("received signal, terminated")
		return
	}
	switch e := err.(type) {
	case *ssh.ExitMissingError:
		DebugPrint("ssh connetion to %s but remote didn't send exit status: %s", sc.host, e)
	case *ssh.ExitError:
		DebugPrint("ssh connetion to %s with error: %s", sc.host, err)
	default:
		DebugPrint("ssh connetion to %s with unknown error: %s", sc.host, err)
	}
}

// SendEnv todo
func (sc *SSHClient) SendEnv() error {
	if len(sc.env) == 0 {
		return nil
	}
	for k, v := range sc.env {
		DebugPrint("SetEnv %s=%s", k, v)
		if err := sc.sess.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

// WatchSignals todo
func (sc *SSHClient) WatchSignals() chan os.Signal {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	return sigC
}

type windowChangeReq struct {
	W, H, Wpx, Hpx uint32
}

func (sc *SSHClient) invokeResizeTerminal(ctx context.Context) {
	ch := sc.watchTerminalResize(ctx)
	sc.wg.Add(1)
	go func() {
		defer sc.wg.Done()
		w, h, err := pty.GetWinSize()
		if err != nil {
			DebugPrint("failed get windows size %v", err)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
			}
			nw, nh, err := pty.GetWinSize()
			if err != nil {
				continue
			}
			if nw == w && nh == h {
				continue
			}
			_, err = sc.sess.SendRequest("window-change", false, ssh.Marshal(
				windowChangeReq{W: uint32(nw), H: uint32(nh)},
			))
			if err != nil {
				continue
			}
			w = nw
			h = nh
		}
	}()
}

// RunInteractive to open a shell
func (sc *SSHClient) RunInteractive() error {
	DebugPrint("ssh shell mode. host: %s", sc.host)
	sc.sess.Stdout = os.Stdout
	sc.sess.Stderr = os.Stderr
	sc.sess.Stdin = os.Stdin
	sc.sys = &sysInfo{}
	if sc.mode == TerminalModeForce {
		x, y, err := pty.GetWinSize()
		if err != nil {
			return err
		}
		if termstate, err := terminal.MakeRaw(int(os.Stdout.Fd())); err == nil {
			defer terminal.Restore(int(os.Stdout.Fd()), termstate)
		}
		if termstate, err := terminal.MakeRaw(int(os.Stderr.Fd())); err == nil {
			defer terminal.Restore(int(os.Stderr.Fd()), termstate)
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
	if err := sc.changeLocalTerminalMode(); err != nil {
		return err
	}
	sigC := sc.WatchSignals()
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		signal.Stop(sigC)
		cancel()
		sc.wg.Wait()
	}()
	sc.invokeResizeTerminal(ctx)
	sessC := make(chan error)
	go func() {
		sessC <- sc.sess.Wait()
	}()
	select {
	case <-sigC:
		return ErrGotSignal
	case err := <-sessC:
		return err
	}
}

// Loop todo
func (sc *SSHClient) Loop() error {
	_ = sc.SendEnv()
	if len(sc.argv) == 0 {
		return sc.RunInteractive()
	}
	sc.sess.Stdout = os.Stdout
	sc.sess.Stderr = os.Stderr
	sc.sess.Stdin = os.Stdin
	// git escape argv done
	args := strings.Join(sc.argv, " ")
	DebugPrint("Exec cmd: %s", args)
	sigC := sc.WatchSignals()
	defer func() {
		signal.Stop(sigC)
	}()
	sessC := make(chan error)
	go func() {
		sessC <- sc.sess.Run(args)
	}()
	select {
	case <-sigC:
		return ErrGotSignal
	case err := <-sessC:
		return err
	}
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
