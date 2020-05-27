package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

func main() {
	msg := fmt.Sprintf(
		"The authenticity of host '%s (%s)' can't be established.\r\n%s key fingerprint is %s\r\nAre you sure you want to continue connecting (yes/no)? ",
		"github.com",
		"51.142.56.2",
		"RSA",
		"SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8",
	)
	var cmd *exec.Cmd
	if len(os.Args) < 2 {
		cmd = exec.Command("ssh-askpass", msg)
	} else {
		cmd = exec.Command(os.Args[1], msg)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "StdoutPipe: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()
	// like ssh-askpass-sublime.exe is GUI subsystem. so
	// ssh-askpass-sublime.exe --> call sublime-merge.exe
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(os.Stdout, out)
	}()
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
	wg.Wait()
}
