package snips

import (
	"context"
	"errors"
	"testing"
)

func TestClientTripsProviderCooldownAfterRepeatedCompletionFailures(t *testing.T) {
	failing := &plainCompletionProvider{err: errors.New("upstream down")}
	client := &Client{providers: []Provider{failing}}

	for i := 0; i < 2; i++ {
		if _, err := client.CompleteText(context.Background(), "prompt"); !errors.Is(err, ErrAllProvidersFailed) {
			t.Fatalf("attempt %d error = %v, want ErrAllProvidersFailed", i+1, err)
		}
	}
	if failing.plainCalls != 2 {
		t.Fatalf("plainCalls = %d, want 2 before cooldown", failing.plainCalls)
	}
	if _, err := client.CompleteText(context.Background(), "prompt"); !errors.Is(err, ErrAllProvidersFailed) {
		t.Fatalf("cooldown error = %v, want ErrAllProvidersFailed", err)
	}
	if failing.plainCalls != 2 {
		t.Fatalf("plainCalls = %d, want cooldown to skip provider", failing.plainCalls)
	}
}

func TestClientClearsProviderCooldownAfterSuccess(t *testing.T) {
	provider := &plainCompletionProvider{err: errors.New("first failure")}
	client := &Client{providers: []Provider{provider}}

	_, _ = client.CompleteText(context.Background(), "prompt")
	provider.err = nil
	provider.result = "recovered"
	if got, err := client.CompleteText(context.Background(), "prompt"); err != nil || got != "recovered" {
		if err != nil {
			t.Fatalf("recovery completion: %v", err)
		}
		t.Fatalf("recovery result = %q, want recovered", got)
	}
	if !client.providerReady(provider) {
		t.Fatal("provider should remain available after a successful attempt")
	}
}
