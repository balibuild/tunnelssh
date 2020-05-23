// +build windows

package sshconfig

import (
	"os"
)

// C:\ProgramData\ssh\ssh_config
func systemConfigFinder() string {
	return os.ExpandEnv("$ProgramData\\ssh\\ssh_config")
}
