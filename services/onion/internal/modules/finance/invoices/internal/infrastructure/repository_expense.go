package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) GetExpenseByID(ctx context.Context, businessID, expenseID int64) (*domain.Expense, error) {
	row, err := r.queries.GetExpenseByID(ctx, sqlc.GetExpenseByIDParams{
		ID:         expenseID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrExpenseNotFound
		}
		logger.Ctx(ctx).Error("Failed to get expense by ID", logger.Error(err))

		return nil, err
	}

	e := expenseFromSqlc(row)

	return &e, nil
}

func (r *Repository) CreateExpense(ctx context.Context, businessID int64, e domain.Expense) (int64, error) {
	result, err := r.queries.CreateExpense(ctx, sqlc.CreateExpenseParams{
		OwnerID:         businessID,
		BusinessID:      businessID,
		CategoryID:      sqlNullInt64(e.CategoryID),
		Description:     e.Description,
		Amount:          fmtFloat(e.Amount),
		Currency:        sqlNullString(e.Currency),
		ExpenseDate:     e.ExpenseDate,
		Vendor:          sqlNullString(e.Vendor),
		PaymentMethod:   sqlc.NullFinanceExpensesPaymentMethod{FinanceExpensesPaymentMethod: sqlc.FinanceExpensesPaymentMethod(e.PaymentMethod), Valid: e.PaymentMethod != ""},
		ReceiptUrl:      sqlNullString(e.ReceiptURL),
		IsTaxDeductible: sql.NullBool{Bool: e.IsTaxDeductible, Valid: true},
		TaxAmount:       sqlNullString(fmtFloat(e.TaxAmount)),
		Notes:           sqlNullString(e.Notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create expense", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created expense ID", logger.Error(err))

		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateExpense(ctx context.Context, businessID int64, e domain.Expense) error {
	err := r.queries.UpdateExpense(ctx, sqlc.UpdateExpenseParams{
		CategoryID:      sqlNullInt64(e.CategoryID),
		Description:     e.Description,
		Amount:          fmtFloat(e.Amount),
		Currency:        sqlNullString(e.Currency),
		ExpenseDate:     e.ExpenseDate,
		Vendor:          sqlNullString(e.Vendor),
		PaymentMethod:   sqlc.NullFinanceExpensesPaymentMethod{FinanceExpensesPaymentMethod: sqlc.FinanceExpensesPaymentMethod(e.PaymentMethod), Valid: e.PaymentMethod != ""},
		ReceiptUrl:      sqlNullString(e.ReceiptURL),
		IsTaxDeductible: sql.NullBool{Bool: e.IsTaxDeductible, Valid: true},
		TaxAmount:       sqlNullString(fmtFloat(e.TaxAmount)),
		Notes:           sqlNullString(e.Notes),
		ID:              e.ID,
		BusinessID:      businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update expense", logger.Error(err))
	}

	return err
}

func (r *Repository) DeleteExpense(ctx context.Context, businessID, expenseID int64) error {
	err := r.queries.DeleteExpense(ctx, sqlc.DeleteExpenseParams{
		ID:         expenseID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete expense", logger.Error(err))
	}

	return err
}

// ---------------------------------------------------------------------------
// Expense Categories
// ---------------------------------------------------------------------------

func (r *Repository) ListExpenseCategories(ctx context.Context, businessID int64) ([]domain.ExpenseCategory, error) {
	rows, err := r.queries.ListExpenseCategories(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expense categories", logger.Error(err))

		return nil, err
	}

	categories := make([]domain.ExpenseCategory, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, domain.ExpenseCategory{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Color:      nullStr(row.Color, ""),
			Icon:       nullStr(row.Icon, ""),
			CreatedAt:  nullTime(row.CreatedAt),
		})
	}

	return categories, nil
}

func (r *Repository) CreateExpenseCategory(ctx context.Context, businessID int64, c domain.ExpenseCategory) (int64, error) {
	result, err := r.queries.CreateExpenseCategory(ctx, sqlc.CreateExpenseCategoryParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		Name:       c.Name,
		Color:      sqlNullString(c.Color),
		Icon:       sqlNullString(c.Icon),
	})
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		// Category names are unique per business; reuse the one that exists.
		var id int64
		lookupErr := r.pool.QueryRowContext(ctx,
			`SELECT id FROM finance_expense_categories WHERE business_id = ? AND name = ?`,
			businessID, c.Name).Scan(&id)
		return id, lookupErr
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create expense category", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created expense category ID", logger.Error(err))

		return 0, err
	}

	return id, nil
}

// ---------------------------------------------------------------------------
// Tax Config
// ---------------------------------------------------------------------------

func (r *Repository) ListTaxConfigs(ctx context.Context, businessID int64) ([]domain.TaxConfig, error) {
	rows, err := r.queries.ListTaxConfigs(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list tax configs", logger.Error(err))

		return nil, err
	}

	configs := make([]domain.TaxConfig, 0, len(rows))
	for _, row := range rows {
		configs = append(configs, domain.TaxConfig{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Rate:       parseFloat(row.Rate),
			Type:       nullEnumStr(row.Type.FinanceTaxConfigType, row.Type.Valid, "vat"),
			IsDefault:  nullBool(row.IsDefault),
			IsActive:   nullBool(row.IsActive),
			CreatedAt:  nullTime(row.CreatedAt),
		})
	}

	return configs, nil
}

func (r *Repository) CreateTaxConfig(ctx context.Context, businessID int64, t domain.TaxConfig) (int64, error) {
	result, err := r.queries.CreateTaxConfig(ctx, sqlc.CreateTaxConfigParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		Name:       t.Name,
		Rate:       fmtFloat(t.Rate),
		Type:       sqlc.NullFinanceTaxConfigType{FinanceTaxConfigType: sqlc.FinanceTaxConfigType(t.Type), Valid: t.Type != ""},
		IsDefault:  sql.NullBool{Bool: t.IsDefault, Valid: true},
		IsActive:   sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create tax config", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created tax config ID", logger.Error(err))

		return 0, err
	}

	return id, nil
}

// ---------------------------------------------------------------------------
// Reports
// ---------------------------------------------------------------------------

func (r *Repository) SumPaidInvoices(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	result, err := r.queries.SumPaidInvoices(ctx, sqlc.SumPaidInvoicesParams{
		BusinessID:   businessID,
		FromPaidDate: sqlNullTimeStr(dateFrom),
		ToPaidDate:   sqlNullTimeStr(dateTo),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum paid invoices", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}
