package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/quotes/internal/domain"
)

type mockRepo struct {
	list      []domain.Quote
	getResult domain.Quote
	getErr    error
	createID  int64
	convertID int64
	createErr error
	updErr    error
	statusErr error
	delErr    error
	convErr   error
}

func (m *mockRepo) List(_ context.Context, _ int64, _, _ int32) ([]domain.Quote, error) {
	return m.list, nil
}

func (m *mockRepo) Get(_ context.Context, _, _ int64) (domain.Quote, error) {
	return m.getResult, m.getErr
}

func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.QuoteInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _, _ int64, _ domain.QuoteInput) error {
	return m.updErr
}

func (m *mockRepo) UpdateStatus(_ context.Context, _, _ int64, _ string) error {
	return m.statusErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}

func (m *mockRepo) ConvertToInvoice(_ context.Context, _, _ int64) (int64, error) {
	return m.convertID, m.convErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Quote{{ID: 1}}}}
	got, _, _ := s.List(context.Background(), 1, 1, 20)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreateUpdateDelete_PropagateErrors(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want, updErr: want, statusErr: want, delErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.QuoteInput{}); !errors.Is(err, want) {
		t.Errorf("create: %v", err)
	}
	if err := s.Update(context.Background(), 1, 5, domain.QuoteInput{}); !errors.Is(err, want) {
		t.Errorf("update: %v", err)
	}
	if err := s.UpdateStatus(context.Background(), 1, 5, "accepted"); !errors.Is(err, want) {
		t.Errorf("update status: %v", err)
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

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 77}}
	id, err := s.Create(context.Background(), 1, domain.QuoteInput{CustomerName: "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 77 {
		t.Errorf("expected id=77, got %d", id)
	}
}

func TestList_EmptyRepoReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Quote{}}}
	got, hasMore, err := s.List(context.Background(), 1, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 || hasMore {
		t.Errorf("expected empty non-paged result, got len=%d hasMore=%v", len(got), hasMore)
	}
}

func TestList_HasMoreWhenResultExceedsPageSize(t *testing.T) {
	quotes := make([]domain.Quote, 21)
	for i := range quotes {
		quotes[i].ID = int64(i + 1)
	}
	s := &Service{repo: &mockRepo{list: quotes}}
	got, hasMore, err := s.List(context.Background(), 1, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasMore {
		t.Error("expected hasMore=true when repo returns limit+1 rows")
	}
	if len(got) != 20 {
		t.Errorf("expected 20 rows, got %d", len(got))
	}
}

func TestUpdate_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), 1, 5, domain.QuoteInput{CustomerName: "Updated"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateStatus_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.UpdateStatus(context.Background(), 1, 5, "accepted"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConvertToInvoice_PropagatesError(t *testing.T) {
	want := errors.New("already converted")
	s := &Service{repo: &mockRepo{convErr: want}}
	_, err := s.ConvertToInvoice(context.Background(), 1, 5)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
