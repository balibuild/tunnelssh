// +build !windows

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/balibuild/tunnelssh/cli"
	"golang.org/x/crypto/ssh/terminal"
)

// MakeAgent make agent
func (ka *KeyAgent) MakeAgent() error {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if len(sock) == 0 {
		return cli.ErrorCat("ssh agent not initialized")
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return err
	}
	ka.conn = conn
	return nil
}

type sysInfo struct {
	origMode *terminal.State
}

func (sc *SSHClient) changeLocalTerminalMode() (err error) {
	if sc.sys.origMode, err = terminal.MakeRaw(int(os.Stdin.Fd())); err != nil {
		return fmt.Errorf("failed to set stdin to raw mode: %s", err)
	}

	return nil
}

func (sc *SSHClient) restoreLocalTerminalMode() error {
	if sc.sys.origMode != nil {
		return terminal.Restore(int(os.Stdin.Fd()), sc.sys.origMode)
	}
	return nil
}

func (sc *SSHClient) watchTerminalResize(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGWINCH)

	sc.wg.Add(1)
	go func() {
		defer func() {
			signal.Reset(syscall.SIGWINCH)
			signal.Stop(sigC)
			close(ch)
			sc.wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-sigC:
				ch <- struct{}{}
			}
		}
	}()

	return ch
}
