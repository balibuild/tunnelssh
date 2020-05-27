package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/balibuild/tunnelssh/cli"
)

// GIT_SSH=xxx

// version info
var (
	VERSION     = "1.0"
	BUILDTIME   string
	BUILDCOMMIT string
	BUILDBRANCH string
	GOVERSION   string
)

func version() {
	fmt.Fprint(os.Stdout, "git-tunnel - Tunnel SSH git command wapper\nversion:       ", VERSION, "\n",
		"build branch:  ", BUILDBRANCH, "\n",
		"build commit:  ", BUILDCOMMIT, "\n",
		"build time:    ", BUILDTIME, "\n",
		"go version:    ", GOVERSION, "\n")
}

// -o option
// Can be used to give options in the format used in the configuration file.
// -4' Forces ssh to use IPv4 addresses only.
// -6' Forces ssh to use IPv6 addresses only.
//-p port Port to connect to on the remote host. This can be specified on a per-host basis in the configuration file.
func usage() {
	fmt.Fprintf(os.Stdout, `git-tunnel - Tunnel SSH git command wapper
usage: %s <option> command args...
  -h|--help        Show usage text and quit
  -v|--version     Show version number and quit
  -V|--verbose     Make the operation more talkative

example
  git-tunnel clone git@github.com:git/git.git

`, os.Args[0])
}

type option struct {
}

func (o *option) Invoke(val int, oa, raw string) error {
	switch val {
	case 'h':
		usage()
		os.Exit(0)
	case 'v':
		version()
		os.Exit(0)
	case 'V':
		os.Setenv("TUNNEL_DEBUG", "YES")
	default:
	}
	return nil
}

// ParseArgv todo
func ParseArgv() ([]string, error) {
	var o option
	ae := &cli.ArgvParser{SubcmdMode: true}
	ae.Add("help", cli.NOARG, 'h')
	ae.Add("version", cli.NOARG, 'v')
	ae.Add("verbose", cli.NOARG, 'V')
	if err := ae.Execute(os.Args, &o); err != nil {
		return nil, err
	}
	if len(ae.Unresolved()) == 0 {
		usage()
		os.Exit(1)
	}
	return ae.Unresolved(), nil
}

// git-tunel
func main() {
	args, err := ParseArgv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ParseArgv: %v\n", err)
		usage()
		os.Exit(1)
	}
	if err := InitializeEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "\x1b[31minitialize env error: %s\x1b[0m\n", err)
		os.Exit(1)
	}
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if cmd.ProcessState.Exited() {
			os.Exit(cmd.ProcessState.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
