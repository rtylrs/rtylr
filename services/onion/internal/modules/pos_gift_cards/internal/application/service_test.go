package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/domain"
)

type mockRepo struct {
	list      []domain.GiftCard
	createID  int64
	card      *domain.GiftCard
	createErr error
	getErr    error
	redErr    error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.GiftCard, error) {
	return m.list, nil
}

func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.GiftCardInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) GetByCode(_ context.Context, _ int64, _ string) (*domain.GiftCard, error) {
	return m.card, m.getErr
}

func (m *mockRepo) Redeem(_ context.Context, _ int64, _ domain.RedeemInput) (*domain.GiftCard, error) {
	return m.card, m.redErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.GiftCard{{}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.GiftCardInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetByCode_PropagatesError(t *testing.T) {
	want := errors.New("not found")
	s := &Service{repo: &mockRepo{getErr: want}}
	if _, err := s.GetByCode(context.Background(), 1, "abc"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestRedeem_PropagatesError(t *testing.T) {
	want := errors.New("invalid")
	s := &Service{repo: &mockRepo{redErr: want}}
	if _, err := s.Redeem(context.Background(), 1, domain.RedeemInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
