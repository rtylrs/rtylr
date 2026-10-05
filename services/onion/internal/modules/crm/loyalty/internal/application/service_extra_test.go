package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/loyalty/internal/domain"
)

func TestListTransactions_ReturnsTotal(t *testing.T) {
	items := []domain.LoyaltyTransaction{{ID: 1}, {ID: 2}}
	s := &Service{repo: &mockRepo{transactions: items, totalTx: 2}}
	_, total, err := s.ListTransactions(context.Background(), 1, 0, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}
}

func TestAwardPoints_ZeroPoints_NoError(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.AwardPoints(context.Background(), domain.AwardInput{OwnerID: 1, CustomerID: 1, Points: 0})
	if err != nil {
		t.Errorf("unexpected error for zero points: %v", err)
	}
}

func TestRedeemPoints_ZeroPoints_NoError(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.RedeemPoints(context.Background(), domain.RedeemInput{OwnerID: 1, CustomerID: 1, Points: 0})
	if err != nil {
		t.Errorf("unexpected error for zero points: %v", err)
	}
}

func TestGetConfig_ReturnsIsActive(t *testing.T) {
	cfg := domain.LoyaltyConfig{ID: 1, IsActive: true}
	s := &Service{repo: &mockRepo{config: cfg}}
	got, err := s.GetConfig(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.IsActive {
		t.Error("expected IsActive=true")
	}
}

func TestGetConfig_ReturnsOwnerID(t *testing.T) {
	cfg := domain.LoyaltyConfig{ID: 1, OwnerID: 42}
	s := &Service{repo: &mockRepo{config: cfg}}
	got, err := s.GetConfig(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.OwnerID != 42 {
		t.Errorf("expected OwnerID=42, got %d", got.OwnerID)
	}
}

func TestListMembers_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{members: []domain.LoyaltyMember{}}}
	got, err := s.ListMembers(context.Background(), 1, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestListRewards_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{rewards: []domain.LoyaltyReward{}}}
	got, err := s.ListRewards(context.Background(), 1, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestUpsertReward_ReturnsZeroOnError(t *testing.T) {
	want := errors.New("upsert fail")
	s := &Service{repo: &mockRepo{err: want}}
	id, err := s.UpsertReward(context.Background(), domain.LoyaltyReward{OwnerID: 1})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
	if id != 0 {
		t.Errorf("expected id=0 on error, got %d", id)
	}
}

func TestUpdateConfig_ZeroOwner_Rejected(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.UpdateConfig(context.Background(), domain.LoyaltyConfigInput{OwnerID: 0})
	if err == nil {
		t.Error("expected validation error for zero owner_id")
	}
}
