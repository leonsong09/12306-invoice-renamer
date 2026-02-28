//go:build windows

package gui

import (
	"os"
	"runtime"
)

func Run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	app := newApp(defaultStartPaths())
	if err := app.run(); err != nil {
		showErrorBox("启动失败", err.Error())
		os.Exit(1)
	}
}

type startPaths struct {
	inputDir  string
	outputDir string
}

func defaultStartPaths() startPaths {
	wd, err := os.Getwd()
	if err != nil {
		return startPaths{}
	}
	return startPaths{
		inputDir:  wd,
		outputDir: wd,
	}
}

