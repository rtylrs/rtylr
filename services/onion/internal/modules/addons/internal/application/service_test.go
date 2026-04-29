package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/addons/internal/domain"
)

type mockRepo struct {
	all     []domain.Addon
	enabled []domain.BusinessAddon
	allErr  error
	bizErr  error
	enErr   error
	disErr  error
}

func (m *mockRepo) ListAll(_ context.Context) ([]domain.Addon, error) {
	return m.all, m.allErr
}
func (m *mockRepo) ListForBusiness(_ context.Context, _ int64) ([]domain.BusinessAddon, error) {
	return m.enabled, m.bizErr
}
func (m *mockRepo) Enable(_ context.Context, _ int64, _ string) error {
	return m.enErr
}
func (m *mockRepo) Disable(_ context.Context, _ int64, _ string) error {
	return m.disErr
}

func TestListAll_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{all: []domain.Addon{{ID: "x"}}}}
	got, _ := s.ListAll(context.Background())
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestListForBusiness_MergesCatalogWithEnabled(t *testing.T) {
	all := []domain.Addon{
		{ID: "a", Name: "Addon A", Price: 10, IsFree: false},
		{ID: "b", Name: "Addon B", Price: 0, IsFree: true},
	}
	enabled := []domain.BusinessAddon{
		{AddonID: "a", Name: "Addon A", IsEnabled: true},
	}
	s := &Service{repo: &mockRepo{all: all, enabled: enabled}}

	got, err := s.ListForBusiness(context.Background(), 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	// Find a and b
	var aGot, bGot *domain.BusinessAddon
	for i := range got {
		if got[i].AddonID == "a" {
			aGot = &got[i]
		}
		if got[i].AddonID == "b" {
			bGot = &got[i]
		}
	}
	if aGot == nil || !aGot.IsEnabled {
		t.Errorf("a should be enabled, got %+v", aGot)
	}
	if bGot == nil || bGot.IsEnabled {
		t.Errorf("b should be present and disabled, got %+v", bGot)
	}
}

func TestListForBusiness_PropagatesErrors(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{allErr: want}}
	if _, err := s.ListForBusiness(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel from ListAll, got %v", err)
	}

	s = &Service{repo: &mockRepo{bizErr: want}}
	if _, err := s.ListForBusiness(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel from ListForBusiness, got %v", err)
	}
}

func TestEnableDisable_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Enable(context.Background(), 1, "x"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.Disable(context.Background(), 1, "x"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestIsEnabled(t *testing.T) {
	s := &Service{repo: &mockRepo{enabled: []domain.BusinessAddon{
		{AddonID: "y", IsEnabled: true},
		{AddonID: "z", IsEnabled: false},
	}}}
	if !s.IsEnabled(context.Background(), 1, "y") {
		t.Error("expected y enabled")
	}
	if s.IsEnabled(context.Background(), 1, "z") {
		t.Error("expected z disabled")
	}
	if s.IsEnabled(context.Background(), 1, "missing") {
		t.Error("expected missing not enabled")
	}
}

func TestIsEnabled_RepoError(t *testing.T) {
	s := &Service{repo: &mockRepo{bizErr: errors.New("db")}}
	if s.IsEnabled(context.Background(), 1, "x") {
		t.Error("expected false on repo error")
	}
}
