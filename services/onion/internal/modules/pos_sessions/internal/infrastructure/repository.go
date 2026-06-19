package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos_sessions/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_sessions/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Session, error) {
	rows, err := r.queries.ListSessionsByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list sessions", logger.Error(err))
		return nil, err
	}

	sessions := make([]domain.Session, 0, len(rows))
	for _, row := range rows {
		s := domain.Session{
			ID:           row.ID,
			OwnerID:      row.OwnerID,
			BusinessID:   row.BusinessID,
			OpeningFloat: decimalToFloat(row.OpeningCash.String),
			ClosingFloat: decimalToFloat(row.ClosingCash.String),
			TotalSales:   decimalToFloat(row.TotalSales.String),
			Status:       string(row.Status.PosSessionsStatus),
			OpenedAt:     row.OpenedAt.Time,
			CreatedAt:    row.OpenedAt.Time,
		}
		if row.LocationID.Valid {
			s.LocationID = &row.LocationID.Int64
		}
		if row.StaffID != 0 {
			s.StaffID = &row.StaffID
		}
		if row.ClosedAt.Valid {
			t := row.ClosedAt.Time
			s.ClosedAt = &t
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *Repository) Get(ctx context.Context, businessID, id int64) (*domain.Session, error) {
	row, err := r.queries.GetSessionByID(ctx, sqlc.GetSessionByIDParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		logger.Ctx(ctx).Error("Failed to get session", logger.Error(err))
		return nil, err
	}
	s := domain.Session{
		ID:           row.ID,
		OwnerID:      row.OwnerID,
		BusinessID:   row.BusinessID,
		OpeningFloat: decimalToFloat(row.OpeningCash.String),
		ClosingFloat: decimalToFloat(row.ClosingCash.String),
		TotalSales:   decimalToFloat(row.TotalSales.String),
		Status:       string(row.Status.PosSessionsStatus),
		OpenedAt:     row.OpenedAt.Time,
		CreatedAt:    row.OpenedAt.Time,
	}
	if row.LocationID.Valid {
		s.LocationID = &row.LocationID.Int64
	}
	if row.StaffID != 0 {
		s.StaffID = &row.StaffID
	}
	if row.ClosedAt.Valid {
		t := row.ClosedAt.Time
		s.ClosedAt = &t
	}
	return &s, nil
}

func (r *Repository) Open(ctx context.Context, input domain.OpenSessionInput) (int64, error) {
	var locID sql.NullInt64
	if input.LocationID != nil {
		locID = sql.NullInt64{Int64: *input.LocationID, Valid: true}
	}

	var staffID int64
	if input.StaffID != nil {
		staffID = *input.StaffID
	}

	result, err := r.queries.OpenSession(ctx, sqlc.OpenSessionParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		LocationID:  locID,
		StaffID:     staffID,
		OpeningCash: sql.NullString{String: floatToDecimal(input.OpeningFloat), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to open session", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Close(ctx context.Context, input domain.CloseSessionInput) error {
	result, err := r.queries.CloseSession(ctx, sqlc.CloseSessionParams{
		ClosingCash: sql.NullString{String: floatToDecimal(input.ClosingFloat), Valid: true},
		ID:          input.ID,
		BusinessID:  input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to close session", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func decimalToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse session decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

func floatToDecimal(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
