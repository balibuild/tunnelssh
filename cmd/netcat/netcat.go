package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/balibuild/tunnelssh/cli"
	"github.com/balibuild/tunnelssh/tunnel"
)

// IsDebugMode todo
var IsDebugMode bool

func main() {
	if len(os.Args) < 2 {
		name := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "usage: %s host port\n       %s host:port", name, name)
		os.Exit(1)
	}
	if cli.IsTrue(os.Getenv("TUNNEL_DEBUG")) {
		IsDebugMode = true
	}
	var bm tunnel.BoringMachine
	if IsDebugMode {
		bm.Debug = func(msg string) {
			_, _ = os.Stderr.WriteString(cli.StrCat("\x1b[33m", msg, "\x1b[0m\n"))
		}
	}
	var address string
	if len(os.Args) == 2 {
		address = os.Args[1]
	} else {
		address = net.JoinHostPort(os.Args[1], os.Args[2])
	}

	bm.DebugPrint("Use netcat to connect: %s", address)
	_ = bm.Initialize()
	conn, err := bm.Dial("tcp", address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable dial %s %v\n", address, err)
		os.Exit(1)
	}
	bm.DebugPrint("Address: %s reomote: %s", address, conn.RemoteAddr().String())
	defer conn.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		io.Copy(conn, os.Stdin)
		wg.Done()
	}()
	go func() {
		io.Copy(os.Stdout, conn)
		wg.Done()
	}()
	wg.Wait()
}
