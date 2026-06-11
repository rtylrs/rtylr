package snips

import (
	"context"
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
