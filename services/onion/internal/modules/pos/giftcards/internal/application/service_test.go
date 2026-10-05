package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/giftcards/internal/domain"
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

func (m *mockRepo) CreateForOwner(_ context.Context, _, _ int64, _ domain.GiftCardInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) GetByCode(_ context.Context, _ int64, _ string) (*domain.GiftCard, error) {
	return m.card, m.getErr
}

func (m *mockRepo) RedeemForOrder(_ context.Context, _, _ int64, _ domain.RedeemInput) (*domain.GiftCard, error) {
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
	input := domain.GiftCardInput{PaidOrderID: 1, Code: "GC-1", InitialAmount: 50}
	if _, err := s.CreateForOwner(context.Background(), 1, 1, input); !errors.Is(err, want) {
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
	input := domain.RedeemInput{OrderID: 1, Code: "GC-1", Amount: 10}
	if _, err := s.RedeemForOrder(context.Background(), 1, 1, input); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreate_RejectsBlankCode(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.CreateForOwner(context.Background(), 1, 1, domain.GiftCardInput{PaidOrderID: 1, InitialAmount: 50})
	if err == nil {
		t.Fatal("expected error for blank code")
	}
}

func TestCreate_RejectsZeroAmount(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.CreateForOwner(context.Background(), 1, 1, domain.GiftCardInput{PaidOrderID: 1, Code: "GC1", InitialAmount: 0})
	if err == nil {
		t.Fatal("expected error for zero initial amount")
	}
}

func TestRedeem_RejectsBlankCode(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.RedeemForOrder(context.Background(), 1, 1, domain.RedeemInput{OrderID: 1, Amount: 10})
	if err == nil {
		t.Fatal("expected error for blank code")
	}
}

func TestRedeem_RejectsZeroAmount(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.RedeemForOrder(context.Background(), 1, 1, domain.RedeemInput{OrderID: 1, Code: "GC1", Amount: 0})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}
