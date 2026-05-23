package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Expense, error) {
	rows, err := r.queries.ListExpenses(ctx, ownerID)
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

func (r *Repository) GetByID(ctx context.Context, ownerID, id int64) (domain.Expense, error) {
	row, err := r.queries.GetExpenseByID(ctx, sqlc.GetExpenseByIDParams{
		ID:      id,
		OwnerID: ownerID,
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
		OwnerID:     input.OwnerID,
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

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	result, err := r.queries.DeleteExpense(ctx, sqlc.DeleteExpenseParams{
		ID:      id,
		OwnerID: ownerID,
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

func (r *Repository) UpdateStatus(ctx context.Context, ownerID, id int64, status string) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE expenses SET status = ? WHERE id = ? AND owner_id = ?`,
		status, id, ownerID,
	)
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

func (r *Repository) ListCategories(ctx context.Context, ownerID int64) ([]domain.ExpenseCategory, error) {
	rows, err := r.queries.ListExpenseCategories(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expense categories", logger.Error(err))
		return nil, err
	}

	cats := make([]domain.ExpenseCategory, 0, len(rows))
	for _, row := range rows {
		cats = append(cats, domain.ExpenseCategory{
			ID:        row.ID,
			OwnerID:   row.OwnerID,
			Name:      row.Name,
			Color:     nullStr(row.Color),
			CreatedAt: nullTime(row.CreatedAt),
		})
	}
	return cats, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func expenseFromRow(row sqlc.ListExpensesRow) domain.Expense {
	e := domain.Expense{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
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
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// Ensure these are used.
var _ = fmt.Sprintf
