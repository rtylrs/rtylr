package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/booking/services/internal/domain"
	"github.com/voxire/onion/internal/modules/booking/services/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.BookingService, error) {
	rows, err := r.queries.ListBookingServices(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list booking services", logger.Error(err))
		return nil, err
	}
	items := make([]domain.BookingService, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(ctx, row))
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.BookingServiceInput) (int64, error) {
	result, err := r.queries.CreateBookingService(ctx, sqlc.CreateBookingServiceParams{
		OwnerID: ownerID, BusinessID: ownerID, Name: input.Name,
		Description:     sql.NullString{String: input.Description, Valid: input.Description != ""},
		DurationMinutes: int32(input.DurationMinutes), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Price:           input.Price.String(),
		Color:           sql.NullString{String: input.Color, Valid: input.Color != ""},
		IsActive:        sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create booking service", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to retrieve last insert ID for booking service", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.BookingServiceInput) error {
	err := r.queries.UpdateBookingService(ctx, sqlc.UpdateBookingServiceParams{
		Name:            input.Name,
		Description:     sql.NullString{String: input.Description, Valid: input.Description != ""},
		DurationMinutes: int32(input.DurationMinutes), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Price:           input.Price.String(),
		Color:           sql.NullString{String: input.Color, Valid: input.Color != ""},
		IsActive:        sql.NullBool{Bool: input.IsActive, Valid: true},
		BusinessID:      businessID, ID: id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update booking service", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteBookingService(ctx, sqlc.DeleteBookingServiceParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete booking service", logger.Error(err))
		return err
	}
	return nil
}

func fromRow(ctx context.Context, row sqlc.BookingService) domain.BookingService {
	s := domain.BookingService{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Name: row.Name, DurationMinutes: int(row.DurationMinutes)}
	if row.Description.Valid {
		s.Description = row.Description.String
	}
	if price, err := money.FromString(row.Price); err != nil {
		logger.Ctx(ctx).Warn("failed to parse price from database",
			logger.String("price_str", row.Price),
			logger.Error(err),
		)
		s.Price = 0
	} else {
		s.Price = price
	}
	if row.Color.Valid {
		s.Color = row.Color.String
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
