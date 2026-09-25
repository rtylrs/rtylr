package application

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
)

// ---- mock ----

type mockRepo struct {
	invoices        []domain.Invoice
	billableTime    []domain.BillableTimeEntry
	billableTotal   int64
	billableMinutes int64
	invoice         *domain.Invoice
	expenses        []domain.Expense
	expense         *domain.Expense
	categories      []domain.ExpenseCategory
	taxConfigs      []domain.TaxConfig
	items           []domain.InvoiceItem
	payments        []domain.Payment
	generatedID     int64
	sumVal          float64
	seqCount        int64
	err             error
	writeErr        error
	created         *domain.Invoice

	createWithItemsCalls int
	lastCreatedItems     []domain.InvoiceItem

	updateWithItemsCalls int
	lastUpdatedItems     []domain.InvoiceItem
	sumCalls             atomic.Int64
	delay                time.Duration
	activeCalls          atomic.Int64
	maxActive            atomic.Int64
}

func (m *mockRepo) ListInvoices(_ context.Context, _ int64, _, _, _, _ string, _, _ int) ([]domain.Invoice, error) {
	return m.invoices, m.err
}

func (m *mockRepo) ListInvoicesForCustomer(_ context.Context, _, _ int64, _, _, _, _ string, _, _ int) ([]domain.Invoice, error) {
	return m.invoices, m.err
}

func (m *mockRepo) ListUnbilledBillableTimeEntries(_ context.Context, _ int64, _, _ int) ([]domain.BillableTimeEntry, error) {
	return m.billableTime, m.err
}

func (m *mockRepo) CountUnbilledBillableTimeEntries(_ context.Context, _ int64) (int64, int64, error) {
	return m.billableTotal, m.billableMinutes, m.err
}

func (m *mockRepo) GetInvoiceByID(_ context.Context, _, _ int64) (*domain.Invoice, error) {
	return m.invoice, m.err
}

func (m *mockRepo) CreateInvoiceWithItems(_ context.Context, _ int64, inv domain.Invoice, items []domain.InvoiceItem) (int64, error) {
	m.createWithItemsCalls++
	m.lastCreatedItems = items
	m.created = &inv
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.generatedID, m.err
}

func (m *mockRepo) GetInvoiceItems(_ context.Context, _ int64) ([]domain.InvoiceItem, error) {
	return m.items, m.err
}

func (m *mockRepo) UpdateInvoiceWithItems(_ context.Context, _ int64, _ domain.Invoice, items []domain.InvoiceItem) error {
	m.updateWithItemsCalls++
	m.lastUpdatedItems = items
	if m.writeErr != nil {
		return m.writeErr
	}
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

func (m *mockRepo) RecordPayment(_ context.Context, _ int64, _ domain.PaymentInput) (*domain.Payment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.Payment{ID: m.generatedID}, nil
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

func (m *mockRepo) DeleteExpense(_ context.Context, _, _ int64) error { return m.err }

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
	done := m.trackSumCall()
	defer done()

	return m.sumVal, m.err
}

func (m *mockRepo) SumExpenses(_ context.Context, _ int64, _, _ string) (float64, error) {
	done := m.trackSumCall()
	defer done()

	return m.sumVal, m.err
}

func (m *mockRepo) SumTotalReceivable(_ context.Context, _ int64) (float64, error) {
	done := m.trackSumCall()
	defer done()

	return m.sumVal, m.err
}

func (m *mockRepo) SumOverdueAmount(_ context.Context, _ int64) (float64, error) {
	done := m.trackSumCall()
	defer done()

	return m.sumVal, m.err
}

func (m *mockRepo) SumTaxCollected(_ context.Context, _ int64, _, _ string) (float64, error) {
	done := m.trackSumCall()
	defer done()

	return m.sumVal, m.err
}

func (m *mockRepo) SumTaxOwed(_ context.Context, _ int64, _, _ string) (float64, error) {
	done := m.trackSumCall()
	defer done()

	return m.sumVal, m.err
}

func (m *mockRepo) trackSumCall() func() {
	m.sumCalls.Add(1)
	if m.delay <= 0 {
		return func() {}
	}

	active := m.activeCalls.Add(1)
	for {
		maxActive := m.maxActive.Load()
		if active <= maxActive || m.maxActive.CompareAndSwap(maxActive, active) {
			break
		}
	}

	time.Sleep(m.delay)

	return func() {
		m.activeCalls.Add(-1)
	}
}

// ---- paginateInvoices ----

func TestPaginateInvoices_DefaultsPage1(t *testing.T) {
	limit, offset := paginateInvoices(0, 10)
	if limit != 10 || offset != 0 {
		t.Errorf("expected limit=10 offset=0, got %d/%d", limit, offset)
	}
}
