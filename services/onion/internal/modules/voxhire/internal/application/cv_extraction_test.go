package application

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
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
	if len(result.Profile.Functions) == 0 || !containsString(result.Profile.Functions, "Operations") {
		t.Fatalf("profile functions = %#v", result.Profile.Functions)
	}
	if !containsString(result.Profile.Industries, "Retail") {
		t.Fatalf("profile industries = %#v", result.Profile.Industries)
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

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// Short keyword terms must only match whole words: "hr" is not in "three",
// "ui" is not in "built", and "commerce" is not in "e-commerce".
func TestBuildCVProfileMatchesKeywordsOnWordBoundaries(t *testing.T) {
	profile := buildCVProfile("built data pipelines for three years of experience in e-commerce")

	if containsString(profile.Skills, "HR") {
		t.Errorf("Skills = %#v, want no HR from 'three'", profile.Skills)
	}
	if containsString(profile.Functions, "Design") {
		t.Errorf("Functions = %#v, want no Design from 'built'", profile.Functions)
	}
	if containsString(profile.Education, "Business") {
		t.Errorf("Education = %#v, want no Business from 'e-commerce'", profile.Education)
	}
}

// Genuine whole-word and punctuation-bearing terms still match.
func TestBuildCVProfileMatchesRealKeywords(t *testing.T) {
	profile := buildCVProfile("Human resources generalist with a bachelor degree in business administration, skilled in ui design and node.js")

	for _, tc := range []struct {
		name   string
		values []string
		want   string
	}{
		{"Skills", profile.Skills, "HR"},
		{"Skills", profile.Skills, "Design"},
		{"Skills", profile.Skills, "JavaScript"},
		{"Functions", profile.Functions, "Design"},
		{"Education", profile.Education, "Bachelor's degree"},
		{"Education", profile.Education, "Business"},
	} {
		if !containsString(tc.values, tc.want) {
			t.Errorf("%s = %#v, want %q", tc.name, tc.values, tc.want)
		}
	}
}

// A DOCX whose word/document.xml inflates to many megabytes must not be
// materialized in full: extraction keeps at most maxCVExtractedTextBytes.
func TestExtractDOCXTextBoundsDecompressedOutput(t *testing.T) {
	raw := newTestDOCX(t, strings.Repeat("A", 8<<20))
	if len(raw) > 1<<20 {
		t.Fatalf("fixture is not a compression bomb: %d compressed bytes", len(raw))
	}

	text, err := extractCVText("resume.docx", raw)
	if err != nil {
		t.Fatalf("extractCVText: %v", err)
	}
	if len(text) == 0 {
		t.Fatal("expected some extracted text")
	}
	if len(text) > maxCVExtractedTextBytes {
		t.Fatalf("extracted %d bytes, want at most %d", len(text), maxCVExtractedTextBytes)
	}
}

// A PDF page whose content stream inflates far beyond the extraction budget
// must not be materialized in full.
func TestExtractPDFTextBoundsDecompressedStreams(t *testing.T) {
	const bombBytes = 32 << 20
	raw := newTestPDF(t, []string{newTestPDFText(strings.Repeat("A", bombBytes))})

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	text, err := extractCVText("resume.pdf", raw)
	runtime.ReadMemStats(&after)

	if err == nil {
		t.Fatalf("expected an over-limit extraction error, got %d bytes of text", len(text))
	}
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > bombBytes/2 {
		t.Fatalf("extraction allocated %d bytes for a %d byte decompression bomb", allocated, bombBytes)
	}
}

// A normal page must still be extracted even when another page has to be
// skipped for exceeding the budget.
func TestExtractPDFTextSkipsOnlyOversizedPages(t *testing.T) {
	raw := newTestPDF(t, []string{
		newTestPDFText(strings.Repeat("B", 8<<20)),
		newTestPDFText("Retail operations manager with payroll experience"),
	})

	text, err := extractCVText("resume.pdf", raw)
	if err != nil {
		t.Fatalf("extractCVText: %v", err)
	}
	if !strings.Contains(text, "Retail operations manager") {
		t.Fatalf("extracted %d bytes, missing candidate text", len(text))
	}
}

// Normal multi-page CVs keep extracting every page.
func TestExtractPDFTextExtractsAllPages(t *testing.T) {
	raw := newTestPDF(t, []string{
		newTestPDFText("First page summary"),
		newTestPDFText("Second page work history"),
		newTestPDFText("Third page education"),
	})

	text, err := extractCVText("resume.pdf", raw)
	if err != nil {
		t.Fatalf("extractCVText: %v", err)
	}
	for _, want := range []string{"First page summary", "Second page work history", "Third page education"} {
		if !strings.Contains(text, want) {
			t.Fatalf("extracted %d bytes, missing %q", len(text), want)
		}
	}
}

func newTestPDFText(text string) string {
	return fmt.Sprintf("BT /F1 12 Tf 72 720 Td (%s) Tj ET", text)
}

// newTestPDF builds a PDF whose pages carry the given decoded content streams,
// each stored as a Flate-compressed stream.
func newTestPDF(t *testing.T, pageContents []string) []byte {
	t.Helper()

	const (
		catalogID   = 1
		pagesID     = 2
		fontID      = 3
		firstPageID = 4
	)
	firstContentID := firstPageID + len(pageContents)

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, firstContentID+len(pageContents))
	writeObject := func(id int, body string) {
		offsets[id] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", id, body)
	}

	kids := make([]string, 0, len(pageContents))
	for i := range pageContents {
		kids = append(kids, fmt.Sprintf("%d 0 R", firstPageID+i))
	}
	writeObject(catalogID, fmt.Sprintf("<< /Type /Catalog /Pages %d 0 R >>", pagesID))
	writeObject(pagesID, fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(pageContents)))
	writeObject(fontID, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	for i := range pageContents {
		writeObject(firstPageID+i, fmt.Sprintf(
			"<< /Type /Page /Parent %d 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 %d 0 R >> >> /Contents %d 0 R >>",
			pagesID, fontID, firstContentID+i))
	}
	for i, content := range pageContents {
		var compressed bytes.Buffer
		zw := zlib.NewWriter(&compressed)
		if _, err := zw.Write([]byte(content)); err != nil {
			t.Fatalf("compress page %d: %v", i, err)
		}
		if err := zw.Close(); err != nil {
			t.Fatalf("compress page %d: %v", i, err)
		}
		offsets[firstContentID+i] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n<< /Length %d /Filter /FlateDecode >>\nstream\n", firstContentID+i, compressed.Len())
		buf.Write(compressed.Bytes())
		buf.WriteString("\nendstream\nendobj\n")
	}

	xrefOffset := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(offsets))
	buf.WriteString("0000000000 65535 f \n")
	for id := 1; id < len(offsets); id++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[id])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), catalogID, xrefOffset)
	return buf.Bytes()
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
