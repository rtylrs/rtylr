package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	"github.com/voxire/onion/internal/modules/suppliers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListSuppliers(ctx context.Context, ownerID int64, limit, offset int) ([]domain.Supplier, bool, error) {
	const q = `
		SELECT id, owner_id, name, email, phone, address, city, country,
		       tax_id, payment_terms, notes, is_active, created_at, updated_at
		FROM erp_suppliers
		WHERE owner_id = ?
		ORDER BY name ASC
		LIMIT ? OFFSET ?`

	rows, err := r.pool.QueryContext(ctx, q, ownerID, limit+1, offset)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list suppliers", logger.Error(err))
		return nil, false, err
	}
	defer rows.Close()

	var suppliers []domain.Supplier
	for rows.Next() {
		var s domain.Supplier
		var email, phone, address, city, country, taxID, paymentTerms, notes sql.NullString
		var isActive sql.NullBool
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(
			&s.ID, &s.OwnerID, &s.Name,
			&email, &phone, &address, &city, &country,
			&taxID, &paymentTerms, &notes, &isActive,
			&createdAt, &updatedAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan supplier row", logger.Error(err))
			return nil, false, err
		}

		if email.Valid {
			s.Email = email.String
		}
		if phone.Valid {
			s.Phone = phone.String
		}
		if address.Valid {
			s.Address = address.String
		}
		if city.Valid {
			s.City = city.String
		}
		if country.Valid {
			s.Country = country.String
		}
		if taxID.Valid {
			s.TaxID = taxID.String
		}
		if paymentTerms.Valid {
			s.PaymentTerms = paymentTerms.String
		}
		if notes.Valid {
			s.Notes = notes.String
		}
		if isActive.Valid {
			s.IsActive = isActive.Bool
		}
		if createdAt.Valid {
			s.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			s.UpdatedAt = updatedAt.Time
		}

		suppliers = append(suppliers, s)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(suppliers) > limit
	if hasMore {
		suppliers = suppliers[:limit]
	}

	return suppliers, hasMore, nil
}

func (r *Repository) CreateSupplier(ctx context.Context, ownerID int64, input domain.SupplierInput) (int64, error) {
	result, err := r.queries.CreateSupplier(ctx, sqlc.CreateSupplierParams{
		OwnerID:      ownerID,
		Name:         input.Name,
		Email:        nullString(input.Email),
		Phone:        nullString(input.Phone),
		Address:      nullString(input.Address),
		City:         nullString(input.City),
		Country:      nullString(input.Country),
		TaxID:        nullString(input.TaxID),
		PaymentTerms: nullString(input.PaymentTerms),
		Notes:        nullString(input.Notes),
		IsActive:     sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create supplier", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateSupplier(ctx context.Context, ownerID, id int64, input domain.SupplierInput) error {
	err := r.queries.UpdateSupplier(ctx, sqlc.UpdateSupplierParams{
		Name:         input.Name,
		Email:        nullString(input.Email),
		Phone:        nullString(input.Phone),
		Address:      nullString(input.Address),
		City:         nullString(input.City),
		Country:      nullString(input.Country),
		TaxID:        nullString(input.TaxID),
		PaymentTerms: nullString(input.PaymentTerms),
		Notes:        nullString(input.Notes),
		IsActive:     sql.NullBool{Bool: input.IsActive, Valid: true},
		OwnerID:      ownerID,
		ID:           id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update supplier", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) DeleteSupplier(ctx context.Context, ownerID, id int64) error {
	err := r.queries.DeleteSupplier(ctx, sqlc.DeleteSupplierParams{
		OwnerID: ownerID,
		ID:      id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete supplier",
			logger.Any("owner_id", ownerID),
			logger.Any("supplier_id", id),
			logger.Error(err),
		)

		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
