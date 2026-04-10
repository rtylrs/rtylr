package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	"github.com/voxire/onion/internal/modules/suppliers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListSuppliers(ctx context.Context, ownerID int64) ([]domain.Supplier, error) {
	rows, err := r.queries.ListSuppliers(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list suppliers", logger.Error(err))
		return nil, err
	}

	var suppliers []domain.Supplier
	for _, row := range rows {
		var s domain.Supplier
		s.ID = row.ID
		s.OwnerID = row.OwnerID
		s.Name = row.Name

		if row.Email.Valid {
			s.Email = row.Email.String
		}
		if row.Phone.Valid {
			s.Phone = row.Phone.String
		}
		if row.Address.Valid {
			s.Address = row.Address.String
		}
		if row.City.Valid {
			s.City = row.City.String
		}
		if row.Country.Valid {
			s.Country = row.Country.String
		}
		if row.TaxID.Valid {
			s.TaxID = row.TaxID.String
		}
		if row.PaymentTerms.Valid {
			s.PaymentTerms = row.PaymentTerms.String
		}
		if row.Notes.Valid {
			s.Notes = row.Notes.String
		}
		if row.IsActive.Valid {
			s.IsActive = row.IsActive.Bool
		}
		if row.CreatedAt.Valid {
			s.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			s.UpdatedAt = row.UpdatedAt.Time
		}

		suppliers = append(suppliers, s)
	}

	return suppliers, nil
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

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
