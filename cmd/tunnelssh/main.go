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
func usage() {
	fmt.Fprintf(os.Stdout, `tunnelssh - A witty ssh client that automatically accesses a remote server through a proxy
usage: %s <option> args ...
  -h|--help        Show usage text and quit
  -v|--version     Show version number and quit
`, os.Args[0])
}

// https://github.com/git/git/blob/e870325ee8575d5c3d7afe0ba2c9be072c692b65/connect.c#L1113
// SetEnv

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
	case 'T':
		c.forcenotty = true
	case 't':
		c.forcetty = true
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
	var ae cli.ArgvParser
	ae.Add("help", cli.NOARG, 'h')
	ae.Add("version", cli.NOARG, 'v')
	ae.Add("verbose", cli.NOARG, 'V')
	ae.Add("port", cli.REQUIRED, 'p')
	ae.Add("no-tty", cli.OPTIONAL, 'T') // default no tty
	ae.Add("force-tty", cli.OPTIONAL, 't')
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
	c.config = &ssh.ClientConfig{}
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
	//
}
