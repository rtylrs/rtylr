package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application/identity"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) ListWorkspaceMembers(ctx context.Context, workspaceID, limit, offset int64) ([]domain.WorkspaceMember, int64, error) {
	total, err := r.queries.CountWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("CountWorkspaceMembers failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListWorkspaceMembers(ctx, sqlc.ListWorkspaceMembersParams{
		WorkspaceID: workspaceID,
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListWorkspaceMembers failed", logger.Error(err))
		return nil, 0, err
	}

	keys := make([]identity.IdentityKey, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, identity.IdentityKey{BusinessID: row.BusinessID, UserID: row.UserID})
	}
	profiles, err := r.readIdentities(ctx, keys)
	if err != nil {
		return nil, 0, err
	}
	members := make([]domain.WorkspaceMember, 0, len(rows))
	for _, row := range rows {
		profile := profiles[identity.IdentityKey{BusinessID: row.BusinessID, UserID: row.UserID}]
		members = append(members, domain.WorkspaceMember{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			UserID:      row.UserID,
			Role:        string(row.Role.FlowWorkspaceMembersRole),
			JoinedAt:    row.JoinedAt.Time,
			Name:        profile.Name,
			Email:       profile.Email,
		})
	}
	return members, total, nil
}

func (r *Repository) UpdateWorkspaceMemberRole(ctx context.Context, input domain.UpdateWorkspaceMemberRoleInput) error {
	err := r.queries.UpdateWorkspaceMemberRole(ctx, sqlc.UpdateWorkspaceMemberRoleParams{
		Role:        sqlc.NullFlowWorkspaceMembersRole{FlowWorkspaceMembersRole: sqlc.FlowWorkspaceMembersRole(input.Role), Valid: true},
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateWorkspaceMemberRole failed", logger.Error(err))
	}
	return err
}

func (r *Repository) CreatePersonalWorkspace(ctx context.Context, userID, businessID int64) (int64, error) {
	existing, err := r.queries.GetPersonalWorkspaceByOwner(ctx, sqlc.GetPersonalWorkspaceByOwnerParams{
		OwnerID:    userID,
		BusinessID: businessID,
	})
	if err == nil && existing > 0 {
		return existing, nil
	}
	// Only a genuine miss may fall through to the INSERT. flow_workspaces has
	// no unique key on owner_id/business_id/is_personal, so treating a
	// transient lookup failure (pool exhaustion, statement timeout) as "no
	// workspace yet" inserts a second 'Personal' workspace for the same owner.
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace lookup failed", logger.Error(err))
		return 0, err
	}

	res, err := r.queries.CreatePersonalWorkspace(ctx, sqlc.CreatePersonalWorkspaceParams{
		OwnerID:    userID,
		BusinessID: businessID,
		CreatedBy:  userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace failed", logger.Error(err))
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace last insert id failed", logger.Error(err))
		return 0, err
	}

	if addErr := r.queries.AddWorkspaceMember(ctx, sqlc.AddWorkspaceMemberParams{
		WorkspaceID: id,
		UserID:      userID,
		Role:        sqlc.NullFlowWorkspaceMembersRole{FlowWorkspaceMembersRole: sqlc.FlowWorkspaceMembersRoleAdmin, Valid: true},
	}); addErr != nil {
		// A workspace whose owner is not a member is not usable; reporting the
		// insert as a success left the caller with a half-created workspace.
		logger.Ctx(ctx).Error("CreatePersonalWorkspace add member failed", logger.Error(addErr))
		return 0, addErr
	}
	return id, nil
}

// ─────────────────────────────────────────────
// Project Members
// ─────────────────────────────────────────────

func (r *Repository) ListProjectMembers(ctx context.Context, projectID, limit, offset int64) ([]domain.ProjectMember, int64, error) {
	total, err := r.queries.CountProjectMembers(ctx, projectID)
	if err != nil {
		logger.Ctx(ctx).Error("CountProjectMembers failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListProjectMembers(ctx, sqlc.ListProjectMembersParams{
		ProjectID: projectID,
		Limit:     int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:    int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListProjectMembers failed", logger.Error(err))
		return nil, 0, err
	}

	keys := make([]identity.IdentityKey, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, identity.IdentityKey{BusinessID: row.BusinessID, UserID: row.UserID})
	}
	profiles, err := r.readIdentities(ctx, keys)
	if err != nil {
		return nil, 0, err
	}
	members := make([]domain.ProjectMember, 0, len(rows))
	for _, row := range rows {
		profile := profiles[identity.IdentityKey{BusinessID: row.BusinessID, UserID: row.UserID}]
		members = append(members, domain.ProjectMember{
			ID:        row.ID,
			ProjectID: row.ProjectID,
			UserID:    row.UserID,
			Role:      string(row.Role.FlowProjectMembersRole),
			AddedAt:   row.JoinedAt.Time,
			Name:      profile.Name,
			Email:     profile.Email,
		})
	}
	return members, total, nil
}

func (r *Repository) AddProjectMember(ctx context.Context, input domain.AddProjectMemberInput) error {
	role := input.Role
	if role == "" {
		role = "member"
	}
	err := r.queries.AddProjectMember(ctx, sqlc.AddProjectMemberParams{
		ProjectID: input.ProjectID,
		UserID:    input.UserID,
		Role:      sqlc.NullFlowProjectMembersRole{FlowProjectMembersRole: sqlc.FlowProjectMembersRole(role), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("AddProjectMember failed", logger.Error(err))
	}
	return err
}

func (r *Repository) RemoveProjectMember(ctx context.Context, input domain.RemoveProjectMemberInput) error {
	err := r.queries.RemoveProjectMember(ctx, sqlc.RemoveProjectMemberParams{
		ProjectID: input.ProjectID,
		UserID:    input.UserID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("RemoveProjectMember failed", logger.Error(err))
	}
	return err
}

func (r *Repository) UpdateProjectMemberRole(ctx context.Context, input domain.UpdateProjectMemberRoleInput) error {
	err := r.queries.UpdateProjectMemberRole(ctx, sqlc.UpdateProjectMemberRoleParams{
		Role:      sqlc.NullFlowProjectMembersRole{FlowProjectMembersRole: sqlc.FlowProjectMembersRole(input.Role), Valid: true},
		ProjectID: input.ProjectID,
		UserID:    input.UserID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateProjectMemberRole failed", logger.Error(err))
	}
	return err
}

func (r *Repository) TransferProjectOwnership(ctx context.Context, input domain.TransferProjectOwnershipInput) error {
	err := r.queries.TransferProjectOwnership(ctx, sqlc.TransferProjectOwnershipParams{
		OwnerID:   input.NewOwnerID,
		ID:        input.ProjectID,
		OwnerID_2: input.CallerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("TransferProjectOwnership failed", logger.Error(err))
	}
	return err
}

func (r *Repository) MoveProject(ctx context.Context, input domain.MoveProjectInput) error {
	err := r.queries.MoveProject(ctx, sqlc.MoveProjectParams{
		WorkspaceID: input.WorkspaceID,
		ID:          input.ProjectID,
		OwnerID:     input.CallerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("MoveProject failed", logger.Error(err))
	}
	return err
}

func (r *Repository) GetUserWorkspaceRole(ctx context.Context, workspaceID, userID int64) (string, error) {
	role, err := r.queries.GetUserWorkspaceRole(ctx, sqlc.GetUserWorkspaceRoleParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return "", err
	}
	if !role.Valid {
		return "", sql.ErrNoRows
	}
	return string(role.FlowWorkspaceMembersRole), nil
}

// GetProjectMemberRole returns the caller's role on a project, or
// sql.ErrNoRows when they hold no explicit membership row. A project owner and
// a workspace member both reach a project without one, so a not-found here is
// not the same as "no access" — see Service.RequireProjectRole.
func (r *Repository) GetProjectMemberRole(ctx context.Context, projectID, userID int64) (string, error) {
	role, err := r.queries.GetProjectMemberRole(ctx, sqlc.GetProjectMemberRoleParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return "", err
	}
	if !role.Valid {
		return "", sql.ErrNoRows
	}
	return string(role.FlowProjectMembersRole), nil
}

// GetProjectOwnerWorkspace returns a project's owner and workspace, the two
// facts needed to fall back from an absent project membership row.
func (r *Repository) GetProjectOwnerWorkspace(ctx context.Context, projectID int64) (int64, int64, error) {
	row, err := r.queries.GetProjectOwnerWorkspace(ctx, projectID)
	if err != nil {
		return 0, 0, err
	}
	return row.OwnerID, row.WorkspaceID, nil
}

// GetProjectProxyFacts returns the project facts the dept-manager visibility
// proxy needs, by id alone. projectRoleFor has no business in hand, and the
// proxy must resolve the project's real business rather than trust one supplied
// by the caller.
func (r *Repository) GetProjectProxyFacts(ctx context.Context, projectID int64) (domain.Project, error) {
	row, err := r.queries.GetProjectProxyFacts(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	return domain.Project{
		ID:         projectID,
		BusinessID: row.BusinessID,
		OwnerID:    row.OwnerID,
		CreatedBy:  row.CreatedBy,
		IsPersonal: row.IsPersonal,
	}, nil
}

// GetWorkspaceOwner returns a workspace's owner. A workspace owner has no
// flow_workspace_members row of their own, so the roster read alone would miss
// them.
func (r *Repository) GetWorkspaceOwner(ctx context.Context, workspaceID int64) (int64, error) {
	return r.queries.GetWorkspaceOwner(ctx, workspaceID)
}

// ─── Project Columns ─────────────────────────────────────────────────────────
