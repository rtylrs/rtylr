package snips

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/openai/openai-go/v2/option"
	"github.com/voxire/onion/internal/platform/config"
)

func TestNvidiaReturnsVisibleAnswerWithoutReasoning(t *testing.T) {
	for _, streaming := range []bool{false, true} {
		t.Run(map[bool]string{false: "completion", true: "stream"}[streaming], func(t *testing.T) {
			useOpenCodeConfig(t, config.AppConfig{NvidiaNIMAPIKey: "fixture-key"})
			provider, err := NewNvidiaNIMProvider()
			if err != nil {
				t.Fatal(err)
			}
			provider.client.Chat.Completions.Options = append(provider.client.Chat.Completions.Options, option.WithHTTPClient(&http.Client{Transport: openCodeTransport(func(r *http.Request) (*http.Response, error) {
				defer r.Body.Close()
				if r.URL.String() != "https://integrate.api.nvidia.com/v1/chat/completions" {
					t.Errorf("unexpected endpoint: %s", r.URL)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					return nil, err
				}
				if body["model"] != "nvidia/nemotron-3.5-lightning-30b-a3b" {
					t.Errorf("retired model requested: %v", body["model"])
				}
				opts, _ := body["chat_template_kwargs"].(map[string]any)
				if opts["enable_thinking"] != false {
					t.Error("reasoning was not disabled")
				}
				content := `{"choices":[{"message":{"content":"Visible answer"}}]}`
				kind := "application/json"
				if streaming {
					kind = "text/event-stream"
					content = "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"Private reasoning\"}}]}\n\n" +
						"data: {\"choices\":[{\"delta\":{\"content\":\"Visible answer\"}}]}\n\n" +
						"data: {\"choices\":[],\"usage\":{\"prompt_tokens\":7,\"completion_tokens\":3,\"total_tokens\":10}}\n\n" +
						"data: [DONE]\n\n"
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {kind}}, Body: io.NopCloser(strings.NewReader(content)), Request: r}, nil
			})}))
			if !streaming {
				answer, err := provider.CompleteText(context.Background(), "Synthetic question")
				if err != nil || answer != "Visible answer" {
					t.Fatalf("answer=%q err=%v", answer, err)
				}
				return
			}
			tokens, err := provider.CompleteTextStream(context.Background(), "Synthetic question")
			if err != nil {
				t.Fatal(err)
			}
			var answer strings.Builder
			var done bool
			for token := range tokens {
				if token.Error != "" {
					t.Fatal(token.Error)
				}
				answer.WriteString(token.Text)
				if token.Done {
					done = true
					if token.PromptTokens != 7 || token.CompletionTokens != 3 {
						t.Error("missing provider usage")
					}
				}
			}
			if answer.String() != "Visible answer" || !done {
				t.Fatalf("answer=%q done=%v", answer.String(), done)
			}
		})
	}
}
