package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_settings/internal/domain"
)

func TestGet_ReturnsTimezone(t *testing.T) {
	s := &Service{repo: &mockRepo{settings: domain.Settings{OwnerID: 1, Timezone: "UTC"}}}
	got, err := s.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Timezone != "UTC" {
		t.Errorf("expected Timezone='UTC', got %q", got.Timezone)
	}
}

func TestGet_PropagatesError(t *testing.T) {
	want := errors.New("db error")
	s := &Service{repo: &mockRepo{getErr: want}}
	_, err := s.Get(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpsert_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.Upsert(context.Background(), domain.SettingsInput{
		OwnerID:  1,
		Timezone: "Europe/Paris",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGet_ReturnsFromEmail(t *testing.T) {
	s := &Service{repo: &mockRepo{settings: domain.Settings{OwnerID: 5, FromEmail: "no-reply@test.com"}}}
	got, err := s.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.FromEmail != "no-reply@test.com" {
		t.Errorf("FromEmail = %q, want no-reply@test.com", got.FromEmail)
	}
}

func TestGet_ReturnsDefaultConsent(t *testing.T) {
	s := &Service{repo: &mockRepo{settings: domain.Settings{OwnerID: 1, DefaultConsent: true}}}
	got, err := s.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.DefaultConsent {
		t.Error("expected DefaultConsent=true")
	}
}

func TestUpsert_ZeroOwnerID_NoError(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.Upsert(context.Background(), domain.SettingsInput{OwnerID: 0})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
