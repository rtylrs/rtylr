package application

import (
	"context"
	"errors"
	"testing"
)

// Service.aiClient == nil is the only state reachable without network calls,
// and it is a valid production edge case (AI provider keys absent). All tests
// below exercise that code path exclusively.

func newNilClientService() *Service {
	return &Service{aiClient: nil}
}

func TestIsAvailable_NilClient_ReturnsFalse(t *testing.T) {
	s := newNilClientService()
	if s.IsAvailable() {
		t.Error("expected false when aiClient is nil")
	}
}

func TestGetProviderStatus_NilClient_ReturnsEmptyMap(t *testing.T) {
	s := newNilClientService()
	got := s.GetProviderStatus()
	if got == nil {
		t.Fatal("expected non-nil map")
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestGetPrimaryProvider_NilClient_ReturnsEmpty(t *testing.T) {
	s := newNilClientService()
	got := s.GetPrimaryProvider()
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestGetProviderDetails_NilClient_ReturnsUnavailable(t *testing.T) {
	s := newNilClientService()
	got := s.GetProviderDetails(context.Background())
	if got.Available {
		t.Error("expected Available=false")
	}
	if got.PrimaryProvider != "" {
		t.Errorf("expected empty PrimaryProvider, got %q", got.PrimaryProvider)
	}
	if got.Providers == nil {
		t.Error("expected non-nil Providers map")
	}
}

func TestGetProviderDetails_NilClient_ProvidersMapIsEmpty(t *testing.T) {
	s := newNilClientService()
	got := s.GetProviderDetails(context.Background())
	if len(got.Providers) != 0 {
		t.Errorf("expected empty providers map, got %v", got.Providers)
	}
}

func TestGenerateTextStream_NilClient_ReturnsError(t *testing.T) {
	s := newNilClientService()
	_, err := s.GenerateTextStream(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error from nil client")
	}
	if !errors.Is(err, ErrClientNotInitialized) {
		t.Errorf("expected ErrClientNotInitialized, got %v", err)
	}
}

func TestGenerateTextStream_NilClient_ReturnsNilChannel(t *testing.T) {
	s := newNilClientService()
	ch, _ := s.GenerateTextStream(context.Background(), "test prompt")
	if ch != nil {
		t.Error("expected nil channel on error")
	}
}

func TestIsAvailable_ReturnsBool(_ *testing.T) {
	s := newNilClientService()
	var _ bool = s.IsAvailable()
}

func TestErrClientNotInitialized_IsNotNil(t *testing.T) {
	if ErrClientNotInitialized == nil {
		t.Error("ErrClientNotInitialized sentinel should not be nil")
	}
}

func TestErrClientNotInitialized_HasMessage(t *testing.T) {
	if ErrClientNotInitialized.Error() == "" {
		t.Error("ErrClientNotInitialized should have a non-empty message")
	}
}

func TestGetProviderDetails_NilClient_ContextDoesNotPanic(_ *testing.T) {
	s := newNilClientService()
	s.GetProviderDetails(context.Background())
}

func TestGetProviderDetails_NilClient_ProvidersCanBeIterated(_ *testing.T) {
	s := newNilClientService()
	got := s.GetProviderDetails(context.Background())
	for k, v := range got.Providers {
		_ = k
		_ = v
	}
}

func TestGetProviderStatus_NilClient_CanBeIteratedWithoutPanic(_ *testing.T) {
	s := newNilClientService()
	m := s.GetProviderStatus()
	for k, v := range m {
		_ = k
		_ = v
	}
}

func TestGetPrimaryProvider_NilClient_EmptyStringIsNotNil(_ *testing.T) {
	s := newNilClientService()
	result := s.GetPrimaryProvider()
	_ = len(result) // should not panic
}

func TestGenerateTextStream_EmptyPrompt_NilClient(t *testing.T) {
	s := newNilClientService()
	_, err := s.GenerateTextStream(context.Background(), "")
	if !errors.Is(err, ErrClientNotInitialized) {
		t.Errorf("expected ErrClientNotInitialized for empty prompt with nil client, got %v", err)
	}
}

func TestGetProviderDetails_NilClient_AvailableIsFalse(t *testing.T) {
	s := newNilClientService()
	details := s.GetProviderDetails(context.Background())
	if details.Available != false {
		t.Error("Available must be false when aiClient is nil")
	}
}
