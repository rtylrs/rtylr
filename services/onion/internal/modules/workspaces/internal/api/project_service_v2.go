package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/wsbus"
	flowprojectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"
)

// ProjectAPIV2 serves the workspace-free flow.project.v2 contracts. Tenancy
// comes from the caller's auth context; there is no workspace id on the wire.
//
// Deliberate authorization notes:
//   - createProject stays permissive (any authenticated business user may
//     create projects). v1 had no backend gate either; this is now documented
//     policy rather than an accident.
//   - public visibility never widens read access: public projects surface as
//     metadata on profiles, not as openable boards.
type ProjectAPIV2 struct {
	Service *application.Service
}

func NewProjectAPIV2(service *application.Service) *ProjectAPIV2 {
	return &ProjectAPIV2{Service: service}
}

func resolveBusinessIDV2(ctx context.Context) (int64, *authPkg.AuthUser, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, nil, onekiterrors.PermissionDenied()
	}
	return businessID, authUser, nil
}

// publishBusinessEvent fans lifecycle events to the business-wide realtime
// topic introduced by the workspace removal, alongside the per-project topic
// subscribers already listen to.
func publishBusinessEvent(eventType string, businessID, entityID, actorID int64) {
	if businessID == 0 {
		return
	}
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("business:%d", businessID),
		Type:     eventType,
		EntityID: entityID,
		ActorID:  actorID,
	})
}

func toProjectV2(p domain.Project) *flowprojectv2.Project {
	out := &flowprojectv2.Project{
		Id:             p.ID,
		OwnerId:        p.OwnerID,
		Name:           p.Name,
		Description:    p.Description,
		Icon:           p.Icon,
		Color:          p.Color,
		Status:         p.Status,
		Visibility:     p.Visibility,
		IsPersonal:     p.IsPersonal,
		CreatedBy:      p.CreatedBy,
		TaskCount:      p.TaskCount,
		CompletedCount: p.CompletedCount,
		MemberCount:    p.MemberCount,
		Code:           p.Code,

		// Per-caller projection: set by the application layer, not stored.
		VisibleViaReport: p.VisibleViaReport,
	}
	if p.StartDate != nil {
		out.StartDate = p.StartDate.Format("2006-01-02")
	}
	if p.DueDate != nil {
		out.DueDate = p.DueDate.Format("2006-01-02")
	}
	out.CreatedAt = p.CreatedAt.Format("2006-01-02T15:04:05Z")
	out.UpdatedAt = p.UpdatedAt.Format("2006-01-02T15:04:05Z")
	if p.ArchivedAt != nil {
		s := p.ArchivedAt.Format("2006-01-02T15:04:05Z")
		out.ArchivedAt = &s
	}
	if p.PurgeAt != nil {
		d := p.PurgeAt.Format("2006-01-02T15:04:05Z")
		out.PurgeAt = &d
	}
	return out
}

// mapProjectV2Error turns a service rule into the status a caller should see,
// and returns nil for anything it does not recognise so the caller falls
// through to its own message.
//
// It used to return the unrecognised error unchanged. Because it is called as
// `if err := mapProjectV2Error(err); err != nil`, that made every caller's own
// fallback dead code and handed an unclassified error straight to the generated
// server, which renders it as a 500 with the reason discarded. That is how
// purging an unarchived project reported a server fault for a client mistake.
func mapProjectV2Error(err error) error {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return onekiterrors.PermissionDenied()
	case errors.Is(err, application.ErrProjectNotArchived):
		// 409: the project exists and the caller may act on it, but not in the
		// state it is in. Archiving it first makes the same request valid.
		return onekiterrors.New(http.StatusConflict, application.ErrProjectNotArchived.Error())
	default:
		return nil
	}
}

func (a *ProjectAPIV2) resolveBusinessID(ctx context.Context) (int64, *authPkg.AuthUser, error) {
	return resolveBusinessIDV2(ctx)
}

func (a *ProjectAPIV2) GetProject(ctx context.Context, req *flowprojectv2.GetProjectRequest) (*flowprojectv2.GetProjectResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:view"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	project, err := a.Service.GetProjectV2(ctx, req.GetId(), authUser.UserID, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, onekiterrors.New(http.StatusNotFound, "Project not found")
		}
		if err := mapProjectV2Error(err); err != nil {
			return nil, err
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get project")
	}
	return &flowprojectv2.GetProjectResponse{Project: toProjectV2(project)}, nil
}

func (a *ProjectAPIV2) ListProjects(ctx context.Context, req *flowprojectv2.ListProjectsRequest) (*flowprojectv2.ListProjectsResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:view"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	projects, total, err := a.Service.ListProjectsV2(
		ctx, authUser.UserID, businessID,
		req.GetStatus(), req.GetVisibility(), req.GetQ(),
		int64(limit), offset,
	)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list projects")
	}

	out := make([]*flowprojectv2.Project, 0, len(projects))
	for _, p := range projects {
		out = append(out, toProjectV2(p))
	}
	return &flowprojectv2.ListProjectsResponse{
		Projects:   out,
		Pagination: buildPagination(page, limit, total, len(out)),
	}, nil
}

func (a *ProjectAPIV2) CreateProject(ctx context.Context, req *flowprojectv2.CreateProjectRequest) (*flowprojectv2.CreateProjectResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:create"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	visibility := req.GetVisibility()
	if visibility == "" {
		visibility = domain.VisibilityPrivate
	}
	if !domain.ValidVisibility(visibility) {
		return nil, onekiterrors.New(http.StatusBadRequest, "visibility must be private or public")
	}

	startDate, err := parseProjectDate(req.GetStartDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date")
	}
	dueDate, err := parseProjectDate(req.GetDueDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid due_date")
	}

	input := domain.ProjectInput{
		OwnerID:     authUser.UserID,
		BusinessID:  businessID,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Icon:        req.GetIcon(),
		Color:       req.GetColor(),
		Visibility:  visibility,
		StartDate:   startDate,
		DueDate:     dueDate,
		CreatedBy:   authUser.UserID,
	}

	project, err := a.Service.CreateProject(ctx, input)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create project")
	}

	publishBusinessEvent("project.created", businessID, project.ID, authUser.UserID)
	return &flowprojectv2.CreateProjectResponse{Project: toProjectV2(project)}, nil
}

func (a *ProjectAPIV2) UpdateProject(ctx context.Context, req *flowprojectv2.UpdateProjectRequest) (*flowprojectv2.UpdateProjectResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	body := req.GetBody()
	startDate, err := parseProjectDate(body.GetStartDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date")
	}
	dueDate, err := parseProjectDate(body.GetDueDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid due_date")
	}

	status := body.GetStatus()
	if status == "" {
		status = "active"
	}

	input := domain.UpdateProjectInput{
		ID:          req.GetId(),
		OwnerID:     authUser.UserID,
		BusinessID:  businessID,
		Name:        body.GetName(),
		Description: body.GetDescription(),
		Icon:        body.GetIcon(),
		Color:       body.GetColor(),
		Status:      status,
		StartDate:   startDate,
		DueDate:     dueDate,
	}

	if err := a.Service.UpdateProject(ctx, input); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update project")
	}

	publishBusinessEvent("project.updated", businessID, req.GetId(), authUser.UserID)
	return &flowprojectv2.UpdateProjectResponse{Message: "Project updated successfully"}, nil
}

func (a *ProjectAPIV2) ArchiveProject(ctx context.Context, req *flowprojectv2.ArchiveProjectRequest) (*flowprojectv2.ArchiveProjectResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.DeleteProject(ctx, req.GetId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to archive project")
	}

	project, err := a.Service.GetProjectV2(ctx, req.GetId(), authUser.UserID, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to read archived project")
	}

	publishBusinessEvent("project.archived", businessID, req.GetId(), authUser.UserID)
	return &flowprojectv2.ArchiveProjectResponse{Project: toProjectV2(project)}, nil
}

func (a *ProjectAPIV2) SetProjectVisibility(ctx context.Context, req *flowprojectv2.SetProjectVisibilityRequest) (*flowprojectv2.SetProjectVisibilityResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	project, err := a.Service.SetProjectVisibility(ctx, req.GetId(), authUser.UserID, businessID, req.GetVisibility())
	if err != nil {
		if err := mapProjectV2Error(err); err != nil {
			return nil, err
		}
		return nil, onekiterrors.New(http.StatusBadRequest, "Failed to set visibility")
	}

	publishBusinessEvent("project.updated", businessID, req.GetId(), authUser.UserID)
	return &flowprojectv2.SetProjectVisibilityResponse{Project: toProjectV2(project)}, nil
}

func (a *ProjectAPIV2) ListArchivedProjects(ctx context.Context, req *flowprojectv2.ListArchivedProjectsRequest) (*flowprojectv2.ListArchivedProjectsResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:view"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	projects, total, err := a.Service.ListArchivedProjectsV2(ctx, authUser.UserID, businessID, int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list archived projects")
	}

	out := make([]*flowprojectv2.Project, 0, len(projects))
	for _, p := range projects {
		out = append(out, toProjectV2(p))
	}
	return &flowprojectv2.ListArchivedProjectsResponse{
		Projects:   out,
		Pagination: buildPagination(page, limit, total, len(out)),
	}, nil
}

func (a *ProjectAPIV2) RestoreProject(ctx context.Context, req *flowprojectv2.RestoreProjectRequest) (*flowprojectv2.RestoreProjectResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	project, err := a.Service.RestoreProjectV2(ctx, req.GetId(), authUser.UserID, businessID)
	if err != nil {
		if err := mapProjectV2Error(err); err != nil {
			return nil, err
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to restore project")
	}

	publishBusinessEvent("project.restored", businessID, req.GetId(), authUser.UserID)
	return &flowprojectv2.RestoreProjectResponse{Project: toProjectV2(project)}, nil
}

func (a *ProjectAPIV2) PurgeProject(ctx context.Context, req *flowprojectv2.PurgeProjectRequest) (*flowprojectv2.PurgeProjectResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	urls, err := a.Service.PurgeProjectV2(ctx, req.GetId(), authUser.UserID, businessID)
	if err != nil {
		if err := mapProjectV2Error(err); err != nil {
			return nil, err
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to purge project")
	}

	// Rows are gone; the stored files are not yet. No shared object-delete
	// helper exists today, so the URLs go to the log where the upcoming
	// storage-sweep worker picks them up. Tracked as Phase 4 work alongside
	// the column drops.
	logger.Ctx(ctx).Warn("purged project left attachment files for sweep",
		logger.Int("files", len(urls)),
		logger.Int("project_id", int(req.GetId())),
	)

	publishBusinessEvent("project.purged", businessID, req.GetId(), authUser.UserID)
	return &flowprojectv2.PurgeProjectResponse{Message: "Project permanently deleted"}, nil
}

func toProjectColumnV2(c domain.ProjectColumn) *flowprojectv2.ProjectColumn {
	return &flowprojectv2.ProjectColumn{
		Id:        c.ID,
		ProjectId: c.ProjectID,
		Name:      c.Name,
		Color:     c.Color,
		SortOrder: int32(c.SortOrder), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toProjectMemberV2(m domain.ProjectMember) *flowprojectv2.ProjectMember {
	role, ok := domain.CanonicalFlowMemberRole(m.Role)
	if !ok {
		role = m.Role
	}
	out := &flowprojectv2.ProjectMember{
		Id:        m.ID,
		ProjectId: m.ProjectID,
		UserId:    m.UserID,
		Name:      m.Name,
		Email:     m.Email,
		Role:      role,
		AddedAt:   m.AddedAt.Format("2006-01-02T15:04:05Z"),
	}
	if m.AddedBy != nil {
		out.AddedBy = *m.AddedBy
	}
	return out
}

func (a *ProjectAPIV2) ListProjectColumns(ctx context.Context, req *flowprojectv2.ListProjectColumnsRequest) (*flowprojectv2.ListProjectColumnsResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:view"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleViewer); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	cols, total, err := a.Service.GetColumnsByProject(ctx, req.GetProjectId(), int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list columns")
	}
	out := make([]*flowprojectv2.ProjectColumn, 0, len(cols))
	for _, c := range cols {
		out = append(out, toProjectColumnV2(c))
	}
	return &flowprojectv2.ListProjectColumnsResponse{
		Columns:    out,
		Pagination: buildPagination(page, limit, total, len(cols)),
	}, nil
}

func (a *ProjectAPIV2) CreateProjectColumn(ctx context.Context, req *flowprojectv2.CreateProjectColumnRequest) (*flowprojectv2.CreateProjectColumnResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleMember); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	// The columns table still carries a NOT NULL workspace_id until Phase 4
	// drops it; the server derives it from the project instead of trusting a
	// client echo.
	workspaceID := a.Service.GetProjectWorkspace(ctx, req.GetProjectId())

	color := req.GetColor()
	if color == "" {
		color = "#6b7280"
	}

	col, err := a.Service.CreateColumn(ctx, domain.ProjectColumnInput{
		ProjectID:   req.GetProjectId(),
		WorkspaceID: workspaceID,
		Name:        req.GetName(),
		Color:       color,
		SortOrder:   int(req.GetSortOrder()),
		CreatedBy:   authUser.UserID,
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create column")
	}
	publishColumnEvent("column.created", col.ProjectID, col.ID, authUser.UserID)
	publishBusinessEvent("column.created", businessID, col.ProjectID, authUser.UserID)
	return &flowprojectv2.CreateProjectColumnResponse{Column: toProjectColumnV2(col)}, nil
}

func (a *ProjectAPIV2) UpdateProjectColumn(ctx context.Context, req *flowprojectv2.UpdateProjectColumnRequest) (*flowprojectv2.UpdateProjectColumnResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	err := a.Service.UpdateColumn(ctx, domain.UpdateProjectColumnInput{
		ID:        req.GetId(),
		ActorID:   authUser.UserID,
		Name:      req.GetName(),
		Color:     req.GetColor(),
		SortOrder: int(req.GetSortOrder()),
	})
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update column")
	}
	projectID := a.Service.GetColumnProject(ctx, req.GetId())
	businessID, _, bizErr := a.resolveBusinessID(ctx)
	if bizErr == nil {
		publishBusinessEvent("column.updated", businessID, projectID, authUser.UserID)
	}
	publishColumnEvent("column.updated", projectID, req.GetId(), authUser.UserID)
	return &flowprojectv2.UpdateProjectColumnResponse{Message: "Column updated"}, nil
}

func (a *ProjectAPIV2) DeleteProjectColumn(ctx context.Context, req *flowprojectv2.DeleteProjectColumnRequest) (*flowprojectv2.DeleteProjectColumnResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	projectID := a.Service.GetColumnProject(ctx, req.GetId())
	if err := a.Service.DeleteColumn(ctx, req.GetId(), authUser.UserID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete column")
	}
	businessID, _, bizErr := a.resolveBusinessID(ctx)
	if bizErr == nil {
		publishBusinessEvent("column.deleted", businessID, projectID, authUser.UserID)
	}
	publishColumnEvent("column.deleted", projectID, req.GetId(), authUser.UserID)
	return &flowprojectv2.DeleteProjectColumnResponse{Message: "Column deleted"}, nil
}

func (a *ProjectAPIV2) MoveTaskToColumn(ctx context.Context, req *flowprojectv2.MoveTaskToColumnRequest) (*flowprojectv2.MoveTaskToColumnResponse, error) {
	businessID, authUser, err := resolveBusinessIDV2(ctx)
	if err != nil {
		return nil, err
	}

	// This handler had no role check of its own and leaned entirely on the
	// workspace guard inside the UPDATE. That guard cannot pass for a v2 task
	// (workspace_id = 0), so the endpoint always failed; fixing the statement
	// without adding a check here would have turned "always denied" into
	// "never checked". RequireTaskRole resolves a v2 task through its project.
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), authUser.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.MoveTaskToColumn(ctx, req.GetTaskId(), req.GetColumnId(), authUser.UserID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to move task")
	}
	projectID := a.Service.GetColumnProject(ctx, req.GetColumnId())
	publishTaskEventByProject("task.updated", projectID, req.GetTaskId(), authUser.UserID)
	return &flowprojectv2.MoveTaskToColumnResponse{Message: "Task moved"}, nil
}

func (a *ProjectAPIV2) ListProjectMembers(ctx context.Context, req *flowprojectv2.ListProjectMembersRequest) (*flowprojectv2.ListProjectMembersResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:view"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsProjectMemberOrBypass(ctx, authUser.UserID, businessID, req.GetProjectId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	members, total, err := a.Service.ListProjectMembers(ctx, req.GetProjectId(), int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list project members")
	}
	out := make([]*flowprojectv2.ProjectMember, 0, len(members))
	for _, m := range members {
		out = append(out, toProjectMemberV2(m))
	}
	return &flowprojectv2.ListProjectMembersResponse{
		Members:    out,
		Pagination: buildPagination(page, limit, total, len(members)),
	}, nil
}

func (a *ProjectAPIV2) AddProjectMember(ctx context.Context, req *flowprojectv2.AddProjectMemberRequest) (*flowprojectv2.AddProjectMemberResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:assign"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	role := req.GetRole()
	if role == "" {
		role = "member"
	}
	canonicalRole, ok := domain.CanonicalFlowMemberRole(role)
	if !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "role must be admin, member, or viewer")
	}
	if err := a.Service.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: req.GetProjectId(),
		UserID:    req.GetUserId(),
		Role:      canonicalRole,
		AddedBy:   authUser.UserID,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add project member")
	}
	return &flowprojectv2.AddProjectMemberResponse{Message: "Member added"}, nil
}

func (a *ProjectAPIV2) RemoveProjectMember(ctx context.Context, req *flowprojectv2.RemoveProjectMemberRequest) (*flowprojectv2.RemoveProjectMemberResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:assign"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.RemoveProjectMember(ctx, domain.RemoveProjectMemberInput{
		ProjectID: req.GetProjectId(),
		UserID:    req.GetUserId(),
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to remove project member")
	}
	return &flowprojectv2.RemoveProjectMemberResponse{Message: "Member removed"}, nil
}

func (a *ProjectAPIV2) UpdateProjectMemberRole(ctx context.Context, req *flowprojectv2.UpdateProjectMemberRoleRequest) (*flowprojectv2.UpdateProjectMemberRoleResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:assign"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	role, ok := domain.CanonicalFlowMemberRole(req.GetRole())
	if !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "role must be admin, member, or viewer")
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.UpdateProjectMemberRole(ctx, domain.UpdateProjectMemberRoleInput{
		ProjectID: req.GetProjectId(),
		UserID:    req.GetUserId(),
		Role:      role,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update role")
	}
	return &flowprojectv2.UpdateProjectMemberRoleResponse{Message: "Role updated"}, nil
}

func (a *ProjectAPIV2) InviteProjectMember(ctx context.Context, req *flowprojectv2.InviteProjectMemberRequest) (*flowprojectv2.InviteProjectMemberResponse, error) {
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	role := req.GetRole()
	if role == "" {
		role = "member"
	}
	canonicalRole, ok := domain.CanonicalFlowMemberRole(role)
	if !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "role must be admin, member, or viewer")
	}
	found, err := a.Service.InviteProjectMemberByEmail(ctx, req.GetProjectId(), req.GetEmail(), canonicalRole, authUser.UserID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to invite member")
	}
	if !found {
		return nil, onekiterrors.New(http.StatusNotFound, "No user found with that email")
	}
	return &flowprojectv2.InviteProjectMemberResponse{Message: "Member invited"}, nil
}

func (a *ProjectAPIV2) TransferProjectOwnership(ctx context.Context, req *flowprojectv2.TransferProjectOwnershipRequest) (*flowprojectv2.TransferProjectOwnershipResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.TransferProjectOwnership(ctx, domain.TransferProjectOwnershipInput{
		ProjectID:  req.GetProjectId(),
		NewOwnerID: req.GetNewOwnerId(),
		CallerID:   authUser.UserID,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to transfer ownership")
	}
	return &flowprojectv2.TransferProjectOwnershipResponse{Message: "Ownership transferred"}, nil
}

// GetProjectLabourCost returns the labour cost of billable logged time on a
// project. Admin only: it exposes aggregate spend, and that aggregate comes from
// HR rates which must not become readable through Flow.
//
// Ported from the v1 handler unchanged in behaviour. The zero-id check v1 did by
// hand is gone because the v2 contract declares id @required @gt(0), so a zero
// never reaches here.
func (a *ProjectAPIV2) GetProjectLabourCost(ctx context.Context, req *flowprojectv2.GetProjectLabourCostRequest) (*flowprojectv2.GetProjectLabourCostResponse, error) {
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	projectID := req.GetId()
	if roleErr := a.Service.RequireProjectRole(ctx, projectID, authUser.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	cost, costErr := a.Service.GetProjectLabourCost(ctx, businessID, projectID)
	if costErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load project labour cost")
	}

	// UncostedEmployeeCount is wider than the contract field. Clamping rather
	// than letting the conversion wrap keeps a large count reading as "very
	// many" instead of flipping negative.
	uncostedEmployees := cost.UncostedEmployeeCount
	if uncostedEmployees > math.MaxInt32 {
		uncostedEmployees = math.MaxInt32
	}

	return &flowprojectv2.GetProjectLabourCostResponse{
		Cost: &flowprojectv2.ProjectLabourCost{
			ProjectId:             cost.ProjectID,
			CostMinorUnits:        cost.CostCents,
			CostedMinutes:         cost.CostedMinutes,
			UncostedMinutes:       cost.UncostedMinutes,
			UncostedEmployeeCount: int32(uncostedEmployees),
		},
	}, nil
}
