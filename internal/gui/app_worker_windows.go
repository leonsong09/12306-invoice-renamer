//go:build windows

package gui

import (
	"TrainTicketsTool/internal/invoice"
	"TrainTicketsTool/internal/processor"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func runWorker(cfg processor.Config, w *worker) {
	start := time.Now()
	sum, err := processor.Run(cfg, func(line string) {
		w.logCh <- line
	})
	w.logCh <- fmt.Sprintf("用时：%s", time.Since(start).Round(time.Millisecond))
	close(w.logCh)
	w.doneCh <- workerDone{sum: sum, err: err}
}

func buildConfig(inputDir string, outputDir string, useTravelDate bool) (processor.Config, error) {
	if err := validateRequiredText(inputDir, "输入目录"); err != nil {
		return processor.Config{}, err
	}
	if err := validateRequiredText(outputDir, "输出目录"); err != nil {
		return processor.Config{}, err
	}

	inAbs, err := filepath.Abs(strings.TrimSpace(inputDir))
	if err != nil {
		return processor.Config{}, fmt.Errorf("输入目录无效: %w", err)
	}
	outAbs, err := filepath.Abs(strings.TrimSpace(outputDir))
	if err != nil {
		return processor.Config{}, fmt.Errorf("输出目录无效: %w", err)
	}

	field := invoice.DateFieldIssue
	if useTravelDate {
		field = invoice.DateFieldTravel
	}

	if field != invoice.DateFieldTravel && field != invoice.DateFieldIssue {
		return processor.Config{}, errors.New("未知日期字段")
	}

	return processor.Config{
		InputDir:  inAbs,
		OutputDir: outAbs,
		DateField: field,
	}, nil
}

