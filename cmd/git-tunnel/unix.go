// +build !windows

package main

// define
const (
	SSHEnv = "GIT_SSH=tunnelssh"
	SSHVARIANT="GIT_SSH_VARIANT=ssh"
)

func InitializeEnv() error {

	return nil
}
