//go:build windows

package gui

import (
	"TrainTicketsTool/internal/processor"
	"errors"
	"fmt"
	"strings"
	"syscall"
)

const (
	uiMargin    int32 = 12
	uiLabelW    int32 = 80
	uiEditW     int32 = 560
	uiBrowseW   int32 = 110
	uiRowH      int32 = 26
	uiRowGap    int32 = 10
	uiWindowH   int32 = 520
	uiLogH      int32 = 360
	uiTopStartY int32 = 14
	uiGapSmall  int32 = 8
	uiRadioW    int32 = 220
	uiRadioGap  int32 = 240
	uiStartBtnW int32 = 120
	uiLogLabelY int32 = 4
)

var uiWindowW int32 = uiMargin + uiLabelW + uiGapSmall + uiEditW + uiGapSmall + uiBrowseW + uiMargin

func (a *app) onCreate(hwnd syscall.Handle) error {
	if err := a.initIcons(hwnd); err != nil {
		return err
	}
	setFixedWindowSize(hwnd, uiWindowW, uiWindowH)

	y := uiTopStartY
	y = a.createInputRow(hwnd, y)
	y = a.createOutputRow(hwnd, y)
	y = a.createDateRow(hwnd, y)
	y = a.createStartRow(hwnd, y)
	a.createLogArea(hwnd, y)

	a.applyInitialPaths()
	if err := a.handleFirstRun(hwnd); err != nil {
		return err
	}
	return nil
}

func (a *app) createInputRow(hwnd syscall.Handle, y int32) int32 {
	createStatic(hwnd, "输入目录：", uiMargin, y+uiLogLabelY, uiLabelW, uiRowH)
	a.inputEdit = createEdit(hwnd, idInputEdit, uiMargin+uiLabelW+uiGapSmall, y, uiEditW, uiRowH)
	a.inputBrowse = createButton(hwnd, idInputBrowse, "选择…", uiMargin+uiLabelW+uiGapSmall+uiEditW+uiGapSmall, y, uiBrowseW, uiRowH)
	return y + uiRowH + uiRowGap
}

func (a *app) createOutputRow(hwnd syscall.Handle, y int32) int32 {
	createStatic(hwnd, "输出目录：", uiMargin, y+uiLogLabelY, uiLabelW, uiRowH)
	a.outputEdit = createEdit(hwnd, idOutputEdit, uiMargin+uiLabelW+uiGapSmall, y, uiEditW, uiRowH)
	a.outputBrowse = createButton(hwnd, idOutputBrowse, "选择…", uiMargin+uiLabelW+uiGapSmall+uiEditW+uiGapSmall, y, uiBrowseW, uiRowH)
	return y + uiRowH + uiRowGap
}

func (a *app) createDateRow(hwnd syscall.Handle, y int32) int32 {
	createStatic(hwnd, "日期字段：", uiMargin, y+uiLogLabelY, uiLabelW, uiRowH)
	a.dateTravel = createRadio(hwnd, idDateTravel, "乘车日期(TravelDate)", uiMargin+uiLabelW+uiGapSmall, y, uiRadioW, uiRowH, true)
	a.dateIssue = createRadio(hwnd, idDateIssue, "开票日期(DateOfIssue)", uiMargin+uiLabelW+uiGapSmall+uiRadioGap, y, uiRadioW, uiRowH, false)
	return y + uiRowH + uiRowGap
}

func (a *app) createStartRow(hwnd syscall.Handle, y int32) int32 {
	a.startButton = createButton(hwnd, idStartButton, "开始处理", uiMargin, y, uiStartBtnW, uiRowH)
	createStatic(hwnd, "日志：", uiMargin, y+uiRowH+uiRowGap+uiLogLabelY, uiLabelW, uiRowH)
	return y + uiRowH + uiRowGap + uiRowH
}

func (a *app) createLogArea(hwnd syscall.Handle, y int32) {
	a.logEdit = createLogEdit(hwnd, idLogEdit, uiMargin, y, uiWindowW-(uiMargin*2), uiLogH)
}

func (a *app) applyInitialPaths() {
	if strings.TrimSpace(a.startInputDir) != "" {
		setWindowText(a.inputEdit, a.startInputDir)
	}
	if strings.TrimSpace(a.startOutputDir) != "" {
		setWindowText(a.outputEdit, a.startOutputDir)
	}
}

func (a *app) onCommand(hwnd syscall.Handle, wParam uintptr) {
	switch loword(wParam) {
	case idInputBrowse:
		if dir, ok := browseForFolder(hwnd, "选择发票所在文件夹"); ok {
			setWindowText(a.inputEdit, dir)
			a.trySaveSettingsFromEdits(hwnd)
		}
	case idOutputBrowse:
		if dir, ok := browseForFolder(hwnd, "选择输出文件夹"); ok {
			setWindowText(a.outputEdit, dir)
			a.trySaveSettingsFromEdits(hwnd)
		}
	case idStartButton:
		a.startProcessing(hwnd)
	default:
		return
	}
}

func (a *app) startProcessing(hwnd syscall.Handle) {
	if a.worker != nil {
		return
	}

	cfg, err := a.buildConfigFromEdits()
	if err != nil {
		return
	}
	if err := a.ensureDirsForRun(hwnd, cfg); err != nil {
		showErrorBox("目录错误", err.Error())
		return
	}
	a.saveSettingsFromConfig(cfg)

	clearLog(a.logEdit)
	disableControls(a, true)

	w := &worker{
		logCh:  make(chan string, logBufferSize),
		doneCh: make(chan workerDone, 1),
	}
	a.worker = w

	setTimer(hwnd, timerID, uiPollIntervalMs)
	go runWorker(cfg, w)
}

func (a *app) onTimer(hwnd syscall.Handle) {
	if a.worker == nil {
		return
	}
	drainLog(a.logEdit, a.worker.logCh)

	select {
	case done := <-a.worker.doneCh:
		killTimer(hwnd, timerID)
		drainLog(a.logEdit, a.worker.logCh)
		a.worker = nil
		disableControls(a, false)
		a.showDone(done)
	default:
		return
	}
}

func (a *app) showDone(done workerDone) {
	if done.err != nil {
		showErrorBox("处理失败", done.err.Error())
		return
	}
	msg := fmt.Sprintf("完成。\n发现 PDF：%d\n成功：%d\n失败：%d", done.sum.FoundPDF, done.sum.Succeeded, done.sum.Failed)
	showInfoBox("处理完成", msg)
}

func validateRequiredText(value string, field string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(field + "为空")
	}
	return nil
}

func (a *app) handleFirstRun(hwnd syscall.Handle) error {
	if !a.firstRun {
		return nil
	}
	a.firstRun = false

	if err := promptCreateMissingDefaultDirs(hwnd, a.startInputDir, a.startOutputDir); err != nil {
		return err
	}
	if err := saveSettingsAtomic(a.settingsPath, settingsV1{InputDir: a.startInputDir, OutputDir: a.startOutputDir}); err != nil {
		showErrorBox("保存设置失败", err.Error())
	}
	return nil
}

func (a *app) buildConfigFromEdits() (processor.Config, error) {
	inputDir, err := getWindowText(a.inputEdit)
	if err != nil {
		showErrorBox("读取输入目录失败", err.Error())
		return processor.Config{}, err
	}
	outputDir, err := getWindowText(a.outputEdit)
	if err != nil {
		showErrorBox("读取输出目录失败", err.Error())
		return processor.Config{}, err
	}
	cfg, err := buildConfig(inputDir, outputDir, isChecked(a.dateTravel))
	if err != nil {
		showErrorBox("参数错误", err.Error())
		return processor.Config{}, err
	}
	return cfg, nil
}

func (a *app) ensureDirsForRun(hwnd syscall.Handle, cfg processor.Config) error {
	if err := ensureDirExistsOrPromptCreate(hwnd, "输入目录", cfg.InputDir); err != nil {
		return err
	}
	return ensureDirExistsOrPromptCreate(hwnd, "输出目录", cfg.OutputDir)
}

func (a *app) saveSettingsFromConfig(cfg processor.Config) {
	if err := saveSettingsAtomic(a.settingsPath, settingsV1{InputDir: cfg.InputDir, OutputDir: cfg.OutputDir}); err != nil {
		showErrorBox("保存设置失败", err.Error())
	}
}

func (a *app) trySaveSettingsFromEdits(hwnd syscall.Handle) {
	inputDir, err := getWindowText(a.inputEdit)
	if err != nil {
		showErrorBox("读取输入目录失败", err.Error())
		return
	}
	outputDir, err := getWindowText(a.outputEdit)
	if err != nil {
		showErrorBox("读取输出目录失败", err.Error())
		return
	}
	if strings.TrimSpace(inputDir) == "" || strings.TrimSpace(outputDir) == "" {
		return
	}

	cfg, err := buildConfig(inputDir, outputDir, isChecked(a.dateTravel))
	if err != nil {
		showErrorBox("参数错误", err.Error())
		return
	}
	a.saveSettingsFromConfig(cfg)
}
