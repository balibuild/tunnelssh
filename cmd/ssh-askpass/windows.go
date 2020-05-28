// +build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/mattn/go-isatty"
	"golang.org/x/sys/windows"
)

const (
	errorAccessDenied     syscall.Errno = 5
	errorInvalidHandle    syscall.Errno = 6
	errorInvalidParameter syscall.Errno = 87
)

var (
	kernel32                            = syscall.NewLazyDLL("kernel32.dll")
	user32                              = syscall.NewLazyDLL("user32.dll")
	credui                              = syscall.NewLazyDLL("Credui.dll")
	comctl32                            = syscall.NewLazyDLL("comctl32.dll")
	pGetModuleHandleW                   = kernel32.NewProc("GetModuleHandleW")
	pSetForegroundWindow                = user32.NewProc("SetForegroundWindow")
	pGetActiveWindow                    = user32.NewProc("GetActiveWindow")
	pCredUIPromptForCredentialsW        = credui.NewProc("CredUIPromptForCredentialsW")
	pCredUIPromptForWindowsCredentialsW = credui.NewProc("CredUIPromptForWindowsCredentialsW")
	pCredUICmdLinePromptForCredentialsW = credui.NewProc("CredUICmdLinePromptForCredentialsW")
	pCredPackAuthenticationBufferW      = credui.NewProc("CredPackAuthenticationBufferW")
	pCredUnPackAuthenticationBufferW    = credui.NewProc("CredUnPackAuthenticationBufferW")
	pTaskDialog                         = comctl32.NewProc("TaskDialog")
	pTaskDialogIndirect                 = comctl32.NewProc("TaskDialogIndirect")
	pGetConsoleMode                     = kernel32.NewProc("GetConsoleMode")
	pSetConsoleMode                     = kernel32.NewProc("SetConsoleMode")
	pGetConsoleScreenBufferInfo         = kernel32.NewProc("GetConsoleScreenBufferInfo")
	pReadConsoleInput                   = kernel32.NewProc("ReadConsoleInputW")
	pAttachConsole                      = kernel32.NewProc("AttachConsole")
	pAllocConsole                       = kernel32.NewProc("AllocConsole")
)

// defined
const (
	IDOK              = 1
	IDCANCEL          = 2
	IDABORT           = 3
	IDRETRY           = 4
	IDIGNORE          = 5
	IDYES             = 6
	IDNO              = 7
	TDCBFOKBUTTON     = 0x0001 // selected control return value IDOK
	TDCBFYESBUTTON    = 0x0002 // selected control return value IDYES
	TDCBFNOBUTTON     = 0x0004 // selected control return value IDNO
	TDCBFCANCELBUTTON = 0x0008 // selected control return value IDCANCEL
	TDCBFRETRYBUTTON  = 0x0010 // selected control return value IDRETRY
	TDCBFCLOSEBUTTON  = 0x0020 // selected control return value IDCLOSE
	//

	TDFENABLEHYPERLINKS         = 0x0001
	TDFUSEHICONMAIN             = 0x0002
	TDFUSEHICONFOOTER           = 0x0004
	TDFALLOWDIALOGCANCELLATION  = 0x0008
	TDFUSECOMMANDLINKS          = 0x0010
	TDFUSECOMMANDLINKSNOICON    = 0x0020
	TDFEXPANDFOOTERAREA         = 0x0040
	TDFEXPANDEDBYDEFAULT        = 0x0080
	TDFVERIFICATIONFLAGCHECKED  = 0x0100
	TDFSHOWPROGRESSBAR          = 0x0200
	TDFSHOWMARQUEEPROGRESSBAR   = 0x0400
	TDFCALLBACKTIMER            = 0x0800
	TDFPOSITIONRELATIVETOWINDOW = 0x1000
	TDFRTLLAYOUT                = 0x2000
	TDFNODEFAULTRADIOBUTTON     = 0x4000
	TDFCANBEMINIMIZED           = 0x8000
	TDFNOSETFOREGROUND          = 0x00010000 // Don't call SetForegroundWindow() when activating the dialog
	TDFSIZETOCONTENT            = 0x01000000 // used by ShellMessageBox to emulate MessageBox sizing behavior
)

// GetActiveWindow todo
func GetActiveWindow() windows.Handle {
	h, _, e1 := syscall.Syscall(pGetActiveWindow.Addr(), 0, 0, 0, 0)
	if e1 != 0 {
		return 0
	}
	return windows.Handle(h)
}

// MessageBox todo
func MessageBox(hwnd uintptr, caption, title string, flags uint) int {
	ret, _ := windows.MessageBox(
		windows.Handle(hwnd),
		syscall.StringToUTF16Ptr(caption),
		syscall.StringToUTF16Ptr(title),
		uint32(flags))
	return int(ret)
}

// HresultOK todo
// HRESULT 0 -1 success
func HresultOK(i int) bool {
	return i == 0 || i == -1
}

// MakeIntreSource todo
func MakeIntreSource(i int) uintptr {
	return uintptr(uint16(i))
}

// typedef struct _TASKDIALOG_BUTTON
// {
//     int     nButtonID;
//     PCWSTR  pszButtonText;
// } TASKDIALOG_BUTTON;

type taskButton struct {
	nButtonID     int
	pszButtonText *uint16
}

type taskDialogConfig struct {
	cbSize                  uint32
	hwndParent              uintptr
	hInst                   uintptr
	dwFlags                 int32
	dwCommonButtons         int32
	pszWindowTitle          *uint16
	pszMainIcon             uintptr //avoid convert
	pszMainInstruction      *uint16
	pszContent              *uint16
	cButtons                int32
	pButtons                uintptr
	nDefaultButton          int32
	cRadioButtons           uint32
	pDefaultRadioButton     uintptr
	nDefaultRadioButton     int32
	pszVerificationText     *uint16
	pszExpandedInformation  *uint16
	pszExpandedControlText  *uint16
	pszCollapsedControlText *uint16
	pszFooterIcon           *uint16
	pszFooter               *uint16
	pfCallback              uintptr
	lpCallbackData          uintptr
	cxWidth                 uint32
}

func taskdialogcallback(hwnd windows.Handle, msg uint, wParam uintptr, lParam, lpRefData uintptr) uint {
	if msg == 0 {
		_, _, _ = pSetForegroundWindow.Call(uintptr(hwnd))
	}
	return 0
}

//TaskDialogIndirect todo
func TaskDialogIndirect(caption, title string) int {
	var tc taskDialogConfig
	var nButton, nRadioButton int
	tc.cbSize = uint32(unsafe.Sizeof(tc))
	fmt.Fprintf(os.Stderr, "%d \n", tc.cbSize)
	tc.hwndParent = uintptr(GetActiveWindow())
	h, _, _ := pGetModuleHandleW.Call(NULL)
	tc.hInst = h
	tc.pszWindowTitle = syscall.StringToUTF16Ptr(title)
	tc.pszContent = syscall.StringToUTF16Ptr(caption)
	tc.dwFlags = TDFALLOWDIALOGCANCELLATION | TDFPOSITIONRELATIVETOWINDOW | TDFSIZETOCONTENT
	tc.pszCollapsedControlText = syscall.StringToUTF16Ptr("More information")
	tc.pszExpandedControlText = syscall.StringToUTF16Ptr("Less information")
	tc.pszMainIcon = MakeIntreSource(-3)
	tc.pfCallback = syscall.NewCallback(taskdialogcallback)
	r, _, _ := pTaskDialogIndirect.Call(
		uintptr(unsafe.Pointer(&tc)),
		uintptr(unsafe.Pointer(&nButton)),
		uintptr(unsafe.Pointer(&nRadioButton)),
		0,
	)
	fmt.Fprintf(os.Stderr, "%v \n", windows.GetLastError())
	if r != 0 {
		return IDNO
	}
	return int(r)
}

//HRESULT WINAPI TaskDialogIndirect(_In_ const TASKDIALOGCONFIG *pTaskConfig,
//_Out_opt_ int *pnButton, _Out_opt_ int *pnRadioButton, _Out_opt_ BOOL *pfVerificationFlagChecked);

// TaskDialog impl
// Must use bali build it
func TaskDialog(caption, title string) int {
	var nButtonPressed int
	h, _, _ := pGetModuleHandleW.Call(NULL)
	r, _, _ := pTaskDialog.Call(
		uintptr(GetActiveWindow()),
		uintptr(h), // Mode
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Askpass Utility Confirm"))), //icon
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(caption))),                   //icon
		uintptr(TDCBFYESBUTTON|TDCBFNOBUTTON),
		MakeIntreSource(-1), //WARNING ICON
		uintptr(unsafe.Pointer(&nButtonPressed)),
	)
	if !HresultOK(int(r)) {
		return -1
	}
	return nButtonPressed
}

const (
	conin  string = "CONIN$"
	conout string = "CONOUT$"
)

func attachConsole(pid uint32) (err error) {
	r1, _, e1 := syscall.Syscall(pAttachConsole.Addr(), 1, uintptr(pid), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}

	return

}

func allocConsole() (err error) {
	r1, _, e1 := syscall.Syscall(pAllocConsole.Addr(), 0, 0, 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}

	return
}

// AskYesConsole todo
func AskYesConsole(caption, title string) int {
	err := attachConsole(uint32(os.Getpid()))
	if err != nil && err == error(errorInvalidHandle) {
		err = allocConsole()
		if err != nil {
			fmt.Fprintf(os.Stderr, "allocConsole %v\n", err)
			return 1
		}
	}
	in, err := os.OpenFile(conin, os.O_RDONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable open CONIN$ %v\n", err)
		return 1
	}
	defer in.Close()
	fmt.Fprintf(os.Stderr, "%s", caption)
	br := bufio.NewReader(in)
	for {
		answer, err := br.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable read string %v", err)
			return 1
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "yes" {
			fmt.Fprintf(os.Stdout, "yes")
			return 0
		}
		if answer == "no" {
			fmt.Fprintf(os.Stdout, "no")
			return 0
		}
		fmt.Fprintf(os.Stderr, "Please type 'yes' or 'no': ")
	}
}

// AskYes todo
func AskYes(caption, title string) int {
	if isatty.IsTerminal(os.Stderr.Fd()) {
		return AskYesConsole(caption, title)
	}
	pressed := MessageBox(uintptr(GetActiveWindow()), caption, title, windows.MB_YESNO|windows.MB_ICONWARNING)
	//pressed := TaskDialogIndirect(caption, title)
	if pressed == IDYES {
		fmt.Fprintf(os.Stdout, "Yes\n")
	} else {
		fmt.Fprintf(os.Stdout, "No\n")
	}
	return 0
}
