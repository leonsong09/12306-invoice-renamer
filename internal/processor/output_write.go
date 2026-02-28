package processor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultFileMode    = 0o644
	defaultDirMode     = 0o755
	firstCollisionNum  = 2
)

func EnsureDir(path string) error {
	return os.MkdirAll(path, defaultDirMode)
}

func WritePDF(outputDir string, fileName string, pdfBytes []byte) (string, error) {
	if err := EnsureDir(outputDir); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	outPath, err := UniqueOutputPath(outputDir, fileName)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(outPath, pdfBytes, defaultFileMode); err != nil {
		return "", fmt.Errorf("写入输出文件失败: %w", err)
	}
	return outPath, nil
}

func UniqueOutputPath(outputDir string, fileName string) (string, error) {
	basePath := filepath.Join(outputDir, fileName)
	exists, err := fileExists(basePath)
	if err != nil {
		return "", err
	}
	if !exists {
		return basePath, nil
	}

	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	for i := firstCollisionNum; ; i++ {
		tryName := fmt.Sprintf("%s-%d%s", base, i, ext)
		tryPath := filepath.Join(outputDir, tryName)
		exists, err := fileExists(tryPath)
		if err != nil {
			return "", err
		}
		if !exists {
			return tryPath, nil
		}
	}
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func ValidateOutputFileName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("输出文件名为空")
	}
	if filepath.Base(name) != name {
		return errors.New("输出文件名不能包含路径分隔符")
	}
	return nil
}
