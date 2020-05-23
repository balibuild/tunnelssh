// +build !windows

package sshconfig

import "path/filepath"

// /etc/ssh/ssh_config
func systemConfigFinder() string {
	return filepath.Join("/", "etc", "ssh", "ssh_config")
}
