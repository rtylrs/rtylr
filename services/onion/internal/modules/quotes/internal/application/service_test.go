package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
)

type mockRepo struct {
	list      []domain.Quote
	createID  int64
	convertID int64
	createErr error
	updErr    error
	delErr    error
	convErr   error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Quote, error) {
	return m.list, nil
}
func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.QuoteInput) (int64, error) {
	return m.createID, m.createErr
}
func (m *mockRepo) Update(_ context.Context, _, _ int64, _ domain.QuoteInput) error {
	return m.updErr
}
func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}
func (m *mockRepo) ConvertToInvoice(_ context.Context, _, _ int64) (int64, error) {
	return m.convertID, m.convErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Quote{{ID: 1}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreateUpdateDelete_PropagateErrors(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want, updErr: want, delErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.QuoteInput{}); !errors.Is(err, want) {
		t.Errorf("create: %v", err)
	}
	if err := s.Update(context.Background(), 1, 5, domain.QuoteInput{}); !errors.Is(err, want) {
		t.Errorf("update: %v", err)
	}
	if err := s.Delete(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("delete: %v", err)
	}
}

func TestConvertToInvoice_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{convertID: 99}}
	id, err := s.ConvertToInvoice(context.Background(), 1, 5)
	if err != nil || id != 99 {
		t.Errorf("id=%d err=%v", id, err)
	}
}
