package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/balibuild/tunnelssh/cli"
	"github.com/balibuild/tunnelssh/pty"
	"github.com/balibuild/tunnelssh/tunnel"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

var defaultKnownhosts = tunnel.PathConvert("~/.ssh/known_hosts")

func keyTypeName(key ssh.PublicKey) string {
	kt := key.Type()
	switch kt {
	case "ssh-rsa":
		return "RSA"
	case "ssh-dss":
		return "DSA"
	case "ssh-ed25519":
		return "ED25519"
	default:
		if strings.HasPrefix(kt, "ecdsa-sha2-") {
			return "ECDSA"
		}
	}
	return kt
}

func askAddingUnknownHostKey(address string, remote net.Addr, key ssh.PublicKey) (bool, error) {
	stopC := make(chan struct{})
	defer func() {
		close(stopC)
	}()

	go func() {
		sigC := make(chan os.Signal, 1)
		signal.Notify(sigC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case <-sigC:
			os.Exit(1)
		case <-stopC:
		}
	}()
	msg := fmt.Sprintf("The authenticity of host '%s (%s)' can't be established.\n%s key fingerprint is %s\nAre you sure you want to continue connecting (yes/no)? ",
		address, remote.String(),
		keyTypeName(key),
		ssh.FingerprintSHA256(key))
	if pty.IsTerminal(os.Stdin) {
		_, _ = os.Stderr.WriteString(msg)
		b := bufio.NewReader(os.Stdin)
		for {
			answer, err := b.ReadString('\n')
			if err != nil {
				return false, fmt.Errorf("failed to read answer: %s", err)
			}
			answer = strings.ToLower(strings.TrimSpace(answer))
			if answer == "yes" {
				return true, nil
			}
			if answer == "no" {
				return false, nil
			}
			fmt.Print("Please type 'yes' or 'no': ")
		}
	}
	answer, err := readAskPass(msg, "", false)
	if err != nil {
		return false, err
	}
	if strings.ToLower(answer) == "yes" {
		return true, nil
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
func (sc *SSHClient) HostKeyCallback(hostname string, remote net.Addr, key ssh.PublicKey) error {
	DebugPrint("Server %s host key: %s %s", hostname, keyTypeName(key), ssh.FingerprintSHA256(key))
	if _, err := os.Stat(defaultKnownhosts); err == nil {
		DebugPrint("Found %s", defaultKnownhosts)
		hostKeyCallback, err := knownhosts.New(defaultKnownhosts)
		if err != nil {
			return cli.ErrorCat("failed to load knownhosts files: %s", err.Error())
		}
		err = hostKeyCallback(hostname, remote, key)
		if err == nil {
			return nil
		}
		keyErr, ok := err.(*knownhosts.KeyError)
		if !ok || len(keyErr.Want) > 0 {
			DebugPrint("Verify KnownHosts %v", err)
			return err
		}
	} else if !os.IsNotExist(err) {
		// if not exists
		return err
	}
	if answer, err := askAddingUnknownHostKey(hostname, remote, key); err != nil || !answer {
		msg := "host key verification failed"
		if err != nil {
			msg = cli.StrCat(msg, ": ", err.Error())
		}
		return errors.New(msg)
	}
	f, err := os.OpenFile(defaultKnownhosts, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to add new host key: %s", err)
	}
	defer f.Close()

	var addrs []string
	if remote.String() == hostname {
		addrs = []string{hostname}
	} else {
		addrs = []string{hostname, remote.String()}
	}

	entry := knownhosts.Line(addrs, key)
	if _, err = f.WriteString(entry + "\n"); err != nil {
		return fmt.Errorf("failed to add new host key: %s", err)
	}
	return nil
}

func (sc *SSHClient) openPrivateKey(kf string) (ssh.Signer, error) {
	fd, err := os.Open(kf)
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
	DebugPrint("Offering public key: %s %s", kf, ssh.FingerprintSHA256(key))
	return sig, nil
}

// SearchKey todo
func (sc *SSHClient) SearchKey(name string) (ssh.Signer, error) {
	file := filepath.Join(sc.home, ".ssh", name)
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			DebugPrint("Trying private key: %s: no such identity", file)
		} else {
			DebugPrint("%v", err)
		}
		return nil, err
	}
	return sc.openPrivateKey(file)
}

// PublicKeys todo
func (sc *SSHClient) PublicKeys() ([]ssh.Signer, error) {
	if len(sc.IdentityFile) != 0 {
		sig, err := sc.openPrivateKey(tunnel.PathConvert(sc.IdentityFile))
		if err != nil {
			return nil, errors.New("not found host matched keys")
		}
		return []ssh.Signer{sig}, nil
	}
	// We drop id_dsa key support
	// http://www.openssh.com/txt/release-6.5
	keys := []string{"id_ed25519", "id_ecdsa", "id_rsa"} // keys
	signers := make([]ssh.Signer, 0, len(keys))
	for _, k := range keys {
		sig, err := sc.SearchKey(k)
		if err == nil {
			signers = append(signers, sig)
		}
	}
	return signers, nil
}
