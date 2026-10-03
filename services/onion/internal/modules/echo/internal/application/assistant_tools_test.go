package application

import (
	"context"
	"encoding/json"
	"errors"
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
	if _, executed, err := executor.ExecuteCaretTool(ctx, "web.search", map[string]any{"query": "Gemini release notes from 2026-10-03"}); err != nil || !executed {
		t.Fatalf("first public search returned executed=%v err=%v", executed, err)
	}
	if _, _, err := executor.ExecuteCaretTool(ctx, "web.search", map[string]any{"query": "another query"}); err == nil {
		t.Fatal("expected the second search in one turn to be rejected")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("search calls = %d, want 1", got)
	}
}

func TestParseGeminiGroundedSearchRequiresSafeSourcesAndSuggestions(t *testing.T) {
	payload, err := json.Marshal(map[string]any{"steps": []any{
		map[string]any{"type": "google_search_result", "result": []any{map[string]any{"search_suggestions": `<div><a href="https://www.google.com/search?q=gemini">Gemini</a></div>`}}},
		map[string]any{"type": "model_output", "content": []any{map[string]any{"type": "text", "text": "Gemini supports grounded search.", "annotations": []any{
			map[string]any{"type": "url_citation", "url": "https://ai.google.dev/gemini-api/docs/google-search", "title": "Gemini Search docs"},
			map[string]any{"type": "url_citation", "url": "javascript:alert(1)", "title": "Unsafe"},
			map[string]any{"type": "url_citation", "url": "https://user:pass@example.com/private", "title": "Credentials"},
		}}}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	output, err := parseGeminiGroundedSearch(payload, "public query")
	if err != nil {
		t.Fatalf("parse grounded result: %v", err)
	}
	if output["answer"] != "Gemini supports grounded search." || output["search_query"] != "public query" || output["search_suggestions_html"] == "" {
		t.Fatalf("unexpected result: %#v", output)
	}
	sources, ok := output["sources"].([]map[string]string)
	if !ok || len(sources) != 1 || sources[0]["url"] != "https://ai.google.dev/gemini-api/docs/google-search" {
		t.Fatalf("unsafe or missing citation sources: %#v", output["sources"])
	}

	withoutSuggestions := []byte(`{"steps":[{"type":"google_search_result","result":[{"search_suggestions":""}]},{"type":"model_output","content":[{"type":"text","text":"Answer","annotations":[{"type":"url_citation","url":"https://example.com","title":"Example"}]}]}]}`)
	if _, err := parseGeminiGroundedSearch(withoutSuggestions, "public query"); err == nil {
		t.Fatal("expected a grounded answer without associated search suggestions to fail")
	}
}

func TestGeminiWebSearchUsesInteractionsGoogleSearchAndReturnsGrounding(t *testing.T) {
	previousConfig := config.GlobalConfig
	config.GlobalConfig = &config.AppConfig{GeminiAPIKey: "test-key"}
	t.Cleanup(func() { config.GlobalConfig = previousConfig })
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.Host != "generativelanguage.googleapis.com" || request.URL.Path != "/v1beta/interactions" {
			t.Fatalf("unexpected request target: %s %s", request.Method, request.URL.String())
		}
		if request.Header.Get("x-goog-api-key") != "test-key" {
			t.Fatal("Gemini API key header was not set")
		}
		var body struct {
			Model string `json:"model"`
			Input string `json:"input"`
			Tools []struct {
				Type string `json:"type"`
			} `json:"tools"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode search request: %v", err)
		}
		if body.Model != "gemini-3.8-flash" || body.Input != "public query" || len(body.Tools) != 1 || body.Tools[0].Type != "google_search" {
			t.Fatalf("unexpected search request body: %#v", body)
		}
		payload := `{"steps":[{"type":"google_search_result","result":[{"search_suggestions":"<a href=\"https://www.google.com/search?q=related\">Related</a>"}]},{"type":"model_output","content":[{"type":"text","text":"Grounded answer.","annotations":[{"type":"url_citation","url":"https://example.com/source","title":"Source"}]}]}]}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(payload))}, nil
	})}
	result, err := (GeminiWebSearch{Client: client}).Search(context.Background(), "public query")
	if err != nil {
		t.Fatalf("Gemini search: %v", err)
	}
	if result["answer"] != "Grounded answer." || result["search_suggestions_html"] == "" {
		t.Fatalf("grounded result is incomplete: %#v", result)
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
