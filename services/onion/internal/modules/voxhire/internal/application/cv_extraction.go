package application

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/storage"
)

const (
	cvProcessingBatchSize = 25
	maxCVExtractionBytes  = 10 << 20
	maxCVStoredTextBytes  = 128 << 10
	maxCVSummaryRunes     = 700
)

var whitespaceRE = regexp.MustCompile(`\s+`)

type cvStorage interface {
	Download(ctx context.Context, key string) (io.ReadCloser, int64, error)
}

// ProcessPendingCVs downloads queued CV files, extracts searchable text,
// and writes the result back onto the application row.
func (s *Service) ProcessPendingCVs(ctx context.Context) error {
	targets, err := s.repo.ListPendingCVProcessingTargets(ctx, cvProcessingBatchSize)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}

	client, err := storage.Default()
	if err != nil {
		return fmt.Errorf("cv storage unavailable: %w", err)
	}

	for _, target := range targets {
		result := s.processCVTarget(ctx, client, target)
		if err := s.repo.UpdateCVProcessingResult(ctx, result); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) processCVTarget(ctx context.Context, store cvStorage, target domain.CVProcessingTarget) domain.CVProcessingResult {
	result := domain.CVProcessingResult{
		Source: target.Source,
		ID:     target.ID,
		Status: "processed",
	}

	body, size, err := store.Download(ctx, target.CVKey)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	defer body.Close()

	if size > maxCVExtractionBytes {
		result.Status = "error"
		result.Error = "CV is too large to extract"
		return result
	}

	raw, err := io.ReadAll(io.LimitReader(body, maxCVExtractionBytes+1))
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	if len(raw) > maxCVExtractionBytes {
		result.Status = "error"
		result.Error = "CV is too large to extract"
		return result
	}

	text, err := extractCVText(target.CVKey, raw)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		logger.Ctx(ctx).Warn("voxhire cv extraction failed",
			logger.String("source", target.Source),
			logger.Any("id", target.ID),
			logger.String("key", target.CVKey),
			logger.Error(err),
		)
		return result
	}

	text = normalizeCVText(text)
	result.ExtractedText = truncateBytes(text, maxCVStoredTextBytes)
	result.Summary = summarizeCVText(result.ExtractedText)
	return result
}

func extractCVText(key string, raw []byte) (string, error) {
	lower := strings.ToLower(key)
	switch {
	case strings.HasSuffix(lower, ".pdf"):
		return extractPDFText(raw)
	case strings.HasSuffix(lower, ".docx"):
		return extractDOCXText(raw)
	case strings.HasSuffix(lower, ".doc"):
		return "", fmt.Errorf("legacy .doc CV parsing is not supported; ask candidate for PDF or DOCX")
	default:
		if bytes.HasPrefix(raw, []byte("%PDF")) {
			return extractPDFText(raw)
		}
		if len(raw) >= 4 && raw[0] == 'P' && raw[1] == 'K' {
			return extractDOCXText(raw)
		}
		return "", fmt.Errorf("unsupported CV format")
	}
}

func extractPDFText(raw []byte) (string, error) {
	tmp, err := os.CreateTemp("", "voxhire-cv-*.pdf")
	if err != nil {
		return "", err
	}
	path := tmp.Name()
	defer os.Remove(path)

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	plain, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	out, err := io.ReadAll(io.LimitReader(plain, maxCVStoredTextBytes+1))
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func extractDOCXText(raw []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return "", err
	}
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		return extractWordXMLText(rc)
	}
	return "", fmt.Errorf("DOCX missing word/document.xml")
}

func extractWordXMLText(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var out strings.Builder
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.CharData:
			out.Write([]byte(t))
			out.WriteByte(' ')
		}
	}
	return out.String(), nil
}

func normalizeCVText(text string) string {
	return strings.TrimSpace(whitespaceRE.ReplaceAllString(text, " "))
}

func summarizeCVText(text string) string {
	runes := []rune(text)
	if len(runes) > maxCVSummaryRunes {
		runes = runes[:maxCVSummaryRunes]
	}
	return strings.TrimSpace(string(runes))
}

func truncateBytes(text string, max int) string {
	if len(text) <= max {
		return text
	}
	cut := 0
	for i := range text {
		if i > max {
			break
		}
		cut = i
	}
	if cut == 0 {
		return ""
	}
	return strings.TrimSpace(text[:cut])
}
