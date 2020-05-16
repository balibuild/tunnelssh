// +build !windows

package main

import "os"

// ResolveProxy todo
func ResolveProxy() (string, error) {
	if s := os.Getenv("SSH_PROXY"); len(s) > 0 {
		return s, nil
	}
	if s := os.Getenv("HTTPS_PROXY"); len(s) > 0 {
		return s, nil
	}
	if s := os.Getenv("HTTP_PROXY"); len(s) > 0 {
		return s, nil
	}
	if s := os.Getenv("ALL_PROXY"); len(s) > 0 {
		return s, nil
	}
	return "", ErrProxyNotConfigured
}
