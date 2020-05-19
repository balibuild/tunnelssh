package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	fd := int(os.Stdin.Fd())
	xterm := os.Getenv("TERM")
	fmt.Fprintf(os.Stderr, "IsTerminal: %v %s\n", terminal.IsTerminal(fd), xterm)
	fmt.Fprintf(os.Stderr, "IsCygwinTerminal %v\n", isatty.IsCygwinTerminal(os.Stdin.Fd()))

}
