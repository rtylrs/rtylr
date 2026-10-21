package application

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestExtractPageTextKeepsArticleAndDropsChrome(t *testing.T) {
	page := `<html><head><title> Pricing  plans </title><meta name="description" content="Plans for teams"><script>var x=1</script></head>
<body><nav>Home About</nav><main><h1>Plans</h1><p>The <b>Pro</b> plan costs $20.</p><ul><li>Unlimited chats</li><li>Priority</li></ul></main><footer>© 2026</footer></body></html>`
	title, description, text := extractPageText(page)
	if title != "Pricing plans" || description != "Plans for teams" {
		t.Fatalf("title=%q description=%q", title, description)
	}
	for _, want := range []string{"## Plans", "The Pro plan costs $20.", "- Unlimited chats"} {
		if !strings.Contains(text, want) {
			t.Fatalf("text missing %q:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{"var x", "Home About", "© 2026"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("text kept %q:\n%s", unwanted, text)
		}
	}
}

func TestFocusPageTextKeepsMatchingParagraphs(t *testing.T) {
	text := strings.Repeat("filler paragraph about nothing\n", 200) + "the refund policy allows 30 days\n" + strings.Repeat("more filler\n", 50)
	focused := focusPageText(text, "refund policy", 400)
	if !strings.Contains(focused, "refund policy allows 30 days") {
		t.Fatalf("focused text dropped the matching paragraph:\n%s", focused)
	}
	if len([]rune(focused)) > 500 {
		t.Fatalf("focused text is %d runes", len([]rune(focused)))
	}
}

func TestPublicPageURLRejectsUnsafeTargets(t *testing.T) {
	for _, raw := range []string{"", "file:///etc/passwd", "ftp://example.com", "https://user:pw@example.com", "http://localhost:8080", "http://metadata.internal"} {
		if _, err := publicPageURL(raw); err == nil {
			t.Errorf("publicPageURL(%q) accepted", raw)
		}
	}
	if parsed, err := publicPageURL("example.com/docs#intro"); err != nil || parsed.String() != "https://example.com/docs" {
		t.Fatalf("bare domain: %v %v", parsed, err)
	}
}

func TestWebFetchReturnsTextAndSource(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200, Header: http.Header{"Content-Type": {"text/html; charset=utf-8"}}, Request: r,
			Body: io.NopCloser(strings.NewReader(`<title>Docs</title><p>Hello world</p>`)),
		}, nil
	})}}
	output, executed, err := executor.ExecuteCaretTool(context.Background(), "web.fetch", map[string]any{"url": "https://example.com/docs"})
	if err != nil || !executed {
		t.Fatalf("web.fetch: executed=%v err=%v", executed, err)
	}
	if output["text"] != "Hello world" || output["title"] != "Docs" {
		t.Fatalf("output = %#v", output)
	}
	if sources, _ := output["sources"].([]map[string]string); len(sources) != 1 || sources[0]["url"] != "https://example.com/docs" {
		t.Fatalf("sources = %#v", output["sources"])
	}
}

func TestExportableDocumentURL(t *testing.T) {
	for raw, want := range map[string]string{
		"https://docs.google.com/document/d/abc_123/edit?usp=sharing": "https://docs.google.com/document/d/abc_123/export?format=txt",
		"https://docs.google.com/spreadsheets/d/XYZ/edit#gid=42":      "https://docs.google.com/spreadsheets/d/XYZ/export?format=csv&gid=42",
		"https://docs.google.com/presentation/d/P1/edit":              "https://docs.google.com/presentation/d/P1/export/txt",
		"https://example.com/document/d/abc/edit":                     "https://example.com/document/d/abc/edit",
	} {
		parsed, err := publicPageURL(raw)
		if err != nil {
			t.Fatal(err)
		}
		if got := exportableDocumentURL(parsed).String(); got != want {
			t.Errorf("%s -> %s, want %s", raw, got, want)
		}
	}
}
