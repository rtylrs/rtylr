package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) Search(ctx context.Context, ownerID int64, query string) ([]domain.Customer, error) {
	searchTerm := "%" + query + "%"

	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, first_name, COALESCE(last_name,''), COALESCE(email,''), COALESCE(phone,'')
		 FROM crm_customers
		 WHERE owner_id = ? AND (phone LIKE ? OR email LIKE ? OR first_name LIKE ? OR last_name LIKE ?)
		 ORDER BY first_name LIMIT 20`,
		ownerID, searchTerm, searchTerm, searchTerm, searchTerm,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to search POS customers", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	customers := make([]domain.Customer, 0)
	for rows.Next() {
		var c domain.Customer
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone); err != nil {
			logger.Ctx(ctx).Error("Failed to scan POS customer", logger.Error(err))
			return nil, err
		}
		customers = append(customers, c)
	}
	return customers, rows.Err()
}
