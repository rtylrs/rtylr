package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
)

type mockRepo struct {
	transactions []domain.LoyaltyTransaction
	totalTx      int
	config       domain.LoyaltyConfig
	balance      int64
	members      []domain.LoyaltyMember
	rewards      []domain.LoyaltyReward
	upsertID     int64
	err          error
}

func (m *mockRepo) ListTransactions(_ context.Context, _, _ int64, _, _ int) ([]domain.LoyaltyTransaction, int, error) {
	return m.transactions, m.totalTx, m.err
}

func (m *mockRepo) CreateTransaction(_ context.Context, _ domain.AwardInput, _ string) error {
	return m.err
}

func (m *mockRepo) CreateRedeemTransaction(_ context.Context, _ domain.RedeemInput) error {
	return m.err
}

func (m *mockRepo) GetCustomerPointsBalance(_ context.Context, _, _ int64) (int64, error) {
	return m.balance, m.err
}

func (m *mockRepo) GetConfig(_ context.Context, _ int64) (domain.LoyaltyConfig, error) {
	return m.config, m.err
}

func (m *mockRepo) UpsertConfig(_ context.Context, _ domain.LoyaltyConfigInput) error {
	return m.err
}

func (m *mockRepo) ListMembers(_ context.Context, _ int64, _, _ int) ([]domain.LoyaltyMember, error) {
	return m.members, m.err
}

func (m *mockRepo) ListRewards(_ context.Context, _ int64, _ bool) ([]domain.LoyaltyReward, error) {
	return m.rewards, m.err
}

func (m *mockRepo) UpsertReward(_ context.Context, _ domain.LoyaltyReward) (int64, error) {
	return m.upsertID, m.err
}

func TestListTransactions_ReturnsDelegatedItems(t *testing.T) {
	items := []domain.LoyaltyTransaction{
		{ID: 1, Points: 100, Type: "award"},
	}
	s := &Service{repo: &mockRepo{transactions: items, totalTx: 1}}
	got, total, err := s.ListTransactions(context.Background(), 1, 1, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || total != 1 {
		t.Errorf("expected 1 item, got %d (total=%d)", len(got), total)
	}
}

func TestListTransactions_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	got, total, err := s.ListTransactions(context.Background(), 1, 1, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 || total != 0 {
		t.Errorf("expected empty, got len=%d total=%d", len(got), total)
	}
}

func TestListTransactions_PropagatesRepoError(t *testing.T) {
	want := errors.New("db error")
	s := &Service{repo: &mockRepo{err: want}}
	_, _, err := s.ListTransactions(context.Background(), 1, 1, 1, 10)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestAwardPoints_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.AwardPoints(context.Background(), domain.AwardInput{OwnerID: 1, CustomerID: 1, Points: 50})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAwardPoints_PropagatesRepoError(t *testing.T) {
	want := errors.New("award error")
	s := &Service{repo: &mockRepo{err: want}}
	err := s.AwardPoints(context.Background(), domain.AwardInput{OwnerID: 1, CustomerID: 1, Points: 50})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestRedeemPoints_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{balance: 100}}
	err := s.RedeemPoints(context.Background(), domain.RedeemInput{OwnerID: 1, CustomerID: 1, Points: 20})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRedeemPoints_RejectsInsufficientBalance(t *testing.T) {
	s := &Service{repo: &mockRepo{balance: 10}}
	err := s.RedeemPoints(context.Background(), domain.RedeemInput{OwnerID: 1, CustomerID: 1, Points: 20})
	if !errors.Is(err, domain.ErrInsufficientPoints) {
		t.Errorf("expected ErrInsufficientPoints, got %v", err)
	}
}

func TestRedeemPoints_RejectsBelowMinRedemption(t *testing.T) {
	s := &Service{repo: &mockRepo{
		config:  domain.LoyaltyConfig{MinRedemptionPoints: 50},
		balance: 100,
	}}
	err := s.RedeemPoints(context.Background(), domain.RedeemInput{OwnerID: 1, CustomerID: 1, Points: 20})
	if !errors.Is(err, domain.ErrInvalidLoyaltyConfig) {
		t.Errorf("expected ErrInvalidLoyaltyConfig, got %v", err)
	}
}

func TestRedeemPoints_PropagatesRepoError(t *testing.T) {
	want := errors.New("redeem error")
	s := &Service{repo: &mockRepo{err: want}}
	err := s.RedeemPoints(context.Background(), domain.RedeemInput{OwnerID: 1, CustomerID: 1, Points: 20})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetConfig_ReturnsDelegatedConfig(t *testing.T) {
	cfg := domain.LoyaltyConfig{ID: 1, PointsPerDollar: 1.5, IsActive: true}
	s := &Service{repo: &mockRepo{config: cfg}}
	got, err := s.GetConfig(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PointsPerDollar != 1.5 {
		t.Errorf("expected 1.5, got %f", got.PointsPerDollar)
	}
}

func TestGetConfig_PropagatesRepoError(t *testing.T) {
	want := domain.ErrLoyaltyConfigNotFound
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GetConfig(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateConfig_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.UpdateConfig(context.Background(), domain.LoyaltyConfigInput{OwnerID: 1, PointsPerDollar: 2})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateConfig_PropagatesRepoError(t *testing.T) {
	want := errors.New("upsert error")
	s := &Service{repo: &mockRepo{err: want}}
	err := s.UpdateConfig(context.Background(), domain.LoyaltyConfigInput{OwnerID: 1})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestListMembers_ReturnsDelegatedItems(t *testing.T) {
	members := []domain.LoyaltyMember{{ID: 1, Name: "Alice", Points: 200}}
	s := &Service{repo: &mockRepo{members: members}}
	got, err := s.ListMembers(context.Background(), 1, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Alice" {
		t.Errorf("expected Alice, got %+v", got)
	}
}

func TestListMembers_PropagatesRepoError(t *testing.T) {
	want := errors.New("members error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.ListMembers(context.Background(), 1, 10, 0)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestListRewards_ReturnsDelegatedItems(t *testing.T) {
	rewards := []domain.LoyaltyReward{{ID: 1, Name: "Free Coffee", PointsCost: 100}}
	s := &Service{repo: &mockRepo{rewards: rewards}}
	got, err := s.ListRewards(context.Background(), 1, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Free Coffee" {
		t.Errorf("expected Free Coffee reward, got %+v", got)
	}
}

func TestListRewards_PropagatesRepoError(t *testing.T) {
	want := errors.New("rewards error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.ListRewards(context.Background(), 1, false)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpsertReward_ReturnsGeneratedID(t *testing.T) {
	s := &Service{repo: &mockRepo{upsertID: 42}}
	id, err := s.UpsertReward(context.Background(), domain.LoyaltyReward{OwnerID: 1, Name: "Discount"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected 42, got %d", id)
	}
}

func TestUpsertReward_PropagatesRepoError(t *testing.T) {
	want := errors.New("upsert reward error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.UpsertReward(context.Background(), domain.LoyaltyReward{OwnerID: 1})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
