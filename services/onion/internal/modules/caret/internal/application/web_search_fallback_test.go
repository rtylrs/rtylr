package application

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubWebSearch struct {
	result map[string]any
	err    error
	calls  int
}

func (s *stubWebSearch) Search(context.Context, string) (map[string]any, error) {
	s.calls++
	return s.result, s.err
}

func TestFallbackWebSearchUsesFallbackOnlyWhenPrimaryFails(t *testing.T) {
	primaryResult := map[string]any{"from": "primary"}
	primary := &stubWebSearch{result: primaryResult}
	fallback := &stubWebSearch{result: map[string]any{"from": "fallback"}}
	got, err := FallbackWebSearch{Primary: primary, Fallback: fallback}.Search(context.Background(), "q")
	if err != nil || got["from"] != "primary" || fallback.calls != 0 {
		t.Fatalf("healthy primary: got %v, err %v, fallback calls %d", got, err, fallback.calls)
	}

	primary = &stubWebSearch{err: errors.New("web search returned no usable results")}
	got, err = FallbackWebSearch{Primary: primary, Fallback: fallback}.Search(context.Background(), "q")
	if err != nil || got["from"] != "fallback" {
		t.Fatalf("failed primary: got %v, err %v", got, err)
	}
}

func TestFallbackWebSearchKeepsPrimaryErrorWhenFallbackUnconfigured(t *testing.T) {
	primaryErr := errors.New("web search returned no usable results")
	_, err := FallbackWebSearch{
		Primary:  &stubWebSearch{err: primaryErr},
		Fallback: &stubWebSearch{err: errWebSearchNotConfigured},
	}.Search(context.Background(), "q")
	if !errors.Is(err, primaryErr) {
		t.Fatalf("err = %v, want primary error", err)
	}
}

func TestBraveWebSearchSendsKeyAndStripsHighlighting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Subscription-Token") != "key" || r.URL.Query().Get("q") != "plague news" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"web":{"results":[
			{"title":"Suspected <strong>plague</strong> case","url":"https://example.com/a","description":"Lab worker &amp; quarantine"},
			{"title":"Insecure","url":"http://example.com/b","description":"dropped"}
		]}}`))
	}))
	defer server.Close()

	got, err := BraveWebSearch{Client: server.Client(), APIKey: "key", Endpoint: server.URL}.Search(context.Background(), "plague news")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	sources := got["sources"].([]map[string]string)
	if len(sources) != 1 || sources[0]["title"] != "Suspected plague case" || sources[0]["snippet"] != "Lab worker & quarantine" {
		t.Fatalf("sources = %v", sources)
	}
}

func TestBraveWebSearchIsInertWithoutKey(t *testing.T) {
	t.Setenv("BRAVE_SEARCH_API_KEY", "")
	_, err := BraveWebSearch{}.Search(context.Background(), "q")
	if !errors.Is(err, errWebSearchNotConfigured) {
		t.Fatalf("err = %v, want not configured", err)
	}
}
