package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/booking_services/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_services/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.BookingService, error) {
	rows, err := r.queries.ListBookingServices(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list booking services", logger.Error(err))
		return nil, err
	}
	items := make([]domain.BookingService, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(row))
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.BookingServiceInput) (int64, error) {
	result, err := r.queries.CreateBookingService(ctx, sqlc.CreateBookingServiceParams{
		OwnerID: ownerID, Name: input.Name,
		Description:     sql.NullString{String: input.Description, Valid: input.Description != ""},
		DurationMinutes: int32(input.DurationMinutes),
		Price:           strconv.FormatFloat(input.Price, 'f', 2, 64),
		Color:           sql.NullString{String: input.Color, Valid: input.Color != ""},
		IsActive:        sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create booking service", logger.Error(err))
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.BookingServiceInput) error {
	err := r.queries.UpdateBookingService(ctx, sqlc.UpdateBookingServiceParams{
		Name:            input.Name,
		Description:     sql.NullString{String: input.Description, Valid: input.Description != ""},
		DurationMinutes: int32(input.DurationMinutes),
		Price:           strconv.FormatFloat(input.Price, 'f', 2, 64),
		Color:           sql.NullString{String: input.Color, Valid: input.Color != ""},
		IsActive:        sql.NullBool{Bool: input.IsActive, Valid: true},
		OwnerID:         ownerID, ID: id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update booking service", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	err := r.queries.DeleteBookingService(ctx, sqlc.DeleteBookingServiceParams{OwnerID: ownerID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete booking service", logger.Error(err))
		return err
	}
	return nil
}

func fromRow(row sqlc.BookingService) domain.BookingService {
	s := domain.BookingService{ID: row.ID, OwnerID: row.OwnerID, Name: row.Name, DurationMinutes: int(row.DurationMinutes)}
	if row.Description.Valid {
		s.Description = row.Description.String
	}
	s.Price, _ = strconv.ParseFloat(row.Price, 64)
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
