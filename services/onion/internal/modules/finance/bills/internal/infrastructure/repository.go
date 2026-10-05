package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/finance/bills/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/bills/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

type Repository struct {
	purchaseOrderExists func(context.Context, int64, int64) (bool, error)
	pool                *sql.DB
	queries             *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
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
	bill, err := r.Get(ctx, businessID, id)
	if err != nil {
		return err
	}
	if bill.IsPaid() {
		return domain.ErrPaidBillImmutable
	}

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
	if err := r.validateBillReferences(ctx, businessID, input.PurchaseOrderID); err != nil {
		return 0, err
	}
	status := input.Status
	if status == "" {
		status = "pending"
	} else if status != "pending" {
		return 0, domain.ErrInvalidStatusTransition
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
		OwnerID:         businessID,
		BusinessID:      businessID,
		SupplierID:      sql.NullInt64{Int64: input.SupplierID, Valid: input.SupplierID != 0},
		SupplierName:    input.SupplierName,
		PurchaseOrderID: nullableInt64FromInput(input.PurchaseOrderID),
		Amount:          formatAmount(input.Amount),
		Currency:        nullableString(currency),
		DueDate:         dueDate,
		Status:          sqlc.NullFinanceBillsStatus{FinanceBillsStatus: sqlc.FinanceBillsStatus(status), Valid: true},
		Description:     nullableString(input.Description),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create bill", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.BillInput) error {
	current, err := r.Get(ctx, businessID, id)
	if err != nil {
		return err
	}
	if current.IsPaid() {
		return domain.ErrPaidBillImmutable
	}
	nextStatus := input.Status
	if nextStatus == "" {
		nextStatus = current.Status
	}
	if err := domain.ValidateBillStatusTransition(current.Status, nextStatus); err != nil {
		return err
	}
	input.Status = nextStatus
	// The update body is partial: fields the caller omitted keep the stored
	// value instead of clearing the vendor snapshot or the order link.
	mergeBillInput(&input, current)
	if err := r.validateBillReferences(ctx, businessID, input.PurchaseOrderID); err != nil {
		return err
	}

	dueDate, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil && input.DueDate != "" {
		logger.Ctx(ctx).Error("Failed to parse due_date", logger.Error(err))
		return err
	}

	res, err := r.queries.UpdateBill(ctx, sqlc.UpdateBillParams{
		SupplierID:      sql.NullInt64{Int64: input.SupplierID, Valid: input.SupplierID != 0},
		SupplierName:    input.SupplierName,
		PurchaseOrderID: nullableInt64FromInput(input.PurchaseOrderID),
		Amount:          formatAmount(input.Amount),
		Currency:        nullableString(input.Currency),
		DueDate:         dueDate,
		Status:          sqlc.NullFinanceBillsStatus{FinanceBillsStatus: sqlc.FinanceBillsStatus(input.Status), Valid: input.Status != ""},
		Description:     nullableString(input.Description),
		CurrentStatus:   sqlc.NullFinanceBillsStatus{FinanceBillsStatus: sqlc.FinanceBillsStatus(current.Status), Valid: true},
		ID:              id,
		BusinessID:      businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update bill", logger.Error(err))
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		// MySQL reports changed rows, not matched rows: storing identical
		// values matches the row but affects none. The pre-read above already
		// proved the bill exists for this business, so only report a missing
		// bill when it has actually disappeared since.
		if _, getErr := r.Get(ctx, businessID, id); getErr != nil {
			return getErr
		}
	}

	return nil
}

// mergeBillInput keeps stored values when the caller omitted them from the
// partial update body.
func mergeBillInput(input *domain.BillInput, current domain.Bill) {
	if input.SupplierID == 0 {
		input.SupplierID = current.SupplierID
	}
	if input.SupplierName == "" {
		input.SupplierName = current.SupplierName
	}
	if input.PurchaseOrderID == nil {
		input.PurchaseOrderID = current.PurchaseOrderID
	}
	if input.Amount == 0 {
		input.Amount = current.Amount
	}
	if input.Currency == "" {
		input.Currency = current.Currency
	}
	if input.DueDate == "" {
		input.DueDate = current.DueDate.Format("2006-01-02")
	}
	if input.Description == "" {
		input.Description = current.Description
	}
}

func (r *Repository) validateBillReferences(ctx context.Context, businessID int64, purchaseOrderID *int64) error {
	if purchaseOrderID == nil {
		return nil
	}
	if *purchaseOrderID <= 0 {
		return domain.ErrBillReferenceNotFound
	}
	if r.purchaseOrderExists == nil {
		return errors.New("finance purchase order reader is not configured")
	}
	present, err := r.purchaseOrderExists(ctx, businessID, *purchaseOrderID)
	if err != nil {
		return err
	}
	if !present {
		return domain.ErrBillReferenceNotFound
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
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		SupplierID:      toInt64(row.SupplierID),
		SupplierName:    row.SupplierName,
		PurchaseOrderID: nullableInt64(row.PurchaseOrderID),
		Amount:          parseFloat(row.Amount),
		Currency:        nullStr(row.Currency),
		DueDate:         row.DueDate,
		Status:          string(row.Status.FinanceBillsStatus),
		Description:     nullStr(row.Description),
		CreatedAt:       nullTime(row.CreatedAt),
		UpdatedAt:       nullTime(row.UpdatedAt),
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

func nullableInt64(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	out := v.Int64
	return &out
}

func nullableInt64FromInput(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

// ListByPurchaseOrder pages every bill referencing one ERP order.
func (r *Repository) ListByPurchaseOrder(ctx context.Context, businessID, purchaseOrderID int64) ([]domain.Bill, error) {
	rows, err := r.queries.ListBillsByPurchaseOrder(ctx, sqlc.ListBillsByPurchaseOrderParams{
		BusinessID:      businessID,
		PurchaseOrderID: sql.NullInt64{Int64: purchaseOrderID, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list bills by purchase order", logger.Any("business_id", businessID), logger.Any("purchase_order_id", purchaseOrderID), logger.Error(err))
		return nil, err
	}
	out := make([]domain.Bill, 0, len(rows))
	for _, row := range rows {
		out = append(out, billFromSqlc(row))
	}
	return out, nil
}

// CountByPurchaseOrder reports how many bills reference one ERP order.
func (r *Repository) CountByPurchaseOrder(ctx context.Context, businessID, purchaseOrderID int64) (int64, error) {
	return r.queries.CountBillsByPurchaseOrder(ctx, sqlc.CountBillsByPurchaseOrderParams{
		BusinessID:      businessID,
		PurchaseOrderID: sql.NullInt64{Int64: purchaseOrderID, Valid: true},
	})
}

func (r *Repository) WithPurchaseOrderReader(read func(context.Context, int64, int64) (bool, error)) *Repository {
	r.purchaseOrderExists = read
	return r
}
