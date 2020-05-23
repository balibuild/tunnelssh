package main

import (
	"fmt"
	"os"

	"github.com/balibuild/tunnelssh/pty"
)

func main() {
	password, err := pty.ReadPassword("Password")
	if err == nil {
		fmt.Fprintf(os.Stderr, "\nPassword %s\n", password)
	}
	x, y, err := pty.GetWinSize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetWinSize: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "WinSize: %d %d\n", x, y)
}
