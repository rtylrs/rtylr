package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) Search(ctx context.Context, businessID int64, query string) ([]domain.Customer, error) {
	searchTerm := "%" + query + "%"

	rows, err := r.queries.SearchPOSCustomers(ctx, sqlc.SearchPOSCustomersParams{
		BusinessID: businessID,
		Phone:      sql.NullString{String: searchTerm, Valid: true},
		Email:      sql.NullString{String: searchTerm, Valid: true},
		FirstName:  searchTerm,
		LastName:   sql.NullString{String: searchTerm, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to search POS customers", logger.Error(err))
		return nil, err
	}

	customers := make([]domain.Customer, 0, len(rows))
	for _, row := range rows {
		customers = append(customers, domain.Customer{
			ID:        row.ID,
			FirstName: row.FirstName,
			LastName:  row.LastName.String,
			Email:     row.Email.String,
			Phone:     row.Phone.String,
		})
	}
	return customers, nil
}
