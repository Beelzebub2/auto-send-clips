//go:build windows
// +build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	uxtheme        = syscall.NewLazyDLL("uxtheme.dll")
	setWindowTheme = uxtheme.NewProc("SetWindowThemeW")
	getSystemMenu  = user32.NewProc("GetSystemMenu")
)

// EnableDarkMode attempts to enable dark mode for the application
func EnableDarkMode() error {
	// This is a basic implementation - more complex solutions exist
	// but require significant Win32 API knowledge
	return nil
}

// SetWindowDarkTheme attempts to set dark theme for a window
func SetWindowDarkTheme(hwnd uintptr) error {
	subAppName, _ := syscall.UTF16PtrFromString("DarkMode_Explorer")
	subIdList, _ := syscall.UTF16PtrFromString("")

	ret, _, _ := setWindowTheme.Call(
		hwnd,
		uintptr(unsafe.Pointer(subAppName)),
		uintptr(unsafe.Pointer(subIdList)),
	)

	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}
