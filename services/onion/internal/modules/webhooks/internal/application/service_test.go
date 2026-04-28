package application

import (
	"context"
	"testing"
)

func TestGenerateSecret_HexAndUnique(t *testing.T) {
	a, err := generateSecret()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	b, _ := generateSecret()
	if a == b {
		t.Fatal("two secrets should differ")
	}
	if len(a) != 64 {
		t.Errorf("expected 64-char hex secret, got %d", len(a))
	}
}

func TestCreate_RequiresURL(t *testing.T) {
	s := &Service{}
	if _, err := s.Create(context.Background(), 1, "", nil); err == nil {
		t.Fatalf("expected error for empty url")
	}
}

func TestDelete_RequiresWebhookID(t *testing.T) {
	s := &Service{}
	if err := s.Delete(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error for missing webhook id")
	}
}
