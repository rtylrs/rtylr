package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// User lookup
// ─────────────────────────────────────────────

func (r *Repository) LookupUserByEmail(ctx context.Context, email string) (int64, bool, error) {
	id, err := r.queries.LookupUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		logger.Ctx(ctx).Error("LookupUserByEmail failed", logger.Error(err))
		return 0, false, err
	}
	return id, true, nil
}

func (r *Repository) LookupUserByUsername(ctx context.Context, username string, businessID int64) (int64, bool, error) {
	id, err := r.queries.LookupUserByUsername(ctx, sqlc.LookupUserByUsernameParams{
		Username:     sql.NullString{String: username, Valid: true},
		BusinessID:   sql.NullInt64{Int64: businessID, Valid: businessID != 0},
		BusinessID_2: businessID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		logger.Ctx(ctx).Error("LookupUserByUsername failed", logger.Error(err))
		return 0, false, err
	}
	return id, true, nil
}

// Membership checks (used by realtime topic auth).
// Kept cheap: a single indexed EXISTS query each; callers (the WS hub)
// run them per subscribe, which is infrequent compared to per-event dispatch.

// IsWorkspaceMember returns true when the user either owns the workspace
// or has a row in flow_workspace_members.
func (r *Repository) IsWorkspaceMember(ctx context.Context, userID, workspaceID int64) bool {
	if userID == 0 || workspaceID == 0 {
		return false
	}
	member, err := r.queries.IsWorkspaceMember(ctx, sqlc.IsWorkspaceMemberParams{
		ID:          workspaceID,
		OwnerID:     userID,
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("IsWorkspaceMember check failed", logger.Error(err))
	}
	return member.Bool
}

func (r *Repository) GetMyTasksProject(ctx context.Context, userID, workspaceID int64) (int64, string) {
	row, err := r.queries.GetMyTasksProject(ctx, sqlc.GetMyTasksProjectParams{
		OwnerID:     userID,
		WorkspaceID: workspaceID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetMyTasksProject failed", logger.Error(err))
		return 0, ""
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ""
	}
	return row.ID, row.Name
}

func (r *Repository) RenameProject(ctx context.Context, projectID, ownerID, workspaceID int64, name string) error {
	return r.queries.RenameProject(ctx, sqlc.RenameProjectParams{
		Name:        name,
		ID:          projectID,
		OwnerID:     ownerID,
		WorkspaceID: workspaceID,
	})
}

func (r *Repository) GetProjectWorkspace(ctx context.Context, projectID int64) int64 {
	workspaceID, err := r.queries.GetProjectWorkspace(ctx, projectID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetProjectWorkspace failed", logger.Error(err))
	}
	return workspaceID
}

func (r *Repository) GetColumnProject(ctx context.Context, columnID int64) int64 {
	projectID, err := r.queries.GetColumnProject(ctx, columnID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetColumnProject failed", logger.Error(err))
	}
	return projectID
}

func (r *Repository) IsProjectMember(ctx context.Context, userID, projectID int64) bool {
	if userID == 0 || projectID == 0 {
		return false
	}
	member, err := r.queries.IsProjectMember(ctx, sqlc.IsProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
		UserID_2:  userID,
		ID:        projectID,
		OwnerID:   userID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("IsProjectMember check failed", logger.Error(err))
	}
	return member.Bool
}
