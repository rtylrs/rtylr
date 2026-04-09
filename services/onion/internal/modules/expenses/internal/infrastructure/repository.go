package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/expenses/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Expense, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, category_id, amount, currency, description, vendor, receipt_url, expense_date, status, created_at, updated_at
		 FROM finance_expenses WHERE owner_id = ? ORDER BY expense_date DESC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expenses", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	expenses := make([]domain.Expense, 0)
	for rows.Next() {
		e, scanErr := scanExpense(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		expenses = append(expenses, e)
	}
	return expenses, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, ownerID, id int64) (domain.Expense, error) {
	row := r.pool.QueryRowContext(ctx,
		`SELECT id, owner_id, category_id, amount, currency, description, vendor, receipt_url, expense_date, status, created_at, updated_at
		 FROM finance_expenses WHERE id = ? AND owner_id = ?`, id, ownerID,
	)
	var e domain.Expense
	var catID sql.NullInt64
	var currency, desc, vendor, receipt, status sql.NullString
	if err := row.Scan(&e.ID, &e.OwnerID, &catID, &e.Amount, &currency, &desc, &vendor, &receipt, &e.ExpenseDate, &status, &e.CreatedAt, &e.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Expense{}, domain.ErrExpenseNotFound
		}
		return domain.Expense{}, err
	}
	if catID.Valid {
		e.CategoryID = &catID.Int64
	}
	e.Currency = currency.String
	e.Description = desc.String
	e.Vendor = vendor.String
	e.ReceiptURL = receipt.String
	e.Status = status.String
	return e, nil
}

func (r *Repository) Create(ctx context.Context, input domain.ExpenseInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO finance_expenses (owner_id, category_id, amount, currency, description, vendor, receipt_url, expense_date, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending')`,
		input.OwnerID, input.CategoryID, input.Amount,
		nullableString(input.Currency), nullableString(input.Description),
		nullableString(input.Vendor), nullableString(input.ReceiptURL), input.ExpenseDate,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create expense", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.ExpenseUpdateInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE finance_expenses SET category_id = ?, amount = ?, currency = ?, description = ?, vendor = ?, receipt_url = ?, expense_date = ?
		 WHERE id = ? AND owner_id = ?`,
		input.CategoryID, input.Amount,
		nullableString(input.Currency), nullableString(input.Description),
		nullableString(input.Vendor), nullableString(input.ReceiptURL), input.ExpenseDate,
		input.ID, input.OwnerID,
	)
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
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM finance_expenses WHERE id = ? AND owner_id = ?", id, ownerID,
	)
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

func (r *Repository) ListCategories(ctx context.Context, ownerID int64) ([]domain.ExpenseCategory, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, color, created_at FROM finance_expense_categories WHERE owner_id = ? ORDER BY name`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expense categories", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	cats := make([]domain.ExpenseCategory, 0)
	for rows.Next() {
		var cat domain.ExpenseCategory
		var color sql.NullString
		if scanErr := rows.Scan(&cat.ID, &cat.OwnerID, &cat.Name, &color, &cat.CreatedAt); scanErr != nil {
			return nil, scanErr
		}
		cat.Color = color.String
		cats = append(cats, cat)
	}
	return cats, rows.Err()
}

func scanExpense(rows *sql.Rows) (domain.Expense, error) {
	var e domain.Expense
	var catID sql.NullInt64
	var currency, desc, vendor, receipt, status sql.NullString
	if err := rows.Scan(&e.ID, &e.OwnerID, &catID, &e.Amount, &currency, &desc, &vendor, &receipt, &e.ExpenseDate, &status, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return domain.Expense{}, err
	}
	if catID.Valid {
		e.CategoryID = &catID.Int64
	}
	e.Currency = currency.String
	e.Description = desc.String
	e.Vendor = vendor.String
	e.ReceiptURL = receipt.String
	e.Status = status.String
	return e, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
