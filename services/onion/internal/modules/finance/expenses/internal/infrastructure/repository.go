package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/finance/expenses/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/expenses/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Expense, error) {
	rows, err := r.queries.ListExpenses(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expenses", logger.Error(err))
		return nil, err
	}

	expenses := make([]domain.Expense, 0, len(rows))
	for _, row := range rows {
		expenses = append(expenses, expenseFromRow(row))
	}
	return expenses, nil
}

// ListPaginated returns a single page of expenses for the business, ordered by
// expense date. Limit and offset are applied as-is; callers clamp them.
func (r *Repository) ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Expense, error) {
	rows, err := r.queries.ListExpensesPaginated(ctx, sqlc.ListExpensesPaginatedParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expenses page", logger.Error(err))
		return nil, err
	}

	expenses := make([]domain.Expense, 0, len(rows))
	for _, row := range rows {
		expenses = append(expenses, expenseFromRow(row))
	}
	return expenses, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (domain.Expense, error) {
	row, err := r.queries.GetExpenseByID(ctx, sqlc.GetExpenseByIDParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Expense{}, domain.ErrExpenseNotFound
		}
		return domain.Expense{}, err
	}
	return expenseFromGetRow(row), nil
}

func (r *Repository) Create(ctx context.Context, input domain.ExpenseInput) (int64, error) {
	businessID := input.BusinessID
	if businessID == 0 {
		businessID = input.OwnerID
	}

	result, err := r.queries.CreateExpense(ctx, sqlc.CreateExpenseParams{
		OwnerID:         input.OwnerID,
		BusinessID:      businessID,
		CategoryID:      nullableInt64Ptr(input.CategoryID),
		Amount:          fmtFloat(input.Amount),
		Currency:        nullableString(input.Currency),
		Description:     input.Description,
		Vendor:          nullableString(input.Vendor),
		PaymentMethod:   sqlc.NullFinanceExpensesPaymentMethod{FinanceExpensesPaymentMethod: sqlc.FinanceExpensesPaymentMethod(input.PaymentMethod), Valid: input.PaymentMethod != ""},
		ReceiptUrl:      nullableString(input.ReceiptURL),
		IsTaxDeductible: sql.NullBool{Bool: input.IsTaxDeductible, Valid: true},
		TaxAmount:       nullableString(fmtFloat(input.TaxAmount)),
		Notes:           nullableString(input.Notes),
		ExpenseDate:     input.ExpenseDate,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create expense", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.ExpenseUpdateInput) error {
	current, err := r.GetByID(ctx, input.BusinessID, input.ID)
	if err != nil {
		return err
	}
	if current.Status == "approved" {
		return domain.ErrApprovedExpenseImmutable
	}

	result, err := r.queries.UpdateExpense(ctx, sqlc.UpdateExpenseParams{
		CategoryID:      nullableInt64Ptr(input.CategoryID),
		Amount:          fmtFloat(input.Amount),
		Currency:        nullableString(input.Currency),
		Description:     input.Description,
		Vendor:          nullableString(input.Vendor),
		PaymentMethod:   sqlc.NullFinanceExpensesPaymentMethod{FinanceExpensesPaymentMethod: sqlc.FinanceExpensesPaymentMethod(input.PaymentMethod), Valid: input.PaymentMethod != ""},
		ReceiptUrl:      nullableString(input.ReceiptURL),
		IsTaxDeductible: sql.NullBool{Bool: input.IsTaxDeductible, Valid: true},
		TaxAmount:       nullableString(fmtFloat(input.TaxAmount)),
		Notes:           nullableString(input.Notes),
		ExpenseDate:     input.ExpenseDate,
		ID:              input.ID,
		BusinessID:      input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update expense", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrExpenseNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	current, err := r.GetByID(ctx, businessID, id)
	if err != nil {
		return err
	}
	if current.Status == "approved" {
		return domain.ErrApprovedExpenseImmutable
	}

	result, err := r.queries.DeleteExpense(ctx, sqlc.DeleteExpenseParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete expense", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrExpenseNotFound
	}
	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, businessID, id int64, status string) error {
	current, err := r.GetByID(ctx, businessID, id)
	if err != nil {
		return err
	}
	if err := domain.ValidateExpenseStatusTransition(current.Status, status); err != nil {
		return err
	}

	result, err := r.queries.UpdateExpenseStatus(ctx, sqlc.UpdateExpenseStatusParams{
		Status:        sql.NullString{String: status, Valid: true},
		ID:            id,
		BusinessID:    businessID,
		CurrentStatus: sql.NullString{String: current.Status, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update expense status", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrExpenseNotFound
	}
	return nil
}

func (r *Repository) ListCategories(ctx context.Context, businessID int64) ([]domain.ExpenseCategory, error) {
	rows, err := r.queries.ListExpenseCategories(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expense categories", logger.Error(err))
		return nil, err
	}

	cats := make([]domain.ExpenseCategory, 0, len(rows))
	for _, row := range rows {
		cats = append(cats, domain.ExpenseCategory{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Color:      nullStr(row.Color),
			CreatedAt:  nullTime(row.CreatedAt),
		})
	}
	return cats, nil
}

// SumByOwner returns the total expense amount for the given business, regardless
// of category or date.
func (r *Repository) SumByOwner(ctx context.Context, businessID int64) (float64, error) {
	total, err := r.queries.SumExpensesByBusiness(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum expenses", logger.Error(err))
		return 0, err
	}
	return interfaceToFloat(total), nil
}

func interfaceToFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case []byte:
		return sqlutil.ParseDecimal(string(val))
	case string:
		return sqlutil.ParseDecimal(val)
	default:
		return 0
	}
}

func (r *Repository) CountByCategory(ctx context.Context, businessID int64) (map[string]int64, error) {
	rows, err := r.queries.CountExpensesByCategory(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count expenses by category", logger.Error(err))
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		var key string
		switch value := row.CategoryKey.(type) {
		case []byte:
			key = string(value)
		case string:
			key = value
		default:
			key = fmt.Sprint(value)
		}
		counts[key] = row.Cnt
	}
	return counts, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func expenseFromRow(row sqlc.FinanceExpense) domain.Expense {
	e := domain.Expense{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		Amount:          parseFloat(row.Amount),
		Currency:        nullStr(row.Currency),
		Description:     row.Description,
		Vendor:          nullStr(row.Vendor),
		PaymentMethod:   nullEnumStr(row.PaymentMethod.FinanceExpensesPaymentMethod, row.PaymentMethod.Valid, "cash"),
		ReceiptURL:      nullStr(row.ReceiptUrl),
		IsTaxDeductible: nullBool(row.IsTaxDeductible),
		TaxAmount:       parseFloatNull(row.TaxAmount),
		Notes:           nullStr(row.Notes),
		ExpenseDate:     row.ExpenseDate,
		Status:          nullStr(row.Status),
		CreatedAt:       nullTime(row.CreatedAt),
		UpdatedAt:       nullTime(row.UpdatedAt),
	}
	if row.CategoryID.Valid {
		e.CategoryID = &row.CategoryID.Int64
	}
	return e
}

func expenseFromGetRow(row sqlc.FinanceExpense) domain.Expense {
	e := domain.Expense{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		Amount:          parseFloat(row.Amount),
		Currency:        nullStr(row.Currency),
		Description:     row.Description,
		Vendor:          nullStr(row.Vendor),
		PaymentMethod:   nullEnumStr(row.PaymentMethod.FinanceExpensesPaymentMethod, row.PaymentMethod.Valid, "cash"),
		ReceiptURL:      nullStr(row.ReceiptUrl),
		IsTaxDeductible: nullBool(row.IsTaxDeductible),
		TaxAmount:       parseFloatNull(row.TaxAmount),
		Notes:           nullStr(row.Notes),
		ExpenseDate:     row.ExpenseDate,
		Status:          nullStr(row.Status),
		CreatedAt:       nullTime(row.CreatedAt),
		UpdatedAt:       nullTime(row.UpdatedAt),
	}
	if row.CategoryID.Valid {
		e.CategoryID = &row.CategoryID.Int64
	}
	return e
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullableInt64Ptr(p *int64) sql.NullInt64 {
	if p == nil || *p == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func nullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func nullBool(b sql.NullBool) bool {
	return b.Valid && b.Bool
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid {
		return 0
	}
	return parseFloat(s.String)
}

func nullEnumStr[T ~string](value T, valid bool, fallback string) string {
	if valid {
		return string(value)
	}
	return fallback
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse expense decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return v
}

// Ensure these are used.
var _ = fmt.Sprintf
