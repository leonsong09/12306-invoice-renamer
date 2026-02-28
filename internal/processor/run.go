package processor

import (
	"TrainTicketsTool/internal/invoice"
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	pdfExt = ".pdf"
	zipExt = ".zip"
)

func Run(cfg Config, logLine func(string)) (Summary, error) {
	if logLine == nil {
		return Summary{}, errors.New("logLine 不能为空")
	}
	if strings.TrimSpace(cfg.InputDir) == "" {
		return Summary{}, errors.New("未选择输入目录")
	}
	if strings.TrimSpace(cfg.OutputDir) == "" {
		return Summary{}, errors.New("未选择输出目录")
	}

	normalizedCfg, err := normalizeConfigDirs(cfg)
	if err != nil {
		return Summary{}, err
	}
	if sameDir(normalizedCfg.InputDir, normalizedCfg.OutputDir) {
		return Summary{}, errors.New("输出目录不能与输入目录相同")
	}

	skipOutputDir := ""
	if isChildDir(normalizedCfg.OutputDir, normalizedCfg.InputDir) {
		skipOutputDir = normalizedCfg.OutputDir
		logLine(fmt.Sprintf("INFO: 输出目录位于输入目录下，扫描时将跳过输出目录: %s", skipOutputDir))
	}

	if err := EnsureDir(normalizedCfg.OutputDir); err != nil {
		return Summary{}, err
	}

	sum := Summary{}
	r := runner{
		cfg:          normalizedCfg,
		logLine:      logLine,
		sum:          &sum,
		skipDir:      skipOutputDir,
		seenPDFNames: make(map[string]struct{}),
	}
	if err := r.walk(); err != nil {
		return sum, err
	}
	return sum, nil
}

type runner struct {
	cfg          Config
	logLine      func(string)
	sum          *Summary
	skipDir      string
	seenPDFNames map[string]struct{}
}

func (r *runner) walk() error {
	return filepath.WalkDir(r.cfg.InputDir, func(path string, d fs.DirEntry, err error) error {
		return r.onWalk(path, d, err)
	})
}

func (r *runner) onWalk(path string, d fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		r.logLine(fmt.Sprintf("ERR: 访问失败: %s: %v", path, walkErr))
		r.sum.Failed++
		return nil
	}
	if d.IsDir() {
		if r.skipDir != "" && sameDir(path, r.skipDir) {
			return filepath.SkipDir
		}
		return nil
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case pdfExt:
		key := pdfDedupKeyFromFilePath(path)
		if !r.shouldProcessPDFByKey(key) {
			r.logSkipDuplicatePDF(path)
			return nil
		}
		r.sum.FoundPDF++
		if err := r.processPDFFile(path); err != nil {
			r.logLine(fmt.Sprintf("ERR: %s: %v", path, err))
			r.sum.Failed++
			return nil
		}
		r.sum.Succeeded++
	case zipExt:
		if err := r.processZipFile(path); err != nil {
			r.logLine(fmt.Sprintf("ERR: %s: %v", path, err))
			r.sum.Failed++
		}
	default:
		return nil
	}
	return nil
}

func (r *runner) processPDFFile(path string) error {
	pdfBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 PDF 失败: %w", err)
	}
	return r.processPDFBytes(path, pdfBytes)
}

func (r *runner) processPDFBytes(source string, pdfBytes []byte) error {
	info, err := invoice.ExtractInvoiceInfoFromPDFBytes(pdfBytes)
	if err != nil {
		return err
	}
	dateStr, err := pickDate(info, r.cfg.DateField)
	if err != nil {
		return err
	}
	date, err := NormalizeDate(dateStr)
	if err != nil {
		return err
	}

	dep := SanitizeFileNamePart(info.DepartureStation)
	dst := SanitizeFileNamePart(info.DestinationStation)
	if dep == "" || dst == "" {
		return errors.New("站点为空")
	}

	fileName := fmt.Sprintf("%s-%s-%s.pdf", date, dep, dst)
	if err := ValidateOutputFileName(fileName); err != nil {
		return err
	}
	outPath, err := WritePDF(r.cfg.OutputDir, fileName, pdfBytes)
	if err != nil {
		return err
	}
	r.logLine(fmt.Sprintf("OK: %s -> %s", source, outPath))
	return nil
}

func (r *runner) processZipFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开 ZIP 失败: %w", err)
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return fmt.Errorf("读取 ZIP 信息失败: %w", err)
	}
	zr, err := zip.NewReader(f, st.Size())
	if err != nil {
		return fmt.Errorf("解析 ZIP 失败: %w", err)
	}
	return r.processZipReader(path, zr)
}

func (r *runner) processZipReader(ctxPath string, zr *zip.Reader) error {
	for _, entry := range zr.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		if err := r.processZipEntry(ctxPath, entry); err != nil {
			r.sum.Failed++
			r.logLine(fmt.Sprintf("ERR: %s!%s: %v", ctxPath, entry.Name, err))
		}
	}
	return nil
}

func (r *runner) processZipEntry(ctxPath string, entry *zip.File) error {
	switch strings.ToLower(filepath.Ext(entry.Name)) {
	case pdfExt:
		key := pdfDedupKeyFromZipEntryName(entry.Name)
		if !r.shouldProcessPDFByKey(key) {
			source := fmt.Sprintf("%s!%s", ctxPath, entry.Name)
			r.logSkipDuplicatePDF(source)
			return nil
		}
		r.sum.FoundPDF++
		if err := r.processZipPDFEntry(ctxPath, entry); err != nil {
			return err
		}
		r.sum.Succeeded++
		return nil
	case zipExt:
		return r.processZipZipEntry(ctxPath, entry)
	default:
		return nil
	}
}

func (r *runner) processZipPDFEntry(ctxPath string, entry *zip.File) error {
	rc, err := entry.Open()
	if err != nil {
		return fmt.Errorf("打开 ZIP 内 PDF 失败: %w", err)
	}
	defer rc.Close()

	pdfBytes, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("读取 ZIP 内 PDF 失败: %w", err)
	}
	source := fmt.Sprintf("%s!%s", ctxPath, entry.Name)
	return r.processPDFBytes(source, pdfBytes)
}

func (r *runner) processZipZipEntry(ctxPath string, entry *zip.File) error {
	rc, err := entry.Open()
	if err != nil {
		return fmt.Errorf("打开 ZIP 内 ZIP 失败: %w", err)
	}
	defer rc.Close()

	b, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("读取 ZIP 内 ZIP 失败: %w", err)
	}
	nested, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return fmt.Errorf("解析内层 ZIP 失败: %w", err)
	}
	nextCtx := fmt.Sprintf("%s!%s", ctxPath, entry.Name)
	return r.processZipReader(nextCtx, nested)
}

func pickDate(info invoice.InvoiceInfo, field invoice.DateField) (string, error) {
	switch field {
	case invoice.DateFieldTravel:
		if strings.TrimSpace(info.TravelDate) == "" {
			return "", errors.New("缺少乘车日期(TravelDate)")
		}
		return info.TravelDate, nil
	case invoice.DateFieldIssue:
		if strings.TrimSpace(info.DateOfIssue) == "" {
			return "", errors.New("缺少开票日期(DateOfIssue)")
		}
		return info.DateOfIssue, nil
	default:
		return "", errors.New("未知日期字段")
	}
}
