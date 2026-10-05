package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/campaigns/internal/domain"
)

type mockRepo struct {
	list      []domain.Campaign
	get       domain.Campaign
	createID  int64
	createErr error
	updErr    error
	updStErr  error
	delErr    error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Campaign, error) {
	return m.list, nil
}

func (m *mockRepo) ListPaginated(_ context.Context, _ int64, _, _ int32) ([]domain.Campaign, error) {
	return m.list, nil
}

func (m *mockRepo) Count(_ context.Context, _ int64) (int64, error) {
	return int64(len(m.list)), nil
}

func (m *mockRepo) GetByID(_ context.Context, _, _ int64) (domain.Campaign, error) {
	return m.get, nil
}

func (m *mockRepo) Create(_ context.Context, _ domain.CampaignInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ domain.CampaignUpdateInput) error {
	return m.updErr
}

func (m *mockRepo) UpdateStatus(_ context.Context, _, _ int64, _ string) error {
	return m.updStErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Campaign{{ID: 1}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestGetByID_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{get: domain.Campaign{ID: 7}}}
	got, _ := s.GetByID(context.Background(), 1, 7)
	if got.ID != 7 {
		t.Errorf("got %+v", got)
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	input := domain.CampaignInput{Name: "Promo", Type: domain.TypeEmail}
	if _, err := s.Create(context.Background(), input); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateAndStatus_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	upd := domain.CampaignUpdateInput{ID: 5, OwnerID: 1, Name: "Promo"}
	if err := s.Update(context.Background(), upd); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.UpdateStatus(context.Background(), 1, 5, "active"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
