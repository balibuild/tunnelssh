// +build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/balibuild/tunnelssh/cli"
	"golang.org/x/sys/windows"
)

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
