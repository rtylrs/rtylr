package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/support_settings/internal/domain"
)

type mockRepo struct {
	settings *domain.Settings
	getErr   error
	upErr    error
}

func (m *mockRepo) Get(_ context.Context, _ int64) (*domain.Settings, error) {
	return m.settings, m.getErr
}

func (m *mockRepo) Upsert(_ context.Context, _ domain.SettingsInput) error {
	return m.upErr
}

func TestGet_Delegates(t *testing.T) {
	want := &domain.Settings{}
	s := &Service{repo: &mockRepo{settings: want}}
	got, _ := s.Get(context.Background(), 1)
	if got != want {
		t.Errorf("got %+v", got)
	}
}

func TestUpsert_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{upErr: want}}
	if err := s.Upsert(context.Background(), domain.SettingsInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
