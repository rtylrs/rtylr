package api

import (
	"context"
	"errors"
	"math"
	"net/http"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowprojectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"
)

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

	if err := a.Service.MoveTaskToColumnOrdered(ctx, req.GetTaskId(), req.GetColumnId(), authUser.UserID, req.GetOrderedTaskIds()); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrInvalidTaskOrder) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid ordered_task_ids")
		}
		if errors.Is(err, domain.ErrTaskOrderOutsideProject) {
			return nil, onekiterrors.New(http.StatusBadRequest, "ordered_task_ids must belong to the task's project")
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
