package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
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
