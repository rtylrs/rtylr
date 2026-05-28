package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm_settings/internal/domain"
)

func TestGet_ReturnsFromName(t *testing.T) {
	s := &Service{repo: &mockRepo{settings: domain.Settings{OwnerID: 1, FromName: "My Shop"}}}
	got, err := s.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.FromName != "My Shop" {
		t.Errorf("FromName = %q, want %q", got.FromName, "My Shop")
	}
}

func TestGet_ReturnsSmsSenderID(t *testing.T) {
	s := &Service{repo: &mockRepo{settings: domain.Settings{OwnerID: 1, SmsSenderID: "MYSHOP"}}}
	got, err := s.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SmsSenderID != "MYSHOP" {
		t.Errorf("SmsSenderID = %q, want %q", got.SmsSenderID, "MYSHOP")
	}
}

func TestGet_ReturnsOwnerID(t *testing.T) {
	s := &Service{repo: &mockRepo{settings: domain.Settings{OwnerID: 99}}}
	got, err := s.Get(context.Background(), 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.OwnerID != 99 {
		t.Errorf("OwnerID = %d, want 99", got.OwnerID)
	}
}

func TestGet_DefaultConsentFalse(t *testing.T) {
	s := &Service{repo: &mockRepo{settings: domain.Settings{OwnerID: 1, DefaultConsent: false}}}
	got, err := s.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.DefaultConsent {
		t.Error("expected DefaultConsent=false")
	}
}
