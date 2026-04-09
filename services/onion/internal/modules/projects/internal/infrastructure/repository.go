package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/projects/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Project, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, workspace_id, name, description, status, color, created_at, updated_at
		 FROM flow_projects WHERE owner_id = ? ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list projects", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	projects := make([]domain.Project, 0)
	for rows.Next() {
		p, scanErr := scanProject(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input domain.ProjectInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_projects (owner_id, workspace_id, name, description, status, color)
		 VALUES (?, ?, ?, ?, 'active', ?)`,
		input.OwnerID, input.WorkspaceID, input.Name,
		nullableString(input.Description), nullableString(input.Color),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create project", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.ProjectUpdateInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE flow_projects SET workspace_id = ?, name = ?, description = ?, status = ?, color = ?
		 WHERE id = ? AND owner_id = ?`,
		input.WorkspaceID, input.Name, nullableString(input.Description),
		input.Status, nullableString(input.Color), input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update project", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM flow_projects WHERE id = ? AND owner_id = ?", id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete project", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func scanProject(rows *sql.Rows) (domain.Project, error) {
	var p domain.Project
	var workspaceID sql.NullInt64
	var desc, color sql.NullString
	if err := rows.Scan(&p.ID, &p.OwnerID, &workspaceID, &p.Name, &desc, &p.Status, &color, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return domain.Project{}, err
	}
	if workspaceID.Valid {
		p.WorkspaceID = &workspaceID.Int64
	}
	p.Description = desc.String
	p.Color = color.String
	return p, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
