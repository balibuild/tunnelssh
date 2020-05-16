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

func main() {
	//
}
