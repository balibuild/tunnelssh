package main

import (
	"errors"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/balibuild/tunnelssh/cli"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// PathConvert todo
func PathConvert(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("USERPROFILE"), p[1:])
	}
	return filepath.Join(os.Getenv("HOME"), p[1:])
}

// HomeDir todo
func HomeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}

// search keys

func khnormalize(addr net.Addr, k ssh.PublicKey) string {
	return cli.StrCat(knownhosts.Normalize(addr.String()), " ", k.Type(), " ", string(k.Marshal()), "\n")
}

var defaultKnownhost = PathConvert("~/.ssh/known_hosts")

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
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			DebugPrint("Trying private key: %s: no such identity", file)
		} else {
			DebugPrint("%v", err)
		}
		return nil, err
	}
	fd, err := os.Open(file)
	if err != nil {
		DebugPrint("%v", err)
		return nil, err
	}
	defer fd.Close()
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		DebugPrint("%v", err)
		return nil, err
	}
	sig, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		DebugPrint("%v", err)
		return nil, err
	}
	key := sig.PublicKey()
	DebugPrint("Offering public key: %s %s", file, ssh.FingerprintSHA256(key))
	return sig, nil
}

// MatchPublicKeys todo
func (c *client) MatchPublicKeys() ([]ssh.Signer, error) {
	return nil, errors.New("not found host matched keys")
}

// PublicKeys todo
func (c *client) PublicKeys() ([]ssh.Signer, error) {

	if sigs, err := c.MatchPublicKeys(); err == nil {
		return sigs, nil
	}
	ks := KeySearcher{home: HomeDir()}
	// We drop id_dsa key support
	// http://www.openssh.com/txt/release-6.5
	keys := []string{"id_ed25519", "id_ecdsa", "id_rsa"} // keys
	signers := make([]ssh.Signer, 0, len(keys))
	for _, k := range keys {
		sig, err := ks.Search(k)
		if err == nil {
			signers = append(signers, sig)
		}
	}
	return signers, nil
}
