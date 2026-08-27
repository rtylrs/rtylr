package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowcyclev1 "github.com/voxire/voxie/gen/go/flow/cycle/v1"
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
		Scope:       int32(c.Scope),     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Completed:   int32(c.Completed), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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
		resolved, err := a.Service.ResolveWorkspaceID(ctx, authUser.UserID, businessID, req.GetWorkspaceId())
		if err != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
		}
		if resolved != 0 {
			wsID = resolved
		}
	} else if !a.Service.IsWorkspaceMember(ctx, authUser.UserID, wsID) {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	cycles, total, err := a.Service.ListCycles(ctx, wsID, int64(limit), offset)
	if err != nil {
		return nil, workspaceInternalError(ctx, "ListCycles failed", err)
	}

	out := make([]*flowcyclev1.Cycle, 0, len(cycles))
	for _, c := range cycles {
		out = append(out, toCycle(c))
	}
	return &flowcyclev1.ListCyclesResponse{
		Cycles:     out,
		Pagination: buildPagination(page, limit, total, len(cycles)),
	}, nil
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
		resolved, err := a.Service.ResolveWorkspaceID(ctx, authUser.UserID, businessID, req.GetWorkspaceId())
		if err != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
		}
		if resolved != 0 {
			wsID = resolved
		}
	} else if !a.Service.IsWorkspaceMember(ctx, authUser.UserID, wsID) {
		return nil, onekiterrors.PermissionDenied()
	}

	var start time.Time
	if req.GetStartDate() != "" {
		parsed, parseErr := time.Parse("2006-01-02", req.GetStartDate())
		if parseErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date")
		}
		start = parsed
	}
	var end time.Time
	if req.GetEndDate() != "" {
		parsed, parseErr := time.Parse("2006-01-02", req.GetEndDate())
		if parseErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid end_date")
		}
		end = parsed
	}
	if start.IsZero() {
		// A zero time would fail the DATETIME insert (or persist a year-0001
		// row); default to today when only an end date is supplied.
		start = time.Now().UTC().Truncate(24 * time.Hour)
	}
	if end.IsZero() {
		end = start.AddDate(0, 0, 14)
	}
	if end.Before(start) {
		return nil, onekiterrors.New(http.StatusBadRequest, "end_date cannot be before start_date")
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
		return nil, workspaceInternalError(ctx, "CreateCycle failed", err)
	}

	return &flowcyclev1.CreateCycleResponse{Cycle: toCycle(cycle)}, nil
}

func (a *CycleAPI) UpdateCycle(ctx context.Context, req *flowcyclev1.UpdateCycleRequest) (*flowcyclev1.UpdateCycleResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	var end time.Time
	if req.GetEndDate() != "" {
		parsed, parseErr := time.Parse("2006-01-02", req.GetEndDate())
		if parseErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid end_date")
		}
		end = parsed
	}

	if err := a.Service.UpdateCycle(ctx, domain.UpdateCycleInput{
		ID:         req.GetId(),
		ActorID:    authUser.UserID,
		BusinessID: businessID,
		Name:       req.GetName(),
		Status:     req.GetStatus(),
		EndDate:    end,
	}); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, workspaceInternalError(ctx, "UpdateCycle failed", err)
	}
	return &flowcyclev1.UpdateCycleResponse{}, nil
}
