package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/tax/internal/domain"
	"github.com/voxire/onion/internal/modules/tax/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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

func (r *Repository) GetConfig(ctx context.Context, businessID int64) (domain.TaxConfig, error) {
	row, err := r.queries.GetTaxConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.TaxConfig{}, domain.ErrTaxConfigNotFound
		}
		logger.Ctx(ctx).Error("Failed to get tax config", logger.Error(err))
		return domain.TaxConfig{}, err
	}

	return domain.TaxConfig{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		TaxName:     row.Name,
		TaxRate:     parseStringFloat(row.Rate),
		IsInclusive: nullBool(row.IsInclusive),
		IsActive:    nullBool(row.IsActive),
		CreatedAt:   nullTime(row.CreatedAt),
	}, nil
}

func (r *Repository) UpsertConfig(ctx context.Context, input domain.TaxConfigInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := r.queries.WithTx(tx)
	if err = queries.ClearDefaultTaxConfig(ctx, input.BusinessID); err != nil {
		return err
	}
	err = queries.UpsertTaxConfig(ctx, sqlc.UpsertTaxConfigParams{
		OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.TaxName,
		Rate: fmtFloat(input.TaxRate), Type: sqlc.NullFinanceTaxConfigType{FinanceTaxConfigType: sqlc.FinanceTaxConfigType("vat"), Valid: true},
		IsActive: sql.NullBool{Bool: input.IsActive, Valid: true}, IsInclusive: sql.NullBool{Bool: input.IsInclusive, Valid: true},
	})
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert tax config", logger.Error(err))
	}
	return err
}

func (r *Repository) GetPLReport(ctx context.Context, businessID int64) (domain.PLReport, error) {
	var report domain.PLReport

	totalRevenue, err := r.queries.SumPaidInvoicesTotal(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get total revenue", logger.Error(err))
		return domain.PLReport{}, err
	}
	report.TotalRevenue = toFloat64FromInterface(totalRevenue)

	totalExpenses, err := r.queries.SumAllExpensesAmount(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get total expenses", logger.Error(err))
		return domain.PLReport{}, err
	}
	report.TotalExpenses = toFloat64FromInterface(totalExpenses)
	report.GrossProfit = report.TotalRevenue - report.TotalExpenses
	report.NetProfit = report.GrossProfit

	return report, nil
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

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

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func parseStringFloat(value string) float64 {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return v
}

func parseTaxDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse tax decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

func toFloat64FromInterface(v any) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case []byte:
		return parseTaxDecimal(string(val))
	case string:
		return parseTaxDecimal(val)
	default:
		return parseTaxDecimal(fmt.Sprintf("%v", val))
	}
}
