package application

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

func (s *Service) markCVTargetsErrored(ctx context.Context, targets []domain.CVProcessingTarget, message string) error {
	var updateErr error
	for _, target := range targets {
		err := s.repo.UpdateCVProcessingResult(ctx, domain.CVProcessingResult{
			Source: target.Source,
			ID:     target.ID,
			Status: "error",
			Error:  message,
		})
		if err != nil {
			updateErr = errors.Join(updateErr, err)
		}
	}
	return updateErr
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
	result.Profile = buildCVProfile(result.ExtractedText)
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
		if closeErr := tmp.Close(); closeErr != nil {
			return "", fmt.Errorf("write temporary CV: %w; close file: %w", err, closeErr)
		}
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

	return extractPDFPlainText(reader, maxCVExtractedTextBytes)
}

// extractPDFPlainText decodes PDF pages one at a time and stops once the
// extracted text reaches limit. The pdf library materializes a page's full text
// before returning it, so pages whose decompressed content does not fit in the
// remaining budget are skipped instead of decoded: a compression bomb must not
// exhaust worker memory before the caller's limit can be applied.
func extractPDFPlainText(reader *pdf.Reader, limit int) (string, error) {
	var out strings.Builder
	skipped := false
	pages := reader.NumPage()
	if pages > maxCVExtractedPDFPages {
		pages = maxCVExtractedPDFPages
	}
	for i := 1; i <= pages && out.Len() < limit; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		if !pdfContentFits(page, limit-out.Len()) {
			skipped = true
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		out.WriteString(text)
	}
	if skipped && out.Len() == 0 {
		// Nothing could be decoded without buffering more than the limit; fail
		// loudly instead of storing an empty CV.
		return "", fmt.Errorf("CV text is too large to extract")
	}
	return out.String(), nil
}

// pdfContentFits reports whether the decompressed content streams of a page fit
// in budget.
func pdfContentFits(page pdf.Page, budget int) bool {
	contents := page.V.Key("Contents")
	switch contents.Kind() {
	case pdf.Stream:
		return pdfStreamFits(contents, budget)
	case pdf.Array:
		for i := 0; i < contents.Len(); i++ {
			if !pdfStreamFits(contents.Index(i), budget) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

// pdfStreamFits counts a content stream's decompressed bytes up to budget+1.
// The pdf library panics on unreadable streams, so a panic counts as "does not
// fit" and the page is skipped.
func pdfStreamFits(stream pdf.Value, budget int) (fits bool) {
	defer func() {
		if recover() != nil {
			fits = false
		}
	}()
	reader := stream.Reader()
	defer reader.Close()
	n, err := io.CopyN(io.Discard, reader, int64(budget)+1)
	return n <= int64(budget) && errors.Is(err, io.EOF)
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

		// Bound the decompressed entry: a deflate-ratio bomb must not grow the
		// worker's memory before the extracted text is truncated.
		data, err := io.ReadAll(io.LimitReader(rc, maxCVExtractedXMLBytes+1))
		if err != nil {
			return "", err
		}
		truncated := len(data) > maxCVExtractedXMLBytes
		text, err := extractWordXMLText(bytes.NewReader(data))
		if err != nil {
			if !truncated {
				return "", err
			}
			// The bounded prefix may end mid-token; keep what was decoded.
			return text, nil
		}
		return text, nil
	}
	return "", fmt.Errorf("DOCX missing word/document.xml")
}
