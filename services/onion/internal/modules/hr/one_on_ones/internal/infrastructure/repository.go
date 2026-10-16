package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/one_on_ones/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/one_on_ones/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.OneOnOne, error) {
	rows, err := r.queries.ListOneOnOnes(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list one-on-ones", logger.Error(err))
		return nil, err
	}

	items := make([]domain.OneOnOne, 0, len(rows))
	for _, row := range rows {
		var item domain.OneOnOne
		item.ID = row.ID
		item.OwnerID = row.OwnerID
		item.BusinessID = row.BusinessID
		item.ManagerID = row.ManagerID
		item.StaffID = row.StaffID
		item.ScheduledAt = row.ScheduledAt
		if row.Notes.Valid {
			item.Notes = row.Notes.String
		}
		if row.ActionItems != nil {
			item.ActionItems = string(row.ActionItems)
		}
		if row.Status.Valid {
			item.Status = string(row.Status.HrOneOnOnesStatus)
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.OneOnOneInput) (int64, error) {
	status := input.Status
	if status == "" {
		status = "scheduled"
	}

	actionItems := json.RawMessage("[]")
	if input.ActionItems != "" {
		actionItems = json.RawMessage(input.ActionItems)
	}

	if err := r.requireStaff(ctx, businessID, input.ManagerID, input.StaffID); err != nil {
		return 0, err
	}

	res, err := r.queries.CreateOneOnOne(ctx, sqlc.CreateOneOnOneParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		ManagerID:   input.ManagerID,
		StaffID:     input.StaffID,
		ScheduledAt: parseTime(ctx, input.ScheduledAt),
		Notes:       sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		ActionItems: actionItems,
		Status:      sqlc.NullHrOneOnOnesStatus{HrOneOnOnesStatus: sqlc.HrOneOnOnesStatus(status), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create one-on-one", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// parseTime tries each accepted layout and returns the first successful
// parse. When every layout fails it logs at Warn and returns the zero time —
// the handler validates scheduled_at before this is reached, so a failure here
// signals corruption upstream rather than ordinary user input. Logging makes
// that corruption visible instead of silently storing 0001-01-01.
func parseTime(ctx context.Context, s string) time.Time {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		if parsed, err := time.Parse(layout, s); err == nil {
			return parsed
		}
	}
	logger.Ctx(ctx).Warn("failed to parse scheduled_at, defaulting to zero time",
		logger.String("value", s),
	)
	return time.Time{}
}

// requireStaff fails with ErrStaffNotFound unless every id is a staff member of
// the business, so a one-on-one cannot point at another tenant's staff.
func (r *Repository) requireStaff(ctx context.Context, businessID int64, ids ...int64) error {
	unique := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	args := make([]any, 0, len(unique)+1)
	args = append(args, businessID)
	placeholders := make([]string, 0, len(unique))
	for id := range unique {
		args = append(args, id)
		placeholders = append(placeholders, "?")
	}
	var found int
	err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM hr_staff WHERE business_id = ? AND id IN ("+strings.Join(placeholders, ",")+")",
		args...,
	).Scan(&found)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to verify one-on-one staff", logger.Error(err))
		return err
	}
	if found != len(unique) {
		return domain.ErrStaffNotFound
	}
	return nil
}
