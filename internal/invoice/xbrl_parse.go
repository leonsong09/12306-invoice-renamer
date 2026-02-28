package invoice

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

var errMissingRequiredField = errors.New("XBRL 缺少必要字段")

func ExtractInvoiceInfoFromPDFBytes(pdfBytes []byte) (InvoiceInfo, error) {
	xbrl, err := ExtractXbrlFromPDFBytes(pdfBytes)
	if err != nil {
		return InvoiceInfo{}, err
	}
	return ParseInvoiceInfoFromXbrl(xbrl)
}

func ParseInvoiceInfoFromXbrl(xbrlBytes []byte) (InvoiceInfo, error) {
	dec := xml.NewDecoder(bytes.NewReader(xbrlBytes))
	info := InvoiceInfo{}

	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return InvoiceInfo{}, fmt.Errorf("解析 XBRL 失败: %w", err)
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if !isWantedTag(start.Name.Local) {
			continue
		}

		var text string
		if err := dec.DecodeElement(&text, &start); err != nil {
			return InvoiceInfo{}, fmt.Errorf("读取字段 %s 失败: %w", start.Name.Local, err)
		}
		setIfEmpty(&info, start.Name.Local, strings.TrimSpace(text))
	}

	if strings.TrimSpace(info.DepartureStation) == "" || strings.TrimSpace(info.DestinationStation) == "" {
		return InvoiceInfo{}, errMissingRequiredField
	}
	if strings.TrimSpace(info.TravelDate) == "" && strings.TrimSpace(info.DateOfIssue) == "" {
		return InvoiceInfo{}, errMissingRequiredField
	}
	return info, nil
}

func isWantedTag(local string) bool {
	switch local {
	case "TravelDate", "DateOfIssue", "DepartureStation", "DestinationStation":
		return true
	default:
		return false
	}
}

func setIfEmpty(info *InvoiceInfo, local string, value string) {
	if value == "" {
		return
	}
	switch local {
	case "TravelDate":
		if info.TravelDate == "" {
			info.TravelDate = value
		}
	case "DateOfIssue":
		if info.DateOfIssue == "" {
			info.DateOfIssue = value
		}
	case "DepartureStation":
		if info.DepartureStation == "" {
			info.DepartureStation = value
		}
	case "DestinationStation":
		if info.DestinationStation == "" {
			info.DestinationStation = value
		}
	}
}

