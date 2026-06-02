package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
)

// ---- mock ----

type mockRepo struct {
	invoices    []domain.Invoice
	invoice     *domain.Invoice
	expenses    []domain.Expense
	expense     *domain.Expense
	categories  []domain.ExpenseCategory
	taxConfigs  []domain.TaxConfig
	items       []domain.InvoiceItem
	payments    []domain.Payment
	generatedID int64
	sumVal      float64
	seqCount    int64
	err         error
}

func (m *mockRepo) ListInvoices(_ context.Context, _ int64, _, _, _, _ string, _, _ int) ([]domain.Invoice, error) {
	return m.invoices, m.err
}

func (m *mockRepo) GetInvoiceByID(_ context.Context, _, _ int64) (*domain.Invoice, error) {
	return m.invoice, m.err
}

func (m *mockRepo) CreateInvoice(_ context.Context, _ int64, _ domain.Invoice) (int64, error) {
	return m.generatedID, m.err
}

func (m *mockRepo) CreateInvoiceItems(_ context.Context, _ int64, _ []domain.InvoiceItem) error {
	return m.err
}

func (m *mockRepo) GetInvoiceItems(_ context.Context, _ int64) ([]domain.InvoiceItem, error) {
	return m.items, m.err
}
func (m *mockRepo) DeleteInvoiceItems(_ context.Context, _ int64) error { return m.err }
func (m *mockRepo) UpdateInvoice(_ context.Context, _ int64, _ domain.Invoice) error {
	return m.err
}
func (m *mockRepo) UpdateInvoiceStatus(_ context.Context, _, _ int64, _ string) error { return m.err }
func (m *mockRepo) UpdateInvoicePaymentTotals(_ context.Context, _, _ int64, _, _ float64, _ string, _ *time.Time) error {
	return m.err
}
func (m *mockRepo) DeleteInvoice(_ context.Context, _, _ int64) error { return m.err }
func (m *mockRepo) CountInvoiceSequence(_ context.Context, _ int64, _ int) (int64, error) {
	return m.seqCount, m.err
}

func (m *mockRepo) CreatePayment(_ context.Context, _ int64, _ domain.Payment) (int64, error) {
	return m.generatedID, m.err
}

func (m *mockRepo) GetPaymentsByInvoice(_ context.Context, _ int64) ([]domain.Payment, error) {
	return m.payments, m.err
}

func (m *mockRepo) ListExpenses(_ context.Context, _, _ int64, _, _, _ string, _, _ int) ([]domain.Expense, error) {
	return m.expenses, m.err
}

func (m *mockRepo) GetExpenseByID(_ context.Context, _, _ int64) (*domain.Expense, error) {
	return m.expense, m.err
}

func (m *mockRepo) CreateExpense(_ context.Context, _ int64, _ domain.Expense) (int64, error) {
	return m.generatedID, m.err
}
func (m *mockRepo) UpdateExpense(_ context.Context, _ int64, _ domain.Expense) error { return m.err }
func (m *mockRepo) DeleteExpense(_ context.Context, _, _ int64) error                { return m.err }
func (m *mockRepo) ListExpenseCategories(_ context.Context, _ int64) ([]domain.ExpenseCategory, error) {
	return m.categories, m.err
}

func (m *mockRepo) CreateExpenseCategory(_ context.Context, _ int64, _ domain.ExpenseCategory) (int64, error) {
	return m.generatedID, m.err
}

func (m *mockRepo) ListTaxConfigs(_ context.Context, _ int64) ([]domain.TaxConfig, error) {
	return m.taxConfigs, m.err
}

func (m *mockRepo) CreateTaxConfig(_ context.Context, _ int64, _ domain.TaxConfig) (int64, error) {
	return m.generatedID, m.err
}

func (m *mockRepo) SumPaidInvoices(_ context.Context, _ int64, _, _ string) (float64, error) {
	return m.sumVal, m.err
}

func (m *mockRepo) SumExpenses(_ context.Context, _ int64, _, _ string) (float64, error) {
	return m.sumVal, m.err
}

func (m *mockRepo) SumTotalReceivable(_ context.Context, _ int64) (float64, error) {
	return m.sumVal, m.err
}

func (m *mockRepo) SumOverdueAmount(_ context.Context, _ int64) (float64, error) {
	return m.sumVal, m.err
}

func (m *mockRepo) SumTaxCollected(_ context.Context, _ int64, _, _ string) (float64, error) {
	return m.sumVal, m.err
}

func (m *mockRepo) SumTaxOwed(_ context.Context, _ int64, _, _ string) (float64, error) {
	return m.sumVal, m.err
}

// ---- paginateInvoices ----

func TestPaginateInvoices_DefaultsPage1(t *testing.T) {
	limit, offset := paginateInvoices(0, 10)
	if limit != 10 || offset != 0 {
		t.Errorf("expected limit=10 offset=0, got %d/%d", limit, offset)
	}
}

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

func TestCreateExpenseCategory_PropagatesRepoError(t *testing.T) {
	want := errors.New("cat error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.CreateExpenseCategory(context.Background(), 1, domain.ExpenseCategoryInput{Name: "Office"})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- CreateTaxConfig ----

func TestCreateTaxConfig_DefaultsTypeToVAT(t *testing.T) {
	s := &Service{repo: &mockRepo{generatedID: 5}}
	tc, err := s.CreateTaxConfig(context.Background(), 1, domain.TaxConfigInput{Name: "GST", Rate: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.Type != "vat" {
		t.Errorf("expected type=vat, got %s", tc.Type)
	}
}

func TestCreateTaxConfig_PropagatesRepoError(t *testing.T) {
	want := errors.New("tax error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.CreateTaxConfig(context.Background(), 1, domain.TaxConfigInput{Name: "VAT", Rate: 20})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- CreateInvoice date validation ----

func TestCreateInvoice_RejectsDueDateBeforeIssueDate(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		IssueDate: "2024-06-15",
		DueDate:   "2024-06-10",
	})
	if !errors.Is(err, domain.ErrInvalidInvoiceDate) {
		t.Errorf("expected ErrInvalidInvoiceDate, got %v", err)
	}
}

func TestCreateInvoice_RejectsInvalidIssueDate(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		IssueDate: "not-a-date",
		DueDate:   "2024-06-30",
	})
	if err == nil {
		t.Error("expected parse error, got nil")
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
	_, err := s.GetDashboard(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetDashboard_ReturnsZeroValuesForEmptyRepo(t *testing.T) {
	s := &Service{repo: &mockRepo{sumVal: 0}}
	dash, err := s.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dash.TotalReceivable != 0 || dash.OverdueAmount != 0 {
		t.Errorf("expected zero dashboard, got %+v", dash)
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
	s := &Service{repo: &mockRepo{}}
	err := s.UpdateInvoiceStatus(context.Background(), 1, 10, "sent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateInvoiceStatus_PropagatesError(t *testing.T) {
	want := errors.New("status err")
	s := &Service{repo: &mockRepo{err: want}}
	if err := s.UpdateInvoiceStatus(context.Background(), 1, 10, "sent"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
