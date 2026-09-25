package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowprojectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"
)

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
