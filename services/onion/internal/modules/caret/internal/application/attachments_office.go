package application

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	// maxOfficePartBytes bounds one decompressed XML part, so a small
	// archive can't expand into gigabytes.
	maxOfficePartBytes = 24 << 20
	// maxOfficeTextBytes stops reading once enough text is collected.
	maxOfficeTextBytes = 1 << 20
	maxOfficeParts     = 400
)

var errOfficeUnreadable = errors.New("the document could not be read")

// officeExtensions lists the document formats whose text the server can
// extract; they are zip archives of XML.
var officeExtensions = map[string]bool{
	"docx": true, "xlsx": true, "pptx": true,
	"odt": true, "ods": true, "odp": true,
}

func fileExtension(name string) string {
	return strings.ToLower(strings.TrimPrefix(path.Ext(strings.TrimSpace(name)), "."))
}

// extractOfficeText reads the text of a Word, Excel or PowerPoint file, or
// its OpenDocument equivalent.
func extractOfficeText(name string, body []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil || len(zr.File) > maxOfficeParts*5 {
		return "", errOfficeUnreadable
	}
	var text string
	switch ext := fileExtension(name); ext {
	case "docx":
		text, err = extractDocx(zr)
	case "pptx":
		text, err = extractPptx(zr)
	case "xlsx":
		text, err = extractXlsx(zr)
	case "odt", "ods", "odp":
		text, err = extractODF(zr)
	default:
		return "", errors.New("this document type is not supported")
	}
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(tidyOfficeText(text))
	if text == "" {
		return "", errors.New("the document has no extractable text")
	}
	return text, nil
}

func findZipPart(zr *zip.Reader, name string) *zip.File {
	for _, f := range zr.File {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func zipPartDecoder(f *zip.File) (*xml.Decoder, io.Closer, error) {
	if f.UncompressedSize64 > maxOfficePartBytes {
		return nil, nil, errOfficeUnreadable
	}
	rc, err := f.Open()
	if err != nil {
		return nil, nil, errOfficeUnreadable
	}
	return xml.NewDecoder(io.LimitReader(rc, maxOfficePartBytes)), rc, nil
}

// numberedParts returns prefix<N>.xml parts ordered by N.
func numberedParts(zr *zip.Reader, prefix string) []*zip.File {
	var parts []*zip.File
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, prefix) && strings.HasSuffix(f.Name, ".xml") {
			if _, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(f.Name, prefix), ".xml")); err == nil {
				parts = append(parts, f)
			}
		}
	}
	index := func(f *zip.File) int {
		n, _ := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(f.Name, prefix), ".xml"))
		return n
	}
	sort.Slice(parts, func(i, j int) bool { return index(parts[i]) < index(parts[j]) })
	if len(parts) > maxOfficeParts {
		parts = parts[:maxOfficeParts]
	}
	return parts
}

// readXMLText walks one XML part. textElem names the element holding text
// runs; breakAfter and tabAfter name elements that end a line or a cell.
func readXMLText(f *zip.File, textElem string, breakAfter, tabAfter map[string]bool, out *strings.Builder) error {
	dec, closer, err := zipPartDecoder(f)
	if err != nil {
		return err
	}
	defer closer.Close()
	inText, inTabStops, cellDepth := false, false, 0
	for out.Len() < maxOfficeTextBytes {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errOfficeUnreadable
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch name := t.Name.Local; {
			case name == textElem:
				inText = true
			case name == "tc":
				cellDepth++
			case name == "tabs" || name == "tabLst":
				inTabStops = true // tab stop definitions, not content
			case name == "tab" && !inTabStops:
				out.WriteByte('\t')
			case name == "br" || name == "cr":
				out.WriteByte('\n')
			}
		case xml.EndElement:
			name := t.Name.Local
			if name == textElem {
				inText = false
			}
			if name == "tabs" || name == "tabLst" {
				inTabStops = false
			}
			if name == "tc" && cellDepth > 0 {
				cellDepth--
			}
			if name == "p" && cellDepth > 0 {
				out.WriteByte(' ') // keep a table cell on one line
			} else if breakAfter[name] {
				out.WriteByte('\n')
			} else if tabAfter[name] {
				out.WriteByte('\t')
			}
		case xml.CharData:
			if inText {
				out.Write(t)
			}
		}
	}
	return nil
}

func extractDocx(zr *zip.Reader) (string, error) {
	doc := findZipPart(zr, "word/document.xml")
	if doc == nil {
		return "", errOfficeUnreadable
	}
	var out strings.Builder
	err := readXMLText(doc, "t", map[string]bool{"p": true, "tr": true}, map[string]bool{"tc": true}, &out)
	return out.String(), err
}

func extractPptx(zr *zip.Reader) (string, error) {
	slides := numberedParts(zr, "ppt/slides/slide")
	if len(slides) == 0 {
		return "", errOfficeUnreadable
	}
	var out strings.Builder
	for i, slide := range slides {
		out.WriteString("--- Slide " + strconv.Itoa(i+1) + " ---\n")
		if err := readXMLText(slide, "t", map[string]bool{"p": true}, nil, &out); err != nil {
			return "", err
		}
		out.WriteByte('\n')
	}
	return out.String(), nil
}

// extractODF reads content.xml of an OpenDocument text, sheet or slide file.
func extractODF(zr *zip.Reader) (string, error) {
	content := findZipPart(zr, "content.xml")
	if content == nil {
		return "", errOfficeUnreadable
	}
	dec, closer, err := zipPartDecoder(content)
	if err != nil {
		return "", err
	}
	defer closer.Close()
	var out strings.Builder
	depth := 0 // inside a paragraph or heading
	for out.Len() < maxOfficeTextBytes {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", errOfficeUnreadable
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p", "h":
				depth++
			case "tab":
				out.WriteByte('\t')
			case "line-break":
				out.WriteByte('\n')
			case "s":
				out.WriteByte(' ')
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p", "h":
				depth--
				out.WriteByte('\n')
			case "table-cell":
				out.WriteByte('\t')
			case "table-row":
				out.WriteByte('\n')
			}
		case xml.CharData:
			if depth > 0 {
				out.Write(t)
			}
		}
	}
	return out.String(), nil
}

func extractXlsx(zr *zip.Reader) (string, error) {
	shared, err := xlsxSharedStrings(zr)
	if err != nil {
		return "", err
	}
	sheets := xlsxSheets(zr)
	if len(sheets) == 0 {
		return "", errOfficeUnreadable
	}
	var out strings.Builder
	for _, sheet := range sheets {
		if out.Len() >= maxOfficeTextBytes {
			break
		}
		part := findZipPart(zr, sheet.part)
		if part == nil {
			continue
		}
		out.WriteString("--- Sheet: " + sheet.name + " ---\n")
		if err := xlsxSheetText(part, shared, &out); err != nil {
			return "", err
		}
		out.WriteByte('\n')
	}
	return out.String(), nil
}

func xlsxSharedStrings(zr *zip.Reader) ([]string, error) {
	f := findZipPart(zr, "xl/sharedStrings.xml")
	if f == nil {
		return nil, nil
	}
	dec, closer, err := zipPartDecoder(f)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	var items []string
	var cur strings.Builder
	inT := false
	for len(items) < 500000 {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errOfficeUnreadable
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" {
				inT = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inT = false
			case "si":
				items = append(items, cur.String())
				cur.Reset()
			}
		case xml.CharData:
			if inT {
				cur.Write(t)
			}
		}
	}
	return items, nil
}

type xlsxSheet struct{ name, part string }

// xlsxSheets lists worksheets in workbook order. A missing or odd workbook
// falls back to the numbered sheet parts.
func xlsxSheets(zr *zip.Reader) []xlsxSheet {
	var sheets []xlsxSheet
	wb, rels := findZipPart(zr, "xl/workbook.xml"), findZipPart(zr, "xl/_rels/workbook.xml.rels")
	if wb != nil && rels != nil {
		targets := map[string]string{}
		if dec, closer, err := zipPartDecoder(rels); err == nil {
			for {
				tok, err := dec.Token()
				if err != nil {
					break
				}
				if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "Relationship" {
					var id, target string
					for _, a := range se.Attr {
						switch a.Name.Local {
						case "Id":
							id = a.Value
						case "Target":
							target = a.Value
						}
					}
					targets[id] = path.Clean("xl/" + strings.TrimPrefix(target, "/xl/"))
				}
			}
			closer.Close()
		}
		if dec, closer, err := zipPartDecoder(wb); err == nil {
			for {
				tok, err := dec.Token()
				if err != nil {
					break
				}
				if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "sheet" {
					var name, id string
					for _, a := range se.Attr {
						switch a.Name.Local {
						case "name":
							name = a.Value
						case "id":
							id = a.Value
						}
					}
					if p := targets[id]; p != "" {
						sheets = append(sheets, xlsxSheet{name: name, part: p})
					}
				}
			}
			closer.Close()
		}
	}
	if len(sheets) == 0 {
		for i, f := range numberedParts(zr, "xl/worksheets/sheet") {
			sheets = append(sheets, xlsxSheet{name: "Sheet" + strconv.Itoa(i+1), part: f.Name})
		}
	}
	return sheets
}

// xlsxColumn converts the letters of a cell reference such as "C12" to a
// zero-based column index.
func xlsxColumn(ref string) int {
	col := 0
	for _, r := range ref {
		if r < 'A' || r > 'Z' {
			break
		}
		col = col*26 + int(r-'A') + 1
	}
	return col - 1
}

func xlsxSheetText(f *zip.File, shared []string, out *strings.Builder) error {
	dec, closer, err := zipPartDecoder(f)
	if err != nil {
		return err
	}
	defer closer.Close()
	var (
		cellType string
		col      int
		next     int // next column to write in this row
		value    strings.Builder
		inV, inT bool
	)
	for out.Len() < maxOfficeTextBytes {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errOfficeUnreadable
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "row":
				next = 0
			case "c":
				cellType, value = "", strings.Builder{}
				col = next
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "t":
						cellType = a.Value
					case "r":
						col = max(xlsxColumn(a.Value), 0)
					}
				}
			case "v":
				inV = true
			case "t":
				inT = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "v":
				inV = false
			case "t":
				inT = false
			case "c":
				cell := value.String()
				if cellType == "s" {
					if i, err := strconv.Atoi(strings.TrimSpace(cell)); err == nil && i >= 0 && i < len(shared) {
						cell = shared[i]
					} else {
						cell = ""
					}
				}
				if cell == "" {
					continue
				}
				if col > 1000 {
					col = next
				}
				for ; next < col; next++ {
					out.WriteByte('\t')
				}
				out.WriteString(strings.NewReplacer("\t", " ", "\n", " ").Replace(cell))
				out.WriteByte('\t')
				next = col + 1
			case "row":
				out.WriteByte('\n')
			}
		case xml.CharData:
			if inV || inT {
				value.Write(t)
			}
		}
	}
	return nil
}

// tidyOfficeText trims trailing tabs and collapses runs of blank lines.
func tidyOfficeText(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	blank := 0
	for _, line := range lines {
		line = strings.ReplaceAll(strings.TrimRight(line, "\t \r"), " \t", "\t")
		if line == "" {
			blank++
			if blank > 1 {
				continue
			}
		} else {
			blank = 0
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
