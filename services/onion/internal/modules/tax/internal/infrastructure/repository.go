package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/tax/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) GetConfig(ctx context.Context, ownerID int64) (domain.TaxConfig, error) {
	var cfg domain.TaxConfig
	var taxName sql.NullString
	err := r.pool.QueryRowContext(ctx,
		`SELECT id, owner_id, tax_name, tax_rate, is_inclusive, is_active, created_at, updated_at
		 FROM finance_tax_config WHERE owner_id = ?`, ownerID,
	).Scan(&cfg.ID, &cfg.OwnerID, &taxName, &cfg.TaxRate, &cfg.IsInclusive, &cfg.IsActive, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.TaxConfig{}, domain.ErrTaxConfigNotFound
		}
		logger.Ctx(ctx).Error("Failed to get tax config", logger.Error(err))
		return domain.TaxConfig{}, err
	}
	cfg.TaxName = taxName.String
	return cfg, nil
}

func (r *Repository) UpsertConfig(ctx context.Context, input domain.TaxConfigInput) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO finance_tax_config (owner_id, tax_name, tax_rate, is_inclusive, is_active)
		 VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   tax_name = VALUES(tax_name),
		   tax_rate = VALUES(tax_rate),
		   is_inclusive = VALUES(is_inclusive),
		   is_active = VALUES(is_active)`,
		input.OwnerID, nullableString(input.TaxName), input.TaxRate, input.IsInclusive, input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert tax config", logger.Error(err))
	}
	return err
}

func (r *Repository) GetPLReport(ctx context.Context, ownerID int64) (domain.PLReport, error) {
	var report domain.PLReport

	var totalRevenue sql.NullFloat64
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total), 0) FROM finance_invoices WHERE owner_id = ? AND status = 'paid'`, ownerID,
	).Scan(&totalRevenue); err != nil {
		logger.Ctx(ctx).Error("Failed to get total revenue", logger.Error(err))
		return domain.PLReport{}, err
	}
	report.TotalRevenue = totalRevenue.Float64

	var totalExpenses sql.NullFloat64
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM finance_expenses WHERE owner_id = ?`, ownerID,
	).Scan(&totalExpenses); err != nil {
		logger.Ctx(ctx).Error("Failed to get total expenses", logger.Error(err))
		return domain.PLReport{}, err
	}
	report.TotalExpenses = totalExpenses.Float64
	report.GrossProfit = report.TotalRevenue - report.TotalExpenses
	report.NetProfit = report.GrossProfit

	return report, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
