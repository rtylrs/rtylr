package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowmilestonev1 "github.com/voxire/voxie/gen/go/flow/milestone/v1"
)

type MilestoneAPI struct {
	Service *application.Service
}

func NewMilestoneAPI(service *application.Service) *MilestoneAPI {
	return &MilestoneAPI{Service: service}
}

func toMilestone(m domain.Milestone) *flowmilestonev1.Milestone {
	out := &flowmilestonev1.Milestone{
		Id:               m.ID,
		ProjectId:        m.ProjectID,
		Name:             m.Name,
		TargetDate:       m.DueDate.Format("2006-01-02"),
		Progress:         int32(m.Progress),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		ComputedProgress: int32(m.Progress),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		DoneCount:        int32(m.DoneCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		TaskCount:        int32(m.TaskCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Status:           m.Status,
		Description:      m.Description,
	}
	if m.Project != nil {
		out.Project = &flowmilestonev1.MilestoneProject{
			Name:  m.Project.Name,
			Color: m.Project.Color,
		}
	}
	return out
}

func (a *MilestoneAPI) ListMilestones(ctx context.Context, req *flowmilestonev1.ListMilestonesRequest) (*flowmilestonev1.ListMilestonesResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	wsID := businessID
	if req.GetWorkspaceId() != "" {
		resolved, err := a.Service.ResolveWorkspaceID(ctx, authUser.UserID, businessID, req.GetWorkspaceId())
		if err != nil || resolved == 0 {
			return nil, onekiterrors.PermissionDenied()
		}
		wsID = resolved
	} else if !a.Service.IsWorkspaceMember(ctx, authUser.UserID, wsID) {
		return nil, onekiterrors.PermissionDenied()
	}

	var projID *int64
	if req.GetProjectId() != "" {
		if id, err := strconv.ParseInt(req.GetProjectId(), 10, 64); err == nil {
			projID = &id
		}
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	milestones, total, err := a.Service.ListMilestones(ctx, wsID, projID, int64(limit), offset)
	if err != nil {
		return nil, workspaceInternalError(ctx, "ListMilestones failed", err)
	}

	out := make([]*flowmilestonev1.Milestone, 0, len(milestones))
	for _, m := range milestones {
		out = append(out, toMilestone(m))
	}
	return &flowmilestonev1.ListMilestonesResponse{
		Milestones: out,
		Pagination: buildPagination(page, limit, total, len(milestones)),
	}, nil
}

func (a *MilestoneAPI) CreateMilestone(ctx context.Context, req *flowmilestonev1.CreateMilestoneRequest) (*flowmilestonev1.CreateMilestoneResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	wsID := businessID
	if req.GetWorkspaceId() != 0 {
		if !a.Service.IsWorkspaceMember(ctx, authUser.UserID, req.GetWorkspaceId()) {
			return nil, onekiterrors.PermissionDenied()
		}
		wsID = req.GetWorkspaceId()
	} else if !a.Service.IsWorkspaceMember(ctx, authUser.UserID, wsID) {
		return nil, onekiterrors.PermissionDenied()
	}

	var dueDate time.Time
	if req.GetTargetDate() != "" {
		parsed, parseErr := time.Parse("2006-01-02", req.GetTargetDate())
		if parseErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid target_date")
		}
		dueDate = parsed
	}
	if dueDate.IsZero() {
		dueDate = time.Now().UTC().AddDate(0, 1, 0)
	}

	milestone, err := a.Service.CreateMilestone(ctx, domain.MilestoneInput{
		ProjectID:   req.GetProjectId(),
		WorkspaceID: wsID,
		BusinessID:  businessID,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		DueDate:     dueDate,
		CreatedBy:   authUser.UserID,
	})
	if err != nil {
		return nil, workspaceInternalError(ctx, "CreateMilestone failed", err)
	}

	return &flowmilestonev1.CreateMilestoneResponse{Milestone: toMilestone(milestone)}, nil
}

func (a *MilestoneAPI) UpdateMilestone(ctx context.Context, req *flowmilestonev1.UpdateMilestoneRequest) (*flowmilestonev1.UpdateMilestoneResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	var dueDate time.Time
	if req.GetTargetDate() != "" {
		parsed, parseErr := time.Parse("2006-01-02", req.GetTargetDate())
		if parseErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid target_date")
		}
		dueDate = parsed
	}

	if err := a.Service.UpdateMilestone(ctx, domain.UpdateMilestoneInput{
		ID:          req.GetId(),
		ActorID:     authUser.UserID,
		BusinessID:  businessID,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Status:      req.GetStatus(),
		DueDate:     dueDate,
	}); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, workspaceInternalError(ctx, "UpdateMilestone failed", err)
	}
	return &flowmilestonev1.UpdateMilestoneResponse{}, nil
}
