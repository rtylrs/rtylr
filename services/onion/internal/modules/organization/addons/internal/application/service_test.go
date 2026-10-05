package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/addons/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

type mockRepo struct {
	all      []domain.Addon
	enabled  []domain.BusinessAddon
	allErr   error
	bizErr   error
	enErr    error
	disErr   error
	grant    bool
	grantErr error
}

func (m *mockRepo) ListAll(_ context.Context) ([]domain.Addon, error) {
	return m.all, m.allErr
}

func (m *mockRepo) ListForBusiness(_ context.Context, _ int64) ([]domain.BusinessAddon, error) {
	return m.enabled, m.bizErr
}

func (m *mockRepo) HasActiveGrant(_ context.Context, _ int64, _ string) (bool, error) {
	return m.grant, m.grantErr
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

func TestListAll_HidesLegacyAIChatAddon(t *testing.T) {
	s := &Service{repo: &mockRepo{all: []domain.Addon{
		{ID: "menu", Name: "Digital Menu"},
		{ID: "ai_chat", Name: "AI Chat"},
	}}}
	got, err := s.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 1 || got[0].ID != "menu" {
		t.Fatalf("expected only visible addons, got %+v", got)
	}
}

func TestListForBusiness_MergesCatalogWithEnabled(t *testing.T) {
	all := []domain.Addon{
		{ID: "a", Name: "Addon A", Price: money.Amount(1000), IsFree: false},
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

func TestListForBusiness_HidesLegacyAIChatAddon(t *testing.T) {
	s := &Service{repo: &mockRepo{
		all: []domain.Addon{
			{ID: "menu", Name: "Digital Menu"},
			{ID: "ai_chat", Name: "AI Chat"},
		},
		enabled: []domain.BusinessAddon{
			{AddonID: "ai_chat", Name: "AI Chat", IsEnabled: true},
		},
	}}
	got, err := s.ListForBusiness(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 1 || got[0].AddonID != "menu" {
		t.Fatalf("expected hidden ai_chat to be filtered, got %+v", got)
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
	s := &Service{repo: &mockRepo{all: []domain.Addon{{ID: "x", IsFree: true}}}}
	if err := s.Enable(context.Background(), 1, "x"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.Disable(context.Background(), 1, "x"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestIsEnabled(t *testing.T) {
	s := &Service{repo: &mockRepo{
		all: []domain.Addon{
			{ID: "y", Name: "Addon Y"},
			{ID: "z", Name: "Addon Z"},
		},
		enabled: []domain.BusinessAddon{
			{AddonID: "y", IsEnabled: true},
			{AddonID: "z", IsEnabled: false},
		},
	}}
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

func TestListAll_Empty(t *testing.T) {
	s := &Service{repo: &mockRepo{all: []domain.Addon{}}}
	got, err := s.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestListAll_PropagatesError(t *testing.T) {
	want := errors.New("db down")
	s := &Service{repo: &mockRepo{allErr: want}}
	if _, err := s.ListAll(context.Background()); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestEnable_PropagatesError(t *testing.T) {
	want := errors.New("enable fail")
	s := &Service{repo: &mockRepo{all: []domain.Addon{{ID: "x", IsFree: true}}, enErr: want}}
	if err := s.Enable(context.Background(), 1, "x"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestEnable_PaidAddonRequiresActiveGrant(t *testing.T) {
	s := &Service{repo: &mockRepo{all: []domain.Addon{{ID: "paid", Price: money.Amount(1200)}}}}
	if err := s.Enable(context.Background(), 1, "paid"); !errors.Is(err, domain.ErrAddonNotEntitled) {
		t.Fatalf("expected ErrAddonNotEntitled, got %v", err)
	}
}

func TestEnable_PaidAddonWithActiveGrantDelegates(t *testing.T) {
	s := &Service{repo: &mockRepo{
		all:   []domain.Addon{{ID: "paid", Price: money.Amount(1200)}},
		grant: true,
	}}
	if err := s.Enable(context.Background(), 1, "paid"); err != nil {
		t.Fatalf("expected granted addon to enable, got %v", err)
	}
}

func TestDisable_PropagatesError(t *testing.T) {
	want := errors.New("disable fail")
	s := &Service{repo: &mockRepo{disErr: want}}
	if err := s.Disable(context.Background(), 1, "x"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestListForBusiness_Empty(t *testing.T) {
	s := &Service{repo: &mockRepo{all: []domain.Addon{}, enabled: []domain.BusinessAddon{}}}
	got, err := s.ListForBusiness(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d", len(got))
	}
}

func TestListForBusiness_AllDisabled(t *testing.T) {
	all := []domain.Addon{
		{ID: "a", Name: "Addon A"},
		{ID: "b", Name: "Addon B"},
	}
	s := &Service{repo: &mockRepo{all: all, enabled: []domain.BusinessAddon{}}}
	got, err := s.ListForBusiness(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	for _, item := range got {
		if item.IsEnabled {
			t.Errorf("expected all disabled, but %s is enabled", item.AddonID)
		}
	}
}

func TestIsEnabled_NotFound(t *testing.T) {
	s := &Service{repo: &mockRepo{
		all:     []domain.Addon{{ID: "x", Name: "X"}},
		enabled: []domain.BusinessAddon{{AddonID: "x", IsEnabled: true}},
	}}
	if s.IsEnabled(context.Background(), 1, "unknown") {
		t.Error("expected false for unknown addon")
	}
}

func TestListForBusiness_BizErrPropagates(t *testing.T) {
	want := errors.New("biz err")
	s := &Service{repo: &mockRepo{
		all:    []domain.Addon{{ID: "a"}},
		bizErr: want,
	}}
	if _, err := s.ListForBusiness(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel from ListForBusiness repo, got %v", err)
	}
}
