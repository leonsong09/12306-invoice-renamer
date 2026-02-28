//go:build windows

package gui

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
)

const (
	defaultDirMode = 0o755
)

func ensureDirExistsOrPromptCreate(owner syscall.Handle, label string, path string) error {
	dir := strings.TrimSpace(path)
	if dir == "" {
		return errors.New(label + "为空")
	}

	st, err := os.Stat(dir)
	if err == nil {
		if !st.IsDir() {
			return fmt.Errorf("%s不是文件夹: %s", label, dir)
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取%s失败: %w", label, err)
	}

	msg := fmt.Sprintf("%s不存在：\n%s\n\n是否创建？", label, dir)
	if !askYesNo(owner, "创建文件夹", msg) {
		return fmt.Errorf("%s不存在: %s", label, dir)
	}
	if err := os.MkdirAll(dir, defaultDirMode); err != nil {
		return fmt.Errorf("创建%s失败: %w", label, err)
	}
	return nil
}

func promptCreateMissingDefaultDirs(owner syscall.Handle, inputDir string, outputDir string) error {
	missing, err := collectMissingDirs(inputDir, outputDir)
	if err != nil {
		return err
	}
	if len(missing) == 0 {
		return nil
	}

	msg := buildMissingDirsMessage(missing)
	if !askYesNo(owner, "创建文件夹", msg) {
		return nil
	}
	for _, dir := range missing {
		if err := os.MkdirAll(dir.path, defaultDirMode); err != nil {
			return fmt.Errorf("创建%s失败: %w", dir.label, err)
		}
	}
	return nil
}

type labeledDir struct {
	label string
	path  string
}

func collectMissingDirs(inputDir string, outputDir string) ([]labeledDir, error) {
	in, err := missingDir("输入目录", inputDir)
	if err != nil {
		return nil, err
	}
	out, err := missingDir("输出目录", outputDir)
	if err != nil {
		return nil, err
	}
	var missing []labeledDir
	if in != nil {
		missing = append(missing, *in)
	}
	if out != nil {
		missing = append(missing, *out)
	}
	return missing, nil
}

func missingDir(label string, path string) (*labeledDir, error) {
	dir := strings.TrimSpace(path)
	if dir == "" {
		return nil, nil
	}
	st, err := os.Stat(dir)
	if err == nil {
		if !st.IsDir() {
			return nil, fmt.Errorf("%s不是文件夹: %s", label, dir)
		}
		return nil, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return &labeledDir{label: label, path: dir}, nil
	}
	return nil, fmt.Errorf("读取%s失败: %w", label, err)
}

func buildMissingDirsMessage(missing []labeledDir) string {
	lines := []string{"默认目录不存在："}
	for _, d := range missing {
		lines = append(lines, fmt.Sprintf("- %s：%s", d.label, d.path))
	}
	lines = append(lines, "", "是否创建？")
	return strings.Join(lines, "\n")
}
