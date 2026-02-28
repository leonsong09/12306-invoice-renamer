package processor

import (
	"path"
	"path/filepath"
	"strings"
)

func pdfDedupKeyFromFilePath(filePath string) string {
	return normalizePDFDedupKey(filepath.Base(filePath))
}

func pdfDedupKeyFromZipEntryName(entryName string) string {
	return normalizePDFDedupKey(path.Base(entryName))
}

func normalizePDFDedupKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func (r *runner) shouldProcessPDFByKey(key string) bool {
	if key == "" {
		return true
	}
	if _, exists := r.seenPDFNames[key]; exists {
		return false
	}
	r.seenPDFNames[key] = struct{}{}
	return true
}

func (r *runner) logSkipDuplicatePDF(source string) {
	r.logLine("SKIP: 重名 PDF（按文件名去重）: " + source)
}
