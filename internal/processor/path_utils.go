package processor

import (
	"fmt"
	"path/filepath"
	"strings"
)

func normalizeConfigDirs(cfg Config) (Config, error) {
	inAbs, err := absClean(cfg.InputDir)
	if err != nil {
		return Config{}, fmt.Errorf("输入目录无效: %w", err)
	}
	outAbs, err := absClean(cfg.OutputDir)
	if err != nil {
		return Config{}, fmt.Errorf("输出目录无效: %w", err)
	}
	return Config{
		InputDir:  inAbs,
		OutputDir: outAbs,
		DateField: cfg.DateField,
	}, nil
}

func absClean(path string) (string, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return "", fmt.Errorf("路径为空")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func sameDir(a string, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func isChildDir(child string, parent string) bool {
	c := strings.ToLower(filepath.Clean(child))
	p := strings.ToLower(filepath.Clean(parent))
	if c == p {
		return false
	}
	sep := string(filepath.Separator)
	return strings.HasPrefix(c, p+sep)
}

