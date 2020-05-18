package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/balibuild/tunnelssh/cli"
	"golang.org/x/crypto/ssh"
)

// IsDebugMode todo
var IsDebugMode bool

// DebugPrint todo
func DebugPrint(format string, a ...interface{}) {
	if IsDebugMode {
		ss := fmt.Sprintf(format, a...)
		_, _ = os.Stderr.WriteString(cli.StrCat("\x1b[33m* ", ss, "\x1b[0m\n"))
	}
}

// version info
var (
	VERSION     = "1.0"
	BUILDTIME   string
	BUILDCOMMIT string
	BUILDBRANCH string
	GOVERSION   string
)

func version() {
	fmt.Fprint(os.Stdout, "tunnelssh - A witty ssh client that automatically accesses a remote server through a proxy\nversion:       ", VERSION, "\n",
		"build branch:  ", BUILDBRANCH, "\n",
		"build commit:  ", BUILDCOMMIT, "\n",
		"build time:    ", BUILDTIME, "\n",
		"go version:    ", GOVERSION, "\n")
}

// -o option
// Can be used to give options in the format used in the configuration file. This is useful for specifying options for which there is no separate command-line flag. For full details of the options listed below, and their possible values, see ssh_config(5).

// AddressFamily
// BatchMode
// BindAddress
// ChallengeResponseAuthentication
// CheckHostIP
// Cipher
// Ciphers
// ClearAllForwardings
// Compression
// CompressionLevel
// ConnectionAttempts
// ConnectTimeout
// ControlMaster
// ControlPath
// DynamicForward
// EscapeChar
// ExitOnForwardFailure
// ForwardAgent
// ForwardX11
// ForwardX11Trusted
// GatewayPorts
// GlobalKnownHostsFile
// GSSAPIAuthentication
// GSSAPIDelegateCredentials
// HashKnownHosts
// Host'
// HostbasedAuthentication
// HostKeyAlgorithms
// HostKeyAlias
// HostName
// IdentityFile
// IdentitiesOnly
// KbdInteractiveDevices
// LocalCommand
// LocalForward
// LogLevel
// MACs'
// NoHostAuthenticationForLocalhost
// NumberOfPasswordPrompts
// PasswordAuthentication
// PermitLocalCommand
// Port'
// PreferredAuthentications
// Protocol
// ProxyCommand
// PubkeyAuthentication
// RekeyLimit
// RemoteForward
// RhostsRSAAuthentication
// RSAAuthentication
// SendEnv
// ServerAliveInterval
// ServerAliveCountMax
// SmartcardDevice
// StrictHostKeyChecking
// TCPKeepAlive
// Tunnel
// TunnelDevice
// UsePrivilegedPort
// User'
// UserKnownHostsFile
// VerifyHostKeyDNS
// VisualHostKey
// XAuthLocation

// -4' Forces ssh to use IPv4 addresses only.
// -6' Forces ssh to use IPv6 addresses only.
//-p port Port to connect to on the remote host. This can be specified on a per-host basis in the configuration file.
func usage() {
	fmt.Fprintf(os.Stdout, `tunnelssh - A witty ssh client that automatically accesses a remote server through a proxy
usage: %s <option> args ...
  -h|--help        Show usage text and quit
  -v|--version     Show version number and quit
  -V|--verbose     Make the operation more talkative
  -p|--port        Port to connect to on the remote host.
  -o|--option      Only SetEnv is supported
  -T               Disable pseudo-tty allocation.
  -t               Force pseudo-tty allocation.
  -4               Forces ssh to use IPv4 addresses only.
  -6               Forces ssh to use IPv6 addresses only.
`, os.Args[0])
}

// https://github.com/git/git/blob/e870325ee8575d5c3d7afe0ba2c9be072c692b65/connect.c#L1113
// SetEnv

// ParseOption todo
// argv_array_push(args, "SendEnv=" GIT_PROTOCOL_ENVIRONMENT);
// argv_array_pushf(env, GIT_PROTOCOL_ENVIRONMENT "=version=%d",
// 		 version);
func (c *client) ParseOption(option string) bool {
	if strings.HasPrefix(option, "SendEnv=") {
		key := strings.TrimPrefix(option, "SendEnv=")
		val := os.Getenv(key)
		c.env[key] = val
		return true
	}
	return true
}

func (c *client) Invoke(val int, oa, raw string) error {
	switch val {
	case 'h':
		usage()
		os.Exit(0)
	case 'v':
		version()
		os.Exit(0)
	case 'V':
		IsDebugMode = true
	case 'p':
		p, err := strconv.Atoi(oa)
		if err != nil {
			return cli.ErrorCat("invaild port number: ", oa)
		}
		c.port = p
	case 'o':
		if !c.ParseOption(oa) {
			return cli.ErrorCat("option not support '", oa, "'")
		}
	case 'T':
		c.forcenotty = true
	case 't':
		c.forcetty = true
	case '4':
	case '6':
	default:
	}
	return nil
}

// SplitHost todo
// git@xxxx
// xxxx
func (c *client) SplitHost(sshaddr string) error {
	if pos := strings.IndexByte(sshaddr, '@'); pos != -1 {
		c.config.User = sshaddr[0:pos]
		c.host = sshaddr[pos+1:]
		return nil
	}
	c.host = sshaddr
	u, err := user.Current()
	if err != nil {
		return err
	}
	c.config.User = u.Name
	return nil
}

func (c *client) ParseArgv() error {
	c.env = make(map[string]string)
	var ae cli.ArgvParser
	ae.Add("help", cli.NOARG, 'h')
	ae.Add("version", cli.NOARG, 'v')
	ae.Add("verbose", cli.NOARG, 'V')
	ae.Add("port", cli.REQUIRED, 'p')
	ae.Add("option", cli.REQUIRED, 'o')
	ae.Add("no-tty", cli.OPTIONAL, 'T') // default no tty
	ae.Add("force-tty", cli.OPTIONAL, 't')
	ae.Add("ipv4", cli.NOARG, '4')
	ae.Add("ipv6", cli.NOARG, '6')
	if cli.IsTrue(os.Getenv("TUNNEL_DEBUG")) {
		IsDebugMode = true
	}
	c.port = 0
	if err := ae.Execute(os.Args, c); err != nil {
		return err
	}
	if len(ae.Unresolved()) == 0 {
		usage()
		os.Exit(1)
	}
	// not support dsa
	//HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	c.config = &ssh.ClientConfig{
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoRSA,
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoSKECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoED25519,
			ssh.KeyAlgoSKED25519,
		},
	}
	if err := c.SplitHost(ae.Unresolved()[0]); err != nil {
		return cli.ErrorCat("SplitHost: ", err.Error())
	}
	c.argv = ae.Unresolved()[1:]
	c.config.Auth = append(c.config.Auth, ssh.PasswordCallback(sshPasswordPrompt))
	c.ka = &KeyAgent{}
	if c.ka.MakeAgent() == nil {
		c.config.Auth = append(c.config.Auth, c.ka.UseAgent())
	} else {
		c.config.Auth = append(c.config.Auth, ssh.PublicKeysCallback(c.PublicKeys))
	}

	return nil
}

func main() {
	c := &client{}
	if err := c.ParseArgv(); err != nil {

	}
}
