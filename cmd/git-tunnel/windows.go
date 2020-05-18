// +build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/balibuild/tunnelssh/cli"
	"golang.org/x/sys/windows/registry"
)

//InitializeGW todo
func InitializeGW() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\GitForWindows`, registry.QUERY_VALUE)
	if err != nil {
		if k, err = registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\GitForWindows`, registry.QUERY_VALUE); err != nil {
			return "", err
		}
	}
	defer k.Close()
	installPath, _, err := k.GetStringValue("InstallPath")
	if err != nil {
		return "", err
	}
	installPath = filepath.Clean(installPath)
	git := filepath.Join(installPath, "cmd\\git.exe")
	if _, err := os.Stat(git); err != nil {
		return "", err
	}
	return filepath.Join(installPath, "cmd"), nil
}

// InitializeEnv todo
func InitializeEnv() error {
	var gitbin string
	if _, err := exec.LookPath("git"); err != nil {
		if gitbin, err = InitializeGW(); err != nil {
			return cli.ErrorCat("git not installed: ", err.Error())
		}
	}
	p := os.Getenv("PATH")
	pv := strings.Split(p, ";")
	pvv := make([]string, 0, len(pv)+2)
	if len(gitbin) != 0 {
		pvv = append(pvv, filepath.Clean(gitbin))
	}
	for _, s := range pv {
		if len(s) == 0 {
			continue
		}
		pvv = append(pvv, filepath.Clean(s))
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exebin := filepath.Dir(exe)
	if _, err = exec.LookPath("tunnelssh.exe"); err != nil {
		tunnelssh := filepath.Join(exebin, "tunnelssh.exe")
		if _, err := os.Stat(tunnelssh); err != nil {
			return err
		}
		pvv = append(pvv, exebin)
	}
	os.Setenv("PATH", strings.Join(pvv, ";"))
	os.Setenv("GIT_SSH", "tunnelssh.exe")
	os.Setenv("GIT_SSH_VARIANT", "ssh")
	return nil
}
