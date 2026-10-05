package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

func expenseFromSqlc(row sqlc.FinanceExpense) domain.Expense {
	return domain.Expense{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		CategoryID:      toInt64(row.CategoryID),
		Description:     row.Description,
		Amount:          parseFloat(row.Amount),
		Currency:        nullStr(row.Currency, "USD"),
		ExpenseDate:     row.ExpenseDate,
		Vendor:          nullStr(row.Vendor, ""),
		PaymentMethod:   nullEnumStr(row.PaymentMethod.FinanceExpensesPaymentMethod, row.PaymentMethod.Valid, "cash"),
		ReceiptURL:      nullStr(row.ReceiptUrl, ""),
		IsTaxDeductible: nullBool(row.IsTaxDeductible),
		TaxAmount:       parseFloatNull(row.TaxAmount),
		Notes:           nullStr(row.Notes, ""),
		Status:          nullStr(row.Status, "pending"),
		CreatedAt:       nullTime(row.CreatedAt),
		UpdatedAt:       nullTime(row.UpdatedAt),
	}
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func sqlNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func sqlNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: i != 0}
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// paymentAmountToCents rounds an API-supplied payment amount to whole cents.
func paymentAmountToCents(amount float64) money.Amount {
	return money.Amount(int64(math.Round(amount * 100)))
}

// nullDecimalToCents parses a DECIMAL column read back as a nullable string
// into integer cents via the shared money parser, so no float64 rounding
// error is introduced on the way out of the database. Unparseable values
// degrade to zero with a warning rather than failing the payment write.
func nullDecimalToCents(value sql.NullString) money.Amount {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return 0
	}
	cents, err := money.FromString(value.String)
	if err != nil {
		logger.Global().Warn("failed to parse invoice decimal value",
			logger.String("value", value.String),
			logger.Error(err),
		)
		return 0
	}
	return cents
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse float from database",
			logger.String("value", s),
			logger.Error(err),
		)

		return 0
	}

	return v
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid || s.String == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		logger.Global().Warn("failed to parse nullable float from database",
			logger.String("value", s.String),
			logger.Error(err),
		)

		return 0
	}

	return v
}

func nullStr(s sql.NullString, def string) string {
	if s.Valid {
		return s.String
	}

	return def
}

func nullBool(b sql.NullBool) bool {
	if b.Valid {
		return b.Bool
	}

	return false
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

func toInt32(i sql.NullInt32) int32 {
	if i.Valid {
		return i.Int32
	}

	return 0
}

func nullEnumStr[T ~string](val T, valid bool, def string) string {
	if valid {
		return string(val)
	}

	return def
}

func sqlNullTimeStr(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: t, Valid: true}
}

func toFloat64FromInterface(v interface{}) float64 {
	if v == nil {
		return 0
	}

	var str string

	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case []byte:
		str = string(val)
	case string:
		str = val
	default:
		str = fmt.Sprintf("%v", val)
	}

	if str == "" {
		return 0
	}

	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		logger.Global().Warn("failed to parse float from interface value",
			logger.String("value", str),
			logger.Error(err),
		)

		return 0
	}

	return f
}

func nullableInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	out := v.Int64
	return &out
}

func nullableInt64FromPtr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

// ListInvoicesByCustomer pages one customer's invoices off the
// business+customer index.
func (r *Repository) ListInvoicesByCustomer(ctx context.Context, businessID, customerID int64, limit, offset int) ([]domain.Invoice, error) {
	rows, err := r.queries.ListInvoicesByCustomer(ctx, sqlc.ListInvoicesByCustomerParams{
		BusinessID: businessID,
		CustomerID: sql.NullInt64{Int64: customerID, Valid: true},
		Limit:      int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list invoices by customer", logger.Any("business_id", businessID), logger.Any("customer_id", customerID), logger.Error(err))
		return nil, err
	}
	out := make([]domain.Invoice, 0, len(rows))
	for _, row := range rows {
		out = append(out, invoiceFromSqlc(row))
	}
	return out, nil
}

// CountInvoicesByCustomer reports the customer's invoice count.
func (r *Repository) CountInvoicesByCustomer(ctx context.Context, businessID, customerID int64) (int64, error) {
	return r.queries.CountInvoicesByCustomer(ctx, sqlc.CountInvoicesByCustomerParams{
		BusinessID: businessID,
		CustomerID: sql.NullInt64{Int64: customerID, Valid: true},
	})
}
