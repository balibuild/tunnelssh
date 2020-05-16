// +build windows

package main

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

// HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings

// ResolveRegistryProxy todo
func ResolveRegistryProxy() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()
	if d, _, err := k.GetIntegerValue("ProxyEnable"); err == nil && d == 1 {
		if s, _, err := k.GetStringValue("ProxyServer"); err == nil && len(s) > 0 {
			return s, nil
		}
	}
	if s, _, err := k.GetStringValue("AutoConfigURL"); err == nil && len(s) > 0 {
		return s, nil
	}
	return "", ErrProxyNotConfigured
}

// feature read proxy from registry

// ResolveProxy todo
func ResolveProxy() (string, error) {
	if s, err := ResolveRegistryProxy(); err == nil {
		return s, nil
	}
	if s := os.Getenv("SSH_PROXY"); len(s) > 0 {
		return s, nil
	}
	if s := os.Getenv("HTTPS_PROXY"); len(s) > 0 {
		return s, nil
	}
	if s := os.Getenv("HTTP_PROXY"); len(s) > 0 {
		return s, nil
	}
	return "", ErrProxyNotConfigured
}
