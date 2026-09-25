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

func TestGet_ReturnsSettings(t *testing.T) {
	expected := &domain.Settings{
		BusinessID:        5,
		AutoAssign:        true,
		SLAHours:          24,
		NotificationEmail: "support@example.com",
	}
	s := &Service{repo: &mockRepo{settings: expected}}
	got, err := s.Get(context.Background(), 5)
	if err != nil || got.AutoAssign != true || got.SLAHours != 24 {
		t.Errorf("fields not preserved: %+v err=%v", got, err)
	}
}

func TestGet_PropagatesError(t *testing.T) {
	want := errors.New("db fail")
	s := &Service{repo: &mockRepo{getErr: want}}
	if _, err := s.Get(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGet_ReturnsNilWhenNotFound(t *testing.T) {
	want := errors.New("not found")
	s := &Service{repo: &mockRepo{getErr: want}}
	got, err := s.Get(context.Background(), 99)
	if got != nil || !errors.Is(err, want) {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestUpsert_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{upErr: want}}
	if err := s.Upsert(context.Background(), domain.SettingsInput{BusinessID: 1, SLAHours: 24}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpsert_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.Upsert(context.Background(), domain.SettingsInput{
		BusinessID: 1,
		AutoAssign: true,
		SLAHours:   8,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpsert_RejectsZeroSLAHours(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.Upsert(context.Background(), domain.SettingsInput{BusinessID: 1})
	if !errors.Is(err, domain.ErrInvalidSLAHours) {
		t.Errorf("expected ErrInvalidSLAHours for zero SLA hours, got %v", err)
	}
}

func TestUpsert_RejectsNegativeSLAHours(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.Upsert(context.Background(), domain.SettingsInput{BusinessID: 1, SLAHours: -5})
	if !errors.Is(err, domain.ErrInvalidSLAHours) {
		t.Errorf("expected ErrInvalidSLAHours for negative SLA hours, got %v", err)
	}
}

func TestUpsert_RejectsExcessiveSLAHours(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.Upsert(context.Background(), domain.SettingsInput{BusinessID: 1, SLAHours: 1001})
	if !errors.Is(err, domain.ErrInvalidSLAHours) {
		t.Errorf("expected ErrInvalidSLAHours for excessive SLA hours, got %v", err)
	}
}

func TestGet_BusinessIDRoundTrip(t *testing.T) {
	expected := &domain.Settings{BusinessID: 42, SLAHours: 48}
	s := &Service{repo: &mockRepo{settings: expected}}
	got, err := s.Get(context.Background(), 42)
	if err != nil || got.BusinessID != 42 || got.SLAHours != 48 {
		t.Errorf("round trip failed: %+v err=%v", got, err)
	}
}

func TestUpsert_ZeroBusinessID(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Upsert(context.Background(), domain.SettingsInput{BusinessID: 0, SLAHours: 24}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpsert_WithNotificationEmail(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.Upsert(context.Background(), domain.SettingsInput{
		BusinessID:        2,
		SLAHours:          24,
		NotificationEmail: "alerts@co.com",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGet_ReturnsSettingsWithBusinessHours(t *testing.T) {
	expected := &domain.Settings{BusinessHours: "9-5 weekdays"}
	s := &Service{repo: &mockRepo{settings: expected}}
	got, err := s.Get(context.Background(), 1)
	if err != nil || got.BusinessHours != "9-5 weekdays" {
		t.Errorf("business hours not preserved: %+v err=%v", got, err)
	}
}
