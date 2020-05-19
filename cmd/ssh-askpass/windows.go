// +build windows

package main

import (
	"syscall"
	"unsafe"
)

// MessageBox todo
func MessageBox(hwnd uintptr, caption, title string, flags uint) int {
	ret, _, _ := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(caption))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(flags))

	return int(ret)
}

// MessageBoxPlain of Win32 API.
func MessageBoxPlain(title, caption string) int {
	const (
		NULL = 0
		MBOK = 0
	)
	return MessageBox(NULL, caption, title, MB_OK)
}

// https://github.com/jeroen/askpass/blob/master/src/win32/win-askpass.c
// Credui.dll
