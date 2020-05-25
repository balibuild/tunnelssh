package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	msg := fmt.Sprintf(
		"The authenticity of host '%s (%s)' can't be established.\r\n%s key fingerprint is %s\r\nAre you sure you want to continue connecting (yes/no)? ",
		"github.com",
		"51.142.56.2",
		"RSA",
		"SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8",
	)
	cmd := exec.Command("ssh-askpass", msg)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Run()
}
