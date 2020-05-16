// +build !windows

package main

// define
const (
	SSHEnv = "GIT_SSH=tunnelssh"
)

func InitializeEnv() error {

	return nil
}
