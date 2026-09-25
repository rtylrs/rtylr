package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
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
		ID:                 row.ID,
		OwnerID:            row.OwnerID,
		BusinessID:         row.BusinessID,
		TaxName:            row.Name,
		TaxRateBasisPoints: parseRateBasisPoints(row.Rate),
		IsInclusive:        nullBool(row.IsInclusive),
		IsActive:           nullBool(row.IsActive),
		CreatedAt:          nullTime(row.CreatedAt),
	}, nil
}

func (r *Repository) UpsertConfig(ctx context.Context, input domain.TaxConfigInput) error {
	businessID := input.BusinessID
	if businessID == 0 {
		businessID = input.OwnerID
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := r.queries.WithTx(tx)
	if err = queries.ClearDefaultTaxConfig(ctx, businessID); err != nil {
		return err
	}
	err = queries.UpsertTaxConfig(ctx, sqlc.UpsertTaxConfigParams{
		OwnerID: input.OwnerID, BusinessID: businessID, Name: input.TaxName,
		Rate: formatRateBasisPoints(input.TaxRateBasisPoints), Type: sqlc.NullFinanceTaxConfigType{FinanceTaxConfigType: sqlc.FinanceTaxConfigType("vat"), Valid: true},
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
	report.TotalRevenueMinorUnits = minorUnitsFromInterface(totalRevenue)

	totalExpenses, err := r.queries.SumAllExpensesAmount(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get total expenses", logger.Error(err))
		return domain.PLReport{}, err
	}
	report.TotalExpensesMinorUnits = minorUnitsFromInterface(totalExpenses)
	report.GrossProfitMinorUnits = report.TotalRevenueMinorUnits - report.TotalExpensesMinorUnits
	report.NetProfitMinorUnits = report.GrossProfitMinorUnits

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

func parseRateBasisPoints(value string) int64 {
	rate, err := parseMinorUnits(value)
	if err != nil {
		logger.Global().Warn("failed to parse tax rate from database",
			logger.String("value", value), logger.Error(err))
		return 0
	}
	return rate
}

func formatRateBasisPoints(value int64) string {
	if value < 0 {
		return "-" + formatRateBasisPoints(-value)
	}
	return strconv.FormatInt(value/100, 10) + "." + fmt.Sprintf("%02d", value%100)
}

// minorUnitsFromInterface converts a DECIMAL value returned by the SQL driver
// into integer minor units without passing the value through float64.
func minorUnitsFromInterface(v any) int64 {
	if v == nil {
		return 0
	}
	var value string
	switch val := v.(type) {
	case []byte:
		value = string(val)
	case string:
		value = val
	case float64:
		// This is only a compatibility path for drivers that expose DECIMAL as
		// a float; MySQL's normal path is []byte/string and stays exact.
		value = strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		value = strconv.FormatFloat(float64(val), 'f', -1, 32)
	case int:
		value = strconv.FormatInt(int64(val), 10)
	case int32:
		value = strconv.FormatInt(int64(val), 10)
	case int64:
		value = strconv.FormatInt(val, 10)
	case uint:
		value = strconv.FormatUint(uint64(val), 10)
	case uint32:
		value = strconv.FormatUint(uint64(val), 10)
	case uint64:
		value = strconv.FormatUint(val, 10)
	default:
		value = fmt.Sprintf("%v", val)
	}

	minorUnits, err := parseMinorUnits(value)
	if err != nil {
		logger.Global().Warn("failed to parse monetary amount from database",
			logger.String("value", value), logger.Error(err))
		return 0
	}
	return minorUnits
}

// parseMinorUnits parses a base-10 amount with up to two fractional digits
// directly into cents. It intentionally avoids ParseFloat so values such as
// 0.10 remain exactly 10 minor units.
func parseMinorUnits(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	negative := false
	switch value[0] {
	case '-':
		negative = true
		value = value[1:]
	case '+':
		value = value[1:]
	}
	if value == "" {
		return 0, fmt.Errorf("amount has no digits")
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("amount has multiple decimal points")
	}
	wholePart := parts[0]
	if wholePart == "" {
		wholePart = "0"
	}
	whole, err := strconv.ParseUint(wholePart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid whole amount %q: %w", wholePart, err)
	}

	fractionPart := ""
	if len(parts) == 2 {
		fractionPart = parts[1]
	}
	if len(fractionPart) > 2 {
		return 0, fmt.Errorf("amount has more than two fractional digits")
	}
	for len(fractionPart) < 2 {
		fractionPart += "0"
	}
	fraction := uint64(0)
	if fractionPart != "" {
		fraction, err = strconv.ParseUint(fractionPart, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid fractional amount %q: %w", fractionPart, err)
		}
	}

	if whole > math.MaxUint64/100 {
		return 0, fmt.Errorf("amount overflows minor units")
	}
	minorMagnitude := whole*100 + fraction
	if minorMagnitude < whole*100 {
		return 0, fmt.Errorf("amount overflows minor units")
	}
	limit := uint64(math.MaxInt64)
	if negative {
		limit++
	}
	if minorMagnitude > limit {
		return 0, fmt.Errorf("amount overflows int64")
	}
	if negative {
		return -int64(minorMagnitude), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}
	return int64(minorMagnitude), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}
