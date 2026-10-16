package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr/time_off/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/time_off/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) ListTimeOffRequests(ctx context.Context, businessID int64) ([]domain.TimeOffRequest, error) {
	rows, err := r.queries.ListTimeOffRequests(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list time off requests", logger.Error(err))
		return nil, err
	}

	requests := make([]domain.TimeOffRequest, 0, len(rows))
	for _, row := range rows {
		var t domain.TimeOffRequest
		t.ID = row.ID
		t.OwnerID = row.OwnerID
		t.BusinessID = row.BusinessID
		t.StaffID = row.StaffID
		switch v := row.StaffName.(type) {
		case string:
			t.StaffName = v
		case []byte:
			t.StaffName = string(v)
		case nil:
			t.StaffName = ""
		default:
			t.StaffName = fmt.Sprint(v)
		}
		t.Type = string(row.Type)
		t.StartDate = row.StartDate
		t.EndDate = row.EndDate
		t.DaysCount = sqlutil.ParseDecimal(row.DaysCount)
		if row.Status.Valid {
			t.Status = string(row.Status.HrTimeOffRequestsStatus)
		}
		if row.Reason.Valid {
			t.Reason = row.Reason.String
		}
		if row.ApprovedBy.Valid {
			t.ApprovedBy = row.ApprovedBy.Int64
		}
		if row.ApprovedAt.Valid {
			t.ApprovedAt = row.ApprovedAt.Time
		}
		if row.Notes.Valid {
			t.Notes = row.Notes.String
		}
		if row.CreatedAt.Valid {
			t.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			t.UpdatedAt = row.UpdatedAt.Time
		}
		requests = append(requests, t)
	}

	return requests, nil
}

func (r *Repository) CreateTimeOffRequest(ctx context.Context, input domain.TimeOffInput) (int64, error) {
	var found int
	err := r.pool.QueryRowContext(ctx,
		"SELECT 1 FROM hr_staff WHERE id = ? AND business_id = ? LIMIT 1",
		input.StaffID, input.BusinessID,
	).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrStaffNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to verify staff for time off request", logger.Error(err))
		return 0, err
	}
	result, err := r.queries.CreateTimeOffRequest(ctx, sqlc.CreateTimeOffRequestParams{
		OwnerID:    input.OwnerID,
		BusinessID: input.BusinessID,
		StaffID:    input.StaffID,
		Type:       sqlc.HrTimeOffRequestsType(input.Type),
		StartDate:  parseDate(input.StartDate),
		EndDate:    parseDate(input.EndDate),
		DaysCount:  fmt.Sprintf("%g", input.DaysCount),
		Reason:     sql.NullString{String: input.Reason, Valid: input.Reason != ""},
		Notes:      sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create time off request", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) ApproveTimeOffRequest(ctx context.Context, businessID int64, input domain.TimeOffApprovalInput) error {
	err := r.queries.ApproveTimeOffRequest(ctx, sqlc.ApproveTimeOffRequestParams{
		Status:     sqlc.NullHrTimeOffRequestsStatus{HrTimeOffRequestsStatus: sqlc.HrTimeOffRequestsStatus(input.Status), Valid: input.Status != ""},
		ApprovedBy: sql.NullInt64{Int64: input.ApprovedBy, Valid: input.ApprovedBy != 0},
		Notes:      sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		ID:         input.ID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update time off request", logger.Error(err))
		return err
	}

	return nil
}

func parseDate(s string) time.Time {
	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return parsed
	}
	if parsed, err := time.Parse(time.RFC3339, s); err == nil {
		return parsed
	}
	return time.Time{}
}
