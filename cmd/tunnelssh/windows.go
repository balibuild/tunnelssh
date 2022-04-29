//go:build windows
// +build windows

package main

import (
	"context"
	"os"
	"syscall"
	"time"

	winio "github.com/Microsoft/go-winio"
	"github.com/balibuild/tunnelssh/cli"

	"golang.org/x/sys/windows"
)

// const
const (
	EnableVirtualTerminalProcessingMode = 0x4
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode             = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode             = kernel32.NewProc("SetConsoleMode")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procReadConsoleInput           = kernel32.NewProc("ReadConsoleInputW")
	procAttachConsole              = kernel32.NewProc("AttachConsole")
	procAllocConsole               = kernel32.NewProc("AllocConsole")
)

func init() {
	var mode uint32
	// becasue we print message to stderr
	fd := os.Stderr.Fd()
	if windows.GetConsoleMode(windows.Handle(fd), &mode) == nil {
		_ = windows.SetConsoleMode(windows.Handle(fd), mode|EnableVirtualTerminalProcessingMode)
	}
}

// MakeAgent make agent
// Windows use pipe now
// https://github.com/PowerShell/openssh-portable/blob/latestw_all/contrib/win32/win32compat/ssh-agent/agent.c#L40
func (ka *KeyAgent) MakeAgent() error {
	if len(os.Getenv("SSH_AUTH_SOCK")) == 0 {
		return cli.ErrorCat("ssh agent not initialized")
	}
	// \\\\.\\pipe\\openssh-ssh-agent
	conn, err := winio.DialPipe("\\\\.\\pipe\\openssh-ssh-agent", nil)
	if err != nil {
		return err
	}
	ka.conn = conn
	return nil
}

// resize console https://docs.microsoft.com/en-us/windows/console/console-winevents

type (
	short int16
	word  uint16
	dword uint32
	wchar uint16

	coord struct {
		x short
		y short
	}
	smallRect struct {
		left   short
		top    short
		right  short
		bottom short
	}
	consoleScreenBufferInfo struct {
		size              coord
		cursorPosition    coord
		attributes        word
		window            smallRect
		maximumWindowSize coord
	}
	inputRecord struct {
		eventType word
		_         word     // Padding. Event struct is aligned Dword
		event     [16]byte // union struct's largest bytes
	}
	keyEventRecord struct {
		keyDown         int32
		repeatCount     word
		virtualKeyCode  word
		virtualScanCode word
		unicodeChar     wchar
		controlKeyState dword
	}
	mouseEventRecord struct {
		mousePosition   coord
		buttonState     dword
		controlKeyState dword
		eventFlags      dword
	}
	windowBufferSizeRecord struct {
		size coord
	}
)

const (
	enableEchoInput            = 0x0004
	enableExtendedFlags        = 0x0080
	enableInsertMode           = 0x0020
	enableLineInput            = 0x0002
	enableMouseInput           = 0x0010
	enableProcessedInput       = 0x0001
	enableQuickEditMode        = 0x0040
	enableWindowInput          = 0x0008
	enableAutoPosition         = 0x0100 // not in doc but it is said available
	enableVirtualTerminalInput = 0x0200

	enableProcessedOutput           = 0x0001
	enableWrapAtEolOutput           = 0x0002
	enableVirtualTerminalProcessing = 0x0004
	disableNewlineAutoReturn        = 0x0008
	enableLvbGridWorldwide          = 0x0010

	focusEvent            = 0x0010
	keyEvent              = 0x0001
	menuEvent             = 0x0008
	mouseEvent            = 0x0002
	windowBufferSizeEvent = 0x0004

	errorAccessDenied     syscall.Errno = 5
	errorInvalidHandle    syscall.Errno = 6
	errorInvalidParameter syscall.Errno = 87
)

type sysInfo struct {
	inmode   uint32
	outmode  uint32
	errmode  uint32
	lastRune rune
}

func (sc *SSHClient) changeLocalTerminalMode() error {

	if err := windows.GetConsoleMode(windows.Handle(os.Stdin.Fd()), &sc.sys.inmode); err != nil {
		return cli.ErrorCat("failed to get local stdin mode: ", err.Error())
	}

	if err := windows.GetConsoleMode(windows.Handle(os.Stdout.Fd()), &sc.sys.outmode); err != nil {
		return cli.ErrorCat("failed to get local stdout mode: ", err.Error())
	}
	if err := windows.GetConsoleMode(windows.Handle(os.Stderr.Fd()), &sc.sys.errmode); err != nil {
		return cli.ErrorCat("failed to get local stdout mode: ", err.Error())
	}

	newBaseMode := sc.sys.inmode &^ (enableEchoInput | enableProcessedInput | enableLineInput)
	newMode := newBaseMode | enableVirtualTerminalInput
	//err = setConsoleMode(os.Stdin.Fd(), newMode)
	if err := windows.SetConsoleMode(windows.Handle(os.Stdin.Fd()), newMode); err != nil {
		return cli.ErrorCat("failed to set local stdin mode: ", err.Error())
	}

	newMode = sc.sys.outmode | enableVirtualTerminalProcessing | disableNewlineAutoReturn

	if err := windows.SetConsoleMode(windows.Handle(os.Stdout.Fd()), newMode); err != nil {
		return cli.ErrorCat("failed to set local stdout mode: ", err.Error())
	}
	if err := windows.SetConsoleMode(windows.Handle(os.Stderr.Fd()), newMode); err != nil {
		return cli.ErrorCat("failed to set local stderr mode: ", err.Error())
	}
	return nil
}

func (sc *SSHClient) restoreLocalTerminalMode() error {
	if sc.sys.inmode > 0 {
		_ = windows.SetConsoleMode(windows.Handle(os.Stdin.Fd()), sc.sys.inmode)
	}
	if sc.sys.outmode > 0 {
		_ = windows.SetConsoleMode(windows.Handle(os.Stdout.Fd()), sc.sys.outmode)
	}
	if sc.sys.errmode > 0 {
		_ = windows.SetConsoleMode(windows.Handle(os.Stderr.Fd()), sc.sys.errmode)
	}
	return nil
}

func (sc *SSHClient) watchTerminalResize(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)
	sc.wg.Add(1)
	go func() {
		defer sc.wg.Done()

		ticker := time.NewTicker(2 * time.Second)
		defer func() {
			ticker.Stop()
			close(ch)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ch <- struct{}{}
			}
		}
	}()

	return ch
}
