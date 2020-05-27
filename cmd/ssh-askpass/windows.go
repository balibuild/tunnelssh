// +build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/balibuild/tunnelssh/cli"
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
	pGetActiveWindow                    = user32.NewProc("GetActiveWindow")
	pCredUIPromptForCredentialsW        = credui.NewProc("CredUIPromptForCredentialsW")
	pCredUIPromptForWindowsCredentialsW = credui.NewProc("CredUIPromptForWindowsCredentialsW")
	pCredUICmdLinePromptForCredentialsW = credui.NewProc("CredUICmdLinePromptForCredentialsW")
	pCredPackAuthenticationBufferW      = credui.NewProc("CredPackAuthenticationBufferW")
	pCredUnPackAuthenticationBufferW    = credui.NewProc("CredUnPackAuthenticationBufferW")
	pTaskDialog                         = comctl32.NewProc("TaskDialog")
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
	//pressed := MessageBox(uintptr(GetActiveWindow()), caption, title, windows.MB_YESNO|windows.MB_ICONWARNING)
	pressed := TaskDialog(caption, title)
	if pressed == IDYES {
		fmt.Fprintf(os.Stdout, "Yes\n")
	} else {
		fmt.Fprintf(os.Stdout, "No\n")
	}
	return 0
}

// nolint: golint
// const
const (
	NULL                                        = uintptr(0)
	FALSE                                       = 0
	TRUE                                        = 1
	CreduiMaxPasswordLength                     = 256
	CreduiMaxUserNameLength                     = 512 + 1
	CreduiFlagsIncorrectPassword         uint32 = 0x00001 // indicates the username is valid, but password is not
	CreduiFlagsDoNotPersist              uint32 = 0x00002 // Do not show "Save" checkbox, and do not persist credentials
	CreduiFlagsRequestAdministrator      uint32 = 0x00004 // Populate list box with admin accounts
	CreduiFlagsExcludeCertificates       uint32 = 0x00008 // do not include certificates in the drop list
	CreduiFlagsRequireCertificate        uint32 = 0x00010
	CreduiFlagsShowSaveCheckBox          uint32 = 0x00040
	CreduiFlagsAlwaysShowUI              uint32 = 0x00080
	CreduiFlagsRequireSmartcard          uint32 = 0x00100
	CreduiFlagsPasswordOnlyOK            uint32 = 0x00200
	CreduiFlagsValidateUsername          uint32 = 0x00400
	CreduiFlagsCompleteUsername          uint32 = 0x00800 //
	CreduiFlagsPersist                   uint32 = 0x01000 // Do not show "Save" checkbox, but persist credentials anyway
	CreduiFlagsServerCredential          uint32 = 0x04000
	CreduiFlagsExpectConfirmation        uint32 = 0x20000    // do not persist unless caller later confirms credential via CredUIConfirmCredential() api
	CreduiFlagsGenericCredentials        uint32 = 0x40000    // Credential is a generic credential
	CreduiFlagsUsernameTargetCredentials uint32 = 0x80000    // Credential has a username as the target
	CreduiFlagsKeepUsername              uint32 = 0x100000   // don't allow the user to change the supplied username
	CreduiwinGeneric                     uint32 = 0x00000001 // Plain text username/password is being requested
	CreduiwinCheckbox                    uint32 = 0x00000002 // Show the Save Credential checkbox
	CreduiwinAuthpackageOnly             uint32 = 0x00000010 // Only Cred Providers that support the input auth package should enumerate
	CreduiwinInCredOnly                  uint32 = 0x00000020 // Only the incoming cred for the specific auth package should be enumerated
	CreduiwinEnumerateAdmins             uint32 = 0x00000100 // Cred Providers should enumerate administrators only
	CreduiwinEnumerateCurrentUser        uint32 = 0x00000200 // Only the incoming cred for the specific auth package should be enumerated
	CreduiwinSecurePrompt                uint32 = 0x00001000 // The Credui prompt should be displayed on the secure desktop
	CreduiwinPreprompting                uint32 = 0x00002000 // CredUI is invoked by SspiPromptForCredentials and the client is prompting before a prior handshake
	CreduiwinPack32WoW                   uint32 = 0x10000000 // Tell the credential provider it should be packing its Auth Blob 32 bit even though it is running 64 native
	CreduiwinWindowsHello                uint32 = 0x8000000  //Windows Hello credentials will be packed in a smart card auth buffer. This only applies to the face, fingerprint, and PIN credential providers.
	CredPackProtectedCredentials         uint32 = 0x1
	CredPackWOWBuffer                    uint32 = 0x2
	CredPackGenericCredentials           uint32 = 0x4
	CredPackIDProviderCredentials        uint32 = 0x8
)

type creduiinfoa struct {
	cbSize         uint32
	hwnd           windows.Handle
	pszMessageText *byte
	pszCaptionText *byte
	hbmBanner      windows.Handle
}

type creduiinfow struct {
	cbSize         uint32
	hwnd           windows.Handle
	pszMessageText *uint16
	pszCaptionText *uint16
	hbmBanner      windows.Handle
}

// ZeroMemory don't modified
//go:linkname ZeroMemory runtime.memclrNoHeapPointers
func ZeroMemory(ptr unsafe.Pointer, n uintptr)

// CredUIPromptForWindowsCredentials modern UI
// CredUICmdLinePromptForCredentials is xx
func CredUIPromptForWindowsCredentials(prompt, user string) (string, error) {
	var ci creduiinfow
	ci.cbSize = uint32(unsafe.Sizeof(ci))
	ci.pszCaptionText = syscall.StringToUTF16Ptr("Askpass Utility for TunnelSSH")
	ci.pszMessageText = syscall.StringToUTF16Ptr(cli.StrCat(prompt, "\nEnter username '", user, "'"))
	ci.hwnd = GetActiveWindow()
	ci.hbmBanner = windows.Handle(0)
	var authPackage uint32
	var cred *byte
	username := make([]uint16, CreduiMaxUserNameLength+1)
	passwd := make([]uint16, CreduiMaxPasswordLength+1)
	ulen := uint32(CreduiMaxUserNameLength + 1)
	plen := uint32(CreduiMaxPasswordLength + 1)
	var credlen uint32
	fSave := FALSE
	r, _, _ := pCredUIPromptForWindowsCredentialsW.Call(
		uintptr(unsafe.Pointer(&ci)),
		0,
		uintptr(unsafe.Pointer(&authPackage)),
		NULL,
		0,
		uintptr(unsafe.Pointer(&cred)),
		uintptr(unsafe.Pointer(&credlen)),
		uintptr(unsafe.Pointer(&fSave)),
		uintptr(CreduiwinGeneric),
	)
	if r != 0 {
		return "", fmt.Errorf("CredUIPromptForWindowsCredentials %v", windows.GetLastError())
	}
	defer ZeroMemory(unsafe.Pointer(cred), uintptr(credlen))
	defer windows.CoTaskMemFree(unsafe.Pointer(cred))
	//https://docs.microsoft.com/en-us/windows/win32/api/wincred/nf-wincred-credunpackauthenticationbufferw
	r, _, _ = pCredUnPackAuthenticationBufferW.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(cred)),
		uintptr(credlen),
		uintptr(unsafe.Pointer(&username[0])),
		uintptr(unsafe.Pointer(&ulen)),
		NULL,
		NULL,
		uintptr(unsafe.Pointer(&passwd[0])),
		uintptr(unsafe.Pointer(&plen)),
	)
	if r == FALSE {
		return "", fmt.Errorf("CredUnPackAuthenticationBuffer %v", windows.GetLastError())
	}
	return string(utf16.Decode(passwd[:plen])), nil
}

// CredUIPromptForCredentials todo
func CredUIPromptForCredentials(prompt, user string) (string, error) {
	var ci creduiinfow
	ci.cbSize = uint32(unsafe.Sizeof(ci))
	ci.pszCaptionText = syscall.StringToUTF16Ptr("Askpass Utility for TunnelSSH")
	ci.pszMessageText = syscall.StringToUTF16Ptr(prompt)
	ci.hwnd = GetActiveWindow()
	ci.hbmBanner = windows.Handle(0)
	passwd := make([]uint16, CreduiMaxPasswordLength+1)
	username := make([]uint16, 0, CreduiMaxUserNameLength)
	//https://medium.com/jettech/breaking-all-the-rules-using-go-to-call-windows-api-2cbfd8c79724
	userbuf := utf16.Encode([]rune(user + "\x00"))
	username = append(username, userbuf...)

	fSave := FALSE
	//fSave := windows.FALSE
	flags := CreduiFlagsGenericCredentials | CreduiFlagsKeepUsername | CreduiFlagsPasswordOnlyOK | CreduiFlagsAlwaysShowUI | CreduiFlagsDoNotPersist
	r, _, _ := pCredUIPromptForCredentialsW.Call(
		uintptr(unsafe.Pointer(&ci)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("TheServer"))),
		NULL,                                  //Reserved
		0,                                     // Reason
		uintptr(unsafe.Pointer(&username[0])), //User
		uintptr(CreduiMaxPasswordLength),      // Max number of char of user name
		uintptr(unsafe.Pointer(&passwd[0])),   // Password
		uintptr(CreduiMaxPasswordLength+1),    // Max number of password length
		uintptr(unsafe.Pointer(&fSave)),       // State of save check box
		uintptr(flags),                        // flags
	)
	if r != 0 {
		return "", fmt.Errorf("CredUIPromptForCredentials return %d", r)
	}
	var n int
	for ; n < CreduiMaxPasswordLength; n++ {
		if passwd[n] == 0 {
			break
		}
	}
	return string(utf16.Decode(passwd[:n])), nil
}

// https://github.com/jeroen/askpass/blob/master/src/win32/win-askpass.c
// Credui.dll

// AskPassword todo
func AskPassword(caption, user string) int {
	passwd, err := CredUIPromptForWindowsCredentials(caption, user)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Credentials: %s\n", err)
		return 1
	}
	fmt.Fprintf(os.Stdout, "%s\n", passwd)
	return 0
}
