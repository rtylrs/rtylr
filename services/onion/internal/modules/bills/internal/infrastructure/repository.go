package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	"github.com/voxire/onion/internal/modules/bills/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

type Repository struct {
	pool    *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}


// BillExistsInBusiness reports whether the row exists within the business. The write queries
// in this module are :exec and cannot report a zero-row match, so writes
// pre-check existence to avoid silent success on wrong/foreign ids.
func (r *Repository) BillExistsInBusiness(ctx context.Context, businessID, id int64) bool {
	var exists bool
	if err := r.pool.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM finance_bills WHERE id = ? AND business_id = ?)",
		id, businessID,
	).Scan(&exists); err != nil {
		return false
	}
	return exists
}
func (r *Repository) Get(ctx context.Context, businessID, id int64) (domain.Bill, error) {
	row, err := r.queries.GetBill(ctx, sqlc.GetBillParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Bill{}, domain.ErrBillNotFound
		}
		logger.Ctx(ctx).Error("Failed to get bill", logger.Error(err))
		return domain.Bill{}, err
	}
	return billFromSqlc(row), nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	res, err := r.queries.DeleteBill(ctx, sqlc.DeleteBillParams{ID: id, BusinessID: businessID})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete bill", logger.Error(err))
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read delete result", logger.Error(err))
		return err
	}
	if rows == 0 {
		return domain.ErrBillNotFound
	}
	return nil
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Bill, error) {
	rows, err := r.queries.ListBills(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list bills", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Bill, 0, len(rows))
	for _, row := range rows {
		items = append(items, billFromSqlc(row))
	}

	return items, nil
}

// ListPaginated returns a single page of bills for the business, ordered by
// due date. Limit and offset are applied as-is; callers clamp them.
func (r *Repository) ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Bill, error) {
	rows, err := r.queries.ListBillsPaginated(ctx, sqlc.ListBillsPaginatedParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list bills page", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Bill, 0, len(rows))
	for _, row := range rows {
		items = append(items, billFromSqlc(row))
	}

	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.BillInput) (int64, error) {
	status := input.Status
	if status == "" {
		status = "pending"
	}
	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	dueDate, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse due_date", logger.Error(err))
		return 0, err
	}

	res, err := r.queries.CreateBill(ctx, sqlc.CreateBillParams{
		OwnerID:      businessID,
		BusinessID:   businessID,
		SupplierID:   sql.NullInt64{Int64: input.SupplierID, Valid: input.SupplierID != 0},
		SupplierName: input.SupplierName,
		Amount:      formatAmount(input.Amount),
		Currency:    nullableString(currency),
		DueDate:     dueDate,
		Status:      sqlc.NullFinanceBillsStatus{FinanceBillsStatus: sqlc.FinanceBillsStatus(status), Valid: true},
		Description: nullableString(input.Description),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create bill", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.BillInput) error {
	dueDate, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil && input.DueDate != "" {
		logger.Ctx(ctx).Error("Failed to parse due_date", logger.Error(err))
		return err
	}

	err = r.queries.UpdateBill(ctx, sqlc.UpdateBillParams{
		SupplierID:   sql.NullInt64{Int64: input.SupplierID, Valid: input.SupplierID != 0},
		SupplierName: input.SupplierName,
		Amount:      formatAmount(input.Amount),
		Currency:    nullableString(input.Currency),
		DueDate:     dueDate,
		Status:      sqlc.NullFinanceBillsStatus{FinanceBillsStatus: sqlc.FinanceBillsStatus(input.Status), Valid: input.Status != ""},
		Description: nullableString(input.Description),
		ID:          id,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update bill", logger.Error(err))
		return err
	}

	return nil
}

// SumByStatus returns the total bill amount for a given status and owner.
// It returns 0 when the owner has no bills with that status.
func (r *Repository) SumByStatus(ctx context.Context, businessID int64, status string) (money.Amount, error) {
	total, err := r.queries.SumBillsByStatus(ctx, sqlc.SumBillsByStatusParams{
		BusinessID: businessID,
		Status:     sqlc.NullFinanceBillsStatus{FinanceBillsStatus: sqlc.FinanceBillsStatus(status), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum bills by status", logger.Error(err))
		return 0, err
	}

	return parseFloatAny(total), nil
}

// CountByStatus returns the number of bills in each status for the owner.
func (r *Repository) CountByStatus(ctx context.Context, businessID int64) (map[string]int64, error) {
	rows, err := r.queries.CountBillsByStatus(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count bills by status", logger.Error(err))
		return nil, err
	}

	counts := make(map[string]int64)
	for _, row := range rows {
		counts[string(row.Status.FinanceBillsStatus)] = row.Cnt
	}

	return counts, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func billFromSqlc(row sqlc.FinanceBill) domain.Bill {
	return domain.Bill{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		SupplierID:  toInt64(row.SupplierID),
		Amount:      parseFloat(row.Amount),
		Currency:    nullStr(row.Currency),
		DueDate:     row.DueDate,
		Status:      string(row.Status.FinanceBillsStatus),
		Description: nullStr(row.Description),
		CreatedAt:   nullTime(row.CreatedAt),
		UpdatedAt:   nullTime(row.UpdatedAt),
	}
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
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

func toInt64(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}
	return 0
}

func formatAmount(value money.Amount) string {
	return value.String()
}

func parseFloat(s string) money.Amount {
	if s == "" {
		return 0
	}
	v, err := money.FromString(s)
	if err != nil {
		logger.Global().Warn("failed to parse decimal amount from database", logger.String("value", s), logger.Error(err))
		return 0
	}
	return v
}

func parseFloatAny(v any) money.Amount {
	switch val := v.(type) {
	case nil:
		return 0
	case float64:
		parsed, err := money.FromFloat64(val)
		if err != nil {
			return 0
		}
		return parsed
	case float32:
		parsed, err := money.FromFloat64(float64(val))
		if err != nil {
			return 0
		}
		return parsed
	case int64:
		return parseFloat(strconv.FormatInt(val, 10))
	case []byte:
		return parseFloat(string(val))
	case string:
		return parseFloat(val)
	default:
		return parseFloat(fmt.Sprintf("%v", val))
	}
}
