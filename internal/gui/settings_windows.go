//go:build windows

package gui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const (
	settingsFileName   = "settings.json"
	settingsTempSuffix = ".tmp"

	defaultInputDirName  = "input"
	defaultOutputDirName = "output"

	settingsFileMode = 0o644
)

type settingsV1 struct {
	InputDir  string `json:"inputDir"`
	OutputDir string `json:"outputDir"`
}

type settingsLoadErrorKind int

const (
	settingsLoadRead settingsLoadErrorKind = iota + 1
	settingsLoadParse
	settingsLoadInvalid
)

type settingsLoadError struct {
	kind settingsLoadErrorKind
	err  error
}

func (e settingsLoadError) Error() string {
	return e.err.Error()
}

func (e settingsLoadError) Unwrap() error {
	return e.err
}

func defaultSettingsForExeDir(exeDir string) settingsV1 {
	return settingsV1{
		InputDir:  filepath.Join(exeDir, defaultInputDirName),
		OutputDir: filepath.Join(exeDir, defaultOutputDirName),
	}
}

func getExeDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取 exe 路径失败: %w", err)
	}
	abs, err := filepath.Abs(exePath)
	if err != nil {
		return "", fmt.Errorf("解析 exe 绝对路径失败: %w", err)
	}
	return filepath.Dir(abs), nil
}

func settingsPath(exeDir string) string {
	return filepath.Join(exeDir, settingsFileName)
}

func loadSettings(path string) (settingsV1, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return settingsV1{}, false, nil
		}
		return settingsV1{}, true, settingsLoadError{kind: settingsLoadRead, err: fmt.Errorf("读取配置失败: %w", err)}
	}

	var s settingsV1
	if err := json.Unmarshal(b, &s); err != nil {
		return settingsV1{}, true, settingsLoadError{kind: settingsLoadParse, err: fmt.Errorf("解析配置失败: %w", err)}
	}
	normalized, err := normalizeSettings(s)
	if err != nil {
		return settingsV1{}, true, settingsLoadError{kind: settingsLoadInvalid, err: err}
	}
	return normalized, true, nil
}

func saveSettingsAtomic(path string, s settingsV1) error {
	normalized, err := normalizeSettings(s)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	b = append(b, '\n')

	tmp := path + settingsTempSuffix
	if err := os.WriteFile(tmp, b, settingsFileMode); err != nil {
		return fmt.Errorf("写入临时配置失败: %w", err)
	}
	return replaceFile(tmp, path)
}

func normalizeSettings(s settingsV1) (settingsV1, error) {
	in := strings.TrimSpace(s.InputDir)
	out := strings.TrimSpace(s.OutputDir)
	if in == "" || out == "" {
		return settingsV1{}, errors.New("配置缺少 inputDir/outputDir")
	}
	inAbs, err := filepath.Abs(in)
	if err != nil {
		return settingsV1{}, fmt.Errorf("输入目录无效: %w", err)
	}
	outAbs, err := filepath.Abs(out)
	if err != nil {
		return settingsV1{}, fmt.Errorf("输出目录无效: %w", err)
	}
	return settingsV1{
		InputDir:  inAbs,
		OutputDir: outAbs,
	}, nil
}

func replaceFile(tmpPath string, destPath string) error {
	src, err := syscall.UTF16PtrFromString(tmpPath)
	if err != nil {
		return fmt.Errorf("转换临时配置路径失败: %w", err)
	}
	dst, err := syscall.UTF16PtrFromString(destPath)
	if err != nil {
		return fmt.Errorf("转换配置路径失败: %w", err)
	}

	flags := moveFileReplaceExisting | moveFileWriteThrough
	r1, _, callErr := procMoveFileExW.Call(
		uintptr(unsafe.Pointer(src)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(flags),
	)
	if r1 == 0 {
		return fmt.Errorf("写入配置失败: %w", callErr)
	}
	return nil
}
