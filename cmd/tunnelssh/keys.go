package main

import (
	"errors"
	"net"
	"os"

	"github.com/balibuild/tunnelssh/cli"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// search keys

func khnormalize(addr net.Addr, k ssh.PublicKey) string {
	return cli.StrCat(knownhosts.Normalize(addr.String()), " ", k.Type(), " ", string(k.Marshal()), "\n")
}

func addKnownhost(host string, addr net.Addr, k ssh.PublicKey, knownfile string) error {
	if len(knownfile) == 0 {
		knownfile = os.ExpandEnv("$HOME/.ssh/known_hosts")
	}
	fd, err := os.OpenFile(knownfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer fd.Close()
	kh := khnormalize(addr, k)
	if _, err := fd.WriteString(kh); err != nil {
		return err
	}
	return nil
}

// checkKnownhost todo
func checkKnownhost(host string, addr net.Addr, k ssh.PublicKey, knownfile string) (bool, error) {
	if len(knownfile) == 0 {
		knownfile = os.ExpandEnv("$HOME/.ssh/known_hosts")
	}
	callback, err := knownhosts.New(knownfile)
	if err != nil {
		return false, err
	}
	err = callback(host, addr, k)
	if err == nil {
		return true, nil
	}
	var ke *knownhosts.KeyError
	if errors.As(err, &ke) && len(ke.Want) > 0 {
		return true, ke
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

// KeyAgent todo
type KeyAgent struct {
	conn net.Conn
}

// Close todo
func (ka *KeyAgent) Close() error {
	if ka.conn != nil {
		return ka.conn.Close()
	}
	return nil
}

// UseAgent todo
func (ka *KeyAgent) UseAgent() ssh.AuthMethod {
	return ssh.PublicKeysCallback(agent.NewClient(ka.conn).Signers)
}

//HostKeyCallback todo
func (c *client) HostKeyCallback(host string, remote net.Addr, key ssh.PublicKey) error {
	ke, err := checkKnownhost(host, remote, key, "")
	if ke {
		return err
	}
	// trusted check
	// end
	return addKnownhost(host, remote, key, "")
}

// PublicKeys todo
func (c *client) PublicKeys() ([]ssh.Signer, error) {
	keys := []string{"id_ed25519", "id_ecdsa", "id_rsa", "id_dsa"} // keys
	for _, k := range keys {
		DebugPrint("%s", k)
	}
	return nil, nil
}
