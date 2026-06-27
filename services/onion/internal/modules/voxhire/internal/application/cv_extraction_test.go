package application

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
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

func TestProcessPendingCVsMarksTargetsErroredWhenStorageUnavailable(t *testing.T) {
	t.Setenv("FLOCI_ENABLED", "")
	t.Setenv("S3_BUCKET_NAME", "")

	repo := &mockRepo{
		cvTargets: []domain.CVProcessingTarget{
			{Source: "talent", ID: 7, CVKey: "voxhire/applications/ops/resume.pdf"},
			{Source: "career", ID: 11, CVKey: "voxire/applications/42/resume.pdf"},
		},
	}
	svc := &Service{repo: repo}

	err := svc.ProcessPendingCVs(context.Background())
	if err == nil || !strings.Contains(err.Error(), "cv storage unavailable") {
		t.Fatalf("expected storage setup error, got %v", err)
	}
	if len(repo.cvResults) != len(repo.cvTargets) {
		t.Fatalf("stored %d CV results, want %d", len(repo.cvResults), len(repo.cvTargets))
	}
	for i, result := range repo.cvResults {
		if result.Source != repo.cvTargets[i].Source || result.ID != repo.cvTargets[i].ID {
			t.Fatalf("result[%d] = source %q id %d, want source %q id %d", i, result.Source, result.ID, repo.cvTargets[i].Source, repo.cvTargets[i].ID)
		}
		if result.Status != "error" {
			t.Fatalf("result[%d] status = %q, want error", i, result.Status)
		}
		if !strings.Contains(result.Error, "s3 storage is not configured") {
			t.Fatalf("result[%d] error = %q", i, result.Error)
		}
	}
}

func TestExtractCVTextFromPDF(t *testing.T) {
	path := os.Getenv("VOXHIRE_TEST_RESUME_PDF")
	if path == "" {
		t.Skip("set VOXHIRE_TEST_RESUME_PDF to run against a real resume PDF")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text, err := extractCVText("resume.pdf", raw)
	if err != nil {
		t.Fatalf("extractCVText: %v", err)
	}
	text = normalizeCVText(text)
	if text == "" {
		t.Fatal("expected extracted PDF text")
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
