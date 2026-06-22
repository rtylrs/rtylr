package infrastructure

import (
	"context"
	"database/sql"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
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
	query = strings.TrimSpace(query)
	if query == "" {
		return []domain.Customer{}, nil
	}
	searchTerm := searchutil.BooleanPrefixQuery(query)

	rows, err := r.queries.SearchPOSCustomers(ctx, sqlc.SearchPOSCustomersParams{
		BusinessID: businessID,
		Phone:      sql.NullString{String: query, Valid: true},
		Email:      sql.NullString{String: query, Valid: true},
		Search:     searchTerm,
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
