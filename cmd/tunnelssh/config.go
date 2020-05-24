package main

import (
	"os/user"
	"strconv"

	"github.com/balibuild/tunnelssh/external/sshconfig"
)

// InitializeHost todo
func (sc *SSHClient) InitializeHost() {
	host := sc.host
	if sc.IdentityFile = sshconfig.GetEx(host, "IdentityFile"); len(sc.IdentityFile) > 0 {
		DebugPrint("Host: %s IdentityFile %s", host, sc.IdentityFile)
	}
	if hostname := sshconfig.GetEx(host, "HostName"); len(hostname) > 0 {
		sc.host = hostname
		DebugPrint("Host: %s HostName %s", host, hostname)
	}
	if len(sc.config.User) == 0 {
		if user := sshconfig.GetEx(host, "User"); len(user) > 0 {
			sc.config.User = user
			DebugPrint("Host: %s User %s", host, user)
		}
	} else {
		if len(sc.config.User) == 0 {
			if u, err := user.Current(); err == nil {
				sc.config.User = u.Name
			} else {
				sc.config.User = "root"
			}
		}
	}
	// Rebind port
	if sc.port == 0 {
		if port := sshconfig.GetEx(host, "Port"); len(port) > 0 {
			if p, err := strconv.Atoi(port); err == nil {
				if p > 0 && p < 65535 {
					sc.port = p
				}
			}
		}
	}

}
