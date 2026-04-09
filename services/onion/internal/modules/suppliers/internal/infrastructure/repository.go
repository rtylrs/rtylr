package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) ListSuppliers(ctx context.Context, ownerID int64) ([]domain.Supplier, error) {
	query := `
		SELECT id, owner_id, name, email, phone, address, city, country,
		       tax_id, payment_terms, notes, is_active, created_at, updated_at
		FROM erp_suppliers
		WHERE owner_id = ?
		ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list suppliers", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var suppliers []domain.Supplier
	for rows.Next() {
		var s domain.Supplier
		if err := s.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan supplier row", logger.Error(err))
			return nil, err
		}
		suppliers = append(suppliers, s)
	}

	return suppliers, rows.Err()
}

func (r *Repository) CreateSupplier(ctx context.Context, ownerID int64, input domain.SupplierInput) (int64, error) {
	query := `
		INSERT INTO erp_suppliers
		(owner_id, name, email, phone, address, city, country, tax_id, payment_terms, notes, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query,
		ownerID,
		input.Name,
		nullString(input.Email),
		nullString(input.Phone),
		nullString(input.Address),
		nullString(input.City),
		nullString(input.Country),
		nullString(input.TaxID),
		nullString(input.PaymentTerms),
		nullString(input.Notes),
		input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create supplier", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) UpdateSupplier(ctx context.Context, ownerID, id int64, input domain.SupplierInput) error {
	query := `
		UPDATE erp_suppliers
		SET name = ?, email = ?, phone = ?, address = ?, city = ?, country = ?,
		    tax_id = ?, payment_terms = ?, notes = ?, is_active = ?
		WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		input.Name,
		nullString(input.Email),
		nullString(input.Phone),
		nullString(input.Address),
		nullString(input.City),
		nullString(input.Country),
		nullString(input.TaxID),
		nullString(input.PaymentTerms),
		nullString(input.Notes),
		input.IsActive,
		ownerID,
		id,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update supplier", logger.Error(err))
		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
