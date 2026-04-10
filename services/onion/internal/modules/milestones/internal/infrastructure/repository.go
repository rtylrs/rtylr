package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/milestones/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64, projectID *int64) ([]domain.Milestone, error) {
	query := `SELECT id, owner_id, project_id, name, description, target_date, status, created_at, updated_at
	          FROM flow_milestones WHERE owner_id = ?`
	args := []interface{}{ownerID}

	if projectID != nil {
		query += " AND project_id = ?"
		args = append(args, *projectID)
	}

	query += " ORDER BY target_date ASC, created_at ASC"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list milestones", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	milestones := make([]domain.Milestone, 0)
	for rows.Next() {
		m, scanErr := scanMilestone(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		milestones = append(milestones, m)
	}
	return milestones, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input domain.MilestoneInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_milestones (owner_id, project_id, name, description, target_date)
		 VALUES (?, ?, ?, ?, ?)`,
		input.OwnerID, input.ProjectID, input.Name,
		nullableString(input.Description), nullableTime(input.TargetDate),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create milestone", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.MilestoneUpdateInput) error {
	status := input.Status
	if status == "" {
		status = "upcoming"
	}
	result, err := r.pool.ExecContext(ctx,
		`UPDATE flow_milestones SET name = ?, description = ?, target_date = ?, status = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, nullableString(input.Description), nullableTime(input.TargetDate),
		status, input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update milestone", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrMilestoneNotFound
	}
	return nil
}

type milestoneScanner interface {
	Scan(dest ...any) error
}

func scanMilestone(row milestoneScanner) (domain.Milestone, error) {
	var m domain.Milestone
	var desc sql.NullString
	var targetDate sql.NullTime

	err := row.Scan(&m.ID, &m.OwnerID, &m.ProjectID, &m.Name, &desc, &targetDate, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return domain.Milestone{}, err
	}
	m.Description = desc.String
	if targetDate.Valid {
		m.TargetDate = &targetDate.Time
	}
	return m, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
