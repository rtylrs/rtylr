package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_settings/internal/domain"
)

type mockRepo struct {
	settings  domain.Settings
	getErr    error
	upsertErr error
}

func (m *mockRepo) Get(_ context.Context, _ int64) (domain.Settings, error) {
	return m.settings, m.getErr
}

func (m *mockRepo) Upsert(_ context.Context, _ domain.SettingsInput) error {
	return m.upsertErr
}

func TestGet_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{settings: domain.Settings{OwnerID: 7}}}
	got, _ := s.Get(context.Background(), 7)
	if got.OwnerID != 7 {
		t.Errorf("got %+v", got)
	}
}

func TestUpsert_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{upsertErr: want}}
	if err := s.Upsert(context.Background(), domain.SettingsInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
