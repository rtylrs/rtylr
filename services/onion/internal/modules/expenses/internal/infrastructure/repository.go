package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/expenses/internal/domain"
	"github.com/voxire/onion/internal/modules/expenses/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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
	result, err := r.queries.CreateExpense(ctx, sqlc.CreateExpenseParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		CategoryID:  nullableInt64Ptr(input.CategoryID),
		Amount:      fmtFloat(input.Amount),
		Currency:    nullableString(input.Currency),
		Description: input.Description,
		Vendor:      nullableString(input.Vendor),
		ReceiptUrl:  nullableString(input.ReceiptURL),
		ExpenseDate: input.ExpenseDate,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create expense", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.ExpenseUpdateInput) error {
	result, err := r.queries.UpdateExpense(ctx, sqlc.UpdateExpenseParams{
		CategoryID:  nullableInt64Ptr(input.CategoryID),
		Amount:      fmtFloat(input.Amount),
		Currency:    nullableString(input.Currency),
		Description: input.Description,
		Vendor:      nullableString(input.Vendor),
		ReceiptUrl:  nullableString(input.ReceiptURL),
		ExpenseDate: input.ExpenseDate,
		ID:          input.ID,
		BusinessID:  input.BusinessID,
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
	result, err := r.queries.UpdateExpenseStatus(ctx, sqlc.UpdateExpenseStatusParams{
		Status:     sql.NullString{String: status, Valid: true},
		ID:         id,
		BusinessID: businessID,
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
	sqlStr, args, err := sq.Select("COALESCE(SUM(CAST(amount AS REAL)), 0)").
		From("finance_expenses").
		Where(sq.Eq{"business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build sum expenses query", logger.Error(err))
		return 0, err
	}

	var total float64
	if scanErr := r.pool.QueryRowContext(ctx, sqlStr, args...).Scan(&total); scanErr != nil {
		logger.Ctx(ctx).Error("Failed to sum expenses", logger.Error(scanErr))
		return 0, scanErr
	}

	return total, nil
}

// CountByCategory returns the number of expenses in each category for the business.
// The map key is the category_id (as a string) or "uncategorised" for null values.
func (r *Repository) CountByCategory(ctx context.Context, businessID int64) (map[string]int64, error) {
	sqlStr, args, err := sq.Select(
		"COALESCE(CAST(category_id AS TEXT), 'uncategorised')",
		"COUNT(*) AS cnt",
	).
		From("finance_expenses").
		Where(sq.Eq{"business_id": businessID}).
		GroupBy("category_id").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build count expenses by category query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count expenses by category", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var cat string
		var cnt int64
		if scanErr := rows.Scan(&cat, &cnt); scanErr != nil {
			return nil, scanErr
		}
		counts[cat] = cnt
	}

	return counts, rows.Err()
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func expenseFromRow(row sqlc.ListExpensesRow) domain.Expense {
	e := domain.Expense{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Amount:      parseFloat(row.Amount),
		Currency:    nullStr(row.Currency),
		Description: row.Description,
		Vendor:      nullStr(row.Vendor),
		ReceiptURL:  nullStr(row.ReceiptUrl),
		ExpenseDate: row.ExpenseDate,
		Status:      nullStr(row.Status),
		CreatedAt:   nullTime(row.CreatedAt),
		UpdatedAt:   nullTime(row.UpdatedAt),
	}
	if row.CategoryID.Valid {
		e.CategoryID = &row.CategoryID.Int64
	}
	return e
}

func expenseFromGetRow(row sqlc.GetExpenseByIDRow) domain.Expense {
	e := domain.Expense{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Amount:      parseFloat(row.Amount),
		Currency:    nullStr(row.Currency),
		Description: row.Description,
		Vendor:      nullStr(row.Vendor),
		ReceiptURL:  nullStr(row.ReceiptUrl),
		ExpenseDate: row.ExpenseDate,
		Status:      nullStr(row.Status),
		CreatedAt:   nullTime(row.CreatedAt),
		UpdatedAt:   nullTime(row.UpdatedAt),
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
