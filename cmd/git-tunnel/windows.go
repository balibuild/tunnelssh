// +build windows

package main

import (
	"os/exec"

	"github.com/balibuild/tunnelssh/cli"
	"golang.org/x/sys/windows/registry"
)

// define
const (
	SSHEnv = "GIT_SSH=tunnelssh.exe"
	SSHVARIANT="GIT_SSH_VARIANT=ssh"
)

//InitializeGW todo
func InitializeGW() error {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, ``, registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	return nil
}

// InitializeEnv todo
func InitializeEnv() error {
	if _, err := exec.LookPath("git"); err != nil {
		if err := InitializeGW(); err != nil {
			return cli.ErrorCat("git not installed: ", err.Error())
		}
	}

	return nil
}
