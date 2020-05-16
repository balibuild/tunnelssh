package main

import (
	"fmt"
	"os"

	"github.com/balibuild/tunnelssh/cli"
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

func (c *client) Invoke(val int, oa, raw string) error {
	switch val {
	case 'h':
	case 'v':
	case 'V':
	case 'p':
	case 'T':
	case 't':
	default:
	}
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
	return nil
}

func main() {
	//
}
