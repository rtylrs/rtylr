package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/domain"
)

type mockRepo struct {
	list      []domain.RecurringInvoice
	get       domain.RecurringInvoice
	getErr    error
	createID  int64
	createErr error
	updErr    error
	setErr    error
	setGot    struct {
		owner, id int64
		active    bool
	}
}

func (m *mockRepo) ListPage(_ context.Context, _ int64, _, _ int32) ([]domain.RecurringInvoice, error) {
	return m.list, nil
}

func (m *mockRepo) Get(_ context.Context, _, _ int64) (domain.RecurringInvoice, error) {
	return m.get, m.getErr
}

func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.RecurringInvoiceInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _, _ int64, _ domain.RecurringInvoiceInput) error {
	return m.updErr
}

func (m *mockRepo) SetActive(_ context.Context, owner, id int64, active bool) error {
	m.setGot.owner = owner
	m.setGot.id = id
	m.setGot.active = active
	return m.setErr
}

func (m *mockRepo) ProcessDue(_ context.Context, _ time.Time) ([]int64, error) {
	return nil, nil
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.RecurringInvoice{{ID: 1}}}}
	got, hasMore, err := s.List(context.Background(), 1, 1, 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
	if hasMore {
		t.Error("single short page reported hasMore=true")
	}
}

func TestList_ReportsAndTrimsContinuationRow(t *testing.T) {
	rows := make([]domain.RecurringInvoice, 4)
	s := &Service{repo: &mockRepo{list: rows}}
	got, hasMore, err := s.List(context.Background(), 1, 2, 3)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if !hasMore {
		t.Error("expected hasMore=true")
	}
	if len(got) != 3 {
		t.Errorf("got %d rows, want 3", len(got))
	}
}

func TestCreateAndUpdate_PropagateErrors(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want, updErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.RecurringInvoiceInput{}); !errors.Is(err, want) {
		t.Errorf("create: %v", err)
	}
	if err := s.Update(context.Background(), 1, 5, domain.RecurringInvoiceInput{}); !errors.Is(err, want) {
		t.Errorf("update: %v", err)
	}
}

func TestPause_SetsActiveFalse(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if err := s.Pause(context.Background(), 7, 42); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.setGot.active != false || m.setGot.owner != 7 || m.setGot.id != 42 {
		t.Errorf("set got %+v", m.setGot)
	}
}

func TestResume_SetsActiveTrue(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if err := s.Resume(context.Background(), 7, 42); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.setGot.active != true {
		t.Errorf("expected active=true, got %+v", m.setGot)
	}
}

func TestProcessDue_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.ProcessDue(context.Background()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}
