package application

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func buildZip(t *testing.T, parts map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range parts {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func officeDataURL(raw []byte) string {
	return "data:application/octet-stream;base64," + base64.StdEncoding.EncodeToString(raw)
}

func TestExtractDocx(t *testing.T) {
	doc := `<w:document xmlns:w="w"><w:body>
<w:p><w:pPr><w:tabs><w:tab w:val="left"/></w:tabs></w:pPr><w:r><w:t>Quarterly report</w:t></w:r></w:p>
<w:p><w:r><w:t>Revenue </w:t></w:r><w:r><w:t>grew 12%</w:t></w:r></w:p>
<w:tbl><w:tr><w:tc><w:p><w:r><w:t>Region</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>Total</w:t></w:r></w:p></w:tc></w:tr></w:tbl>
</w:body></w:document>`
	text, err := extractOfficeText("report.docx", buildZip(t, map[string]string{"word/document.xml": doc}))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Quarterly report\n", "Revenue grew 12%", "Region\tTotal"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in %q", want, text)
		}
	}
	if strings.HasPrefix(text, "\t") {
		t.Fatalf("tab stop definition leaked into text: %q", text)
	}
}

func TestExtractXlsxKeepsSheetsAndColumns(t *testing.T) {
	parts := map[string]string{
		"xl/workbook.xml":            `<workbook xmlns:r="r"><sheets><sheet name="Sales" sheetId="1" r:id="rId1"/><sheet name="Costs" sheetId="2" r:id="rId2"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships><Relationship Id="rId1" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Target="worksheets/sheet2.xml"/></Relationships>`,
		"xl/sharedStrings.xml":       `<sst><si><t>Month</t></si><si><t>Amount</t></si><si><r><t>Ja</t></r><r><t>n</t></r></si></sst>`,
		"xl/worksheets/sheet1.xml": `<worksheet><sheetData>
<row r="1"><c r="A1" t="s"><v>0</v></c><c r="C1" t="s"><v>1</v></c></row>
<row r="2"><c r="A2" t="s"><v>2</v></c><c r="C2"><v>120.5</v></c></row></sheetData></worksheet>`,
		"xl/worksheets/sheet2.xml": `<worksheet><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>Rent</t></is></c></row></sheetData></worksheet>`,
	}
	text, err := extractOfficeText("book.xlsx", buildZip(t, parts))
	if err != nil {
		t.Fatal(err)
	}
	want := "--- Sheet: Sales ---\nMonth\t\tAmount\nJan\t\t120.5\n\n--- Sheet: Costs ---\nRent"
	if text != want {
		t.Fatalf("got %q want %q", text, want)
	}
}

func TestExtractPptxAndODF(t *testing.T) {
	slide := func(s string) string {
		return `<p:sld xmlns:a="a"><a:p><a:r><a:t>` + s + `</a:t></a:r></a:p></p:sld>`
	}
	text, err := extractOfficeText("deck.pptx", buildZip(t, map[string]string{
		"ppt/slides/slide2.xml":  slide("Second"),
		"ppt/slides/slide10.xml": slide("Tenth"),
		"ppt/slides/slide1.xml":  slide("First"),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if i, j, k := strings.Index(text, "First"), strings.Index(text, "Second"), strings.Index(text, "Tenth"); !(i >= 0 && i < j && j < k) {
		t.Fatalf("slides out of order: %q", text)
	}

	odt := `<office:document-content xmlns:text="t" xmlns:office="o"><office:body><text:p>Hello<text:s/>world</text:p><text:p>Second</text:p></office:body></office:document-content>`
	text, err = extractOfficeText("notes.odt", buildZip(t, map[string]string{"content.xml": odt}))
	if err != nil || text != "Hello world\nSecond" {
		t.Fatalf("odt = %q, %v", text, err)
	}
}

func TestExtractAttachedFilesRoutesOfficeDocuments(t *testing.T) {
	doc := buildZip(t, map[string]string{"word/document.xml": `<w:document xmlns:w="w"><w:p><w:r><w:t>Hi there</w:t></w:r></w:p></w:document>`})
	out, err := ExtractAttachedFiles([]AttachedFile{{Name: "memo.docx", DataURL: officeDataURL(doc)}})
	if err != nil || out[0].Text != "Hi there" {
		t.Fatalf("out = %+v, %v", out, err)
	}
	// Unsupported binary types are rejected, and corrupt archives report a clear error.
	if _, err := ExtractAttachedFiles([]AttachedFile{{Name: "a.exe", DataURL: officeDataURL(doc)}}); err == nil {
		t.Fatal("unsupported type accepted")
	}
	if _, err := ExtractAttachedFiles([]AttachedFile{{Name: "bad.xlsx", DataURL: officeDataURL([]byte("not a zip"))}}); err == nil {
		t.Fatal("corrupt archive accepted")
	}
}
