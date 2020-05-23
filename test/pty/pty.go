package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/balibuild/tunnelssh/pty"
)

// AskPrompt todo
func AskPrompt(prompt string) (string, error) {
	if pty.IsTerminal(os.Stdin) {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		respond, err := pty.ReadInputEx(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(respond), nil
	}
	return "", errors.New("unsupport")
}

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
	yesno, err := AskPrompt("Please say yes/no")
	if err != nil {
		fmt.Fprintf(os.Stderr, "yes/no error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "yes/no: %s\n", yesno)
}
