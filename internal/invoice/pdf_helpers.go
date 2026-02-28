package invoice

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	pdfTokenObj = " obj"
)

func findObjectContentStartBefore(pdfBytes []byte, anchor int) (int, int, error) {
	searchEnd := anchor
	for searchEnd > 0 {
		objIdx := bytes.LastIndex(pdfBytes[:searchEnd], []byte(pdfTokenObj))
		if objIdx < 0 {
			return 0, 0, errors.New("未找到 PDF 对象头")
		}
		lineStart := findLineStart(pdfBytes, objIdx)
		headerEnd := objIdx + len(pdfTokenObj)
		if isObjectHeaderLine(pdfBytes[lineStart:headerEnd]) {
			contentStart := skipWhitespace(pdfBytes, headerEnd)
			return lineStart, contentStart, nil
		}
		searchEnd = lineStart
	}
	return 0, 0, errors.New("未找到 PDF 对象头")
}

func findLineStart(data []byte, before int) int {
	ln := bytes.LastIndexByte(data[:before], '\n')
	cr := bytes.LastIndexByte(data[:before], '\r')
	if ln < 0 && cr < 0 {
		return 0
	}
	if ln > cr {
		return ln + 1
	}
	return cr + 1
}

func isObjectHeaderLine(line []byte) bool {
	s := bytes.TrimSpace(line)
	if !bytes.HasSuffix(s, []byte("obj")) {
		return false
	}
	s = bytes.TrimSpace(s[:len(s)-len("obj")])

	a, next, ok := parseIntAt(s, 0)
	if !ok {
		return false
	}
	next = skipWhitespace(s, next)
	_, next2, ok2 := parseIntAt(s, next)
	if !ok2 {
		return false
	}
	next2 = skipWhitespace(s, next2)
	if next2 != len(s) {
		return false
	}
	return a >= 0
}

func findStreamDataStart(pdfBytes []byte, searchFrom int) (int, int, error) {
	kw := []byte(pdfKeywordStream)
	pos := searchFrom
	for {
		idx := bytes.Index(pdfBytes[pos:], kw)
		if idx < 0 {
			return 0, 0, errors.New("未找到 stream")
		}
		abs := pos + idx
		if isKeywordAt(pdfBytes, abs, kw) {
			dataStart, err := parseStreamDataStart(pdfBytes, abs+len(kw))
			if err != nil {
				return 0, 0, err
			}
			return abs, dataStart, nil
		}
		pos = abs + 1
	}
}

func parseStreamDataStart(pdfBytes []byte, afterKeyword int) (int, error) {
	pos := skipInlineWhitespace(pdfBytes, afterKeyword)
	if pos >= len(pdfBytes) {
		return 0, errors.New("stream 之后缺少换行")
	}
	switch pdfBytes[pos] {
	case '\r':
		pos++
		if pos < len(pdfBytes) && pdfBytes[pos] == '\n' {
			pos++
		}
		return pos, nil
	case '\n':
		return pos + 1, nil
	default:
		return 0, errors.New("stream 之后不是换行")
	}
}

func skipInlineWhitespace(data []byte, i int) int {
	for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
		i++
	}
	return i
}

func isKeywordAt(data []byte, idx int, kw []byte) bool {
	if idx < 0 || idx+len(kw) > len(data) {
		return false
	}
	if !bytes.Equal(data[idx:idx+len(kw)], kw) {
		return false
	}
	if idx > 0 && !isDelimiter(data[idx-1]) {
		return false
	}
	end := idx + len(kw)
	if end < len(data) && !isDelimiter(data[end]) {
		return false
	}
	return true
}

func findNameKey(data []byte, key string) int {
	keyBytes := []byte(key)
	searchFrom := 0
	for {
		idx := bytes.Index(data[searchFrom:], keyBytes)
		if idx < 0 {
			return -1
		}
		abs := searchFrom + idx
		end := abs + len(keyBytes)
		if end == len(data) || isDelimiter(data[end]) {
			return abs
		}
		searchFrom = end
	}
}

func readNameEnd(data []byte, startAfterSlash int) int {
	i := startAfterSlash
	for i < len(data) && !isDelimiter(data[i]) {
		i++
	}
	return i
}

func isWhitespace(b byte) bool {
	switch b {
	case 0x00, 0x09, 0x0A, 0x0C, 0x0D, 0x20:
		return true
	default:
		return false
	}
}

func isDelimiter(b byte) bool {
	if isWhitespace(b) {
		return true
	}
	switch b {
	case '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	default:
		return false
	}
}

func skipWhitespace(data []byte, i int) int {
	for i < len(data) && isWhitespace(data[i]) {
		i++
	}
	return i
}

func parseIntAt(data []byte, pos int) (int, int, bool) {
	i := pos
	if i >= len(data) || data[i] < '0' || data[i] > '9' {
		return 0, pos, false
	}
	v := 0
	for i < len(data) && data[i] >= '0' && data[i] <= '9' {
		v = (v * 10) + int(data[i]-'0')
		i++
	}
	return v, i, true
}

func resolveIndirectIntObject(pdfBytes []byte, ref objectRef) (int, error) {
	pattern := []byte(fmt.Sprintf("%d %d obj", ref.objNum, ref.genNum))
	headerIdx := findObjectHeaderIndex(pdfBytes, pattern)
	if headerIdx < 0 {
		return 0, errors.New("未找到 /Length 的间接对象")
	}

	bodyStart := skipWhitespace(pdfBytes, headerIdx+len(pattern))
	bodyStart = skipWhitespace(pdfBytes, bodyStart)
	v, _, ok := parseIntAt(pdfBytes, bodyStart)
	if !ok {
		return 0, errors.New("间接 /Length 对象不是整数")
	}
	return v, nil
}

func findObjectHeaderIndex(pdfBytes []byte, pattern []byte) int {
	searchFrom := 0
	for {
		idx := bytes.Index(pdfBytes[searchFrom:], pattern)
		if idx < 0 {
			return -1
		}
		abs := searchFrom + idx
		if isObjectHeaderAt(pdfBytes, abs, pattern) {
			return abs
		}
		searchFrom = abs + 1
	}
}

func isObjectHeaderAt(pdfBytes []byte, idx int, pattern []byte) bool {
	if idx > 0 && !isWhitespace(pdfBytes[idx-1]) {
		return false
	}
	end := idx + len(pattern)
	if end < len(pdfBytes) && !isWhitespace(pdfBytes[end]) {
		return false
	}
	return true
}

