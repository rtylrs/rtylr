package infrastructure

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	"github.com/voxire/onion/internal/modules/bills/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{queries: sqlc.New(dbclient.Pool())}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Bill, error) {
	rows, err := r.queries.ListBills(ctx, ownerID)
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

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.BillInput) (int64, error) {
	status := input.Status
	if status == "" {
		status = "pending"
	}
	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	dueDate, _ := time.Parse("2006-01-02", input.DueDate)

	res, err := r.queries.CreateBill(ctx, sqlc.CreateBillParams{
		OwnerID:     ownerID,
		SupplierID:  sql.NullInt64{Int64: input.SupplierID, Valid: input.SupplierID != 0},
		Amount:      fmtFloat(input.Amount),
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

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.BillInput) error {
	dueDate, _ := time.Parse("2006-01-02", input.DueDate)

	err := r.queries.UpdateBill(ctx, sqlc.UpdateBillParams{
		SupplierID:  sql.NullInt64{Int64: input.SupplierID, Valid: input.SupplierID != 0},
		Amount:      fmtFloat(input.Amount),
		Currency:    nullableString(input.Currency),
		DueDate:     dueDate,
		Status:      sqlc.NullFinanceBillsStatus{FinanceBillsStatus: sqlc.FinanceBillsStatus(input.Status), Valid: input.Status != ""},
		Description: nullableString(input.Description),
		ID:          id,
		OwnerID:     ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update bill", logger.Error(err))
		return err
	}

	return nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func billFromSqlc(row sqlc.FinanceBill) domain.Bill {
	return domain.Bill{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
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

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
