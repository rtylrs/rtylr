package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/api_keys/internal/domain"
)

func TestCreate_RequiresName(t *testing.T) {
	s := &Service{}
	if _, _, err := s.Create(context.Background(), 1, domain.CreateAPIKeyInput{}); err == nil {
		t.Fatalf("expected error for empty name")
	}
}

func TestRevoke_RequiresKeyID(t *testing.T) {
	s := &Service{}
	if err := s.Revoke(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error for missing key id")
	}
}
