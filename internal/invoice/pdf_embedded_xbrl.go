package invoice

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	pdfKeyType   = "/Type"
	pdfNameEmbed = "/EmbeddedFile"
	pdfKeyLength = "/Length"

	pdfKeywordStream = "stream"
)

func extractEmbeddedXbrl(pdfBytes []byte) ([]byte, error) {
	typePositions := findEmbeddedFileTypePositions(pdfBytes)
	if len(typePositions) == 0 {
		return nil, errors.New("未在 PDF 中找到可解析的 XBRL 数据")
	}

	var firstErr error
	for _, typePos := range typePositions {
		xbrl, err := extractEmbeddedFileXbrlAt(pdfBytes, typePos)
		if err != nil {
			firstErr = firstNonNil(firstErr, err)
			continue
		}
		return xbrl, nil
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return nil, errNoEmbeddedXbrl
}

func findEmbeddedFileTypePositions(pdfBytes []byte) []int {
	var positions []int
	searchFrom := 0
	for {
		idx := bytes.Index(pdfBytes[searchFrom:], []byte(pdfKeyType))
		if idx < 0 {
			return positions
		}
		abs := searchFrom + idx
		nameStart := skipWhitespace(pdfBytes, abs+len(pdfKeyType))
		if nameStart < len(pdfBytes) && pdfBytes[nameStart] == '/' {
			nameEnd := readNameEnd(pdfBytes, nameStart+1)
			if bytes.Equal(pdfBytes[nameStart:nameEnd], []byte(pdfNameEmbed)) {
				positions = append(positions, abs)
			}
		}
		searchFrom = abs + len(pdfKeyType)
	}
}

func extractEmbeddedFileXbrlAt(pdfBytes []byte, typePos int) ([]byte, error) {
	_, contentStart, err := findObjectContentStartBefore(pdfBytes, typePos)
	if err != nil {
		return nil, err
	}

	streamKw, streamDataStart, err := findStreamDataStart(pdfBytes, typePos)
	if err != nil {
		return nil, err
	}
	if streamKw <= contentStart {
		return nil, errors.New("PDF stream 位置异常")
	}

	spec, err := parseLengthSpec(pdfBytes[contentStart:streamKw])
	if err != nil {
		return nil, err
	}
	length, err := resolveLength(pdfBytes, spec)
	if err != nil {
		return nil, err
	}
	streamEnd := streamDataStart + length
	if streamEnd > len(pdfBytes) {
		return nil, fmt.Errorf("embedded stream 越界: start=%d len=%d pdf=%d", streamDataStart, length, len(pdfBytes))
	}

	raw, err := decompressFlate(pdfBytes[streamDataStart:streamEnd])
	if err != nil {
		return nil, err
	}
	if xbrl, ok := extractPlainXbrl(raw); ok {
		return xbrl, nil
	}
	return nil, errNoEmbeddedXbrl
}

type lengthSpec struct {
	direct     int
	indirect   objectRef
	isIndirect bool
}

type objectRef struct {
	objNum int
	genNum int
}

func parseLengthSpec(dictBytes []byte) (lengthSpec, error) {
	idx := findNameKey(dictBytes, pdfKeyLength)
	if idx < 0 {
		return lengthSpec{}, errors.New("未找到 EmbeddedFile 的 /Length")
	}

	pos := skipWhitespace(dictBytes, idx+len(pdfKeyLength))
	if pos >= len(dictBytes) {
		return lengthSpec{}, errors.New("EmbeddedFile /Length 缺少值")
	}

	objNum, next, ok := parseIntAt(dictBytes, pos)
	if !ok {
		return lengthSpec{}, errors.New("EmbeddedFile /Length 值不是整数")
	}

	genPos := skipWhitespace(dictBytes, next)
	genNum, next2, ok2 := parseIntAt(dictBytes, genPos)
	if ok2 {
		rPos := skipWhitespace(dictBytes, next2)
		if rPos < len(dictBytes) && dictBytes[rPos] == 'R' {
			return lengthSpec{
				indirect:   objectRef{objNum: objNum, genNum: genNum},
				isIndirect: true,
			}, nil
		}
	}

	return lengthSpec{direct: objNum}, nil
}

func resolveLength(pdfBytes []byte, spec lengthSpec) (int, error) {
	if !spec.isIndirect {
		return spec.direct, nil
	}
	return resolveIndirectIntObject(pdfBytes, spec.indirect)
}

