package api

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowprojectv1 "github.com/voxire/voxie/gen/go/flow/project/v1"
)

type ProjectAPI struct {
	Service *application.Service
}

func NewProjectAPI(service *application.Service) *ProjectAPI {
	return &ProjectAPI{Service: service}
}

func toProject(p domain.Project) *flowprojectv1.Project {
	proj := &flowprojectv1.Project{
		Id:             p.ID,
		WorkspaceId:    p.WorkspaceID,
		OwnerId:        p.OwnerID,
		Name:           p.Name,
		Description:    p.Description,
		Icon:           p.Icon,
		Color:          p.Color,
		Status:         p.Status,
		CreatedBy:      p.CreatedBy,
		CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		TaskCount:      p.TaskCount,
		CompletedCount: p.CompletedCount,
		MemberCount:    p.MemberCount,
		Code:           p.Code,
	}
	if p.StartDate != nil {
		proj.StartDate = p.StartDate.Format("2006-01-02")
	}
	if p.DueDate != nil {
		proj.DueDate = p.DueDate.Format("2006-01-02")
	}
	return proj
}

func parseProjectDate(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (a *ProjectAPI) resolveBusinessID(ctx context.Context) (int64, *authPkg.AuthUser, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, nil, onekiterrors.PermissionDenied()
	}
	return businessID, authUser, nil
}

func (a *ProjectAPI) resolveWorkspaceID(ctx context.Context, userID, businessID, workspaceID int64) (int64, error) {
	raw := strconv.FormatInt(workspaceID, 10)
	if workspaceID == 0 {
		raw = "personal"
	}
	resolved, err := a.Service.ResolveWorkspaceID(ctx, userID, businessID, raw)
	if err != nil || resolved == 0 {
		return 0, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
	}
	return resolved, nil
}

func (a *ProjectAPI) ListProjects(ctx context.Context, req *flowprojectv1.ListProjectsRequest) (*flowprojectv1.ListProjectsResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:view"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	workspaceID, err := a.resolveWorkspaceID(ctx, authUser.UserID, businessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	projects, total, err := a.Service.ListProjects(ctx, workspaceID, businessID, int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list projects")
	}

	out := make([]*flowprojectv1.Project, 0, len(projects))
	for _, p := range projects {
		out = append(out, toProject(p))
	}
	return &flowprojectv1.ListProjectsResponse{
		Projects:   out,
		Pagination: buildPagination(page, limit, total, len(projects)),
	}, nil
}

func (a *ProjectAPI) CreateProject(ctx context.Context, req *flowprojectv1.CreateProjectRequest) (*flowprojectv1.CreateProjectResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:create"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	workspaceID, err := a.resolveWorkspaceID(ctx, authUser.UserID, businessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
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
		WorkspaceID: workspaceID,
		OwnerID:     authUser.UserID,
		BusinessID:  businessID,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Icon:        req.GetIcon(),
		Color:       req.GetColor(),
		StartDate:   startDate,
		DueDate:     dueDate,
		CreatedBy:   authUser.UserID,
	}

	project, err := a.Service.CreateProject(ctx, input)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create project")
	}

	return &flowprojectv1.CreateProjectResponse{Project: toProject(project)}, nil
}

func (a *ProjectAPI) UpdateProject(ctx context.Context, req *flowprojectv1.UpdateProjectRequest) (*flowprojectv1.UpdateProjectResponse, error) {
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

	return &flowprojectv1.UpdateProjectResponse{Message: "Project updated successfully"}, nil
}

func (a *ProjectAPI) DeleteProject(ctx context.Context, req *flowprojectv1.DeleteProjectRequest) (*flowprojectv1.DeleteProjectResponse, error) {
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
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete project")
	}

	return &flowprojectv1.DeleteProjectResponse{Message: "Project archived successfully"}, nil
}

// GetProjectLabourCost serves the aggregate declared by flow.project.v1. The
// figures come from the HR-owned labour-cost port (wired in cmd/api/main.go);
// employee hourly rates never cross that boundary, only money and coverage
// totals. UncostedMinutes stays separate from CostMinorUnits so billable time
// belonging to an employee with no configured rate is visible as missing
// coverage rather than silently counted as free work.
//
// Gated at admin: the workspace-replacement proposal's permission matrix does
// not cover labour cost, and a project-level financial total is closer to the
// billing surfaces than to ordinary project reads, so this fails closed.
func (a *ProjectAPI) GetProjectLabourCost(ctx context.Context, req *flowprojectv1.GetProjectLabourCostRequest) (*flowprojectv1.GetProjectLabourCostResponse, error) {
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	projectID := req.GetId()
	if projectID == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	if roleErr := a.Service.RequireProjectRole(ctx, projectID, authUser.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	cost, costErr := a.Service.GetProjectLabourCost(ctx, businessID, projectID)
	if costErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load project labour cost")
	}

	uncostedEmployees := cost.UncostedEmployeeCount
	if uncostedEmployees > math.MaxInt32 {
		uncostedEmployees = math.MaxInt32
	}

	return &flowprojectv1.GetProjectLabourCostResponse{
		Cost: &flowprojectv1.ProjectLabourCost{
			ProjectId:             cost.ProjectID,
			CostMinorUnits:        cost.CostCents,
			CostedMinutes:         cost.CostedMinutes,
			UncostedMinutes:       cost.UncostedMinutes,
			UncostedEmployeeCount: int32(uncostedEmployees),
		},
	}, nil
}
