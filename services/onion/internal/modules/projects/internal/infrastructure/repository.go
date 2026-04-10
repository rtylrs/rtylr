package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/projects/internal/domain"
	"github.com/voxire/onion/internal/modules/projects/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Project, error) {
	rows, err := r.queries.ListProjectsByOwner(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list projects", logger.Error(err))
		return nil, err
	}

	projects := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		p := domain.Project{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			Name:        row.Name,
			Description: row.Description.String,
			Status:      string(row.Status.FlowProjectsStatus),
			Color:       row.Color.String,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
		if row.WorkspaceID != 0 {
			p.WorkspaceID = &row.WorkspaceID
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *Repository) Create(ctx context.Context, input domain.ProjectInput) (int64, error) {
	result, err := r.queries.CreateProject(ctx, sqlc.CreateProjectParams{
		OwnerID:     input.OwnerID,
		WorkspaceID: derefInt64(input.WorkspaceID),
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create project", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.ProjectUpdateInput) error {
	result, err := r.queries.UpdateProject(ctx, sqlc.UpdateProjectParams{
		WorkspaceID: derefInt64(input.WorkspaceID),
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Status:      sqlc.NullFlowProjectsStatus{FlowProjectsStatus: sqlc.FlowProjectsStatus(input.Status), Valid: input.Status != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
		ID:          input.ID,
		OwnerID:     input.OwnerID,
	})
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

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	result, err := r.queries.DeleteProject(ctx, sqlc.DeleteProjectParams{
		ID:      id,
		OwnerID: ownerID,
	})
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
