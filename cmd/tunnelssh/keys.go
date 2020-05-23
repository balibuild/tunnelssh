package main

import (
	"errors"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/balibuild/tunnelssh/cli"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// search keys

func khnormalize(addr net.Addr, k ssh.PublicKey) string {
	return cli.StrCat(knownhosts.Normalize(addr.String()), " ", k.Type(), " ", string(k.Marshal()), "\n")
}

var defaultKnownhost = DefaultKnownHosts()

func addKnownhost(host string, addr net.Addr, k ssh.PublicKey, knownfile string) error {
	if len(knownfile) == 0 {
		knownfile = defaultKnownhost
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
		knownfile = defaultKnownhost
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

// KeySearcher tod
type KeySearcher struct {
	home string
}

// Search todo
func (ks *KeySearcher) Search(name string) (ssh.Signer, error) {
	file := filepath.Join(ks.home, ".ssh", name)
	fd, err := os.Open(file)
	if err != nil {
		DebugPrint("%s %v", file, err)
		return nil, err
	}
	defer fd.Close()
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(buf)
}

// MatchPublicKeys todo
func (c *client) MatchPublicKeys() ([]ssh.Signer, error) {
	return nil, errors.New("not found host matched keys")
}

// PublicKeys todo
func (c *client) PublicKeys() ([]ssh.Signer, error) {
	var ks KeySearcher
	ks.Initialize()
	if sigs, err := c.MatchPublicKeys(); err == nil {
		return sigs, nil
	}
	keys := []string{"id_ed25519", "id_ecdsa", "id_rsa"} // keys
	signers := make([]ssh.Signer, 0, len(keys))
	for _, k := range keys {
		sig, err := ks.Search(k)
		if err == nil {
			key := sig.PublicKey()
			DebugPrint("%s: %s", k, ssh.FingerprintSHA256(key))
			signers = append(signers, sig)
		}
	}
	return signers, nil
}
