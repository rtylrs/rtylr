package application

import (
	"context"
	"testing"
)

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
