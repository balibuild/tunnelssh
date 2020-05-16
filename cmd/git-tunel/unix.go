// +build !windows

package main

// define
const (
	SSHEnv = "GIT_SSH=tunelssh"
)

func InitializeEnv() error {

	return nil
}
