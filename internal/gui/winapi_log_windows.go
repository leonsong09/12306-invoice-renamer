//go:build windows

package gui

import (
	"os"
	"syscall"
	"unsafe"
)

func clearLog(hwnd syscall.Handle) {
	setWindowText(hwnd, "")
}

func appendLogLine(hwnd syscall.Handle, line string) {
	text := line + "\r\n"
	length := getTextLength(hwnd)
	sendMessage(hwnd, emSetSel, uintptr(length), uintptr(length))
	p, _ := syscall.UTF16PtrFromString(text)
	sendMessage(hwnd, emReplaceSel, 0, uintptr(unsafe.Pointer(p)))
	sendMessage(hwnd, emScrollCaret, 0, 0)
}

func getTextLength(hwnd syscall.Handle) int {
	r1, _, _ := procGetWindowTextLenW.Call(uintptr(hwnd))
	return int(r1)
}

func drainLog(logEdit syscall.Handle, ch <-chan string) {
	for {
		select {
		case line, ok := <-ch:
			if !ok {
				return
			}
			appendLogLine(logEdit, line)
		default:
			return
		}
	}
}

func disableControls(a *app, disabled bool) {
	enable := !disabled
	enableWindow(a.inputEdit, enable)
	enableWindow(a.inputBrowse, enable)
	enableWindow(a.outputEdit, enable)
	enableWindow(a.outputBrowse, enable)
	enableWindow(a.dateTravel, enable)
	enableWindow(a.dateIssue, enable)
	enableWindow(a.startButton, enable)
}

func setFixedWindowSize(hwnd syscall.Handle, w int32, h int32) {
	var r rect
	procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	x := r.left
	y := r.top
	procMoveWindow.Call(uintptr(hwnd), uintptr(x), uintptr(y), uintptr(w), uintptr(h), 1)
}

func osGetwd() (string, error) {
	return os.Getwd()
}

