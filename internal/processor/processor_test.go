package processor

import (
	"TrainTicketsTool/internal/invoice"
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const (
	xbrlForProcessor = `<xbrl xmlns:rai="urn:rai"><rai:TravelDate>2026-02-24</rai:TravelDate><rai:DateOfIssue>2026-02-28</rai:DateOfIssue><rai:DepartureStation>郑州东</rai:DepartureStation><rai:DestinationStation>三门峡南</rai:DestinationStation></xbrl>`
)

func TestRun_WritesRenamedPDF(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	pdfPath := filepath.Join(inDir, "a.pdf")
	if err := os.WriteFile(pdfPath, buildPlainPDF(xbrlForProcessor), defaultFileMode); err != nil {
		t.Fatalf("write input pdf: %v", err)
	}

	logs := newLogCollector()
	sum, err := Run(Config{
		InputDir:  inDir,
		OutputDir: outDir,
		DateField: invoice.DateFieldTravel,
	}, logs.Add)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if sum.FoundPDF != 1 || sum.Succeeded != 1 || sum.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", sum)
	}

	want := filepath.Join(outDir, "2026-02-24-郑州东-三门峡南.pdf")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestRun_ZipNestedPDF(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	zipPath := filepath.Join(inDir, "multi.zip")
	if err := writeNestedZipWithPDF(zipPath, "inner.zip", "x.pdf", buildPlainPDF(xbrlForProcessor)); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	logs := newLogCollector()
	sum, err := Run(Config{
		InputDir:  inDir,
		OutputDir: outDir,
		DateField: invoice.DateFieldTravel,
	}, logs.Add)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if sum.FoundPDF != 1 || sum.Succeeded != 1 || sum.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", sum)
	}

	want := filepath.Join(outDir, "2026-02-24-郑州东-三门峡南.pdf")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestRun_CollisionAddsSuffix(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(inDir, "a.pdf"), buildPlainPDF(xbrlForProcessor), defaultFileMode); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "b.pdf"), buildPlainPDF(xbrlForProcessor), defaultFileMode); err != nil {
		t.Fatalf("write b: %v", err)
	}

	logs := newLogCollector()
	sum, err := Run(Config{
		InputDir:  inDir,
		OutputDir: outDir,
		DateField: invoice.DateFieldTravel,
	}, logs.Add)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if sum.FoundPDF != 2 || sum.Succeeded != 2 || sum.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", sum)
	}

	first := filepath.Join(outDir, "2026-02-24-郑州东-三门峡南.pdf")
	second := filepath.Join(outDir, "2026-02-24-郑州东-三门峡南-2.pdf")
	if _, err := os.Stat(first); err != nil {
		t.Fatalf("expected output file 1: %v", err)
	}
	if _, err := os.Stat(second); err != nil {
		t.Fatalf("expected output file 2: %v", err)
	}
}

func TestRun_DedupSameNamedPDFInDirs(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	dirA := filepath.Join(inDir, "a")
	dirB := filepath.Join(inDir, "b")
	if err := os.MkdirAll(dirA, defaultDirMode); err != nil {
		t.Fatalf("mkdir a: %v", err)
	}
	if err := os.MkdirAll(dirB, defaultDirMode); err != nil {
		t.Fatalf("mkdir b: %v", err)
	}

	pdfName := "same.pdf"
	if err := os.WriteFile(filepath.Join(dirA, pdfName), buildPlainPDF(xbrlForProcessor), defaultFileMode); err != nil {
		t.Fatalf("write a/same: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dirB, pdfName), buildPlainPDF(xbrlForProcessor), defaultFileMode); err != nil {
		t.Fatalf("write b/same: %v", err)
	}

	logs := newLogCollector()
	sum, err := Run(Config{
		InputDir:  inDir,
		OutputDir: outDir,
		DateField: invoice.DateFieldTravel,
	}, logs.Add)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if sum.FoundPDF != 1 || sum.Succeeded != 1 || sum.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", sum)
	}

	first := filepath.Join(outDir, "2026-02-24-郑州东-三门峡南.pdf")
	if _, err := os.Stat(first); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
	second := filepath.Join(outDir, "2026-02-24-郑州东-三门峡南-2.pdf")
	if _, err := os.Stat(second); err == nil {
		t.Fatalf("did not expect output file 2")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat output file 2: %v", err)
	}
}

func TestRun_DedupSameNamedPDFInZip(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	zipPath := filepath.Join(inDir, "dup.zip")
	pdfBytes := buildPlainPDF(xbrlForProcessor)
	if err := writeZipWithEntries(zipPath, []zipEntry{
		{name: "x/same.pdf", bytes: pdfBytes},
		{name: "y/same.pdf", bytes: pdfBytes},
	}); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	logs := newLogCollector()
	sum, err := Run(Config{
		InputDir:  inDir,
		OutputDir: outDir,
		DateField: invoice.DateFieldTravel,
	}, logs.Add)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if sum.FoundPDF != 1 || sum.Succeeded != 1 || sum.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", sum)
	}

	first := filepath.Join(outDir, "2026-02-24-郑州东-三门峡南.pdf")
	if _, err := os.Stat(first); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
	second := filepath.Join(outDir, "2026-02-24-郑州东-三门峡南-2.pdf")
	if _, err := os.Stat(second); err == nil {
		t.Fatalf("did not expect output file 2")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat output file 2: %v", err)
	}
}

func buildPlainPDF(xbrl string) []byte {
	return []byte("%PDF-1.7\nstream\n" + xbrl + "\nendstream\n%%EOF\n")
}

type zipEntry struct {
	name  string
	bytes []byte
}

func writeZipWithEntries(zipPath string, entries []zipEntry) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(zipFile)
	for _, entry := range entries {
		entryWriter, err := zipWriter.Create(entry.name)
		if err != nil {
			_ = zipWriter.Close()
			_ = zipFile.Close()
			return err
		}
		if _, err := entryWriter.Write(entry.bytes); err != nil {
			_ = zipWriter.Close()
			_ = zipFile.Close()
			return err
		}
	}

	if err := zipWriter.Close(); err != nil {
		_ = zipFile.Close()
		return err
	}
	return zipFile.Close()
}

func writeNestedZipWithPDF(zipPath string, innerZipName string, pdfName string, pdfBytes []byte) error {
	var inner bytes.Buffer
	innerW := zip.NewWriter(&inner)
	f, err := innerW.Create(pdfName)
	if err != nil {
		return err
	}
	if _, err := f.Write(pdfBytes); err != nil {
		return err
	}
	if err := innerW.Close(); err != nil {
		return err
	}

	outFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	outerW := zip.NewWriter(outFile)
	zf, err := outerW.Create(innerZipName)
	if err != nil {
		return err
	}
	if _, err := zf.Write(inner.Bytes()); err != nil {
		return err
	}
	return outerW.Close()
}

type logCollector struct {
	lines []string
}

func newLogCollector() *logCollector {
	return &logCollector{lines: []string{}}
}

func (c *logCollector) Add(s string) {
	c.lines = append(c.lines, s)
}
