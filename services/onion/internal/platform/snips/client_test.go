package snips

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEnhanceStreamWithProviderAddsProviderName(t *testing.T) {
	source := make(chan StreamToken, 1)
	source <- StreamToken{Text: "hello"}
	close(source)

	out := enhanceStreamWithProvider(context.Background(), "test-provider", source)

	got, ok := <-out
	if !ok {
		t.Fatal("enhanced stream closed before delivering token")
	}
	if got.Provider != "test-provider" {
		t.Fatalf("Provider = %q, want %q", got.Provider, "test-provider")
	}
	if got.Text != "hello" {
		t.Fatalf("Text = %q, want %q", got.Text, "hello")
	}
	if _, ok := <-out; ok {
		t.Fatal("enhanced stream should close after source stream closes")
	}
}

func TestEnhanceStreamWithProviderStopsWhenContextCancelsUnderBackpressure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	source := make(chan StreamToken, 101)
	for i := 0; i < 101; i++ {
		source <- StreamToken{Text: "token"}
	}
	close(source)

	out := enhanceStreamWithProvider(ctx, "test-provider", source)
	waitUntilBuffered(t, out, cap(out))

	cancel()

	for i := 0; i < cap(out); i++ {
		if _, ok := <-out; !ok {
			t.Fatalf("enhanced stream closed after %d buffered tokens, want %d", i, cap(out))
		}
	}

	select {
	case token, ok := <-out:
		if ok {
			t.Fatalf("enhanced stream delivered extra token after cancellation: %+v", token)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("enhanced stream did not close after context cancellation")
	}
}

func TestSendTextTokenWaitsForConsumer(t *testing.T) {
	provider := &OpenAIProvider{}
	tokens := make(chan StreamToken)
	done := make(chan bool, 1)

	go func() {
		done <- provider.sendTextToken(context.Background(), tokens, "hello")
	}()

	select {
	case token := <-tokens:
		if token.Text != "hello" {
			t.Fatalf("Text = %q, want hello", token.Text)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("sendTextToken did not wait for a consumer")
	}

	if ok := <-done; !ok {
		t.Fatal("sendTextToken = false, want true after delivery")
	}
}

func TestSendTextTokenStopsOnContextCancellation(t *testing.T) {
	provider := &OpenAIProvider{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if ok := provider.sendTextToken(ctx, make(chan StreamToken), "hello"); ok {
		t.Fatal("sendTextToken = true, want false after context cancellation")
	}
}

func waitUntilBuffered(t *testing.T, ch <-chan StreamToken, want int) {
	t.Helper()

	deadline := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(ch) == want {
			return
		}
		time.Sleep(time.Millisecond)
	}

	t.Fatalf("channel buffered %d tokens, want %d", len(ch), want)
}

func TestCompleteTextWithOptionsPassesOptionsToProvider(t *testing.T) {
	temperature := 0.0
	provider := &recordingCompletionProvider{
		plainCompletionProvider: plainCompletionProvider{result: "ok"},
	}
	client := &Client{providers: []Provider{provider}}

	got, err := client.CompleteTextWithOptions(context.Background(), "prompt", CompletionOptions{
		Temperature:  &temperature,
		JSONResponse: true,
	})
	if err != nil {
		t.Fatalf("CompleteTextWithOptions: %v", err)
	}
	if got != "ok" {
		t.Fatalf("got %q, want ok", got)
	}
	if provider.optionCalls != 1 {
		t.Fatalf("optionCalls = %d, want 1", provider.optionCalls)
	}
	if provider.plainCalls != 0 {
		t.Fatalf("plainCalls = %d, want 0", provider.plainCalls)
	}
	if provider.lastOptions.Temperature == nil || *provider.lastOptions.Temperature != 0 {
		t.Fatalf("Temperature = %#v, want 0", provider.lastOptions.Temperature)
	}
	if !provider.lastOptions.JSONResponse {
		t.Fatal("JSONResponse option was not forwarded")
	}
}

func TestCompleteTextWithOptionsFallsBackToPlainProvider(t *testing.T) {
	provider := &plainCompletionProvider{result: "plain-ok"}
	client := &Client{providers: []Provider{provider}}

	got, err := client.CompleteTextWithOptions(context.Background(), "prompt", DeterministicJSONCompletionOptions())
	if err != nil {
		t.Fatalf("CompleteTextWithOptions: %v", err)
	}
	if got != "plain-ok" {
		t.Fatalf("got %q, want plain-ok", got)
	}
	if provider.plainCalls != 1 {
		t.Fatalf("plainCalls = %d, want 1", provider.plainCalls)
	}
}

func TestCompleteTextStreamFallsBackWithoutRetryingPrimary(t *testing.T) {
	primary := &streamingTestProvider{streamErr: errors.New("primary unavailable")}
	fallback := &streamingTestProvider{tokens: []StreamToken{{Text: "fallback"}, {Done: true}}}
	client := &Client{
		providers:       []Provider{primary, fallback},
		primaryProvider: primary,
	}

	stream, err := client.CompleteTextStream(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("CompleteTextStream: %v", err)
	}
	first := <-stream
	if first.Text != "fallback" || first.Provider != "stream-test" {
		t.Fatalf("first token = %#v, want fallback token", first)
	}
	if primary.streamCalls != 1 || fallback.streamCalls != 1 {
		t.Fatalf("stream calls: primary=%d fallback=%d, want one each", primary.streamCalls, fallback.streamCalls)
	}
}

func TestCompleteTextStreamFailsOverBeforeVisibleText(t *testing.T) {
	primary := &streamingTestProvider{tokens: []StreamToken{{Error: "primary stream failed"}}}
	fallback := &streamingTestProvider{tokens: []StreamToken{{Text: "fallback"}, {Done: true}}}
	client := &Client{providers: []Provider{primary, fallback}, primaryProvider: primary}

	stream, err := client.CompleteTextStream(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("CompleteTextStream: %v", err)
	}
	var text string
	for token := range stream {
		if token.Error != "" {
			t.Fatalf("failover surfaced an error: %#v", token)
		}
		text += token.Text
	}
	if text != "fallback" || fallback.streamCalls != 1 {
		t.Fatalf("text=%q fallback calls=%d", text, fallback.streamCalls)
	}
}

func TestCompleteTextStreamDoesNotSpliceAfterVisibleProviderFailure(t *testing.T) {
	primary := &streamingTestProvider{tokens: []StreamToken{{Text: "partial"}, {Error: "primary stream failed"}}}
	fallback := &streamingTestProvider{tokens: []StreamToken{{Text: "fallback"}, {Done: true}}}
	client := &Client{providers: []Provider{primary, fallback}, primaryProvider: primary}

	stream, err := client.CompleteTextStream(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("CompleteTextStream: %v", err)
	}
	var tokens []StreamToken
	for token := range stream {
		tokens = append(tokens, token)
	}
	if fallback.streamCalls != 0 || len(tokens) != 2 || tokens[0].Text != "partial" || tokens[1].Error == "" {
		t.Fatalf("tokens=%#v fallback calls=%d", tokens, fallback.streamCalls)
	}
}

func TestCompleteTextStreamContinuesAfterSelectedFallbackWithoutRetryingIt(t *testing.T) {
	primary := &streamingTestProvider{name: "primary", tokens: []StreamToken{{Error: "primary stream failed"}}}
	coolingDown := &streamingTestProvider{name: "cooling-down", tokens: []StreamToken{{Text: "should not run"}, {Done: true}}}
	selected := &streamingTestProvider{name: "selected", tokens: []StreamToken{{Error: "selected stream failed"}}}
	last := &streamingTestProvider{name: "last", tokens: []StreamToken{{Text: "last provider"}, {Done: true}}}
	client := &Client{
		providers:       []Provider{primary, coolingDown, selected, last},
		primaryProvider: primary,
		cooldownUntil:   map[string]time.Time{"cooling-down": time.Now().Add(time.Minute)},
	}

	primaryStream, err := primary.CompleteTextStream(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("start primary stream: %v", err)
	}
	stream := client.superviseProviderStream(context.Background(), "prompt", primary, primaryStream, []Provider{coolingDown, selected, last})
	var text string
	for token := range stream {
		if token.Error != "" {
			t.Fatalf("failover surfaced an error: %#v", token)
		}
		text += token.Text
	}

	if text != "last provider" {
		t.Fatalf("stream text = %q, want last provider", text)
	}
	if primary.streamCalls != 1 || coolingDown.streamCalls != 0 || selected.streamCalls != 1 || last.streamCalls != 1 {
		t.Fatalf("stream calls: primary=%d cooling-down=%d selected=%d last=%d; want 1, 0, 1, 1", primary.streamCalls, coolingDown.streamCalls, selected.streamCalls, last.streamCalls)
	}
}

type plainCompletionProvider struct {
	result     string
	err        error
	plainCalls int
}

func (p *plainCompletionProvider) Name() string      { return "plain" }
func (p *plainCompletionProvider) IsAvailable() bool { return true }
func (p *plainCompletionProvider) CompleteText(_ context.Context, _ string) (string, error) {
	p.plainCalls++
	if p.err != nil {
		return "", p.err
	}
	if p.result == "" {
		return "", errors.New("empty test result")
	}
	return p.result, nil
}

func (p *plainCompletionProvider) CompleteTextStream(_ context.Context, _ string) (<-chan StreamToken, error) {
	return nil, errors.New("not implemented")
}
func (p *plainCompletionProvider) Close() error { return nil }

type recordingCompletionProvider struct {
	plainCompletionProvider
	optionCalls int
	lastOptions CompletionOptions
}

type streamingTestProvider struct {
	plainCompletionProvider
	name        string
	tokens      []StreamToken
	streamErr   error
	streamCalls int
}

func (p *streamingTestProvider) Name() string {
	if p.name != "" {
		return p.name
	}
	return "stream-test"
}

func (p *streamingTestProvider) CompleteTextStream(_ context.Context, _ string) (<-chan StreamToken, error) {
	p.streamCalls++
	if p.streamErr != nil {
		return nil, p.streamErr
	}

	out := make(chan StreamToken, len(p.tokens))
	for _, token := range p.tokens {
		out <- token
	}
	close(out)
	return out, nil
}

func (p *recordingCompletionProvider) CompleteTextWithOptions(_ context.Context, _ string, opts CompletionOptions) (string, error) {
	p.optionCalls++
	p.lastOptions = opts
	if p.err != nil {
		return "", p.err
	}
	if p.result == "" {
		return "", errors.New("empty test result")
	}
	return p.result, nil
}
