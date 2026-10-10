package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/voxire/onion/internal/platform/config"
)

type publicWebSearcherFunc func(context.Context, string) (map[string]any, error)

func (f publicWebSearcherFunc) Search(ctx context.Context, query string) (map[string]any, error) {
	return f(ctx, query)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestAssistantToolsCalculateAndRejectUnsupportedExpressions(t *testing.T) {
	executor := NewAssistantToolExecutor(nil, nil)
	output, executed, err := executor.ExecuteCaretTool(context.Background(), "utilities.calculate", map[string]any{"expression": "(12 + 8) * 3 / 2"})
	if err != nil || !executed || output["result"] != float64(30) {
		t.Fatalf("calculate returned (%v, %v, %v), want 30", output, executed, err)
	}
	for _, expression := range []string{"1/0", "os.Exit(1)", "1**2", strings.Repeat("1", 513)} {
		if _, _, err := executor.ExecuteCaretTool(context.Background(), "utilities.calculate", map[string]any{"expression": expression}); err == nil {
			t.Errorf("calculate(%q) unexpectedly succeeded", expression)
		}
	}
}

func TestAssistantWebSearchRejectsPersonalDataAndLimitsCalls(t *testing.T) {
	var calls atomic.Int32
	executor := NewAssistantToolExecutor(nil, publicWebSearcherFunc(func(_ context.Context, query string) (map[string]any, error) {
		calls.Add(1)
		return map[string]any{"query": query}, nil
	}))
	ctx := WithWebSearchBudget(context.Background())
	if _, _, err := executor.ExecuteCaretTool(ctx, "web.search", map[string]any{"query": "Call me at +1 (555) 123-4567"}); err == nil {
		t.Fatal("expected personal phone number to be rejected")
	}
	for i := 1; i <= MaxWebSearchesPerMessage; i++ {
		if _, executed, err := executor.ExecuteCaretTool(ctx, "web.search", map[string]any{"query": fmt.Sprintf("Gemini release notes part %d", i)}); err != nil || !executed {
			t.Fatalf("search %d returned executed=%v err=%v", i, executed, err)
		}
	}
	if _, _, err := executor.ExecuteCaretTool(ctx, "web.search", map[string]any{"query": "another query"}); err == nil {
		t.Fatal("expected a search past the per-message limit to be rejected")
	}
	if got := calls.Load(); got != MaxWebSearchesPerMessage {
		t.Fatalf("search calls = %d, want %d", got, MaxWebSearchesPerMessage)
	}
}

func TestParseSearXNGResultsKeepsOnlySafeHTTPSResults(t *testing.T) {
	payload := []byte(`{"results":[{"title":"Search result","url":"https://example.com/source","content":"A public snippet."},{"title":"Unsafe scheme","url":"javascript:alert(1)"},{"title":"Credentials","url":"https://user:pass@example.com/private"},{"title":"Insecure","url":"http://example.com/source"}]}`)
	output, err := parseSearXNGResults(payload, "public query")
	if err != nil {
		t.Fatalf("parse results: %v", err)
	}
	if output["search_query"] != "public query" {
		t.Fatalf("search_query = %#v", output["search_query"])
	}
	sources, ok := output["sources"].([]map[string]string)
	if !ok || len(sources) != 1 || sources[0]["url"] != "https://example.com/source" || sources[0]["snippet"] != "A public snippet." {
		t.Fatalf("unsafe or missing search results: %#v", output["sources"])
	}
}

func TestParseSearXNGResultsSkipsOversizedURLAndKeepsLaterResults(t *testing.T) {
	longURL := "https://example.com/" + strings.Repeat("x", 2050)
	payload, err := json.Marshal(map[string]any{"results": []map[string]string{
		{"title": "Too long", "url": longURL, "content": "not usable"},
		{"title": "Valid result", "url": "https://example.com/valid", "content": "usable"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	output, err := parseSearXNGResults(payload, "query")
	if err != nil {
		t.Fatalf("parse results: %v", err)
	}
	sources, ok := output["sources"].([]map[string]string)
	if !ok || len(sources) != 1 || sources[0]["url"] != "https://example.com/valid" {
		t.Fatalf("valid result after oversized URL was lost: %#v", output["sources"])
	}
}

func TestSearXNGWebSearchUsesConfiguredJSONEndpoint(t *testing.T) {
	previousConfig := config.GlobalConfig
	config.GlobalConfig = &config.AppConfig{SearXNGURL: "https://search.example/base/"}
	t.Cleanup(func() { config.GlobalConfig = previousConfig })
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.URL.Host != "search.example" || request.URL.Path != "/base/search" {
			t.Fatalf("unexpected request target: %s %s", request.Method, request.URL.String())
		}
		if request.URL.Query().Get("q") != "public query" || request.URL.Query().Get("format") != "json" ||
			request.URL.Query().Get("categories") != "general,news" {
			t.Fatalf("unexpected search query: %s", request.URL.RawQuery)
		}
		if request.Header.Get("Authorization") != "" {
			t.Fatal("unexpected authorization header on a public search request")
		}
		payload := `{"results":[{"title":"Source","url":"https://example.com/source","content":"Search snippet."}]}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(payload))}, nil
	})}
	result, err := (SearXNGWebSearch{Client: client}).Search(context.Background(), "public query")
	if err != nil {
		t.Fatalf("SearXNG search: %v", err)
	}
	sources, ok := result["sources"].([]map[string]string)
	if !ok || len(sources) != 1 || sources[0]["snippet"] != "Search snippet." {
		t.Fatalf("search result is incomplete: %#v", result)
	}
}

func TestSearXNGWebSearchRequiresConfiguredEndpoint(t *testing.T) {
	previousConfig := config.GlobalConfig
	config.GlobalConfig = &config.AppConfig{}
	t.Cleanup(func() { config.GlobalConfig = previousConfig })
	if _, err := (SearXNGWebSearch{}).Search(context.Background(), "public query"); err == nil {
		t.Fatal("expected an unconfigured SearXNG endpoint to fail")
	}
}

func TestAssistantWebSearchSurfacesProviderFailure(t *testing.T) {
	executor := NewAssistantToolExecutor(nil, publicWebSearcherFunc(func(context.Context, string) (map[string]any, error) {
		return nil, errors.New("upstream failure")
	}))
	_, executed, err := executor.ExecuteCaretTool(context.Background(), "web.search", map[string]any{"query": "public query"})
	if !executed || err == nil || err.Error() != "upstream failure" {
		t.Fatalf("search returned executed=%v err=%v", executed, err)
	}
}

func TestSearXNGUnresponsiveEnginesNamesEachFailure(t *testing.T) {
	payload := []byte(`{"results":[],"unresponsive_engines":[["bing","timeout"],["qwant news","CAPTCHA"]]}`)
	if got := searXNGUnresponsiveEngines(payload); got != "bing: timeout, qwant news: CAPTCHA" {
		t.Fatalf("unresponsive engines = %q", got)
	}
}
