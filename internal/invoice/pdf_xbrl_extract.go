package invoice

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
)

const (
	xbrlStartToken = "<xbrl"
	xbrlEndToken   = "</xbrl>"
)

func ExtractXbrlFromPDFBytes(pdfBytes []byte) ([]byte, error) {
	if xbrl, ok := extractPlainXbrl(pdfBytes); ok {
		return xbrl, nil
	}
	return extractEmbeddedXbrl(pdfBytes)
}

func extractPlainXbrl(pdfBytes []byte) ([]byte, bool) {
	start := bytes.Index(pdfBytes, []byte(xbrlStartToken))
	if start < 0 {
		return nil, false
	}
	rest := pdfBytes[start:]
	endRel := bytes.Index(rest, []byte(xbrlEndToken))
	if endRel < 0 {
		return nil, false
	}
	end := start + endRel + len(xbrlEndToken)
	return pdfBytes[start:end], true
}

func decompressFlate(data []byte) ([]byte, error) {
	if out, err := decompressZlib(data); err == nil {
		return out, nil
	}
	return decompressRawDeflate(data)
}

func decompressZlib(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("zlib 解压失败: %w", err)
	}
	defer r.Close()
	return io.ReadAll(r)
}

func decompressRawDeflate(data []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("deflate 解压失败: %w", err)
	}
	return out, nil
}

func firstNonNil(a, b error) error {
	if a != nil {
		return a
	}
	return b
}

var errNoEmbeddedXbrl = errors.New("未在 PDF EmbeddedFile 中找到 XBRL")

