//go:build windows

package gui

import (
	"errors"
	"fmt"
	"os"
	"runtime"
)

func Run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	paths, err := loadStartPaths()
	if err != nil {
		showErrorBox("启动失败", err.Error())
		os.Exit(1)
	}

	app := newApp(paths)
	if err := app.run(); err != nil {
		showErrorBox("启动失败", err.Error())
		os.Exit(1)
	}
}

type startPaths struct {
	inputDir     string
	outputDir    string
	settingsPath string
	firstRun     bool
}

func loadStartPaths() (startPaths, error) {
	exeDir, err := getExeDir()
	if err != nil {
		return startPaths{}, err
	}

	cfgPath := settingsPath(exeDir)
	s, exists, err := loadSettings(cfgPath)
	if err != nil {
		if !askYesNo(0, settingsResetTitle(err), settingsResetMessage(cfgPath, err)) {
			return startPaths{}, err
		}
		if err := os.Remove(cfgPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return startPaths{}, fmt.Errorf("删除配置失败: %w", err)
		}
		exists = false
	}

	if exists {
		return startPaths{
			inputDir:     s.InputDir,
			outputDir:    s.OutputDir,
			settingsPath: cfgPath,
			firstRun:     false,
		}, nil
	}

	def := defaultSettingsForExeDir(exeDir)
	return startPaths{
		inputDir:     def.InputDir,
		outputDir:    def.OutputDir,
		settingsPath: cfgPath,
		firstRun:     true,
	}, nil
}

func settingsResetTitle(err error) string {
	var sErr settingsLoadError
	if !errors.As(err, &sErr) {
		return "读取配置失败"
	}
	switch sErr.kind {
	case settingsLoadParse, settingsLoadInvalid:
		return "配置文件无效"
	default:
		return "读取配置失败"
	}
}

func settingsResetMessage(path string, err error) string {
	title := settingsResetTitle(err)
	return fmt.Sprintf("%s：\n%s\n\n%v\n\n是否删除并重置？", title, path, err)
}
