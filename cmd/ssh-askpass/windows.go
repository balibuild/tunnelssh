// +build windows

package main

import (
	"syscall"

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

// https://github.com/jeroen/askpass/blob/master/src/win32/win-askpass.c
// Credui.dll

// AskPassword todo
func AskPassword(caption, title string) int {

	return 1
}
