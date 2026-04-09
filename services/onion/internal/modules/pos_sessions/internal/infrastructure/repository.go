package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_sessions/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Session, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, location_id, staff_id, opening_float, closing_float, total_sales, status, opened_at, closed_at, created_at
		 FROM pos_sessions WHERE owner_id = ? ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list sessions", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	sessions := make([]domain.Session, 0)
	for rows.Next() {
		s, scanErr := scanSession(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *Repository) Open(ctx context.Context, input domain.OpenSessionInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO pos_sessions (owner_id, location_id, staff_id, opening_float, status, opened_at)
		 VALUES (?, ?, ?, ?, 'open', NOW())`,
		input.OwnerID, input.LocationID, input.StaffID, input.OpeningFloat,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to open session", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Close(ctx context.Context, input domain.CloseSessionInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE pos_sessions SET closing_float = ?, status = 'closed', closed_at = NOW() WHERE id = ? AND owner_id = ? AND status = 'open'`,
		input.ClosingFloat, input.ID, input.OwnerID,
	)
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

func scanSession(rows *sql.Rows) (domain.Session, error) {
	var s domain.Session
	var locID, staffID sql.NullInt64
	var closedAt sql.NullTime

	if err := rows.Scan(&s.ID, &s.OwnerID, &locID, &staffID, &s.OpeningFloat, &s.ClosingFloat, &s.TotalSales, &s.Status, &s.OpenedAt, &closedAt, &s.CreatedAt); err != nil {
		return domain.Session{}, err
	}
	if locID.Valid {
		s.LocationID = &locID.Int64
	}
	if staffID.Valid {
		s.StaffID = &staffID.Int64
	}
	if closedAt.Valid {
		t := closedAt.Time
		s.ClosedAt = &t
	}
	return s, nil
}
