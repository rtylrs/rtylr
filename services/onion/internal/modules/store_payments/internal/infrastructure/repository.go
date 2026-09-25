package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/store_payments/internal/domain"
	"github.com/voxire/onion/internal/modules/store_payments/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.PaymentMethod, error) {
	rows, err := r.queries.ListPaymentMethods(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list payment methods", logger.Error(err))
		return nil, err
	}
	items := make([]domain.PaymentMethod, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(row))
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID, businessID int64, input domain.PaymentMethodInput) (int64, error) {
	result, err := r.queries.CreatePaymentMethod(ctx, sqlc.CreatePaymentMethodParams{
		OwnerID: ownerID, BusinessID: businessID, Provider: input.Provider,
		Config: nullJSON(input.Config), IsActive: sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create payment method", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.PaymentMethodInput) error {
	err := r.queries.UpdatePaymentMethod(ctx, sqlc.UpdatePaymentMethodParams{
		Provider: input.Provider, Config: nullJSON(input.Config),
		IsActive: sql.NullBool{Bool: input.IsActive, Valid: true}, BusinessID: businessID, ID: id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update payment method", logger.Error(err))
		return err
	}
	return nil
}

func fromRow(row sqlc.StorePaymentMethod) domain.PaymentMethod {
	p := domain.PaymentMethod{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Provider: row.Provider}
	if row.Config != nil {
		p.Config = string(row.Config)
	}
	if row.IsActive.Valid {
		p.IsActive = row.IsActive.Bool
	}
	if row.CreatedAt.Valid {
		p.CreatedAt = row.CreatedAt.Time
	}
	return p
}

func nullJSON(s string) []byte {
	if s == "" {
		return nil
	}
	return []byte(s)
}
