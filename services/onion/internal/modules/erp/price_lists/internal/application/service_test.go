package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/price_lists/internal/domain"
)

type mockRepo struct {
	list      []domain.PriceList
	createID  int64
	createGot domain.PriceListInput
}

func (m *mockRepo) ListPriceLists(_ context.Context, _ int64) ([]domain.PriceList, error) {
	return m.list, nil
}

func (m *mockRepo) CreatePriceList(_ context.Context, _ int64, input domain.PriceListInput) (int64, error) {
	m.createGot = input
	return m.createID, nil
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.PriceList{{}}}}
	got, _ := s.ListPriceLists(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreatePriceList_NormalizesNameAndCurrency(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}

	if _, err := s.CreatePriceList(context.Background(), 1, domain.PriceListInput{
		Name:     "  My List  ",
		Currency: " usd ",
	}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if m.createGot.Name != "My List" {
		t.Errorf("name = %q, want %q", m.createGot.Name, "My List")
	}
	if m.createGot.Currency != "USD" {
		t.Errorf("currency = %q, want USD", m.createGot.Currency)
	}
}

func TestCreatePriceList_DefaultsCurrency(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if _, err := s.CreatePriceList(context.Background(), 1, domain.PriceListInput{Name: "x"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if m.createGot.Currency != "USD" {
		t.Errorf("expected default USD, got %q", m.createGot.Currency)
	}
}
