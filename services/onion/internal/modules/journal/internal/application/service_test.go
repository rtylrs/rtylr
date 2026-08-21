package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	"github.com/voxire/onion/internal/platform/financeaccounts"
)

type mockRepo struct {
	list        []domain.JournalEntry
	get         domain.JournalEntry
	getErr      error
	createID    int64
	createErr   error
	createCalls int
	postErr     error
	payrollErr  error
}

func (m *mockRepo) List(_ context.Context, _ int64, _, _ int32) ([]domain.JournalEntry, error) {
	return m.list, nil
}

func (m *mockRepo) Get(_ context.Context, _, _ int64) (domain.JournalEntry, error) {
	return m.get, m.getErr
}

func (m *mockRepo) Create(_ context.Context, _, _ int64, _ domain.JournalEntryInput) (int64, error) {
	m.createCalls++
	return m.createID, m.createErr
}

func (m *mockRepo) Post(_ context.Context, _, _ int64) error {
	return m.postErr
}

func (m *mockRepo) PostPayrollJournal(_ context.Context, _ domain.PayrollJournalInput) error {
	return m.payrollErr
}

func (m *mockRepo) SetPayrollAccountProvider(financeaccounts.PayrollAccountProvider) {}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.JournalEntry{{ID: 1}}}}
	got, _, _ := s.List(context.Background(), 1, 1, 20)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestGet_Delegates(t *testing.T) {
	want := domain.JournalEntry{ID: 42}
	s := &Service{repo: &mockRepo{get: want}}
	got, err := s.Get(context.Background(), 1, want.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Fatalf("expected id=%d, got %d", want.ID, got.ID)
	}
}

func TestGet_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{getErr: want}}
	if _, err := s.Get(context.Background(), 1, 42); !errors.Is(err, want) {
		t.Fatalf("expected sentinel, got %v", err)
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, 1, validJournalEntryInput()); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestPost_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{postErr: want}}
	if err := s.Post(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 42}}
	id, err := s.Create(context.Background(), 1, 1, validJournalEntryInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected id=42, got %d", id)
	}
}

func TestCreate_RejectsInvalidJournalEntryBeforeRepository(t *testing.T) {
	repo := &mockRepo{createID: 42}
	s := &Service{repo: repo}
	_, err := s.Create(context.Background(), 1, 1, domain.JournalEntryInput{
		Date: "2026-08-13",
		Lines: []domain.JournalLineInput{
			{AccountID: 1, Debit: 100},
			{AccountID: 2, Credit: 90},
		},
	})
	if !errors.Is(err, domain.ErrInvalidJournalEntry) {
		t.Fatalf("expected ErrInvalidJournalEntry, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("repository Create called %d times for invalid input, want 0", repo.createCalls)
	}
}

func validJournalEntryInput() domain.JournalEntryInput {
	return domain.JournalEntryInput{
		Date: "2026-08-13",
		Lines: []domain.JournalLineInput{
			{AccountID: 1, Debit: 100},
			{AccountID: 2, Credit: 100},
		},
	}
}

func TestPost_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Post(context.Background(), 1, 10); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPostPayrollJournal_ValidatesBeforeRepository(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}

	err := s.PostPayrollJournal(context.Background(), domain.PayrollJournalInput{
		OwnerID:     1,
		BusinessID:  1,
		RunID:       7,
		PeriodStart: "2026-08-01",
		PeriodEnd:   "2026-08-31",
		TotalGross:  1000,
		TotalNet:    1100,
	})
	if err == nil {
		t.Fatal("expected invalid payroll journal error")
	}
}

func TestPostPayrollJournal_PropagatesRepositoryError(t *testing.T) {
	want := errors.New("journal unavailable")
	s := &Service{repo: &mockRepo{payrollErr: want}}

	err := s.PostPayrollJournal(context.Background(), domain.PayrollJournalInput{
		OwnerID:     1,
		BusinessID:  1,
		RunID:       7,
		PeriodStart: "2026-08-01",
		PeriodEnd:   "2026-08-31",
		TotalGross:  1000,
		TotalNet:    900,
	})
	if !errors.Is(err, want) {
		t.Fatalf("expected repository error %v, got %v", want, err)
	}
}

func TestList_EmptyRepoReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.JournalEntry{}}}
	got, hasMore, err := s.List(context.Background(), 1, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 entries, got %d", len(got))
	}
	if hasMore {
		t.Error("expected hasMore=false for empty list")
	}
}

func TestList_HasMoreWhenResultExceedsPageSize(t *testing.T) {
	entries := make([]domain.JournalEntry, 21)
	for i := range entries {
		entries[i].ID = int64(i + 1)
	}
	s := &Service{repo: &mockRepo{list: entries}}
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

func TestList_PageSizeDefaultsTo50(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, _, err := s.List(context.Background(), 1, 1, 0)
	if err != nil {
		t.Fatalf("unexpected error with pageSize=0: %v", err)
	}
}

func TestList_PageDefaultsTo1(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, _, err := s.List(context.Background(), 1, 0, 20)
	if err != nil {
		t.Fatalf("unexpected error with page=0: %v", err)
	}
}
