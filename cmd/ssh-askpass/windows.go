// +build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// MessageBox todo
func MessageBox(hwnd uintptr, caption, title string, flags uint) int {
	ret, _ := windows.MessageBox(
		windows.Handle(hwnd),
		syscall.StringToUTF16Ptr(caption),
		syscall.StringToUTF16Ptr(title),
		uint32(flags))
	return int(ret)
}

// defined
const (
	IDYES = 6
)

// AskYes todo
func AskYes(caption, title string) int {
	if MessageBox(0, caption, title, windows.MB_YESNO|windows.MB_ICONWARNING) == IDYES {
		return 0
	}
	return 1
}

var (
	user32                             = syscall.NewLazyDLL("user32.dll")
	credui                             = syscall.NewLazyDLL("Credui.dll")
	pGetActiveWindow                   = user32.NewProc("GetActiveWindow")
	pCredUIPromptForCredentialsA       = credui.NewProc("CredUIPromptForCredentialsA")
	pCredUIPromptForWindowsCredentials = credui.NewProc("CredUIPromptForWindowsCredentialsW")
	pCredUnPackAuthenticationBuffer    = credui.NewProc("CredUnPackAuthenticationBufferW")
)

// GetActiveWindow todo
func GetActiveWindow() windows.Handle {
	h, _, e1 := syscall.Syscall(pGetActiveWindow.Addr(), 0, 0, 0, 0)
	if e1 != 0 {
		return 0
	}
	return windows.Handle(h)
}

/*
CREDUIAPI DWORD CredUIPromptForCredentialsW(
  PCREDUI_INFOW pUiInfo,
  PCWSTR        pszTargetName,
  PCtxtHandle   pContext,
  DWORD         dwAuthError,
  PWSTR         pszUserName,
  ULONG         ulUserNameBufferSize,
  PWSTR         pszPassword,
  ULONG         ulPasswordBufferSize,
  BOOL          *save,
  DWORD         dwFlags
);
*/

// #define CREDUIWIN_GENERIC                   0x00000001  // Plain text username/password is being requested
// #define CREDUIWIN_CHECKBOX                  0x00000002  // Show the Save Credential checkbox
// #define CREDUIWIN_AUTHPACKAGE_ONLY          0x00000010  // Only Cred Providers that support the input auth package should enumerate
// #define CREDUIWIN_IN_CRED_ONLY              0x00000020  // Only the incoming cred for the specific auth package should be enumerated
// #define CREDUIWIN_ENUMERATE_ADMINS          0x00000100  // Cred Providers should enumerate administrators only
// #define CREDUIWIN_ENUMERATE_CURRENT_USER    0x00000200  // Only the incoming cred for the specific auth package should be enumerated
// #define CREDUIWIN_SECURE_PROMPT             0x00001000  // The Credui prompt should be displayed on the secure desktop
// #define CREDUIWIN_PREPROMPTING              0X00002000  // CredUI is invoked by SspiPromptForCredentials and the client is prompting before a prior handshake
// #define CREDUIWIN_PACK_32_WOW               0x10000000  // Tell the credential provider it should be packing its Auth Blob 32 bit even though it is running 64 native

// nolint: golint
// const
const (
	NULL                                        = uintptr(0)
	FALSE                                       = 0
	TRUE                                        = 1
	CreduiMaxPasswordLength                     = 512
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
	CreduiFlagsExpectConfirmation        uint32 = 0x20000  // do not persist unless caller later confirms credential via CredUIConfirmCredential() api
	CreduiFlagsGenericCredentials        uint32 = 0x40000  // Credential is a generic credential
	CreduiFlagsUsernameTargetCredentials uint32 = 0x80000  // Credential has a username as the target
	CreduiFlagsKeepUsername              uint32 = 0x100000 // don't allow the user to change the supplied username

)

type creduiinfoa struct {
	cbSize         uint32
	hwnd           windows.Handle
	pszMessageText *byte
	pszCaptionText *byte
	hbmBanner      windows.Handle
}

// CredUIPromptForCredentials todo
func CredUIPromptForCredentials(prompt, user string) (string, error) {
	var ci creduiinfoa
	ci.cbSize = uint32(unsafe.Sizeof(ci))
	ci.pszCaptionText = syscall.StringBytePtr(prompt)
	ci.pszMessageText = syscall.StringBytePtr("TunnelSSH - Please enter password")
	ci.hwnd = GetActiveWindow()

	ci.hbmBanner = windows.Handle(0)
	passwd := make([]byte, CreduiMaxPasswordLength+1)
	fSave := FALSE
	//fSave := windows.FALSE
	flags := CreduiFlagsGenericCredentials | CreduiFlagsKeepUsername | CreduiFlagsPasswordOnlyOK | CreduiFlagsAlwaysShowUI | CreduiFlagsDoNotPersist
	r, _, _ := pCredUIPromptForCredentialsA.Call(
		uintptr(unsafe.Pointer(&ci)),
		uintptr(unsafe.Pointer(syscall.StringBytePtr("TheServer"))),
		NULL, //Reserved
		0,    // Reason
		uintptr(unsafe.Pointer(syscall.StringBytePtr(user))), //User
		0,                                   // Max number of char of user name
		uintptr(unsafe.Pointer(&passwd[0])), // Password
		CreduiMaxPasswordLength+1,           // Max number of password length
		uintptr(unsafe.Pointer(&fSave)),     // State of save check box
		uintptr(flags),                      // flags
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
	return string(passwd[:n]), nil
}

// https://github.com/jeroen/askpass/blob/master/src/win32/win-askpass.c
// Credui.dll

// AskPassword todo
func AskPassword(caption, title string) int {
	passwd, err := CredUIPromptForCredentials(caption, title)
	if err != nil {
		return 1
	}
	fmt.Fprintf(os.Stdout, "%s\n", passwd)
	return 1
}
