package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

func (s *Service) CreateExpense(ctx context.Context, businessID int64, input domain.ExpenseInput) (*domain.Expense, error) {
	defer invalidateInvoicesScope(ctx, businessID)
	expenseDate, err := time.Parse("2006-01-02", input.ExpenseDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse expense date", logger.Error(err))

		return nil, err
	}

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	paymentMethod := input.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "cash"
	}

	e := domain.Expense{
		OwnerID:         businessID,
		CategoryID:      input.CategoryID,
		Description:     input.Description,
		Amount:          input.Amount,
		Currency:        currency,
		ExpenseDate:     expenseDate,
		Vendor:          input.Vendor,
		PaymentMethod:   paymentMethod,
		ReceiptURL:      input.ReceiptURL,
		IsTaxDeductible: input.IsTaxDeductible,
		TaxAmount:       input.TaxAmount,
		Notes:           input.Notes,
	}

	id, err := s.repo.CreateExpense(ctx, businessID, e)
	if err != nil {
		return nil, err
	}

	e.ID = id

	return &e, nil
}

func (s *Service) UpdateExpense(ctx context.Context, businessID, expenseID int64, input domain.ExpenseInput) (*domain.Expense, error) {
	defer invalidateInvoicesScope(ctx, businessID)
	existing, err := s.repo.GetExpenseByID(ctx, businessID, expenseID)
	if err != nil {
		return nil, err
	}

	expenseDate, err := time.Parse("2006-01-02", input.ExpenseDate)
	if err != nil {
		return nil, err
	}

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	paymentMethod := input.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "cash"
	}

	existing.CategoryID = input.CategoryID
	existing.Description = input.Description
	existing.Amount = input.Amount
	existing.Currency = currency
	existing.ExpenseDate = expenseDate
	existing.Vendor = input.Vendor
	existing.PaymentMethod = paymentMethod
	existing.ReceiptURL = input.ReceiptURL
	existing.IsTaxDeductible = input.IsTaxDeductible
	existing.TaxAmount = input.TaxAmount
	existing.Notes = input.Notes

	if err := s.repo.UpdateExpense(ctx, businessID, *existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *Service) DeleteExpense(ctx context.Context, businessID, expenseID int64) error {
	defer invalidateInvoicesScope(ctx, businessID)
	return s.repo.DeleteExpense(ctx, businessID, expenseID)
}

// ---------------------------------------------------------------------------
// Expense Categories
// ---------------------------------------------------------------------------

func (s *Service) ListExpenseCategories(ctx context.Context, businessID int64) ([]domain.ExpenseCategory, error) {
	key := cacheKeys.Scoped(invoicesScope, businessID, "expense_cats")
	return cacher.Of[[]domain.ExpenseCategory]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.ExpenseCategory, error) {
			return s.listExpenseCategoriesImpl(ctx, businessID)
		},
		cacher.WithExpiry(ctx, 10*time.Minute),
	)
}

func (s *Service) listExpenseCategoriesImpl(ctx context.Context, businessID int64) ([]domain.ExpenseCategory, error) {
	return s.repo.ListExpenseCategories(ctx, businessID)
}

func (s *Service) CreateExpenseCategory(ctx context.Context, businessID int64, input domain.ExpenseCategoryInput) (*domain.ExpenseCategory, error) {
	defer invalidateInvoicesScope(ctx, businessID)
	c := domain.ExpenseCategory{
		OwnerID: businessID,
		Name:    input.Name,
		Color:   input.Color,
		Icon:    input.Icon,
	}

	id, err := s.repo.CreateExpenseCategory(ctx, businessID, c)
	if err != nil {
		return nil, err
	}

	c.ID = id

	return &c, nil
}

// ---------------------------------------------------------------------------
// Tax Config
// ---------------------------------------------------------------------------

func (s *Service) ListTaxConfigs(ctx context.Context, businessID int64) ([]domain.TaxConfig, error) {
	return s.repo.ListTaxConfigs(ctx, businessID)
}

func (s *Service) CreateTaxConfig(ctx context.Context, businessID int64, input domain.TaxConfigInput) (*domain.TaxConfig, error) {
	taxType := input.Type
	if taxType == "" {
		taxType = "vat"
	}

	t := domain.TaxConfig{
		OwnerID:   businessID,
		Name:      input.Name,
		Rate:      input.Rate,
		Type:      taxType,
		IsDefault: input.IsDefault,
		IsActive:  true,
	}

	id, err := s.repo.CreateTaxConfig(ctx, businessID, t)
	if err != nil {
		return nil, err
	}

	t.ID = id

	return &t, nil
}

// ---------------------------------------------------------------------------
// Reports
// ---------------------------------------------------------------------------

func (s *Service) GetPLReport(ctx context.Context, businessID int64, period, date string) (domain.PLReport, error) {
	dateFrom, dateTo := resolvePeriodRange(period, date)

	revenue, err := s.repo.SumPaidInvoices(ctx, businessID, dateFrom, dateTo)
	if err != nil {
		return domain.PLReport{}, err
	}

	expenses, err := s.repo.SumExpenses(ctx, businessID, dateFrom, dateTo)
	if err != nil {
		return domain.PLReport{}, err
	}

	// For simplicity COGS = 0 (can be linked to ERP in the future).
	// Operating expenses = all recorded expenses.
	grossProfit := revenue
	netProfit := grossProfit - expenses

	var marginPct float64
	if revenue > 0 {
		marginPct = (netProfit / revenue) * 100
	}

	return domain.PLReport{
		Period:            period,
		Revenue:           revenue,
		COGS:              0,
		GrossProfit:       grossProfit,
		OperatingExpenses: expenses,
		NetProfit:         netProfit,
		MarginPct:         marginPct,
	}, nil
}
