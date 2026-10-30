package snips

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/openai/openai-go/v2/option"

	"github.com/voxire/onion/internal/platform/config"
)

func useOpenCodeConfig(t *testing.T, cfg config.AppConfig) {
	t.Helper()
	original := config.GlobalConfig
	config.GlobalConfig = &cfg
	t.Cleanup(func() { config.GlobalConfig = original })
}

func TestOpenCodeRegistration(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  config.AppConfig
		want []string
	}{
		{name: "no keys"},
		{name: "alone", cfg: config.AppConfig{OpenCodeAPIKey: "key"}, want: []string{"opencode"}},
		{name: "OpenCode then Gemini defaults", cfg: config.AppConfig{DeepSeekAPIKey: "key", GeminiAPIKey: "key", OpenAIAPIKey: "key", OpenCodeAPIKey: "key"}, want: []string{"opencode", "gemini", "deepseek", "openai"}},
		{name: "Gemini primary", cfg: config.AppConfig{DeepSeekAPIKey: "key", GeminiAPIKey: "key", OpenCodeAPIKey: "key", AIPrimaryProvider: "gemini"}, want: []string{"gemini", "opencode", "deepseek"}},
		{name: "OpenCode primary", cfg: config.AppConfig{DeepSeekAPIKey: "key", GeminiAPIKey: "key", OpenCodeAPIKey: "key", AIPrimaryProvider: "opencode"}, want: []string{"opencode", "gemini", "deepseek"}},
		{name: "unconfigured preference", cfg: config.AppConfig{GeminiAPIKey: "key", OpenCodeAPIKey: "key", AIPrimaryProvider: "deepseek"}, want: []string{"opencode", "gemini"}},
		{name: "retired preference", cfg: config.AppConfig{GeminiAPIKey: "key", AIPrimaryProvider: "nvidia"}, want: []string{"gemini"}},
		{name: "disabled", cfg: config.AppConfig{OpenAIAPIKey: "key"}, want: []string{"openai"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			useOpenCodeConfig(t, tc.cfg)
			client, err := NewClient(context.Background())
			if tc.want == nil {
				if !errors.Is(err, ErrNoProvidersConfigured) {
					t.Fatalf("expected no providers, got %v", err)
				}
				if _, err := NewOpenCodeProvider(); !errors.Is(err, ErrOpenCodeNotConfigured) {
					t.Fatalf("expected missing key error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			var names []string
			for _, p := range client.providers {
				names = append(names, p.Name())
			}
			if !reflect.DeepEqual(names, tc.want) || client.GetPrimaryProvider() != tc.want[0] {
				t.Fatalf("providers=%v primary=%s; want %v", names, client.GetPrimaryProvider(), tc.want)
			}
			if client.GetProviderStatus()["opencode"] != (tc.cfg.OpenCodeAPIKey != "") {
				t.Fatal("incorrect OpenCode status")
			}
		})
	}
}

type openCodeTransport func(*http.Request) (*http.Response, error)

func (f openCodeTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func mockOpenCode(t *testing.T, respond func(map[string]any) (string, string)) *OpenAIProvider {
	t.Helper()
	useOpenCodeConfig(t, config.AppConfig{OpenCodeAPIKey: "test-zen-key"})
	p, err := NewOpenCodeProvider()
	if err != nil {
		t.Fatal(err)
	}
	// Attach the mock to the actual SDK service so no network request can escape.
	p.client.Chat.Completions.Options = append(p.client.Chat.Completions.Options, option.WithHTTPClient(&http.Client{
		Transport: openCodeTransport(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost || r.URL.String() != "https://opencode.ai/zen/go/v1/chat/completions" {
				t.Errorf("unexpected endpoint: %s %s", r.Method, r.URL)
			}
			if r.Header.Get("Authorization") != "Bearer test-zen-key" {
				t.Error("OpenCode key was not used")
			}
			if r.Header.Get("User-Agent") != "rtylr-caret/1.0" || r.Header.Get("x-opencode-session") == "" {
				t.Error("missing OpenCode client/session headers")
			}
			defer r.Body.Close()
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				return nil, err
			}
			if body["model"] != "deepseek-v4.1-flash" || body["max_tokens"] != float64(config.AIMaxOutputTokens) {
				t.Errorf("unexpected model/token cap: %v", body)
			}
			thinking, _ := body["thinking"].(map[string]any)
			if thinking["type"] != "disabled" {
				t.Error("DeepSeek thinking must be disabled")
			}
			contentType, content := respond(body)
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(content)), Request: r}, nil
		}),
	}))
	return p
}

func TestOpenCodeCompletionFallback(t *testing.T) {
	for _, jsonMode := range []bool{false, true} {
		t.Run(map[bool]string{false: "text", true: "JSON"}[jsonMode], func(t *testing.T) {
			calls := 0
			p := mockOpenCode(t, func(body map[string]any) (string, string) {
				calls++
				if jsonMode {
					format, _ := body["response_format"].(map[string]any)
					if body["temperature"] != float64(0) || format["type"] != "json_object" {
						t.Errorf("missing planning options: %v", body)
					}
				}
				return "application/json", `{"choices":[{"message":{"content":"answer"}}]}`
			})
			primary := &plainCompletionProvider{err: errors.New("unavailable")}
			client := &Client{providers: []Provider{primary, p}}
			opts := CompletionOptions{}
			if jsonMode {
				opts = DeterministicJSONCompletionOptions()
			}
			answer, err := client.CompleteTextWithOptions(context.Background(), "prompt", opts)
			if err != nil || answer != "answer" || calls != 1 || primary.plainCalls != 1 {
				t.Fatalf("answer=%q err=%v fallback calls=%d primary calls=%d", answer, err, calls, primary.plainCalls)
			}
		})
	}
}

func TestOpenCodeStreamFallback(t *testing.T) {
	for _, tc := range []struct {
		name     string
		tokens   []StreamToken
		fallback bool
	}{
		{name: "early error", tokens: []StreamToken{{Error: "unavailable"}}, fallback: true},
		{name: "early close", fallback: true},
		{name: "partial response", tokens: []StreamToken{{Text: "partial"}, {Error: "unavailable"}}},
		{name: "primary succeeds", tokens: []StreamToken{{Text: "primary"}, {Done: true}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			p := mockOpenCode(t, func(body map[string]any) (string, string) {
				calls++
				options, _ := body["stream_options"].(map[string]any)
				if body["stream"] != true || options["include_usage"] != true {
					t.Error("stream usage was not requested")
				}
				return "text/event-stream", "data: {\"choices\":[{\"delta\":{\"content\":\"answer\"}}]}\n\ndata: {\"choices\":[],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":5,\"total_tokens\":15}}\n\ndata: [DONE]\n\n"
			})
			primary := &streamingTestProvider{tokens: tc.tokens}
			client := &Client{providers: []Provider{primary, p}, primaryProvider: primary}
			stream, err := client.CompleteTextStream(context.Background(), "prompt")
			if err != nil {
				t.Fatal(err)
			}
			var tokens []StreamToken
			for token := range stream {
				tokens = append(tokens, token)
			}
			if !tc.fallback {
				if calls != 0 || len(tokens) != 2 || tokens[0].Text != tc.tokens[0].Text || tokens[1].Error != tc.tokens[1].Error {
					t.Fatalf("primary stream changed: calls=%d tokens=%v", calls, tokens)
				}
				return
			}
			if calls != 1 || len(tokens) != 2 || tokens[0].Text != "answer" || tokens[0].Provider != "opencode" {
				t.Fatalf("incorrect fallback: calls=%d tokens=%v", calls, tokens)
			}
			if done := tokens[1]; !done.Done || done.Error != "" || done.Provider != "opencode" || done.PromptTokens != 10 || done.CompletionTokens != 5 {
				t.Fatalf("incorrect usage: %v", done)
			}
		})
	}
}

func TestOpenCodeConversationRouting(t *testing.T) {
	p := mockOpenCode(t, func(body map[string]any) (string, string) {
		if body["stream"] == true {
			return "text/event-stream", "data: {\"choices\":[{\"delta\":{\"content\":\"answer\"}}]}\n\ndata: [DONE]\n\n"
		}
		return "application/json", `{"choices":[{"message":{"content":"answer"}}]}`
	})
	var sessions []string
	observer := p.client.Chat.Completions.Options
	p.client.Chat.Completions.Options = append(observer, option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		sessions = append(sessions, r.Header.Get("x-opencode-session"))
		return next(r)
	}))
	for _, id := range []string{"chat-one", "chat-one", "chat-two", "", ""} {
		ctx := WithConversation(context.Background(), id)
		if _, err := p.CompleteText(ctx, "prompt"); err != nil {
			t.Fatal(err)
		}
	}
	stream, err := p.CompleteTextStream(WithConversation(context.Background(), "chat-one"), "prompt")
	if err != nil {
		t.Fatal(err)
	}
	var answer string
	for token := range stream {
		if token.Error != "" {
			t.Fatal(token.Error)
		}
		answer += token.Text
	}
	if answer != "answer" {
		t.Fatalf("unexpected stream answer: %q", answer)
	}
	if len(sessions) != 6 || sessions[5] != sessions[0] || sessions[0] == "chat-one" || sessions[0] != sessions[1] || sessions[0] == sessions[2] || sessions[3] == sessions[4] {
		t.Fatalf("conversation routing is not stable and isolated: %v", sessions)
	}
}

func TestOpenCodeDeepSeekFlashImageInput(t *testing.T) {
	image := ImageInput{DataURL: "data:image/png;base64,aGVsbG8="}
	calls := 0
	p := mockOpenCode(t, func(body map[string]any) (string, string) {
		calls++
		messages := body["messages"].([]any)
		parts := messages[0].(map[string]any)["content"].([]any)
		textPart := parts[0].(map[string]any)
		imagePart := parts[1].(map[string]any)
		if textPart["type"] != "text" || textPart["text"] != "describe" || imagePart["type"] != "image_url" || imagePart["image_url"].(map[string]any)["url"] != image.DataURL {
			t.Errorf("unexpected content: %v", parts)
		}
		if body["stream"] == true {
			return "text/event-stream", "data: {\"choices\":[{\"delta\":{\"content\":\"answer\"}}]}\n\ndata: [DONE]\n\n"
		}
		return "application/json", `{"choices":[{"message":{"content":"answer"}}]}`
	})
	ctx := WithImages(context.Background(), []ImageInput{image})
	client := &Client{providers: []Provider{p}, primaryProvider: p}
	if answer, err := client.CompleteText(ctx, "describe"); err != nil || answer != "answer" {
		t.Fatalf("answer=%q err=%v", answer, err)
	}
	stream, err := client.CompleteTextStream(ctx, "describe")
	if err != nil {
		t.Fatal(err)
	}
	for token := range stream {
		if token.Error != "" {
			t.Fatal(token.Error)
		}
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
	p.model = "deepseek-chat"
	if providerSupportsInput(ctx, p) {
		t.Fatal("legacy text-only model accepted images")
	}
	p.model = "deepseek-v4.1-flash-1017"
	if !providerSupportsInput(ctx, p) {
		t.Fatal("versioned flash model rejected images")
	}
}
