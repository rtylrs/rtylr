package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/business/internal/domain"
)

type mockRepo struct {
	profile domain.BusinessProfile
	err     error
	got     int64
}

func (m *mockRepo) GetProfile(_ context.Context, ownerID int64) (domain.BusinessProfile, error) {
	m.got = ownerID
	return m.profile, m.err
}

func TestGetProfile_Delegates(t *testing.T) {
	want := domain.BusinessProfile{SiteName: "Acme", Currency: "USD"}
	m := &mockRepo{profile: want}
	s := &Service{repo: m}

	got, err := s.GetProfile(context.Background(), 7)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SiteName != "Acme" || got.Currency != "USD" {
		t.Errorf("got %+v", got)
	}
	if m.got != 7 {
		t.Errorf("ownerID didn't pass through, got %d", m.got)
	}
}

func TestGetProfile_PropagatesError(t *testing.T) {
	want := errors.New("missing")
	s := &Service{repo: &mockRepo{err: want}}
	if _, err := s.GetProfile(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
