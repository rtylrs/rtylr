package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/cycles/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64, workspaceID *int64) ([]domain.Cycle, error) {
	query := `SELECT id, owner_id, workspace_id, name, start_date, end_date, status, created_at, updated_at
	          FROM flow_cycles WHERE owner_id = ?`
	args := []interface{}{ownerID}

	if workspaceID != nil {
		query += " AND workspace_id = ?"
		args = append(args, *workspaceID)
	}

	query += " ORDER BY start_date ASC, created_at ASC"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list cycles", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	cycles := make([]domain.Cycle, 0)
	for rows.Next() {
		c, scanErr := scanCycle(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		cycles = append(cycles, c)
	}
	return cycles, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input domain.CycleInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_cycles (owner_id, workspace_id, name, start_date, end_date)
		 VALUES (?, ?, ?, ?, ?)`,
		input.OwnerID, input.WorkspaceID, input.Name,
		nullableTime(input.StartDate), nullableTime(input.EndDate),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create cycle", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.CycleUpdateInput) error {
	status := input.Status
	if status == "" {
		status = "upcoming"
	}
	result, err := r.pool.ExecContext(ctx,
		`UPDATE flow_cycles SET name = ?, start_date = ?, end_date = ?, status = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, nullableTime(input.StartDate), nullableTime(input.EndDate),
		status, input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update cycle", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrCycleNotFound
	}
	return nil
}

type cycleScanner interface {
	Scan(dest ...any) error
}

func scanCycle(row cycleScanner) (domain.Cycle, error) {
	var c domain.Cycle
	var startDate, endDate sql.NullTime

	err := row.Scan(&c.ID, &c.OwnerID, &c.WorkspaceID, &c.Name, &startDate, &endDate, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Cycle{}, err
	}
	if startDate.Valid {
		c.StartDate = &startDate.Time
	}
	if endDate.Valid {
		c.EndDate = &endDate.Time
	}
	return c, nil
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
