package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/invoices/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/pagination"
)

const (
	invoicesScope = "invoices"
	invoicesTTL   = 5 * time.Minute
)

var invoicePagePolicy = pagination.Policy{
	DefaultPageSize: 50,
	MaxPageSize:     500,
	MaxOffset:       10000,
}

func invalidateInvoicesScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(invoicesScope, businessID))
}

type repository interface {
	ListInvoices(ctx context.Context, businessID int64, status, search, dateFrom, dateTo string, limit, offset int) ([]domain.Invoice, error)
	ListInvoicesForCustomer(ctx context.Context, businessID, customerID int64, status, search, dateFrom, dateTo string, limit, offset int) ([]domain.Invoice, error)
	ListUnbilledBillableTimeEntries(ctx context.Context, businessID int64, limit, offset int) ([]domain.BillableTimeEntry, error)
	CountUnbilledBillableTimeEntries(ctx context.Context, businessID int64) (int64, int64, error)
	ListInvoicesByCustomer(ctx context.Context, businessID, customerID int64, limit, offset int) ([]domain.Invoice, error)
	CountInvoicesByCustomer(ctx context.Context, businessID, customerID int64) (int64, error)
	GetInvoiceByID(ctx context.Context, businessID, invoiceID int64) (*domain.Invoice, error)
	CreateInvoiceWithItems(ctx context.Context, businessID int64, inv domain.Invoice, items []domain.InvoiceItem) (int64, error)
	GetInvoiceItems(ctx context.Context, invoiceID int64) ([]domain.InvoiceItem, error)
	UpdateInvoiceWithItems(ctx context.Context, businessID int64, inv domain.Invoice, items []domain.InvoiceItem) error
	UpdateInvoiceStatus(ctx context.Context, businessID, invoiceID int64, status string) error
	UpdateInvoicePaymentTotals(ctx context.Context, businessID, invoiceID int64, amountPaid, amountDue float64, status string, paidDate *time.Time) error
	DeleteInvoice(ctx context.Context, businessID, invoiceID int64) error
	CountInvoiceSequence(ctx context.Context, businessID int64, year int) (int64, error)
	RecordPayment(ctx context.Context, businessID int64, input domain.PaymentInput) (*domain.Payment, error)
	GetPaymentsByInvoice(ctx context.Context, invoiceID int64) ([]domain.Payment, error)
	ListExpenses(ctx context.Context, businessID, categoryID int64, search, dateFrom, dateTo string, limit, offset int) ([]domain.Expense, error)
	GetExpenseByID(ctx context.Context, businessID, expenseID int64) (*domain.Expense, error)
	CreateExpense(ctx context.Context, businessID int64, e domain.Expense) (int64, error)
	UpdateExpense(ctx context.Context, businessID int64, e domain.Expense) error
	DeleteExpense(ctx context.Context, businessID, expenseID int64) error
	ListExpenseCategories(ctx context.Context, businessID int64) ([]domain.ExpenseCategory, error)
	CreateExpenseCategory(ctx context.Context, businessID int64, c domain.ExpenseCategory) (int64, error)
	ListTaxConfigs(ctx context.Context, businessID int64) ([]domain.TaxConfig, error)
	CreateTaxConfig(ctx context.Context, businessID int64, t domain.TaxConfig) (int64, error)
	SumPaidInvoices(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error)
	SumExpenses(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error)
	SumTotalReceivable(ctx context.Context, businessID int64) (float64, error)
	SumOverdueAmount(ctx context.Context, businessID int64) (float64, error)
	SumTaxCollected(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error)
	SumTaxOwed(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error)
}

type Service struct {
	repo              repository
	db                *sql.DB
	customerSnapshots CustomerSnapshotProvider
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
		db:   db,
	}
}

// ---------------------------------------------------------------------------
// Invoices
// ---------------------------------------------------------------------------

func (s *Service) ListInvoices(ctx context.Context, businessID int64, status, search, dateFrom, dateTo string, page, pageSize int64) ([]domain.Invoice, bool, error) {
	limit, offset := paginateInvoices(page, pageSize)
	key := cacheKeys.Scoped(invoicesScope, businessID, fmt.Sprintf("list|s=%s|q=%s|f=%s|t=%s|p=%d|ps=%d", status, search, dateFrom, dateTo, page, pageSize))
	rows, err := cacher.Of[[]domain.Invoice]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Invoice, error) {
			return s.repo.ListInvoices(ctx, businessID, status, search, dateFrom, dateTo, limit+1, offset)
		},
		cacher.WithExpiry(ctx, invoicesTTL),
	)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}

// ListInvoicesForCustomer applies the same pagination and filter semantics as
// ListInvoices while also restricting the result to one CRM customer.
func (s *Service) ListInvoicesForCustomer(ctx context.Context, businessID, customerID, page, pageSize int64, status, search, dateFrom, dateTo string) ([]domain.Invoice, bool, error) {
	limit, offset := paginateInvoices(page, pageSize)
	key := cacheKeys.Scoped(invoicesScope, businessID, fmt.Sprintf("by-customer=%d|s=%s|q=%s|f=%s|t=%s|p=%d|ps=%d", customerID, status, search, dateFrom, dateTo, page, pageSize))
	rows, err := cacher.Of[[]domain.Invoice]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Invoice, error) {
			return s.repo.ListInvoicesForCustomer(ctx, businessID, customerID, status, search, dateFrom, dateTo, limit+1, offset)
		},
		cacher.WithExpiry(ctx, invoicesTTL),
	)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}

func (s *Service) ListUnbilledBillableTimeEntries(ctx context.Context, businessID, page, pageSize int64) ([]domain.BillableTimeEntry, int64, int64, bool, error) {
	limit, offset := paginateInvoices(page, pageSize)

	total, totalMinutes, err := s.repo.CountUnbilledBillableTimeEntries(ctx, businessID)
	if err != nil {
		return nil, 0, 0, false, err
	}

	rows, err := s.repo.ListUnbilledBillableTimeEntries(ctx, businessID, limit+1, offset)
	if err != nil {
		return nil, 0, 0, false, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, total, totalMinutes, hasMore, nil
}

func paginateInvoices(page, pageSize int64) (int, int) {
	normalized := invoicePagePolicy.Normalize(page, pageSize)
	return int(normalized.Size), int(normalized.Offset)
}

func (s *Service) GetInvoice(ctx context.Context, businessID, invoiceID int64) (*domain.Invoice, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, businessID, invoiceID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetInvoiceItems(ctx, inv.ID)
	if err != nil {
		return nil, err
	}

	inv.Items = items

	payments, err := s.repo.GetPaymentsByInvoice(ctx, inv.ID)
	if err != nil {
		return nil, err
	}

	inv.Payments = payments

	return inv, nil
}

// applyHeaderDiscount resolves a header discount against the subtotal shared
// by the create and update paths. Discounts that would drive the invoice
// below zero are rejected instead of persisting negative totals.
func applyHeaderDiscount(subtotal float64, discountType string, discountValue float64) (float64, error) {
	var discountAmount float64
	switch discountType {
	case "percentage":
		if discountValue < 0 || discountValue > 100 {
			return 0, domain.ErrDiscountExceedsTotal
		}
		discountAmount = subtotal * (discountValue / 100)
	default:
		if discountValue < 0 || discountValue > subtotal {
			return 0, domain.ErrDiscountExceedsTotal
		}
		discountAmount = discountValue
	}
	return discountAmount, nil
}

func (s *Service) CreateInvoice(ctx context.Context, businessID int64, input domain.InvoiceInput) (*domain.Invoice, error) {
	defer invalidateInvoicesScope(ctx, businessID)

	// CRM->FIN-02: a supplied customer_id is verified against this business
	// and its name/email/address are snapshotted onto the invoice. The tax id
	// stays caller-supplied (CRM customers carry none today).
	name, email, address, snapErr := s.resolveCustomerSnapshot(ctx, businessID, input.CustomerID)
	if snapErr != nil {
		return nil, snapErr
	}
	if input.CustomerID > 0 {
		input.ClientName = name
		input.ClientEmail = email
		if address != "" {
			input.ClientAddress = address
		}
	}

	issueDate, err := time.Parse("2006-01-02", input.IssueDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse issue date", logger.Error(err))

		return nil, err
	}

	dueDate, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse due date", logger.Error(err))

		return nil, err
	}

	if dueDate.Before(issueDate) {
		return nil, domain.ErrInvalidInvoiceDate
	}

	// Auto-generate invoice number: INV-YYYY-XXXX
	year := issueDate.Year()

	seq, err := s.repo.CountInvoiceSequence(ctx, businessID, year)
	if err != nil {
		return nil, err
	}

	invoiceNumber := fmt.Sprintf("INV-%d-%04d", year, seq+1)

	// Line bases, the header discount, the taxable base, every tax layer and the
	// rounding: all of it lives in computeInvoiceTotals. See money.go for the
	// order of operations; do not re-derive it here.
	discountType := input.DiscountType
	if discountType == "" {
		discountType = "fixed"
	}

	totals, err := computeInvoiceTotals(input.Items, input.TaxRate, discountType, input.DiscountValue)
	if err != nil {
		return nil, err
	}

	items := make([]domain.InvoiceItem, 0, len(input.Items))

	for i, itemInput := range input.Items {
		unit := itemInput.Unit
		if unit == "" {
			unit = "unit"
		}

		items = append(items, domain.InvoiceItem{
			Description: itemInput.Description,
			Quantity:    itemInput.Quantity,
			Unit:        unit,
			UnitPrice:   itemInput.UnitPrice,
			TaxRate:     itemInput.TaxRate,
			DiscountPct: itemInput.DiscountPct,
			Total:       totals.LineTotal(i),
			SortOrder:   i,
		})
	}

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	inv := domain.Invoice{
		OwnerID:        businessID,
		InvoiceNumber:  invoiceNumber,
		ClientName:     input.ClientName,
		ClientEmail:    input.ClientEmail,
		ClientAddress:  input.ClientAddress,
		ClientTaxID:    input.ClientTaxID,
		Status:         "draft",
		IssueDate:      issueDate,
		DueDate:        dueDate,
		Currency:       currency,
		Subtotal:       totals.Subtotal,
		DiscountType:   discountType,
		DiscountValue:  input.DiscountValue,
		DiscountAmount: totals.DiscountAmount,
		TaxRate:        input.TaxRate,
		TaxAmount:      totals.TaxAmount,
		Total:          totals.Total,
		AmountPaid:     0,
		AmountDue:      totals.Total,
		Notes:          input.Notes,
		Terms:          input.Terms,
	}
	// The verified customer link must be persisted, or per-customer reads
	// (ListInvoicesByCustomer) can never return the invoice.
	if input.CustomerID > 0 {
		inv.CustomerID = &input.CustomerID
	}

	invID, err := s.repo.CreateInvoiceWithItems(ctx, businessID, inv, items)
	if err != nil {
		return nil, err
	}

	inv.ID = invID
	inv.Items = items

	return &inv, nil
}
