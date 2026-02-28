//go:build windows

package gui

import (
	"syscall"
	"unsafe"
)

type createStruct struct {
	lpCreateParams unsafe.Pointer
	hInstance      syscall.Handle
	hMenu          syscall.Handle
	hwndParent     syscall.Handle
	cy             int32
	cx             int32
	y              int32
	x              int32
	style          int32
	lpszName       *uint16
	lpszClass      *uint16
	dwExStyle      uint32
}

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     syscall.Handle
	hIcon         syscall.Handle
	hCursor       syscall.Handle
	hbrBackground syscall.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       syscall.Handle
}

type msg struct {
	hwnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type point struct {
	x, y int32
}

type rect struct {
	left, top, right, bottom int32
}

type browseInfo struct {
	hwndOwner      syscall.Handle
	pidlRoot       uintptr
	pszDisplayName *uint16
	lpszTitle      *uint16
	ulFlags        uint32
	lpfn           uintptr
	lParam         uintptr
	iImage         int32
}

type ciexyz struct {
	x, y, z int32
}

type ciexyztriple struct {
	red, green, blue ciexyz
}

type bitmapV5Header struct {
	bV5Size          uint32
	bV5Width         int32
	bV5Height        int32
	bV5Planes        uint16
	bV5BitCount      uint16
	bV5Compression   uint32
	bV5SizeImage     uint32
	bV5XPelsPerMeter int32
	bV5YPelsPerMeter int32
	bV5ClrUsed       uint32
	bV5ClrImportant  uint32
	bV5RedMask       uint32
	bV5GreenMask     uint32
	bV5BlueMask      uint32
	bV5AlphaMask     uint32
	bV5CSType        uint32
	bV5Endpoints     ciexyztriple
	bV5GammaRed      uint32
	bV5GammaGreen    uint32
	bV5GammaBlue     uint32
	bV5Intent        uint32
	bV5ProfileData   uint32
	bV5ProfileSize   uint32
	bV5Reserved      uint32
}

type iconInfo struct {
	fIcon    uint32
	xHotspot uint32
	yHotspot uint32
	hbmMask  syscall.Handle
	hbmColor syscall.Handle
}

const (
	wmNCCreate = 0x0081
	wmCreate   = 0x0001
	wmCommand  = 0x0111
	wmTimer    = 0x0113
	wmDestroy  = 0x0002

	wmSetIcon = 0x0080
	wmSetFont = 0x0030

	wmCtlColorEdit   = 0x0133
	wmCtlColorBtn    = 0x0135
	wmCtlColorStatic = 0x0138

	wsOverlapped   = 0x00000000
	wsCaption      = 0x00C00000
	wsSysMenu      = 0x00080000
	wsMinimizeBox  = 0x00020000
	wsClipChildren = 0x02000000
	wsVisible      = 0x10000000
	wsChild        = 0x40000000

	esLeft        = 0x0000
	esAutoHScroll = 0x0080
	esMultiLine   = 0x0004
	esAutoVScroll = 0x0040
	esReadOnly    = 0x0800

	wsVScroll = 0x00200000

	bsPushButton      = 0x00000000
	bsAutoRadioButton = 0x00000009

	bmGetCheck = 0x00F0
	bmSetCheck = 0x00F1
	bstChecked = 0x0001

	emSetSel      = 0x00B1
	emReplaceSel  = 0x00C2
	emScrollCaret = 0x00B7

	swShow = 5

	gwlpUserData = -21

	mbIconError       = 0x00000010
	mbIconInformation = 0x00000040
	mbOK              = 0x00000000
	mbYesNo           = 0x00000004
	mbIconQuestion    = 0x00000020

	idYes = 6
	idNo  = 7

	bifReturnOnlyFSDirs = 0x00000001
	bifNewDialogStyle   = 0x00000040
	bifEditBox          = 0x00000010

	coinitApartmentThreaded = 0x2

	colorWindow     = 5
	colorWindowText = 8
	idcArrow        = 32512

	iconSmall = 0
	iconBig   = 1

	stockDefaultGuiFont = 17

	dibRGBColors = 0

	biBitFields = 3

	lcsSrgb     = 0x73524742
	lcsGmImages = 4

	moveFileReplaceExisting = 0x00000001
	moveFileWriteThrough    = 0x00000008

	mainWindowStyle = wsOverlapped | wsCaption | wsSysMenu | wsMinimizeBox | wsClipChildren | wsVisible
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
)

var (
	procRegisterClassExW   = user32.NewProc("RegisterClassExW")
	procCreateWindowExW    = user32.NewProc("CreateWindowExW")
	procDefWindowProcW     = user32.NewProc("DefWindowProcW")
	procShowWindow         = user32.NewProc("ShowWindow")
	procUpdateWindow       = user32.NewProc("UpdateWindow")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procTranslateMessage   = user32.NewProc("TranslateMessage")
	procDispatchMessageW   = user32.NewProc("DispatchMessageW")
	procPostQuitMessage    = user32.NewProc("PostQuitMessage")
	procSetWindowTextW     = user32.NewProc("SetWindowTextW")
	procGetWindowTextW     = user32.NewProc("GetWindowTextW")
	procGetWindowTextLenW  = user32.NewProc("GetWindowTextLengthW")
	procSendMessageW       = user32.NewProc("SendMessageW")
	procEnableWindow       = user32.NewProc("EnableWindow")
	procSetTimer           = user32.NewProc("SetTimer")
	procKillTimer          = user32.NewProc("KillTimer")
	procMessageBoxW        = user32.NewProc("MessageBoxW")
	procSetWindowLongPtrW  = user32.NewProc("SetWindowLongPtrW")
	procGetWindowLongPtrW  = user32.NewProc("GetWindowLongPtrW")
	procGetWindowRect      = user32.NewProc("GetWindowRect")
	procMoveWindow         = user32.NewProc("MoveWindow")
	procLoadCursorW        = user32.NewProc("LoadCursorW")
	procDestroyIcon        = user32.NewProc("DestroyIcon")
	procGetSysColor        = user32.NewProc("GetSysColor")
	procGetSysColorBrush   = user32.NewProc("GetSysColorBrush")
	procCreateIconIndirect = user32.NewProc("CreateIconIndirect")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procMoveFileExW      = kernel32.NewProc("MoveFileExW")

	procSHBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")

	procCoInitializeEx = ole32.NewProc("CoInitializeEx")
	procCoUninitialize = ole32.NewProc("CoUninitialize")
	procCoTaskMemFree  = ole32.NewProc("CoTaskMemFree")

	procGetStockObject   = gdi32.NewProc("GetStockObject")
	procCreateDIBSection = gdi32.NewProc("CreateDIBSection")
	procCreateBitmap     = gdi32.NewProc("CreateBitmap")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
	procSetBkColor       = gdi32.NewProc("SetBkColor")
	procSetTextColor     = gdi32.NewProc("SetTextColor")
)
