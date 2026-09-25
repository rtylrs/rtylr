package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

func TestUpdateInvoice_TotalsMatchCreatePath(t *testing.T) {
	repo := &mockRepo{
		invoice: &domain.Invoice{ID: 7, Status: domain.StatusDraft},
	}
	s := &Service{repo: repo}
	inv, err := s.UpdateInvoice(context.Background(), 1, 7, domain.InvoiceUpdateInput{
		IssueDate: "2024-06-15",
		DueDate:   "2024-06-20",
		TaxRate:   10,
		Items: []domain.InvoiceItemInput{
			{Quantity: 1, UnitPrice: 100, TaxRate: 10},
		},
	})
	if err != nil {
		t.Fatalf("UpdateInvoice() error = %v", err)
	}
	if !almostEqualAmount(inv.Subtotal, 100) {
		t.Errorf("Subtotal = %f, want 100", inv.Subtotal)
	}
	if !almostEqualAmount(inv.TaxAmount, 20) {
		t.Errorf("TaxAmount = %f, want 20", inv.TaxAmount)
	}
	if !almostEqualAmount(inv.Total, 120) {
		t.Errorf("Total = %f, want 120", inv.Total)
	}
}

// ---- GetPLReport ----

func TestGetPLReport_NetProfitIsRevenueMinusExpenses(t *testing.T) {
	// sumVal is returned for both SumPaidInvoices and SumExpenses
	// Revenue=200, Expenses=200 => NetProfit=0 (same sumVal)
	// Use different repo calls by using err=nil and sumVal for all sums
	m := &mockRepo{sumVal: 500}
	s := &Service{repo: m}
	report, err := s.GetPLReport(context.Background(), 1, "month", "2024-03-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// revenue=500, expenses=500, netProfit=0
	if report.NetProfit != 0 {
		t.Errorf("expected net profit 0 (revenue=expenses), got %f", report.NetProfit)
	}
	if report.Revenue != 500 {
		t.Errorf("expected revenue=500, got %f", report.Revenue)
	}
}

func TestGetPLReport_PropagatesRevenueError(t *testing.T) {
	want := errors.New("revenue err")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GetPLReport(context.Background(), 1, "month", "2024-03-15")
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetPLReport_PeriodPassedThrough(t *testing.T) {
	s := &Service{repo: &mockRepo{sumVal: 0}}
	report, err := s.GetPLReport(context.Background(), 1, "year", "2024-06-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Period != "year" {
		t.Errorf("expected period=year, got %s", report.Period)
	}
}

func TestGetPLReport_MarginPctZeroWhenNoRevenue(t *testing.T) {
	s := &Service{repo: &mockRepo{sumVal: 0}}
	report, err := s.GetPLReport(context.Background(), 1, "month", "2024-03-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.MarginPct != 0 {
		t.Errorf("expected margin=0 when revenue=0, got %f", report.MarginPct)
	}
}

// ---- GetDashboard ----

func TestGetDashboard_PropagatesReceivableError(t *testing.T) {
	want := errors.New("receivable err")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GetDashboard(context.Background(), 901)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetDashboard_ReturnsZeroValuesForEmptyRepo(t *testing.T) {
	s := &Service{repo: &mockRepo{sumVal: 0}}
	dash, err := s.GetDashboard(context.Background(), 902)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dash.TotalReceivable != 0 || dash.OverdueAmount != 0 {
		t.Errorf("expected zero dashboard, got %+v", dash)
	}
}

func TestGetDashboard_CachesSummary(t *testing.T) {
	ctx := context.Background()
	if err := cacher.InitializeLocalCache(ctx); err != nil {
		t.Fatalf("InitializeLocalCache: %v", err)
	}

	repo := &mockRepo{sumVal: 12}
	s := &Service{repo: repo}

	for i := 0; i < 2; i++ {
		dash, err := s.GetDashboard(ctx, 991)
		if err != nil {
			t.Fatalf("call %d unexpected error: %v", i+1, err)
		}
		if dash.TotalReceivable != 12 || dash.OverdueAmount != 12 || dash.TotalExpensesMTD != 12 {
			t.Fatalf("call %d unexpected dashboard: %+v", i+1, dash)
		}
	}

	if calls := repo.sumCalls.Load(); calls != 4 {
		t.Fatalf("sum calls = %d, want 4", calls)
	}
}

func TestFetchDashboardRunsAggregatesConcurrently(t *testing.T) {
	repo := &mockRepo{sumVal: 12, delay: 20 * time.Millisecond}
	s := &Service{repo: repo}

	if _, err := s.fetchDashboard(context.Background(), 992); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if maxActive := repo.maxActive.Load(); maxActive < 2 {
		t.Fatalf("expected concurrent aggregate calls, max active calls = %d", maxActive)
	}
}

// ---- GetTaxReport ----

func TestGetTaxReport_PropagatesError(t *testing.T) {
	want := errors.New("tax err")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GetTaxReport(context.Background(), 1, "month", "2024-03-15")
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetTaxReport_PeriodPassedThrough(t *testing.T) {
	s := &Service{repo: &mockRepo{sumVal: 0}}
	report, err := s.GetTaxReport(context.Background(), 1, "quarter", "2024-04-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Period != "quarter" {
		t.Errorf("expected period=quarter, got %s", report.Period)
	}
}

// ---- UpdateInvoiceStatus ----

func TestUpdateInvoiceStatus_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{invoice: &domain.Invoice{Status: domain.StatusDraft}}}
	err := s.UpdateInvoiceStatus(context.Background(), 1, 10, "sent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateInvoiceStatus_PropagatesError(t *testing.T) {
	want := errors.New("status err")
	s := &Service{repo: &mockRepo{invoice: &domain.Invoice{Status: domain.StatusDraft}, err: want}}
	if err := s.UpdateInvoiceStatus(context.Background(), 1, 10, "sent"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateInvoice_RejectsNonDraftInvoice(t *testing.T) {
	s := &Service{repo: &mockRepo{invoice: &domain.Invoice{Status: domain.StatusPaid}}}
	if _, err := s.UpdateInvoice(context.Background(), 1, 10, domain.InvoiceUpdateInput{}); !errors.Is(err, domain.ErrInvoiceNotDraft) {
		t.Fatalf("expected ErrInvoiceNotDraft, got %v", err)
	}
}

func TestUpdateInvoiceStatus_RejectsTerminalInvoice(t *testing.T) {
	s := &Service{repo: &mockRepo{invoice: &domain.Invoice{Status: domain.StatusPaid}}}
	if err := s.UpdateInvoiceStatus(context.Background(), 1, 10, domain.StatusDraft); !errors.Is(err, domain.ErrInvoiceNotDraft) {
		t.Fatalf("expected ErrInvoiceNotDraft, got %v", err)
	}
}

func TestUpdateInvoiceStatus_RejectsReopeningInvoice(t *testing.T) {
	s := &Service{repo: &mockRepo{invoice: &domain.Invoice{Status: domain.StatusSent}}}
	if err := s.UpdateInvoiceStatus(context.Background(), 1, 10, domain.StatusDraft); !errors.Is(err, domain.ErrInvoiceNotDraft) {
		t.Fatalf("expected ErrInvoiceNotDraft, got %v", err)
	}
}
