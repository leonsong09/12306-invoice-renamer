//go:build windows

package gui

import (
	"fmt"
	"syscall"
	"unsafe"
)

func registerMainWindowClass(className *uint16, hInstance syscall.Handle, wndProc uintptr) error {
	cur, _, _ := procLoadCursorW.Call(0, idcArrow)
	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   wndProc,
		hInstance:     hInstance,
		hCursor:       syscall.Handle(cur),
		hbrBackground: syscall.Handle(colorWindow + 1),
		lpszClassName: className,
	}
	r1, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if r1 == 0 {
		return fmt.Errorf("RegisterClassExW failed: %w", err)
	}
	return nil
}

func createMainWindow(className *uint16, hInstance syscall.Handle, lpParam unsafe.Pointer) (syscall.Handle, error) {
	title := syscall.StringToUTF16Ptr(appTitle)
	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		mainWindowStyle,
		100,
		100,
		780,
		520,
		0,
		0,
		uintptr(hInstance),
		uintptr(lpParam),
	)
	if hwnd == 0 {
		return 0, fmt.Errorf("CreateWindowExW failed: %w", err)
	}
	return syscall.Handle(hwnd), nil
}

func defWindowProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	r1, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r1
}

func showWindow(hwnd syscall.Handle) {
	procShowWindow.Call(uintptr(hwnd), swShow)
}

func updateWindow(hwnd syscall.Handle) {
	procUpdateWindow.Call(uintptr(hwnd))
}

func messageLoop() error {
	var m msg
	for {
		r1, _, err := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r1) == -1 {
			return fmt.Errorf("GetMessageW failed: %w", err)
		}
		if r1 == 0 {
			return nil
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func postQuitMessage() {
	procPostQuitMessage.Call(0)
}

func setWindowUserData(hwnd syscall.Handle, value uintptr) {
	idx := int32(gwlpUserData)
	procSetWindowLongPtrW.Call(uintptr(hwnd), uintptr(idx), value)
}

func getWindowUserData(hwnd syscall.Handle) uintptr {
	idx := int32(gwlpUserData)
	r1, _, _ := procGetWindowLongPtrW.Call(uintptr(hwnd), uintptr(idx))
	return r1
}

func getModuleHandle() syscall.Handle {
	h, _, _ := procGetModuleHandleW.Call(0)
	return syscall.Handle(h)
}

