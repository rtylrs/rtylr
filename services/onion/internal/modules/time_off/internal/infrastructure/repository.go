package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/time_off/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) ListTimeOffRequests(ctx context.Context, ownerID int64) ([]domain.TimeOffRequest, error) {
	query := `
		SELECT id, owner_id, staff_id, type, start_date, end_date, days_count,
		       status, reason, approved_by, approved_at, notes, created_at, updated_at
		FROM hr_time_off_requests
		WHERE owner_id = ?
		ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list time off requests", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var requests []domain.TimeOffRequest
	for rows.Next() {
		var t domain.TimeOffRequest
		if scanErr := t.FromRow(rows); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan time off row", logger.Error(scanErr))
			return nil, scanErr
		}
		requests = append(requests, t)
	}

	return requests, rows.Err()
}

func (r *Repository) CreateTimeOffRequest(ctx context.Context, input domain.TimeOffInput) (int64, error) {
	query := `
		INSERT INTO hr_time_off_requests
		(owner_id, staff_id, type, start_date, end_date, days_count, reason, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.pool.ExecContext(ctx, query,
		input.OwnerID, input.StaffID, input.Type,
		input.StartDate, input.EndDate, input.DaysCount,
		sql.NullString{String: input.Reason, Valid: input.Reason != ""},
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create time off request", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) ApproveTimeOffRequest(ctx context.Context, ownerID int64, input domain.TimeOffApprovalInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE hr_time_off_requests
		 SET status = ?, approved_by = ?, approved_at = NOW(), notes = ?, updated_at = NOW()
		 WHERE id = ? AND owner_id = ?`,
		input.Status,
		sql.NullInt64{Int64: input.ApprovedBy, Valid: input.ApprovedBy != 0},
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		input.ID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update time off request", logger.Error(err))
		return err
	}

	return nil
}
