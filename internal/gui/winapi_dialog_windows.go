//go:build windows

package gui

import (
	"errors"
	"syscall"
	"unsafe"
)

func showErrorBox(title string, msg string) {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(msg)
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), mbOK|mbIconError)
}

func showInfoBox(title string, msg string) {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(msg)
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), mbOK|mbIconInformation)
}

func askYesNo(owner syscall.Handle, title string, msg string) bool {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(msg)
	r1, _, _ := procMessageBoxW.Call(uintptr(owner), uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), mbYesNo|mbIconQuestion)
	return r1 == idYes
}

func coInitialize() error {
	r1, _, _ := procCoInitializeEx.Call(0, coinitApartmentThreaded)
	if int32(r1) < 0 {
		return errors.New("CoInitializeEx 失败")
	}
	return nil
}

func coUninitialize() {
	procCoUninitialize.Call()
}

func browseForFolder(owner syscall.Handle, title string) (string, bool) {
	displayName := make([]uint16, 260)
	t, _ := syscall.UTF16PtrFromString(title)

	bi := browseInfo{
		hwndOwner:      owner,
		pszDisplayName: &displayName[0],
		lpszTitle:      t,
		ulFlags:        bifReturnOnlyFSDirs | bifNewDialogStyle | bifEditBox,
	}
	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return "", false
	}
	defer procCoTaskMemFree.Call(pidl)

	pathBuf := make([]uint16, 260)
	ok, _, _ := procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&pathBuf[0])))
	if ok == 0 {
		return "", false
	}
	return syscall.UTF16ToString(pathBuf), true
}
