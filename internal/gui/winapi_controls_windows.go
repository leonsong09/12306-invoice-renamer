//go:build windows

package gui

import (
	"fmt"
	"syscall"
	"unsafe"
)

func setWindowText(hwnd syscall.Handle, text string) {
	p, _ := syscall.UTF16PtrFromString(text)
	procSetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(p)))
}

func getWindowText(hwnd syscall.Handle) (string, error) {
	r1, _, err := procGetWindowTextLenW.Call(uintptr(hwnd))
	if r1 == 0 {
		return "", nil
	}
	buf := make([]uint16, r1+1)
	r2, _, err := procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(r1+1))
	if r2 == 0 {
		return "", fmt.Errorf("GetWindowTextW failed: %w", err)
	}
	return syscall.UTF16ToString(buf), nil
}

func enableWindow(hwnd syscall.Handle, enable bool) {
	v := uintptr(0)
	if enable {
		v = 1
	}
	procEnableWindow.Call(uintptr(hwnd), v)
}

func sendMessage(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	r1, _, _ := procSendMessageW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r1
}

func setTimer(hwnd syscall.Handle, id uintptr, elapseMs uint32) {
	procSetTimer.Call(uintptr(hwnd), id, uintptr(elapseMs), 0)
}

func killTimer(hwnd syscall.Handle, id uintptr) {
	procKillTimer.Call(uintptr(hwnd), id)
}

func createControl(class string, text string, style uint32, x, y, w, h int32, parent syscall.Handle, id int) syscall.Handle {
	c, _ := syscall.UTF16PtrFromString(class)
	t, _ := syscall.UTF16PtrFromString(text)
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(t)),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(w),
		uintptr(h),
		uintptr(parent),
		uintptr(id),
		uintptr(getModuleHandle()),
		0,
	)
	return syscall.Handle(hwnd)
}

func createStatic(parent syscall.Handle, text string, x, y, w, h int32) syscall.Handle {
	return createControl("STATIC", text, wsChild|wsVisible, x, y, w, h, parent, 0)
}

func createEdit(parent syscall.Handle, id int, x, y, w, h int32) syscall.Handle {
	style := uint32(wsChild | wsVisible | esLeft | esAutoHScroll)
	return createControl("EDIT", "", style, x, y, w, h, parent, id)
}

func createButton(parent syscall.Handle, id int, text string, x, y, w, h int32) syscall.Handle {
	style := uint32(wsChild | wsVisible | bsPushButton)
	return createControl("BUTTON", text, style, x, y, w, h, parent, id)
}

func createRadio(parent syscall.Handle, id int, text string, x, y, w, h int32, checked bool) syscall.Handle {
	style := uint32(wsChild | wsVisible | bsAutoRadioButton)
	hwnd := createControl("BUTTON", text, style, x, y, w, h, parent, id)
	if checked {
		sendMessage(hwnd, bmSetCheck, bstChecked, 0)
	}
	return hwnd
}

func createLogEdit(parent syscall.Handle, id int, x, y, w, h int32) syscall.Handle {
	style := uint32(wsChild | wsVisible | esMultiLine | esAutoVScroll | esReadOnly | wsVScroll)
	return createControl("EDIT", "", style, x, y, w, h, parent, id)
}

func isChecked(hwnd syscall.Handle) bool {
	v := sendMessage(hwnd, bmGetCheck, 0, 0)
	return v == bstChecked
}

func loword(v uintptr) int {
	return int(uint16(v & 0xFFFF))
}

