package snips

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func geminiTestClient(t *testing.T, transport roundTripFunc) *genai.Client {
	t.Helper()

	client, err := genai.NewClient(context.Background(),
		option.WithAPIKey("test-key"),
		option.WithHTTPClient(&http.Client{Transport: transport}),
	)
	if err != nil {
		t.Fatalf("genai.NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func geminiJSONResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// A candidate without content (finishReason outside the SDK's blocked set,
// e.g. MAX_TOKENS) used to dereference a nil Content and panic the whole
// non-streaming request; it must fall through to the next model instead.
func TestCompleteTextWithOptionsSkipsCandidateWithoutContent(t *testing.T) {
	var calls int32
	client := geminiTestClient(t, func(*http.Request) (*http.Response, error) {
		if atomic.AddInt32(&calls, 1) == 1 {
			return geminiJSONResponse(`{"candidates":[{"content":null,"finishReason":"MAX_TOKENS"}]}`), nil
		}
		return geminiJSONResponse(`{"candidates":[{"content":{"role":"model","parts":[{"text":"second model"}]},"finishReason":"STOP"}]}`), nil
	})

	provider := &GeminiProvider{client: client, models: []string{"first-model", "second-model"}}
	got, err := provider.CompleteTextWithOptions(context.Background(), "prompt", CompletionOptions{})
	if err != nil {
		t.Fatalf("CompleteTextWithOptions: %v", err)
	}
	if got != "second model" {
		t.Fatalf("got %q, want the second model's text", got)
	}
	if calls != 2 {
		t.Fatalf("generate calls = %d, want 2", calls)
	}
}

// Content-less candidates from every model must surface the aggregate error
// rather than panicking or returning an empty success.
func TestCompleteTextWithOptionsContentlessCandidatesReturnError(t *testing.T) {
	client := geminiTestClient(t, func(*http.Request) (*http.Response, error) {
		return geminiJSONResponse(`{"candidates":[{"content":null,"finishReason":"PROHIBITED_CONTENT"}]}`), nil
	})

	provider := &GeminiProvider{client: client, models: []string{"first-model", "second-model"}}
	got, err := provider.CompleteTextWithOptions(context.Background(), "prompt", CompletionOptions{})
	if !errors.Is(err, ErrGeminiAllModelsFailed) {
		t.Fatalf("err = %v, want ErrGeminiAllModelsFailed", err)
	}
	if got != "" {
		t.Fatalf("got %q, want empty result", got)
	}
}
