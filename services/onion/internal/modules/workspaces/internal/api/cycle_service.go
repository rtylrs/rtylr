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

const workspaceLookupPageSize int64 = 500

// listBusinessWorkspaces is used by the resource-level lookup paths below. The
// legacy v1 cycle and label contracts do not carry a workspace id on update,
// while the service only exposes paginated workspace/cycle reads. Resolve the
// owning workspace through those reads before applying a role check or merging
// omitted fields into a partial update.
func listBusinessWorkspaces(ctx context.Context, service *application.Service, businessID int64) ([]domain.Workspace, error) {
	workspaces := make([]domain.Workspace, 0)
	for page := int64(1); ; page++ {
		pageWorkspaces, total, err := service.ListWorkspaces(ctx, businessID, page, workspaceLookupPageSize)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, pageWorkspaces...)
		if len(pageWorkspaces) == 0 || int64(len(workspaces)) >= total {
			return workspaces, nil
		}
	}
}

func findCycleInBusiness(ctx context.Context, service *application.Service, businessID, cycleID int64) (domain.Cycle, error) {
	workspaces, err := listBusinessWorkspaces(ctx, service, businessID)
	if err != nil {
		return domain.Cycle{}, err
	}

	for _, workspace := range workspaces {
		for offset := int64(0); ; offset += workspaceLookupPageSize {
			cycles, total, listErr := service.ListCycles(ctx, workspace.ID, workspaceLookupPageSize, offset)
			if listErr != nil {
				return domain.Cycle{}, listErr
			}
			for _, cycle := range cycles {
				if cycle.ID == cycleID {
					return cycle, nil
				}
			}
			if len(cycles) == 0 || offset+int64(len(cycles)) >= total {
				break
			}
		}
	}

	// Do not disclose whether an id exists in another business. The repository
	// uses the same unauthorized sentinel for an out-of-scope write.
	return domain.Cycle{}, domain.ErrUnauthorized
}

func (a *CycleAPI) resolveCycleWorkspaceID(ctx context.Context, userID, businessID int64, requested string) (int64, error) {
	// An omitted workspace means the caller's personal workspace, not a
	// business id used in the independent workspace-id sequence.
	if requested == "" {
		requested = "personal"
	}
	resolved, err := a.Service.ResolveWorkspaceID(ctx, userID, businessID, requested)
	if err != nil || resolved == 0 {
		return 0, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
	}
	return resolved, nil
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

	wsID, err := a.resolveCycleWorkspaceID(ctx, authUser.UserID, businessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
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

	wsID, err := a.resolveCycleWorkspaceID(ctx, authUser.UserID, businessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireWorkspaceRole(ctx, wsID, authUser.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
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

	existing, lookupErr := findCycleInBusiness(ctx, a.Service, businessID, req.GetId())
	if lookupErr != nil {
		if errors.Is(lookupErr, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, workspaceInternalError(ctx, "Load cycle for update failed", lookupErr)
	}
	var roleErr error
	if existing.ProjectID != nil {
		roleErr = a.Service.RequireProjectRole(ctx, *existing.ProjectID, authUser.UserID, businessID, domain.FlowRoleMember)
	} else {
		roleErr = a.Service.RequireWorkspaceRole(ctx, existing.WorkspaceID, authUser.UserID, businessID, domain.FlowRoleMember)
	}
	if roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if !end.IsZero() && !existing.StartDate.IsZero() && end.Before(existing.StartDate) {
		return nil, onekiterrors.New(http.StatusBadRequest, "end_date cannot be before start_date")
	}

	name := req.GetName()
	if name == "" {
		// The generated v1 request uses scalar strings, so an omitted name and
		// an explicit empty string are indistinguishable. Preserve the stored
		// value rather than allowing a status-only update to erase it.
		name = existing.Name
	}

	if err := a.Service.UpdateCycle(ctx, domain.UpdateCycleInput{
		ID:         req.GetId(),
		ActorID:    authUser.UserID,
		BusinessID: businessID,
		Name:       name,
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
