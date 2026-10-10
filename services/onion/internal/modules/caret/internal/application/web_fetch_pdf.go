package application

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

const (
	maxFetchedPDFPages = 40
	// maxPDFPageStreamBytes bounds one page's decompressed content, so a
	// small file can't expand into gigabytes.
	maxPDFPageStreamBytes = 4 << 20
)

// extractPDFText reads the text of a PDF's first pages. The library panics
// on malformed files, so each step recovers and unreadable pages are skipped.
func extractPDFText(body []byte) (text string, err error) {
	defer func() {
		if recover() != nil {
			text, err = "", errors.New("the PDF could not be read")
		}
	}()
	reader, err := pdf.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", errors.New("the PDF could not be read")
	}
	var out strings.Builder
	pages := min(reader.NumPage(), maxFetchedPDFPages)
	for i := 1; i <= pages && out.Len() < maxFetchedTextRunes*4; i++ {
		page := reader.Page(i)
		if page.V.IsNull() || !pdfPageFits(page) {
			continue
		}
		pageText, err := safePageText(page)
		if err != nil {
			continue
		}
		out.WriteString(strings.TrimSpace(pageText))
		out.WriteString("\n\n")
	}
	if reader.NumPage() > pages {
		out.WriteString("[later pages not read]")
	}
	text = strings.TrimSpace(out.String())
	if text == "" {
		return "", errors.New("the PDF has no extractable text (it may be scanned images)")
	}
	return text, nil
}

func safePageText(page pdf.Page) (text string, err error) {
	defer func() {
		if recover() != nil {
			text, err = "", errors.New("unreadable page")
		}
	}()
	return page.GetPlainText(nil)
}

func pdfPageFits(page pdf.Page) bool {
	contents := page.V.Key("Contents")
	switch contents.Kind() {
	case pdf.Stream:
		return pdfStreamFits(contents)
	case pdf.Array:
		for i := 0; i < contents.Len(); i++ {
			if !pdfStreamFits(contents.Index(i)) {
				return false
			}
		}
	}
	return true
}

func pdfStreamFits(stream pdf.Value) (fits bool) {
	defer func() {
		if recover() != nil {
			fits = false
		}
	}()
	reader := stream.Reader()
	defer reader.Close()
	n, err := io.CopyN(io.Discard, reader, maxPDFPageStreamBytes+1)
	return n <= maxPDFPageStreamBytes && errors.Is(err, io.EOF)
}
