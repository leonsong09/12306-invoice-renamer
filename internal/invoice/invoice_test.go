package invoice

import (
	"bytes"
	"compress/zlib"
	"strconv"
	"testing"
)

const (
	testXbrlXML = `<xbrl xmlns:rai="urn:rai"><rai:TravelDate>2026-02-11</rai:TravelDate><rai:DateOfIssue>2026-02-28</rai:DateOfIssue><rai:DepartureStation>三门峡南</rai:DepartureStation><rai:DestinationStation>郑州</rai:DestinationStation></xbrl>`
)

func TestExtractInvoiceInfoFromPDFBytes_PlainXbrl(t *testing.T) {
	pdf := []byte("%PDF-1.7\nstream\n" + testXbrlXML + "\nendstream\n%%EOF\n")
	info, err := ExtractInvoiceInfoFromPDFBytes(pdf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertInfo(t, info)
}

func TestExtractInvoiceInfoFromPDFBytes_EmbeddedFlateXbrl(t *testing.T) {
	comp := compressZlib([]byte(testXbrlXML))
	length := strconv.Itoa(len(comp))

	pdf := bytes.Join([][]byte{
		[]byte("%PDF-1.7\r\n"),
		[]byte("44 0 obj\r\n<</Filter/FlateDecode/Length "),
		[]byte(length),
		[]byte("/Type/EmbeddedFile>>stream\r\n"),
		comp,
		[]byte("\r\nendstream\r\nendobj\r\n%%EOF\r\n"),
	}, nil)

	info, err := ExtractInvoiceInfoFromPDFBytes(pdf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertInfo(t, info)
}

func TestExtractInvoiceInfoFromPDFBytes_EmbeddedFlateXbrl_IgnoresOtherLength(t *testing.T) {
	comp := compressZlib([]byte(testXbrlXML))
	length := strconv.Itoa(len(comp))

	pdf := bytes.Join([][]byte{
		[]byte("%PDF-1.7\r\n"),
		[]byte("10 0 obj\r\n<</Length 3923>>stream\r\nx\r\nendstream\r\nendobj\r\n"),
		[]byte("44 0 obj\r\n<</Filter/FlateDecode/Length "),
		[]byte(length),
		[]byte("/Type/EmbeddedFile>>stream\r\n"),
		comp,
		[]byte("\r\nendstream\r\nendobj\r\n%%EOF\r\n"),
	}, nil)

	info, err := ExtractInvoiceInfoFromPDFBytes(pdf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertInfo(t, info)
}

func compressZlib(in []byte) []byte {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(in); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func assertInfo(t *testing.T, info InvoiceInfo) {
	t.Helper()
	if info.TravelDate != "2026-02-11" {
		t.Fatalf("TravelDate=%q", info.TravelDate)
	}
	if info.DateOfIssue != "2026-02-28" {
		t.Fatalf("DateOfIssue=%q", info.DateOfIssue)
	}
	if info.DepartureStation != "三门峡南" {
		t.Fatalf("DepartureStation=%q", info.DepartureStation)
	}
	if info.DestinationStation != "郑州" {
		t.Fatalf("DestinationStation=%q", info.DestinationStation)
	}
}
