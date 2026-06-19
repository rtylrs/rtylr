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
		TaxName:     nullStr(row.TaxName),
		TaxRate:     parseFloatNull(row.TaxRate),
		IsInclusive: nullBool(row.IsInclusive),
		IsActive:    nullBool(row.IsActive),
		CreatedAt:   nullTime(row.CreatedAt),
		UpdatedAt:   nullTime(row.UpdatedAt),
	}, nil
}

func (r *Repository) UpsertConfig(ctx context.Context, input domain.TaxConfigInput) error {
	err := r.queries.UpsertTaxConfig(ctx, sqlc.UpsertTaxConfigParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		TaxName:     nullableString(input.TaxName),
		TaxRate:     nullableString(fmtFloat(input.TaxRate)),
		IsInclusive: sql.NullBool{Bool: input.IsInclusive, Valid: true},
		IsActive:    sql.NullBool{Bool: input.IsActive, Valid: true},
	})
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

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
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

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid {
		return 0
	}
	return parseTaxDecimal(s.String)
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
