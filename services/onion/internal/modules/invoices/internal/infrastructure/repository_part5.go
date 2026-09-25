package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/invoices/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) SumExpenses(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	dfTime, err := time.Parse("2006-01-02", dateFrom)
	if err != nil && dateFrom != "" {
		logger.Ctx(ctx).Error("Failed to parse dateFrom for SumExpenses", logger.Error(err))
		return 0, err
	}
	dtTime, err := time.Parse("2006-01-02", dateTo)
	if err != nil && dateTo != "" {
		logger.Ctx(ctx).Error("Failed to parse dateTo for SumExpenses", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.SumExpenses(ctx, sqlc.SumExpensesParams{
		BusinessID:      businessID,
		FromExpenseDate: dfTime,
		ToExpenseDate:   dtTime,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum expenses", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumTotalReceivable(ctx context.Context, businessID int64) (float64, error) {
	result, err := r.queries.SumTotalReceivable(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum total receivable", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumOverdueAmount(ctx context.Context, businessID int64) (float64, error) {
	result, err := r.queries.SumOverdueAmount(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum overdue amount", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumTaxCollected(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	result, err := r.queries.SumTaxCollected(ctx, sqlc.SumTaxCollectedParams{
		BusinessID:   businessID,
		FromPaidDate: sqlNullTimeStr(dateFrom),
		ToPaidDate:   sqlNullTimeStr(dateTo),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum tax collected", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumTaxOwed(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	dfTime, err := time.Parse("2006-01-02", dateFrom)
	if err != nil && dateFrom != "" {
		logger.Ctx(ctx).Error("Failed to parse dateFrom for SumTaxOwed", logger.Error(err))
		return 0, err
	}
	dtTime, err := time.Parse("2006-01-02", dateTo)
	if err != nil && dateTo != "" {
		logger.Ctx(ctx).Error("Failed to parse dateTo for SumTaxOwed", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.SumTaxOwed(ctx, sqlc.SumTaxOwedParams{
		BusinessID:      businessID,
		FromExpenseDate: dfTime,
		ToExpenseDate:   dtTime,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum tax owed", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

// CountInvoicesByStatus returns how many invoices the owner has in each
// status bucket. The returned map uses status strings as keys (e.g. "draft",
// "paid"). An empty map is returned when the owner has no invoices.
func (r *Repository) CountInvoicesByStatus(ctx context.Context, businessID int64) (map[string]int64, error) {
	rows, err := r.queries.CountInvoicesByStatus(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count invoices by status", logger.Error(err))

		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[string(row.Status.FinanceInvoicesStatus)] = row.Cnt
	}

	return counts, nil
}

// GetOverdueInvoices returns one page of unpaid invoices whose due_date is
// before the given cutoff (expressed as YYYY-MM-DD). Pass an empty string to
// use today's date.
func (r *Repository) GetOverdueInvoices(ctx context.Context, businessID int64, beforeDate string, limit, offset int32) ([]domain.Invoice, error) {
	before, err := parseOverdueCutoff(ctx, beforeDate)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListOverdueInvoices(ctx, sqlc.ListOverdueInvoicesParams{
		BusinessID: businessID,
		DueDate:    before,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list overdue invoices", logger.Error(err))

		return nil, err
	}

	invoices := make([]domain.Invoice, 0, len(rows))
	for _, row := range rows {
		invoices = append(invoices, invoiceFromSqlc(row))
	}

	return invoices, nil
}

// CountOverdueInvoices returns the total number of unpaid invoices whose
// due_date is before the given cutoff (expressed as YYYY-MM-DD). Pass an
// empty string to use today's date.
func (r *Repository) CountOverdueInvoices(ctx context.Context, businessID int64, beforeDate string) (int64, error) {
	before, err := parseOverdueCutoff(ctx, beforeDate)
	if err != nil {
		return 0, err
	}

	count, err := r.queries.CountOverdueInvoices(ctx, sqlc.CountOverdueInvoicesParams{
		BusinessID: businessID,
		DueDate:    before,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count overdue invoices", logger.Error(err))

		return 0, err
	}

	return count, nil
}

func parseOverdueCutoff(ctx context.Context, beforeDate string) (time.Time, error) {
	if beforeDate == "" {
		beforeDate = time.Now().Format("2006-01-02")
	}

	before, err := time.Parse("2006-01-02", beforeDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse beforeDate for overdue invoices", logger.Error(err))

		return time.Time{}, err
	}

	return before, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func invoiceFromSqlc(row sqlc.FinanceInvoice) domain.Invoice {
	inv := domain.Invoice{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		BusinessID:     row.BusinessID,
		InvoiceNumber:  row.InvoiceNumber,
		ClientName:     row.ClientName,
		ClientEmail:    nullStr(row.ClientEmail, ""),
		ClientAddress:  nullStr(row.ClientAddress, ""),
		ClientTaxID:    nullStr(row.ClientTaxID, ""),
		CustomerID:     nullableInt64Ptr(row.CustomerID),
		Status:         string(row.Status.FinanceInvoicesStatus),
		IssueDate:      row.IssueDate,
		DueDate:        row.DueDate,
		Currency:       nullStr(row.Currency, "USD"),
		Subtotal:       parseFloatNull(row.Subtotal),
		DiscountType:   nullEnumStr(row.DiscountType.FinanceInvoicesDiscountType, row.DiscountType.Valid, "fixed"),
		DiscountValue:  parseFloatNull(row.DiscountValue),
		DiscountAmount: parseFloatNull(row.DiscountAmount),
		TaxRate:        parseFloatNull(row.TaxRate),
		TaxAmount:      parseFloatNull(row.TaxAmount),
		Total:          parseFloatNull(row.Total),
		AmountPaid:     parseFloatNull(row.AmountPaid),
		AmountDue:      parseFloatNull(row.AmountDue),
		Notes:          nullStr(row.Notes, ""),
		Terms:          nullStr(row.Terms, ""),
		CreatedAt:      nullTime(row.CreatedAt),
		UpdatedAt:      nullTime(row.UpdatedAt),
	}

	if row.PaidDate.Valid {
		t := row.PaidDate.Time
		inv.PaidDate = &t
	}

	return inv
}
