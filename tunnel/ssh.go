package tunnel

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/balibuild/tunnelssh/cli"
	"github.com/balibuild/tunnelssh/external/sshconfig"
	"golang.org/x/crypto/ssh"
)

// IsDebugMode todo
var IsDebugMode bool

// DebugLevel todo
var DebugLevel int

// DebugPrint todo
func DebugPrint(format string, a ...interface{}) {
	if IsDebugMode || DebugLevel == 3 {
		ss := fmt.Sprintf(format, a...)
		_, _ = os.Stderr.WriteString(cli.StrCat("debug3: \x1b[33m", ss, "\x1b[0m\n"))
	}
}

// DebugPrintN todo
func DebugPrintN(level int, format string, a ...interface{}) {
	if level >= DebugLevel {
		ss := fmt.Sprintf(format, a...)
		ns := strconv.Itoa(level)
		_, _ = os.Stderr.WriteString(cli.StrCat("debug", ns, ": \x1b[33m", ss, "\x1b[0m\n"))
	}
}

type sshconn struct {
	client *ssh.Client
	chcon  net.Conn
	host   string
}

// Read reads data from the connection.
func (conn *sshconn) Read(b []byte) (int, error) {
	return conn.chcon.Read(b)
}

// Write writes data
func (conn *sshconn) Write(b []byte) (int, error) {
	return conn.chcon.Write(b)
}

// Close closes the connection.
func (conn *sshconn) Close() error {
	if conn.chcon != nil {
		_ = conn.chcon.Close()
	}
	return conn.client.Close()
}

// LocalAddr returns the local network address.
func (conn *sshconn) LocalAddr() net.Addr {
	return conn.client.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (conn *sshconn) RemoteAddr() net.Addr {
	return conn.client.RemoteAddr()
}

// SetDeadline wapper
func (conn *sshconn) SetDeadline(t time.Time) error {
	return conn.chcon.SetDeadline(t)
}

// SetReadDeadline wapper
func (conn *sshconn) SetReadDeadline(t time.Time) error {
	return conn.chcon.SetReadDeadline(t)
}

// SetWriteDeadline wapper
func (conn *sshconn) SetWriteDeadline(t time.Time) error {
	return conn.chcon.SetWriteDeadline(t)
}

// HomePath todo
func HomePath() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}

// PathConvert todo
func PathConvert(p string) string {
	if p == "~" {
		return HomePath()
	}
	if !strings.HasPrefix(p, "~/") && !strings.HasPrefix(p, "~\\") {
		return os.ExpandEnv(p)
	}
	return filepath.Join(HomePath(), p[2:])
}

func readFromPrivateKey(kf string) (ssh.Signer, error) {
	fd, err := os.Open(kf)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(buf)
}

// publicKeys todo
func (conn *sshconn) publicKeys() ([]ssh.Signer, error) {
	if identityFile := sshconfig.Get(conn.host, "IdentityFile"); len(identityFile) != 0 {
		sig, err := readFromPrivateKey(PathConvert(identityFile))
		if err != nil {
			return nil, errors.New("not found host matched keys")
		}
		return []ssh.Signer{sig}, nil
	}
	home := HomePath()
	keys := []string{"id_ed25519", "id_ecdsa", "id_rsa"} // keys
	signers := make([]ssh.Signer, 0, len(keys))
	for _, k := range keys {
		sig, err := readFromPrivateKey(filepath.Join(home, ".ssh", k))
		if err == nil {
			signers = append(signers, sig)
		}
	}
	return signers, nil
}

// DialTunnelSSH dial ssh tunnel (ssh over ssh)
func (bm *BoringMachine) DialTunnelSSH(u *url.URL, paddr, addr string, timeout time.Duration) (net.Conn, error) {
	config := &ssh.ClientConfig{Timeout: timeout}
	if u.User != nil {
		config.User = u.User.Username()
	} else {
		if config.User = sshconfig.Get(u.Host, "User"); len(config.User) == 0 {
			current, err := user.Current()
			if err != nil {
				return nil, err
			}
			config.User = current.Name
		}
	}
	conn := &sshconn{host: u.Host}
	config.Auth = append(config.Auth, ssh.PublicKeysCallback(conn.publicKeys))
	var err error
	if conn.client, err = ssh.Dial("tcp", paddr, config); err != nil {
		return nil, err
	}
	if conn.chcon, err = conn.client.Dial("tcp", addr); err != nil {
		_ = conn.Close()
		return nil, err
	}
	bm.DebugPrint("Establish connection to proxy(%s): %s", u.Scheme, paddr)
	return conn, nil
}
