package application

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestWebFetchReadsPDFText(t *testing.T) {
	document := newTestPDF(t, []string{"BT /F1 12 Tf 72 720 Td (Menu prices 2026) Tj ET", "BT /F1 12 Tf 72 720 Td (Espresso 3 USD) Tj ET"})
	executor := &AssistantToolExecutor{HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/pdf"}}, Request: r, Body: io.NopCloser(bytes.NewReader(document))}, nil
	})}}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "web.fetch", map[string]any{"url": "https://cafe.example/menu.pdf"})
	if err != nil {
		t.Fatal(err)
	}
	text := output["text"].(string)
	if !strings.Contains(text, "Menu prices 2026") || !strings.Contains(text, "Espresso 3 USD") || output["title"] != "menu.pdf" {
		t.Fatalf("output = %#v", output)
	}
	if _, err := extractPDFText([]byte("%PDF-1.4 garbage")); err == nil {
		t.Fatal("a broken PDF must be an error, not a panic or empty text")
	}
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

func TestExtractAttachedPDF(t *testing.T) {
	document := newTestPDF(t, []string{"BT /F1 12 Tf 72 720 Td (Invoice total 420) Tj ET"})
	files, err := ExtractAttachedFiles([]AttachedFile{{Name: "invoice.pdf", DataURL: "data:application/pdf;base64," + base64.StdEncoding.EncodeToString(document)}, {Name: "a.txt", Text: "plain"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(files[0].Text, "Invoice total 420") || files[0].DataURL != "" || files[1].Text != "plain" {
		t.Fatalf("files = %#v", files)
	}
	if _, err := ExtractAttachedFiles([]AttachedFile{{Name: "x.exe", DataURL: "data:application/octet-stream;base64,AAAA"}}); err == nil {
		t.Fatal("a non-PDF document was accepted")
	}
}
