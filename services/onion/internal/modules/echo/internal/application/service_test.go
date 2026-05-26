package application

import (
	"context"
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
	if err != ErrClientNotInitialized {
		t.Errorf("expected ErrClientNotInitialized, got %v", err)
	}
}
