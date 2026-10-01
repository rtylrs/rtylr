package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

// IsWorkspaceMember reports whether userID is on the workspace's member
// roster or owns it. Used by the realtime hub to authorize
// `workspace:<id>` and `project:<id>` topic subscriptions.
func (s *Service) IsWorkspaceMember(ctx context.Context, userID, workspaceID int64) bool {
	return s.repo.IsWorkspaceMember(ctx, userID, workspaceID)
}

// IsProjectMember reports whether userID can access the project — either
// because they're a project member, a workspace member of the project's
// workspace, or its owner. Used for realtime topic auth.
func (s *Service) IsProjectMember(ctx context.Context, userID, projectID int64) bool {
	return s.repo.IsProjectMember(ctx, userID, projectID)
}

// IsTaskMember reports whether userID can subscribe to a task realtime topic.
func (s *Service) IsTaskMember(ctx context.Context, userID, taskID int64) bool {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return false
	}
	if task.ProjectID != 0 {
		return s.repo.IsProjectMember(ctx, userID, task.ProjectID)
	}
	if task.WorkspaceID != 0 {
		return s.repo.IsWorkspaceMember(ctx, userID, task.WorkspaceID)
	}
	return task.OwnerID == userID
}

// ─────────────────────────────────────────────
// Authorization
// ─────────────────────────────────────────────

// projectRoleFor resolves a caller's effective role on a project, or "" when
// they hold none.
//
// A flow_project_members row is only one of three ways to reach a project:
// IsProjectMember also admits the project's owner and any member of the
// project's workspace, and neither necessarily has a row — CreateProject
// writes none for the owner. Resolving on the members table alone would deny
// both groups access they have today, so the lookup falls back through
// ownership and then the workspace roster.
func (s *Service) projectRoleFor(ctx context.Context, projectID, userID int64) string {
	if projectID == 0 || userID == 0 {
		return ""
	}
	if role, err := s.repo.GetProjectMemberRole(ctx, projectID, userID); err == nil && role != "" {
		return role
	}

	ownerID, workspaceID, err := s.repo.GetProjectOwnerWorkspace(ctx, projectID)
	if err != nil {
		return ""
	}
	if ownerID == userID {
		return domain.FlowRoleAdmin
	}
	if role := s.managerProxyRoleFor(ctx, projectID, userID); role != "" {
		return role
	}
	return s.workspaceRoleFor(ctx, workspaceID, userID)
}

// managerProxyRoleFor grants a department manager member-equivalent rights on a
// project someone reporting to them is involved in (M3), or "" when the proxy
// does not apply.
//
// Member and not admin: the manager can do what a member can, which is the
// decided scope. Admin would let them archive or transfer ownership of a
// project nobody added them to.
//
// This needs the project's business, which projectRoleFor does not carry, so it
// reads it rather than taking it on trust. A project outside the caller's
// business must not resolve through the proxy.
func (s *Service) managerProxyRoleFor(ctx context.Context, projectID, userID int64) string {
	if s.orgchart == nil {
		return ""
	}
	p, err := s.repo.GetProjectProxyFacts(ctx, projectID)
	if err != nil {
		return ""
	}
	if s.managesSomeoneOn(ctx, p, userID, p.BusinessID) {
		return domain.FlowRoleMember
	}
	return ""
}

// workspaceRoleFor resolves a caller's role on a workspace. The owner holds no
// flow_workspace_members row either, so they are resolved separately.
func (s *Service) workspaceRoleFor(ctx context.Context, workspaceID, userID int64) string {
	if workspaceID == 0 || userID == 0 {
		return ""
	}
	if ownerID, err := s.repo.GetWorkspaceOwner(ctx, workspaceID); err == nil && ownerID == userID {
		return domain.FlowRoleAdmin
	}
	role, err := s.repo.GetUserWorkspaceRole(ctx, workspaceID, userID)
	if err != nil {
		return ""
	}
	return role
}

// hubProfileBypasses reports whether central Flow access permits business-wide
// reporting visibility.
func (s *Service) hubProfileBypasses(ctx context.Context, userID, businessID int64) bool {
	if s.hubProfiles == nil {
		return false
	}
	profile, err := s.hubProfiles.GetProfile(ctx, userID, businessID)
	if err != nil {
		return false
	}
	if profile.EffectiveAccess == nil {
		return false
	}
	if profile.EffectiveAccess.IsOwner || profile.EffectiveAccess.IsAdmin {
		return true
	}
	for _, app := range profile.EffectiveAccess.Apps {
		if app.AppID != "flow" {
			continue
		}
		for _, permission := range app.Permissions {
			if permission == "flow:team:manage" {
				return true
			}
		}
	}
	return false
}

func (s *Service) isBusinessAdministrator(ctx context.Context, userID, businessID int64) bool {
	if s.hubProfiles == nil || userID <= 0 || businessID <= 0 {
		return false
	}
	profile, err := s.hubProfiles.GetProfile(ctx, userID, businessID)
	return err == nil && profile.EffectiveAccess != nil && (profile.EffectiveAccess.IsOwner || profile.EffectiveAccess.IsAdmin)
}

// RequireProjectRole reports whether userID holds at least minRole on a
// project, returning domain.ErrUnauthorized when they do not. Flow project
// membership and project scope are the only source of this authorization.
func (s *Service) RequireProjectRole(ctx context.Context, projectID, userID, businessID int64, minRole string) error {
	if s.isBusinessAdministrator(ctx, userID, businessID) {
		return nil
	}
	if domain.RoleAllows(s.projectRoleFor(ctx, projectID, userID), minRole) {
		return nil
	}
	return domain.ErrUnauthorized
}

// RequireWorkspaceRole reports whether userID holds at least minRole on a
// workspace, including the workspace owner who has no membership row.
func (s *Service) RequireWorkspaceRole(ctx context.Context, workspaceID, userID, businessID int64, minRole string) error {
	if s.isBusinessAdministrator(ctx, userID, businessID) {
		return nil
	}
	if domain.RoleAllows(s.workspaceRoleFor(ctx, workspaceID, userID), minRole) {
		return nil
	}
	return domain.ErrUnauthorized
}

// RequireTaskRole is RequireProjectRole addressed by task rather than project.
// A task need not belong to a project, so the fallback order mirrors
// IsTaskMember: project, then workspace, then sole ownership. Business-wide
// visibility does not bypass task, project, or workspace scope.
func (s *Service) RequireTaskRole(ctx context.Context, taskID, userID, businessID int64, minRole string) error {
	if taskID == 0 || userID == 0 {
		return domain.ErrUnauthorized
	}
	if s.isBusinessAdministrator(ctx, userID, businessID) {
		return nil
	}
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return domain.ErrUnauthorized
	}

	if task.ProjectID != 0 {
		return s.RequireProjectRole(ctx, task.ProjectID, userID, businessID, minRole)
	}
	if task.WorkspaceID != 0 {
		if domain.RoleAllows(s.workspaceRoleFor(ctx, task.WorkspaceID, userID), minRole) {
			return nil
		}
		return domain.ErrUnauthorized
	}
	if task.OwnerID == userID {
		return nil
	}
	return domain.ErrUnauthorized
}

// RequireCommentRole authorizes against the project owning a comment's task.
func (s *Service) RequireCommentRole(ctx context.Context, commentID, userID, businessID int64, minRole string) error {
	taskID, err := s.repo.GetCommentTaskID(ctx, commentID)
	if err != nil {
		return domain.ErrUnauthorized
	}
	return s.RequireTaskRole(ctx, taskID, userID, businessID, minRole)
}

// RequireAttachmentRole authorizes against the project owning an attachment's task.
func (s *Service) RequireAttachmentRole(ctx context.Context, attachmentID, userID, businessID int64, minRole string) error {
	taskID, err := s.repo.GetAttachmentTaskID(ctx, attachmentID)
	if err != nil {
		return domain.ErrUnauthorized
	}
	return s.RequireTaskRole(ctx, taskID, userID, businessID, minRole)
}

// RequireTimeLogRole authorizes against the project owning a time log's task.
func (s *Service) RequireTimeLogRole(ctx context.Context, timeLogID, userID, businessID int64, minRole string) error {
	taskID, err := s.repo.GetTimeLogTaskID(ctx, timeLogID)
	if err != nil {
		return domain.ErrUnauthorized
	}
	return s.RequireTaskRole(ctx, taskID, userID, businessID, minRole)
}

// ─────────────────────────────────────────────
// Projects
// ─────────────────────────────────────────────

func (s *Service) ListProjects(ctx context.Context, workspaceID, businessID, limit, offset int64) ([]domain.Project, int64, error) {
	return s.repo.GetProjectsByWorkspace(ctx, workspaceID, businessID, limit, offset)
}

func (s *Service) CreateProject(ctx context.Context, input domain.ProjectInput) (domain.Project, error) {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	// Ownerless projects break RequireProjectRole's ownership fallback, so the
	// service owns this invariant instead of trusting every caller.
	if input.OwnerID == 0 {
		input.OwnerID = input.CreatedBy
	}
	if input.Visibility == "" {
		input.Visibility = domain.VisibilityPrivate
	}
	if !domain.ValidVisibility(input.Visibility) {
		return domain.Project{}, errors.New("invalid project visibility")
	}
	id, err := s.repo.CreateProject(ctx, input)
	if err != nil {
		return domain.Project{}, err
	}

	// Seed default columns for every new project
	_ = s.repo.SeedDefaultColumns(ctx, id, input.WorkspaceID, input.CreatedBy)

	return domain.Project{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		Description: input.Description,
		Icon:        input.Icon,
		Color:       input.Color,
		Status:      "active",
		Visibility:  input.Visibility,
		IsPersonal:  input.IsPersonal,
		StartDate:   input.StartDate,
		DueDate:     input.DueDate,
		CreatedBy:   input.CreatedBy,
	}, nil
}
