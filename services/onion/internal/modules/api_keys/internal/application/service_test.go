package application

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/api_keys/internal/domain"
)

func TestGenerateAPIKey_ProducesUniqueHex(t *testing.T) {
	a, err := generateAPIKey()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	b, err := generateAPIKey()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if a == b {
		t.Fatal("expected two generated keys to differ")
	}
	if len(a) != 64 {
		t.Errorf("expected 64-char hex (32 bytes), got %d", len(a))
	}
	for _, c := range a {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("non-hex character %q in key %q", c, a)
			break
		}
	}
}

func TestHashKey_DeterministicAndDifferentForDifferentInput(t *testing.T) {
	if hashKey("abc") != hashKey("abc") {
		t.Fatal("hashKey should be deterministic for the same input")
	}
	if hashKey("abc") == hashKey("abd") {
		t.Fatal("hashKey should differ for different input")
	}
	// SHA-256 hex is exactly 64 chars.
	if len(hashKey("abc")) != 64 {
		t.Errorf("expected 64-char hex hash, got %d", len(hashKey("abc")))
	}
}

func TestCreateAPIKeyInputFromRequest(t *testing.T) {
	input := CreateAPIKeyInputFromRequest("my key", []string{"read", "write"})
	if input.Name != "my key" || strings.Join(input.Scopes, ",") != "read,write" {
		t.Errorf("unexpected input: %+v", input)
	}
}

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
