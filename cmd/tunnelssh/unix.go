// +build !windows

package main

import (
	"net"
	"os"

	"github.com/balibuild/tunnelssh/cli"
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
