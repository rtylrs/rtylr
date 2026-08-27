package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/support_customers/internal/domain"
	"github.com/voxire/onion/internal/modules/support_customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) List(ctx context.Context, businessID int64, page, pageSize int) ([]domain.Customer, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := r.queries.CountCustomers(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count support customers", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListCustomers(ctx, sqlc.ListCustomersParams{
		BusinessID: businessID,
		Limit:      int32(pageSize),
		Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list support customers", logger.Error(err))
		return nil, 0, err
	}

	items := make([]domain.Customer, 0, len(rows))
	for _, row := range rows {
		c := domain.Customer{
			ID:           row.ID,
			BusinessID:   row.BusinessID,
			Name:         row.Name,
			TotalTickets: int(row.TotalTickets.Int32),
		}
		if row.Email.Valid {
			c.Email = row.Email.String
		}
		if row.Phone.Valid {
			c.Phone = row.Phone.String
		}
		if row.LastTicketAt.Valid {
			t := row.LastTicketAt.Time
			c.LastTicketAt = &t
		}
		if row.CreatedAt.Valid {
			c.CreatedAt = row.CreatedAt.Time
		}
		items = append(items, c)
	}
	return items, int(total), nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (*domain.Customer, error) {
	row, err := r.queries.GetCustomerByID(ctx, sqlc.GetCustomerByIDParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get support customer", logger.Error(err))
		return nil, err
	}
	c := domain.Customer{
		ID:           row.ID,
		BusinessID:   row.BusinessID,
		Name:         row.Name,
		TotalTickets: int(row.TotalTickets.Int32),
	}
	if row.Email.Valid {
		c.Email = row.Email.String
	}
	if row.Phone.Valid {
		c.Phone = row.Phone.String
	}
	if row.LastTicketAt.Valid {
		t := row.LastTicketAt.Time
		c.LastTicketAt = &t
	}
	if row.CreatedAt.Valid {
		c.CreatedAt = row.CreatedAt.Time
	}
	return &c, nil
}
