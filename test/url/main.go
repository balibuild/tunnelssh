package main

import (
	"fmt"
	"net/url"
	"os"
)

func main() {
	sv := []string{"https://127.0.0.1:1080", "127.0.0.1:10041", "http://github.com",
		"http://github.com:433",
		"http://[fe80::9560:d905:cc8f:26b4]:8080",
		"ssh://jack:passwd@github.com:433",
		"ssh://jack@github.com:433"}

	for _, s := range sv {
		u, err := url.Parse(s)
		if err == nil {
			fmt.Fprintf(os.Stderr, "Scheme: %s Host: %s Port: %s Hostname: %s\n", u.Scheme, u.Host, u.Port(), u.Hostname())
			if u.User != nil {
				pw, b := u.User.Password()
				fmt.Fprintf(os.Stderr, "UserInfo: %s [%s, %s, %v]\n", u.User.String(), u.User.Username(), pw, b)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
	}

}
