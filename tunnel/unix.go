// +build !windows

package tunnel

// ResolveProxy todo
func ResolveProxy() (*ProxySettings, error) {
	ps := &ProxySettings{sep: ","}
	ps.ProxyOverride = getEnvAny("NO_PROXY", "no_proxy")
	if ps.ProxyServer = getEnvAny("SSH_PROXY", "ssh_proxy"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	if ps.ProxyServer = getEnvAny("HTTPS_PROXY", "https_proxy"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	if ps.ProxyServer = getEnvAny("HTTP_PROXY", "http_proxy"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	if ps.ProxyServer = getEnvAny("ALL_PROXY", "all_proxy"); len(ps.ProxyServer) > 0 {
		return ps, nil
	}
	return nil, ErrProxyNotConfigured
}
