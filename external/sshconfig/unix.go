// +build !windows

package sshconfig

import (
	"os"
	"path/filepath"
)

// /etc/ssh/ssh_config
func systemConfigFinder() string {
	return filepath.Join("/", "etc", "ssh", "ssh_config")
}

func systemdir() string {
	return os.ExpandEnv("/etc/ssh")
}
