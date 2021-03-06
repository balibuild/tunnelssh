package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/balibuild/tunnelssh/cli"
)

type askPassOption struct {
	PasswordMode bool
	Args         []string
	User         string
}

func usage() {
	fmt.Fprintf(os.Stdout, `ssh-askpass - A witty ssh client that automatically accesses a remote server through a proxy
usage: %s <option> args ...
  -h|--help        Show usage text and quit
  -v|--version     Show version number and quit
  -V|--verbose     Make the operation more talkative
  -u|--user        UserName for password enter
  -p|--password    Prompt user to enter password. default Yes/No confirmation

`, os.Args[0])
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

// Invoke true
func (a *askPassOption) Invoke(val int, oa, raw string) error {
	switch val {
	case 'h':
		usage()
		os.Exit(0)
	case 'v':
		version()
		os.Exit(0)
	case 'u':
		a.User = oa
	case 'p':
		a.PasswordMode = true
	}
	return nil
}

func (a *askPassOption) ParseArgv() error {
	var ae cli.ParseArgs
	ae.Add("help", cli.NOARG, 'h')
	ae.Add("version", cli.NOARG, 'v')
	ae.Add("verbose", cli.NOARG, 'V')
	ae.Add("password", cli.NOARG, 'p')
	ae.Add("user", cli.REQUIRED, 'u')

	if err := ae.Execute(os.Args, a); err != nil {
		return err
	}
	a.Args = ae.Unresolved()
	if len(a.Args) == 0 {
		return errors.New("missing prompt")
	}
	if a.User == "" {
		a.User = "N/A"
	}
	return nil
}

func main() {
	var a askPassOption
	if err := a.ParseArgv(); err != nil {
		fmt.Fprintf(os.Stderr, "ParseArgv: %v\n", err)
		usage()
		os.Exit(1)
	}
	if a.PasswordMode {
		os.Exit(AskPassword(a.Args[0], a.User))
	}
	os.Exit(AskYes(a.Args[0], "TunnelSSH AskPass Yes/No confirm"))
}
