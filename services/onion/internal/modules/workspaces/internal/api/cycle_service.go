package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	flowcyclev1 "github.com/voxire/voxie/gen/go/flow/cycle/v1"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type CycleAPI struct {
	Service *application.Service
}

func NewCycleAPI(service *application.Service) *CycleAPI {
	return &CycleAPI{Service: service}
}

func toCycle(c domain.Cycle) *flowcyclev1.Cycle {
	return &flowcyclev1.Cycle{
		Id:          c.ID,
		WorkspaceId: c.WorkspaceID,
		Name:        c.Name,
		StartDate:   c.StartDate.Format("2006-01-02"),
		EndDate:     c.EndDate.Format("2006-01-02"),
		Status:      c.Status,
		Scope:       int32(c.Scope),
		Completed:   int32(c.Completed),
	}
}

func (a *CycleAPI) ListCycles(ctx context.Context, req *flowcyclev1.ListCyclesRequest) (*flowcyclev1.ListCyclesResponse, error) {
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
		resolved, err := a.Service.ResolveWorkspaceID(ctx, authUser.UserID, req.GetWorkspaceId())
		if err != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
		}
		if resolved != 0 {
			wsID = resolved
		}
	}

	cycles, err := a.Service.ListCycles(ctx, wsID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, err.Error())
	}

	out := make([]*flowcyclev1.Cycle, 0, len(cycles))
	for _, c := range cycles {
		out = append(out, toCycle(c))
	}
	return &flowcyclev1.ListCyclesResponse{Cycles: out}, nil
}

func (a *CycleAPI) CreateCycle(ctx context.Context, req *flowcyclev1.CreateCycleRequest) (*flowcyclev1.CreateCycleResponse, error) {
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
		resolved, err := a.Service.ResolveWorkspaceID(ctx, authUser.UserID, req.GetWorkspaceId())
		if err != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
		}
		if resolved != 0 {
			wsID = resolved
		}
	}

	start, _ := time.Parse("2006-01-02", req.GetStartDate())
	end, _ := time.Parse("2006-01-02", req.GetEndDate())
	if end.IsZero() {
		end = start.AddDate(0, 0, 14)
	}

	cycle, err := a.Service.CreateCycle(ctx, domain.CycleInput{
		BusinessID:  businessID,
		WorkspaceID: wsID,
		Name:        req.GetName(),
		StartDate:   start,
		EndDate:     end,
		CreatedBy:   authUser.UserID,
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, err.Error())
	}

	return &flowcyclev1.CreateCycleResponse{Cycle: toCycle(cycle)}, nil
}

func (a *CycleAPI) UpdateCycle(ctx context.Context, req *flowcyclev1.UpdateCycleRequest) (*flowcyclev1.UpdateCycleResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	_, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	end, _ := time.Parse("2006-01-02", req.GetEndDate())

	if err := a.Service.UpdateCycle(ctx, domain.UpdateCycleInput{
		ID:      req.GetId(),
		ActorID: authUser.UserID,
		Name:    req.GetName(),
		Status:  req.GetStatus(),
		EndDate: end,
	}); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, err.Error())
	}
	return &flowcyclev1.UpdateCycleResponse{}, nil
}
