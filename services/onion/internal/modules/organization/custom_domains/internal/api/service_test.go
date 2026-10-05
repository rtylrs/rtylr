package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/organization/custom_domains/internal/domain"
)

func TestPayloadVerificationTokenVisibility(t *testing.T) {
	item := domain.CustomDomain{ID: 7, BusinessID: 42, Domain: "menu.example.com", VerificationToken: "challenge-token"}

	if got := payload(item, true).GetVerificationToken(); got != item.VerificationToken {
		t.Fatalf("registration payload token = %q, want %q", got, item.VerificationToken)
	}
	if got := payload(item, false).GetVerificationToken(); got != "" {
		t.Fatalf("non-registration payload token = %q, want it redacted", got)
	}
}
