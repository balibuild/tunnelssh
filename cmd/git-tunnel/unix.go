// +build !windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func InitializeEnv() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exebin := filepath.Dir(exe)
	if _, err = exec.LookPath("tunnelssh"); err != nil {
		tunnelssh := filepath.Join(exebin, "tunnelssh.exe")
		if _, err := os.Stat(tunnelssh); err != nil {
			return err
		}
	} else {
		exebin = ""
	}
	if len(exebin) == 0 {
		return nil
	}
	p := os.Getenv("PATH")
	pv := strings.Split(p, ":")
	pvv := make([]string, 0, len(pv)+2)
	for _, s := range pv {
		if len(s) == 0 {
			continue
		}
		pvv = append(pvv, filepath.Clean(s))
	}
	if len(exebin) != 0 {
		pvv = append(pvv, exebin)
	}
	os.Setenv("PATH", strings.Join(pvv, ";"))
	os.Setenv("GIT_SSH", "tunnelssh")
	os.Setenv("GIT_SSH_VARIANT", "ssh")
	return nil
}
