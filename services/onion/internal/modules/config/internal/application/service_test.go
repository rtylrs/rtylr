package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
)

type mockRepo struct {
	config    domain.Config
	theme     domain.Theme
	getErr    error
	createErr error
	updateErr error
	themeErr  error
}

func (m *mockRepo) GetConfig(_ context.Context, _ string) (domain.Config, error) {
	return m.config, m.getErr
}

func (m *mockRepo) CreateConfig(_ context.Context, _ domain.ConfigInput) error {
	return m.createErr
}

func (m *mockRepo) UpdateConfig(_ context.Context, _ domain.ConfigInput) error {
	return m.updateErr
}

func (m *mockRepo) GetThemeByOwnerID(_ context.Context, _ string) (domain.Theme, error) {
	return m.theme, m.themeErr
}

func TestCreateConfig_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if err := s.CreateConfig(context.Background(), domain.ConfigInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetThemeByOwnerID_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{themeErr: want}}
	if _, err := s.GetThemeByOwnerID(context.Background(), "1"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
