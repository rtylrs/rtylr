package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
)

func TestListTransactions_MultipleItems(t *testing.T) {
	items := []domain.LoyaltyTransaction{
		{ID: 1, Points: 100, Type: "award"},
		{ID: 2, Points: 50, Type: "redeem"},
		{ID: 3, Points: 200, Type: "award"},
	}
	s := &Service{repo: &mockRepo{transactions: items, totalTx: 3}}
	got, total, err := s.ListTransactions(context.Background(), 1, 0, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 || total != 3 {
		t.Errorf("expected 3 items (total=3), got len=%d total=%d", len(got), total)
	}
}

func TestGetConfig_ReturnsPointsPerDollar(t *testing.T) {
	cfg := domain.LoyaltyConfig{OwnerID: 1, PointsPerDollar: 2.5}
	s := &Service{repo: &mockRepo{config: cfg}}
	got, err := s.GetConfig(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PointsPerDollar != 2.5 {
		t.Errorf("PointsPerDollar = %f, want 2.5", got.PointsPerDollar)
	}
}

func TestListMembers_SingleMember(t *testing.T) {
	m := []domain.LoyaltyMember{{ID: 7, Name: "Charlie", Points: 350, Tier: "silver"}}
	s := &Service{repo: &mockRepo{members: m}}
	got, err := s.ListMembers(context.Background(), 1, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Charlie" {
		t.Errorf("expected Charlie, got %+v", got)
	}
}

func TestListRewards_MultipleRewards(t *testing.T) {
	rewards := []domain.LoyaltyReward{
		{ID: 1, Name: "Coffee"},
		{ID: 2, Name: "Tea"},
		{ID: 3, Name: "Juice"},
	}
	s := &Service{repo: &mockRepo{rewards: rewards}}
	got, err := s.ListRewards(context.Background(), 1, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 rewards, got %d", len(got))
	}
}

func TestAwardPoints_ReturnsNilOnSuccess(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.AwardPoints(context.Background(), domain.AwardInput{OwnerID: 1, CustomerID: 5, Points: 100}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestUpdateConfig_WithActiveFlag(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.UpdateConfig(context.Background(), domain.LoyaltyConfigInput{OwnerID: 1, IsActive: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpsertReward_WithStock(t *testing.T) {
	s := &Service{repo: &mockRepo{upsertID: 10}}
	stock := 5
	id, err := s.UpsertReward(context.Background(), domain.LoyaltyReward{OwnerID: 1, Name: "Limited", Stock: &stock})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 10 {
		t.Errorf("expected id=10, got %d", id)
	}
}

func TestListTransactions_ZeroCustomerID_ReturnsAll(t *testing.T) {
	items := []domain.LoyaltyTransaction{{ID: 1, CustomerID: 5}, {ID: 2, CustomerID: 6}}
	s := &Service{repo: &mockRepo{transactions: items, totalTx: 2}}
	_, total, err := s.ListTransactions(context.Background(), 1, 0, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}
}

func TestGetConfig_NotFoundReturnsError(t *testing.T) {
	want := domain.ErrLoyaltyConfigNotFound
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GetConfig(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected ErrLoyaltyConfigNotFound, got %v", err)
	}
}
