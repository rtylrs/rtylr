package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
)

func TestPaginateInvoices_DefaultsPageSize50(t *testing.T) {
	limit, _ := paginateInvoices(1, 0)
	if limit != 50 {
		t.Errorf("expected default pageSize=50, got %d", limit)
	}
}

func TestPaginateInvoices_CapsAt500(t *testing.T) {
	limit, _ := paginateInvoices(1, 1000)
	if limit != 500 {
		t.Errorf("expected cap at 500, got %d", limit)
	}
}

func TestPaginateInvoices_Page2Offset(t *testing.T) {
	_, got := paginateInvoices(2, 25)
	if got != 25 {
		t.Errorf("expected offset=25, got %d", got)
	}
}

func TestPaginateInvoices_ClampsDeepOffset(t *testing.T) {
	_, got := paginateInvoices(100000, 500)
	if got != 10000 {
		t.Errorf("expected deep offset to clamp at 10000, got %d", got)
	}
}

// ---- resolvePeriodRange ----

func TestResolvePeriodRange_MonthReturnsFirstAndLastDay(t *testing.T) {
	from, to := resolvePeriodRange("month", "2024-03-15")
	if from != "2024-03-01" {
		t.Errorf("expected 2024-03-01, got %s", from)
	}
	if to != "2024-03-31" {
		t.Errorf("expected 2024-03-31, got %s", to)
	}
}

func TestResolvePeriodRange_QuarterQ1(t *testing.T) {
	from, to := resolvePeriodRange("quarter", "2024-02-10")
	if from != "2024-01-01" {
		t.Errorf("expected 2024-01-01, got %s", from)
	}
	if to != "2024-03-31" {
		t.Errorf("expected 2024-03-31, got %s", to)
	}
}

func TestResolvePeriodRange_YearReturnsFullYear(t *testing.T) {
	from, to := resolvePeriodRange("year", "2024-06-01")
	if from != "2024-01-01" {
		t.Errorf("expected 2024-01-01, got %s", from)
	}
	if to != "2024-12-31" {
		t.Errorf("expected 2024-12-31, got %s", to)
	}
}

func TestResolvePeriodRange_UnknownDefaultsToMonth(t *testing.T) {
	from, _ := resolvePeriodRange("unknown", "2024-03-15")
	if from != "2024-03-01" {
		t.Errorf("expected month default (2024-03-01), got %s", from)
	}
}

// ---- ListInvoices ----

func TestListInvoices_ReturnsDelegatedItems(t *testing.T) {
	rows := []domain.Invoice{{ID: 1}, {ID: 2}}
	s := &Service{repo: &mockRepo{invoices: rows}}
	got, hasMore, err := s.ListInvoices(context.Background(), 1, "", "", "", "", 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || hasMore {
		t.Errorf("expected 2 items no-more, got len=%d hasMore=%v", len(got), hasMore)
	}
}

func TestListUnbilledBillableTimeEntries_PaginatesAndReturnsTotals(t *testing.T) {
	repo := &mockRepo{
		billableTime: []domain.BillableTimeEntry{
			{ID: 1, DurationMinutes: 30},
			{ID: 2, DurationMinutes: 60},
			{ID: 3, DurationMinutes: 90},
		},
		billableTotal:   3,
		billableMinutes: 180,
	}
	s := &Service{repo: repo}
	s.SetBillableTimeReader(repo)

	entries, total, totalMinutes, hasMore, err := s.ListUnbilledBillableTimeEntries(context.Background(), 1001, 1, 2)
	if err != nil {
		t.Fatalf("ListUnbilledBillableTimeEntries returned error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if totalMinutes != 180 {
		t.Errorf("total minutes = %d, want 180", totalMinutes)
	}
	if !hasMore {
		t.Error("hasMore = false, want true")
	}
}

func TestListInvoices_HasMoreWhenExtraItemReturned(t *testing.T) {
	// pageSize=2 but repo returns 3 items
	rows := []domain.Invoice{{ID: 1}, {ID: 2}, {ID: 3}}
	s := &Service{repo: &mockRepo{invoices: rows}}
	got, hasMore, err := s.ListInvoices(context.Background(), 1, "", "", "", "", 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasMore {
		t.Error("expected hasMore=true")
	}
	if len(got) != 2 {
		t.Errorf("expected trimmed to 2, got %d", len(got))
	}
}

func TestListInvoices_PropagatesRepoError(t *testing.T) {
	want := errors.New("list error")
	s := &Service{repo: &mockRepo{err: want}}
	_, _, err := s.ListInvoices(context.Background(), 1, "", "", "", "", 1, 10)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- DeleteInvoice ----

func TestDeleteInvoice_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.DeleteInvoice(context.Background(), 1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteInvoice_PropagatesRepoError(t *testing.T) {
	want := errors.New("delete error")
	s := &Service{repo: &mockRepo{err: want}}
	err := s.DeleteInvoice(context.Background(), 1, 10)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- ListExpenses ----

func TestListExpenses_ReturnsDelegatedItems(t *testing.T) {
	rows := []domain.Expense{{ID: 1, Amount: 50}, {ID: 2, Amount: 75}}
	s := &Service{repo: &mockRepo{expenses: rows}}
	got, hasMore, err := s.ListExpenses(context.Background(), 1, 0, "", "", "", 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || hasMore {
		t.Errorf("expected 2 items no-more, got len=%d hasMore=%v", len(got), hasMore)
	}
}

func TestListExpenses_HasMoreWhenExtraItemReturned(t *testing.T) {
	rows := []domain.Expense{{ID: 1}, {ID: 2}, {ID: 3}}
	s := &Service{repo: &mockRepo{expenses: rows}}
	_, hasMore, err := s.ListExpenses(context.Background(), 1, 0, "", "", "", 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasMore {
		t.Error("expected hasMore=true")
	}
}

// ---- CreateExpenseCategory ----

func TestCreateExpenseCategory_ReturnsGeneratedID(t *testing.T) {
	s := &Service{repo: &mockRepo{generatedID: 99}}
	cat, err := s.CreateExpenseCategory(context.Background(), 1, domain.ExpenseCategoryInput{Name: "Office"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat.ID != 99 {
		t.Errorf("expected ID=99, got %d", cat.ID)
	}
}
