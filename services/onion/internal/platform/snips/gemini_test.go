package snips

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"google.golang.org/genai"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func geminiTestClient(t *testing.T, transport roundTripFunc) *genai.Client {
	t.Helper()

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:     "test-key",
		Backend:    genai.BackendGeminiAPI,
		HTTPClient: &http.Client{Transport: transport},
	})
	if err != nil {
		t.Fatalf("genai.NewClient: %v", err)
	}
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

func TestCompleteTextWithOptionsUsesLowThinkingAndPreservesImageInput(t *testing.T) {
	var imageData bytes.Buffer
	if err := png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	imageInput := ImageInput{Name: "tiny.png", DataURL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData.Bytes())}
	client := geminiTestClient(t, func(request *http.Request) (*http.Response, error) {
		var body struct {
			Contents []struct {
				Parts []map[string]json.RawMessage `json:"parts"`
			} `json:"contents"`
			GenerationConfig struct {
				ThinkingConfig struct {
					ThinkingLevel string `json:"thinkingLevel"`
				} `json:"thinkingConfig"`
				ResponseMIMEType string  `json:"responseMimeType"`
				Temperature      float64 `json:"temperature"`
			} `json:"generationConfig"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.GenerationConfig.ThinkingConfig.ThinkingLevel != string(genai.ThinkingLevelLow) {
			t.Errorf("thinking level = %q, want %q", body.GenerationConfig.ThinkingConfig.ThinkingLevel, genai.ThinkingLevelLow)
		}
		if body.GenerationConfig.ResponseMIMEType != "application/json" || body.GenerationConfig.Temperature != 0.2 {
			t.Errorf("generation config = %+v, want JSON and temperature 0.2", body.GenerationConfig)
		}
		if len(body.Contents) != 1 {
			t.Errorf("contents = %d, want one user message", len(body.Contents))
		} else if len(body.Contents[0].Parts) != 2 {
			t.Errorf("contents parts = %d, want prompt and image", len(body.Contents[0].Parts))
		} else if _, ok := body.Contents[0].Parts[1]["inlineData"]; !ok {
			t.Errorf("second content part does not contain image data: %v", body.Contents[0].Parts[1])
		}
		return geminiJSONResponse(`{"candidates":[{"content":{"role":"model","parts":[{"text":"ready"}]},"finishReason":"STOP"}]}`), nil
	})
	provider := &GeminiProvider{client: client, models: []string{"test-model"}}
	temperature := 0.2
	got, err := provider.CompleteTextWithOptions(WithImages(context.Background(), []ImageInput{imageInput}), "describe this", CompletionOptions{Temperature: &temperature, JSONResponse: true})
	if err != nil {
		t.Fatalf("CompleteTextWithOptions: %v", err)
	}
	if got != "ready" {
		t.Fatalf("got %q, want ready", got)
	}
}

func TestCompleteTextStreamSuppressesThoughtPartsAndCompletes(t *testing.T) {
	client := geminiTestClient(t, func(request *http.Request) (*http.Response, error) {
		if !strings.Contains(request.URL.Path, ":streamGenerateContent") {
			t.Errorf("request path = %q, want streaming endpoint", request.URL.Path)
		}
		body := strings.Join([]string{
			`data: {"candidates":[{"content":{"role":"model","parts":[{"text":"private reasoning","thought":true},{"text":"Hello"}]}}]}`,
			`data: {"candidates":[{"content":{"role":"model","parts":[{"text":" there"}]}}]}`,
		}, "\n\n") + "\n\n"
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	provider := &GeminiProvider{client: client, models: []string{"test-model"}}
	stream, err := provider.CompleteTextStream(context.Background(), "say hello")
	if err != nil {
		t.Fatalf("CompleteTextStream: %v", err)
	}
	var got strings.Builder
	done := false
	for token := range stream {
		if token.Error != "" {
			t.Fatalf("stream error: %s", token.Error)
		}
		if token.Done {
			done = true
		}
		got.WriteString(token.Text)
	}
	if got.String() != "Hello there" {
		t.Fatalf("stream text = %q, want public text only", got.String())
	}
	if !done {
		t.Fatal("stream did not send a completion token")
	}
}
