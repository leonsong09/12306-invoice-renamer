//go:build windows

package gui

import (
	"TrainTicketsTool/internal/processor"
	"syscall"
	"unsafe"
)

const (
	appTitle = "12306 发票重命名工具 - 作者：leonsong09"

	timerID          = 1
	uiPollIntervalMs = 100

	logBufferSize = 256

	wmCreateFailed = ^uintptr(0)
)

const (
	idInputEdit    = 1001
	idInputBrowse  = 1002
	idOutputEdit   = 1003
	idOutputBrowse = 1004
	idDateTravel   = 1005
	idDateIssue    = 1006
	idStartButton  = 1007
	idLogEdit      = 1008
)

type app struct {
	hwnd syscall.Handle

	startInputDir  string
	startOutputDir string
	settingsPath   string
	firstRun       bool

	iconBig   syscall.Handle
	iconSmall syscall.Handle

	inputEdit    syscall.Handle
	inputBrowse  syscall.Handle
	outputEdit   syscall.Handle
	outputBrowse syscall.Handle
	dateTravel   syscall.Handle
	dateIssue    syscall.Handle
	startButton  syscall.Handle
	logEdit      syscall.Handle

	worker *worker
}

type worker struct {
	logCh  chan string
	doneCh chan workerDone
}

type workerDone struct {
	sum processor.Summary
	err error
}

func newApp(paths startPaths) *app {
	return &app{
		startInputDir:  paths.inputDir,
		startOutputDir: paths.outputDir,
		settingsPath:   paths.settingsPath,
		firstRun:       paths.firstRun,
	}
}

func (a *app) run() error {
	if err := coInitialize(); err != nil {
		return err
	}
	defer coUninitialize()

	className := syscall.StringToUTF16Ptr("InvoiceGuiMainWindow")
	hInstance := getModuleHandle()
	windowProc := syscall.NewCallback(wndProc)

	if err := registerMainWindowClass(className, hInstance, windowProc); err != nil {
		return err
	}

	hwnd, err := createMainWindow(className, hInstance, unsafe.Pointer(a))
	if err != nil {
		return err
	}
	a.hwnd = hwnd

	showWindow(hwnd)
	updateWindow(hwnd)

	return messageLoop()
}

func wndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmNCCreate:
		cs := (*createStruct)(unsafe.Pointer(lParam))
		setWindowUserData(hwnd, uintptr(cs.lpCreateParams))
		return defWindowProc(hwnd, msg, wParam, lParam)
	case wmCreate:
		a := getApp(hwnd)
		if a == nil {
			return defWindowProc(hwnd, msg, wParam, lParam)
		}
		if err := a.onCreate(hwnd); err != nil {
			a.onDestroy()
			showErrorBox("初始化失败", err.Error())
			return wmCreateFailed
		}
		return 0
	case wmCtlColorBtn, wmCtlColorEdit, wmCtlColorStatic:
		return handleCtlColor(wParam)
	case wmCommand:
		a := getApp(hwnd)
		if a == nil {
			return defWindowProc(hwnd, msg, wParam, lParam)
		}
		a.onCommand(hwnd, wParam)
		return 0
	case wmTimer:
		a := getApp(hwnd)
		if a == nil {
			return defWindowProc(hwnd, msg, wParam, lParam)
		}
		a.onTimer(hwnd)
		return 0
	case wmDestroy:
		a := getApp(hwnd)
		if a != nil {
			a.onDestroy()
		}
		killTimer(hwnd, timerID)
		postQuitMessage()
		return 0
	default:
		return defWindowProc(hwnd, msg, wParam, lParam)
	}
}

func getApp(hwnd syscall.Handle) *app {
	ptr := getWindowUserData(hwnd)
	if ptr == 0 {
		return nil
	}
	return (*app)(unsafe.Pointer(ptr))
}
