package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application/identity"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type workDayLineStore interface {
	ListWorkDayLines(context.Context, int64, int64, int64) ([]domain.WorkDayLine, int32, error)
	SumWorkDayLineMinutesByDayIDs(context.Context, int64, []int64) (map[int64]int32, error)
	SumWorkDayLineMinutesByMemberAndTaskIDs(context.Context, int64, []int64) ([]domain.WorkDayMemberMinutes, error)
	SumWorkDayLineMinutesByTaskIDs(context.Context, int64, []int64) (map[int64]int64, error)
	CreateWorkDayLine(context.Context, domain.WorkDayLineInput) (domain.WorkDayLine, error)
	UpdateWorkDayLine(context.Context, domain.WorkDayLineInput) (domain.WorkDayLine, error)
	GetWorkDayLineOwner(context.Context, int64, int64, int64) (domain.WorkDayLineOwner, bool, error)
	DeleteWorkDayLine(context.Context, int64, int64, int64) (bool, error)
}

type Repository struct {
	identities             identity.IdentityReader
	accountByEmail         func(context.Context, string) (int64, error)
	businessUserByUsername func(context.Context, int64, string) (int64, error)
	queries                *sqlc.Queries
	pool                   *sql.DB
	workDayLineStore       workDayLineStore
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) SetWorkDayLineStore(store workDayLineStore) {
	r.workDayLineStore = store
}

// ─────────────────────────────────────────────
// Workspaces
// ─────────────────────────────────────────────

func (r *Repository) CountWorkspacesByOwner(ctx context.Context, businessID int64) (int64, error) {
	count, err := r.queries.CountWorkspacesByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("CountWorkspacesByOwner failed", logger.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *Repository) CreateWorkspace(ctx context.Context, input domain.WorkspaceInput, isDefault bool) (int64, error) {
	icon := input.Icon
	if icon == "" {
		icon = "🚀"
	}
	color := input.Color
	if color == "" {
		color = "#4583C4"
	}

	res, err := r.queries.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Icon:        sql.NullString{String: icon, Valid: true},
		Color:       sql.NullString{String: color, Valid: true},
		IsDefault:   sql.NullBool{Bool: isDefault, Valid: true},
		CreatedBy:   input.CreatedBy,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateWorkspace failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetWorkspacesByOwner(ctx context.Context, businessID, limit, offset int64) ([]domain.Workspace, int64, error) {
	total, err := r.queries.CountWorkspacesByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("CountWorkspacesByOwner failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.GetWorkspacesByOwner(ctx, sqlc.GetWorkspacesByOwnerParams{
		BusinessID: businessID,
		Limit:      int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetWorkspacesByOwner failed", logger.Error(err))
		return nil, 0, err
	}

	workspaces := make([]domain.Workspace, 0, len(rows))
	for _, row := range rows {
		workspaces = append(workspaces, domain.Workspace{
			ID:          row.ID,
			BusinessID:  row.BusinessID,
			OwnerID:     row.OwnerID,
			Name:        row.Name,
			Description: row.Description.String,
			Icon:        row.Icon.String,
			Color:       row.Color.String,
			IsDefault:   row.IsDefault.Bool,
			CreatedBy:   row.CreatedBy,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return workspaces, total, nil
}

func (r *Repository) UpdateWorkspace(ctx context.Context, input domain.UpdateWorkspaceInput) error {
	err := r.queries.UpdateWorkspace(ctx, sqlc.UpdateWorkspaceParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Icon:        sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
		ID:          input.ID,
		BusinessID:  input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateWorkspace failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteWorkspace(ctx context.Context, id, businessID int64) error {
	err := r.queries.DeleteWorkspace(ctx, sqlc.DeleteWorkspaceParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteWorkspace failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) AddWorkspaceMember(ctx context.Context, input domain.AddWorkspaceMemberInput) error {
	role := input.Role
	if role == "" {
		role = "member"
	}

	err := r.queries.AddWorkspaceMember(ctx, sqlc.AddWorkspaceMemberParams{
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
		Role:        sqlc.NullFlowWorkspaceMembersRole{FlowWorkspaceMembersRole: sqlc.FlowWorkspaceMembersRole(role), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("AddWorkspaceMember failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) RemoveWorkspaceMember(ctx context.Context, input domain.RemoveWorkspaceMemberInput) error {
	err := r.queries.RemoveWorkspaceMember(ctx, sqlc.RemoveWorkspaceMemberParams{
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("RemoveWorkspaceMember failed", logger.Error(err))
		return err
	}
	return nil
}

// ─────────────────────────────────────────────
// Projects
// ─────────────────────────────────────────────

func (r *Repository) CreateProject(ctx context.Context, input domain.ProjectInput) (int64, error) {
	icon := input.Icon
	if icon == "" {
		icon = "📋"
	}
	color := input.Color
	if color == "" {
		color = "#4583C4"
	}

	visibility := input.Visibility
	if visibility == "" {
		visibility = domain.VisibilityPrivate
	}
	res, err := r.queries.CreateProject(ctx, sqlc.CreateProjectParams{
		WorkspaceID: input.WorkspaceID,
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Visibility:  sqlc.FlowProjectsVisibility(visibility),
		IsPersonal:  input.IsPersonal,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Icon:        sql.NullString{String: icon, Valid: true},
		Color:       sql.NullString{String: color, Valid: true},
		StartDate:   nullTimePtrToSql(input.StartDate),
		DueDate:     nullTimePtrToSql(input.DueDate),
		CreatedBy:   input.CreatedBy,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateProject failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetProjectsByWorkspace(ctx context.Context, workspaceID, businessID, limit, offset int64) ([]domain.Project, int64, error) {
	total, err := r.queries.CountProjectsByWorkspace(ctx, sqlc.CountProjectsByWorkspaceParams{
		WorkspaceID: workspaceID,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CountProjectsByWorkspace failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.GetProjectsByWorkspace(ctx, sqlc.GetProjectsByWorkspaceParams{
		WorkspaceID: workspaceID,
		BusinessID:  businessID,
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetProjectsByWorkspace failed", logger.Error(err))
		return nil, 0, err
	}

	projects := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		p := domain.Project{
			ID:             row.ID,
			WorkspaceID:    row.WorkspaceID,
			OwnerID:        row.OwnerID,
			BusinessID:     row.BusinessID,
			Name:           row.Name,
			Description:    row.Description,
			Icon:           row.Icon,
			Color:          row.Color,
			Status:         string(row.Status),
			CreatedBy:      row.CreatedBy,
			TaskCount:      int32(countIfaceToInt64(row.TaskCount)),      // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			CompletedCount: int32(countIfaceToInt64(row.CompletedCount)), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			MemberCount:    int32(countIfaceToInt64(row.MemberCount)),    // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		}
		if row.StartDate.Valid {
			p.StartDate = &row.StartDate.Time
		}
		if row.DueDate.Valid {
			p.DueDate = &row.DueDate.Time
		}
		if row.CreatedAt.Valid {
			p.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			p.UpdatedAt = row.UpdatedAt.Time
		}
		projects = append(projects, p)
	}
	return projects, total, nil
}
