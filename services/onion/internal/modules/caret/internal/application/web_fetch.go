package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// maxFetchedTextRunes is how much page text one web.fetch returns: enough to
// read an article, small enough to leave room for the rest of the turn.
const maxFetchedTextRunes = 12000

func init() {
	registerGeneralTool("web.fetch", (*AssistantToolExecutor).fetchPage)
}

func (e *AssistantToolExecutor) fetchPage(ctx context.Context, input map[string]any) (map[string]any, error) {
	target, err := publicPageURL(inputText(input, "url", "link", "page", "href"))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, errors.New("could not build the request")
	}
	req.Header.Set("User-Agent", caretUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain;q=0.9,application/json;q=0.8")
	resp, err := e.httpClient().Do(req)
	if err != nil {
		if errors.Is(err, errPrivateAddress) || strings.Contains(err.Error(), errPrivateAddress.Error()) {
			return nil, errPrivateAddress
		}
		return nil, errors.New("the page did not respond")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("the page returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxPublicResponseBytes))
	if err != nil {
		return nil, errors.New("could not read the page")
	}
	mediaType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	var title, description, text string
	switch {
	case mediaType == "" || strings.Contains(mediaType, "html"):
		title, description, text = extractPageText(string(body))
	case strings.HasPrefix(mediaType, "text/") || strings.Contains(mediaType, "json") || strings.Contains(mediaType, "xml"):
		text = string(body)
	default:
		return nil, fmt.Errorf("the page is %s, which Caret can't read as text", mediaType)
	}
	if !utf8.ValidString(text) {
		text = strings.ToValidUTF8(text, "")
	}
	text = focusPageText(text, inputText(input, "query", "focus", "question"), maxFetchedTextRunes)
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("the page has no readable text (it may need JavaScript)")
	}
	final := resp.Request.URL.String()
	if title == "" {
		title = resp.Request.URL.Host
	}
	output := map[string]any{
		"url":   final,
		"title": truncateSearchText(title, 300),
		"text":  text,
	}
	if description != "" {
		output["description"] = truncateSearchText(description, 500)
	}
	if strings.HasPrefix(final, "https://") {
		output["sources"] = []map[string]string{{"title": truncateSearchText(title, 300), "url": final}}
	}
	return output, nil
}

// publicPageURL accepts an absolute http(s) URL with no credentials. A bare
// domain gets https.
func publicPageURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("give the page URL to read")
	}
	if len(raw) > 2048 {
		return nil, errors.New("the URL is too long")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" || parsed.User != nil {
		return nil, errors.New("use a public http or https URL")
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".internal") || strings.HasSuffix(host, ".local") {
		return nil, errPrivateAddress
	}
	parsed.Fragment = ""
	return parsed, nil
}

var skippedPageElements = map[atom.Atom]bool{
	atom.Script: true, atom.Style: true, atom.Noscript: true, atom.Svg: true, atom.Iframe: true,
	atom.Nav: true, atom.Footer: true, atom.Form: true, atom.Button: true, atom.Select: true, atom.Template: true,
}

var blockPageElements = map[atom.Atom]bool{
	atom.P: true, atom.Div: true, atom.Section: true, atom.Article: true, atom.Main: true, atom.Br: true,
	atom.H1: true, atom.H2: true, atom.H3: true, atom.H4: true, atom.H5: true, atom.H6: true,
	atom.Li: true, atom.Tr: true, atom.Table: true, atom.Ul: true, atom.Ol: true, atom.Blockquote: true,
	atom.Pre: true, atom.Header: true, atom.Aside: true, atom.Dt: true, atom.Dd: true, atom.Figcaption: true,
}

// extractPageText reduces HTML to its title, meta description and readable
// text, one block per line, with headings and list items marked.
func extractPageText(document string) (title, description, text string) {
	root, err := html.Parse(strings.NewReader(document))
	if err != nil {
		return "", "", ""
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.DataAtom {
			case atom.Title:
				if title == "" && n.FirstChild != nil {
					title = strings.Join(strings.Fields(n.FirstChild.Data), " ")
				}
				return
			case atom.Meta:
				name := strings.ToLower(htmlAttr(n, "name") + htmlAttr(n, "property"))
				if description == "" && (name == "description" || name == "og:description") {
					description = strings.TrimSpace(htmlAttr(n, "content"))
				}
				return
			}
			if skippedPageElements[n.DataAtom] || htmlAttr(n, "aria-hidden") == "true" || htmlAttr(n, "hidden") != "" {
				return
			}
			if blockPageElements[n.DataAtom] {
				b.WriteString("\n")
				switch n.DataAtom {
				case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
					b.WriteString("## ")
				case atom.Li:
					b.WriteString("- ")
				}
			}
		}
		if n.Type == html.TextNode {
			if words := strings.Fields(n.Data); len(words) > 0 {
				if b.Len() > 0 {
					last := b.String()[b.Len()-1]
					if last != '\n' && last != ' ' {
						b.WriteByte(' ')
					}
				}
				b.WriteString(strings.Join(words, " "))
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
		if n.Type == html.ElementNode && blockPageElements[n.DataAtom] {
			b.WriteString("\n")
		}
	}
	walk(root)
	lines := strings.Split(b.String(), "\n")
	kept := lines[:0]
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "-" || line == "##" {
			continue
		}
		kept = append(kept, line)
	}
	return title, description, strings.Join(kept, "\n")
}

func htmlAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if strings.EqualFold(attr.Key, key) {
			return attr.Val
		}
	}
	return ""
}

// focusPageText fits text into limit runes. With a query, the paragraphs that
// mention its words are kept first (in page order), so the part of a long
// page the question is about survives the cut.
func focusPageText(text, query string, limit int) string {
	if utf8.RuneCountInString(text) <= limit {
		return text
	}
	paragraphs := strings.Split(text, "\n")
	terms := []string{}
	for _, word := range strings.Fields(strings.ToLower(query)) {
		if utf8.RuneCountInString(word) >= 3 {
			terms = append(terms, word)
		}
	}
	keep := make([]bool, len(paragraphs))
	used := 0
	take := func(i int) bool {
		size := utf8.RuneCountInString(paragraphs[i]) + 1
		if used+size > limit {
			return false
		}
		keep[i] = true
		used += size
		return true
	}
	if len(terms) > 0 {
		for i, paragraph := range paragraphs {
			lower := strings.ToLower(paragraph)
			for _, term := range terms {
				if strings.Contains(lower, term) {
					take(i)
					break
				}
			}
		}
	}
	for i := range paragraphs {
		if !keep[i] && !take(i) {
			break
		}
	}
	var b strings.Builder
	skipped := false
	for i, paragraph := range paragraphs {
		if !keep[i] {
			skipped = true
			continue
		}
		if skipped && b.Len() > 0 {
			b.WriteString("[...]\n")
		}
		skipped = false
		b.WriteString(paragraph + "\n")
	}
	if skipped {
		b.WriteString("[page continues]")
	}
	return strings.TrimSpace(b.String())
}
