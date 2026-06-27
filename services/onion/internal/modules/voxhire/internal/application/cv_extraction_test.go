package application

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
)

type fakeCVStorage struct {
	body []byte
	size int64
	err  error
}

func (s fakeCVStorage) Download(context.Context, string) (io.ReadCloser, int64, error) {
	if s.err != nil {
		return nil, 0, s.err
	}
	return io.NopCloser(bytes.NewReader(s.body)), s.size, nil
}

func TestProcessCVTargetExtractsDOCXText(t *testing.T) {
	raw := newTestDOCX(t, "Retail operations manager with payroll and inventory experience.")
	svc := &Service{}

	result := svc.processCVTarget(context.Background(), fakeCVStorage{body: raw, size: int64(len(raw))}, domain.CVProcessingTarget{
		Source: "talent",
		ID:     7,
		CVKey:  "voxhire/applications/operations/resume.docx",
	})

	if result.Status != "processed" {
		t.Fatalf("status = %q, error = %q", result.Status, result.Error)
	}
	if !strings.Contains(result.ExtractedText, "Retail operations manager") {
		t.Fatalf("extracted text = %q", result.ExtractedText)
	}
	if !strings.Contains(result.Summary, "payroll") {
		t.Fatalf("summary = %q", result.Summary)
	}
}

func TestTruncateBytesDoesNotSplitRunes(t *testing.T) {
	got := truncateBytes("abcédef", 4)
	if got != "abc" {
		t.Fatalf("truncateBytes split a rune: %q", got)
	}
}

func newTestDOCX(t *testing.T, text string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatalf("Create document.xml: %v", err)
	}
	if _, err := f.Write([]byte(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>` + text + `</w:t></w:r></w:p></w:body></w:document>`)); err != nil {
		t.Fatalf("write document.xml: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("Close zip: %v", err)
	}
	return buf.Bytes()
}
