package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/segments/internal/domain"
)

type mockRepo struct {
	list          []domain.Segment
	getByIDResult domain.Segment
	getByIDErr    error
	createID      int64
	createInput   domain.SegmentInput
	createErr     error
	updErr        error
	delErr        error
	customers     []domain.SegmentCustomer
	customersErr  error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Segment, error) {
	return m.list, nil
}

func (m *mockRepo) GetByID(_ context.Context, _, _ int64) (domain.Segment, error) {
	return m.getByIDResult, m.getByIDErr
}

func (m *mockRepo) Create(_ context.Context, input domain.SegmentInput) (int64, error) {
	m.createInput = input
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ domain.SegmentUpdateInput) error {
	return m.updErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}

func (m *mockRepo) ListCustomers(_ context.Context, _ int64, _ []byte, _, _ int) ([]domain.SegmentCustomer, int, error) {
	return m.customers, len(m.customers), m.customersErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Segment{{}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), domain.SegmentInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateAndDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.SegmentUpdateInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
