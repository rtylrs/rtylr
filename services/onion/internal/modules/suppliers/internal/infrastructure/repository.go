package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	"github.com/voxire/onion/internal/modules/suppliers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) ListSuppliers(ctx context.Context, businessID int64, limit, offset int) ([]domain.Supplier, bool, error) {
	rows, err := r.queries.ListSuppliersPaginated(ctx, sqlc.ListSuppliersPaginatedParams{
		BusinessID: businessID,
		Limit:      int32(limit + 1), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:     int32(offset),    // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list suppliers", logger.Error(err))
		return nil, false, err
	}

	suppliers := make([]domain.Supplier, 0, limit+1)
	for _, row := range rows {
		suppliers = append(suppliers, supplierFromRow(row))
	}

	hasMore := len(suppliers) > limit
	if hasMore {
		suppliers = suppliers[:limit]
	}

	return suppliers, hasMore, nil
}

func (r *Repository) CreateSupplier(ctx context.Context, businessID int64, input domain.SupplierInput) (int64, error) {
	result, err := r.queries.CreateSupplier(ctx, sqlc.CreateSupplierParams{
		OwnerID:      businessID,
		BusinessID:   businessID,
		Name:         input.Name,
		Email:        sqlutil.NullString(input.Email),
		Phone:        sqlutil.NullString(input.Phone),
		Address:      sqlutil.NullString(input.Address),
		City:         sqlutil.NullString(input.City),
		Country:      sqlutil.NullString(input.Country),
		TaxID:        sqlutil.NullString(input.TaxID),
		PaymentTerms: sqlutil.NullString(input.PaymentTerms),
		Notes:        sqlutil.NullString(input.Notes),
		IsActive:     sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create supplier", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateSupplier(ctx context.Context, businessID, id int64, input domain.SupplierInput) error {
	err := r.queries.UpdateSupplier(ctx, sqlc.UpdateSupplierParams{
		Name:         input.Name,
		Email:        sqlutil.NullString(input.Email),
		Phone:        sqlutil.NullString(input.Phone),
		Address:      sqlutil.NullString(input.Address),
		City:         sqlutil.NullString(input.City),
		Country:      sqlutil.NullString(input.Country),
		TaxID:        sqlutil.NullString(input.TaxID),
		PaymentTerms: sqlutil.NullString(input.PaymentTerms),
		Notes:        sqlutil.NullString(input.Notes),
		IsActive:     sql.NullBool{Bool: input.IsActive, Valid: true},
		BusinessID:   businessID,
		ID:           id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update supplier", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) DeleteSupplier(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteSupplier(ctx, sqlc.DeleteSupplierParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete supplier",
			logger.Any("business_id", businessID),
			logger.Any("supplier_id", id),
			logger.Error(err),
		)

		return err
	}

	return nil
}

// CountSuppliers returns the total number of suppliers for the given business.
// Pass activeOnly=true to count only active suppliers.
func (r *Repository) CountSuppliers(ctx context.Context, businessID int64, activeOnly bool) (int64, error) {
	var count int64
	var err error

	if activeOnly {
		count, err = r.queries.CountActiveSuppliers(ctx, businessID)
	} else {
		count, err = r.queries.CountSuppliers(ctx, businessID)
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to count suppliers", logger.Error(err))
		return 0, err
	}

	return count, nil
}

// GetSupplierByID fetches a single supplier by its primary key for the given business.
// Returns ErrSupplierNotFound when no matching row exists.
func (r *Repository) GetSupplierByID(ctx context.Context, businessID, id int64) (domain.Supplier, error) {
	row, err := r.queries.GetSupplierByID(ctx, sqlc.GetSupplierByIDParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Supplier{}, domain.ErrSupplierNotFound
		}
		logger.Ctx(ctx).Error("Failed to get supplier by ID", logger.Error(err))
		return domain.Supplier{}, err
	}

	return supplierFromRow(row), nil
}

// supplierFromRow maps a sqlc-generated erp_suppliers row to the domain model,
// collapsing nullable columns to their zero value when NULL.
func supplierFromRow(row sqlc.ErpSupplier) domain.Supplier {
	s := domain.Supplier{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Name:       row.Name,
	}

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

	return s
}
