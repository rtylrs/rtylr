package api

import (
	"context"
	"errors"
	"net/http"

	flowstatusv2 "github.com/voxire/voxie/gen/go/flow/status/v2"

	"github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

// StatusAPIV2 serves F-A's per-project custom statuses.
type StatusAPIV2 struct{ Service *application.Service }

func NewStatusAPIV2(s *application.Service) *StatusAPIV2 { return &StatusAPIV2{Service: s} }

func toStatusPayload(s domain.Status) *flowstatusv2.Status {
	return &flowstatusv2.Status{
		Id:           s.ID,
		ProjectId:    s.ProjectID,
		Name:         s.Name,
		Bucket:       s.Bucket,
		LegacyStatus: s.LegacyStatus,
		Color:        s.Color,
		SortOrder:    s.SortOrder,
		IsDefault:    s.IsDefault,
		CreatedAt:    s.CreatedAt,
	}
}

// statusError maps the domain errors that are the caller's fault onto the right
// status code.
//
// The validation errors have to be 400 and the duplicate name has to be 409. A
// blanket 500 would tell the user "something went wrong" for a typo in a bucket
// name, which is the same unhelpful outcome CreateCycle had before onion#923
// mapped ErrInvalidCycleDates.
//
// The messages are written out rather than taken from err.Error(), which
// TestWorkspaceHandlersDoNotExposeRawErrors forbids in this package. The rule is
// lexical and blunt on purpose: these particular errors are safe to show, but
// nothing stops a later edit to a domain error from leaking something, and the
// guard cannot tell the difference.
func statusError(err error, fallback string) error {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return onekiterrors.PermissionDenied()
	case errors.Is(err, domain.ErrStatusNotFound):
		return onekiterrors.New(http.StatusNotFound, "Status not found")
	case errors.Is(err, domain.ErrStatusNameTaken):
		return onekiterrors.New(http.StatusConflict, "A status with that name already exists in this project")
	case errors.Is(err, domain.ErrInvalidStatusBucket):
		return onekiterrors.New(http.StatusBadRequest, "bucket must be one of backlog, active, done, cancelled")
	case errors.Is(err, domain.ErrInvalidLegacyStatus):
		return onekiterrors.New(http.StatusBadRequest, "legacy_status must be one of backlog, todo, in_progress, in_review, done, cancelled")
	case errors.Is(err, domain.ErrStatusNameRequired):
		return onekiterrors.New(http.StatusBadRequest, "Status name is required")
	}
	return onekiterrors.New(http.StatusInternalServerError, fallback)
}

func (a *StatusAPIV2) ListStatuses(ctx context.Context, req *flowstatusv2.ListStatusesRequest) (*flowstatusv2.ListStatusesResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	statuses, err := a.Service.ListProjectStatuses(ctx, req.GetProjectId(), caller.UserID, businessID)
	if err != nil {
		return nil, statusError(err, "Failed to list statuses")
	}
	// An empty set is the answer for an unconfigured project, so the slice is
	// always non-nil: a JSON null here would make every consumer write its own
	// null check to reach the same "use the built-ins" conclusion.
	out := make([]*flowstatusv2.Status, 0, len(statuses))
	for _, s := range statuses {
		out = append(out, toStatusPayload(s))
	}
	return &flowstatusv2.ListStatusesResponse{Statuses: out}, nil
}

func (a *StatusAPIV2) CreateStatus(ctx context.Context, req *flowstatusv2.CreateStatusRequest) (*flowstatusv2.CreateStatusResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	status, err := a.Service.CreateProjectStatus(ctx, caller.UserID, businessID, domain.StatusInput{
		ProjectID:    req.GetProjectId(),
		Name:         req.GetName(),
		Bucket:       req.GetBucket(),
		LegacyStatus: req.GetLegacyStatus(),
		Color:        req.GetColor(),
		SortOrder:    req.GetSortOrder(),
		// A caller that omits sort_order gets appended. Zero is a legitimate
		// position, so "was it supplied" cannot be inferred from the value.
		SortOrderSet: req.GetSortOrder() != 0,
		IsDefault:    req.GetIsDefault(),
	})
	if err != nil {
		return nil, statusError(err, "Failed to create status")
	}
	return &flowstatusv2.CreateStatusResponse{Status: toStatusPayload(status)}, nil
}

func (a *StatusAPIV2) UpdateStatus(ctx context.Context, req *flowstatusv2.UpdateStatusRequest) (*flowstatusv2.UpdateStatusResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "body is required")
	}
	status, err := a.Service.UpdateProjectStatus(ctx, req.GetId(), caller.UserID, businessID, domain.StatusInput{
		Name:         body.GetName(),
		Bucket:       body.GetBucket(),
		LegacyStatus: body.GetLegacyStatus(),
		Color:        body.GetColor(),
		SortOrder:    body.GetSortOrder(),
		SortOrderSet: true,
		IsDefault:    body.GetIsDefault(),
	})
	if err != nil {
		return nil, statusError(err, "Failed to update status")
	}
	return &flowstatusv2.UpdateStatusResponse{Status: toStatusPayload(status)}, nil
}

func (a *StatusAPIV2) DeleteStatus(ctx context.Context, req *flowstatusv2.DeleteStatusRequest) (*flowstatusv2.DeleteStatusResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	moved, err := a.Service.DeleteProjectStatus(ctx, req.GetId(), caller.UserID, businessID)
	if err != nil {
		return nil, statusError(err, "Failed to delete status")
	}
	return &flowstatusv2.DeleteStatusResponse{
		Message:             "Status deleted",
		ReassignedTaskCount: int32(moved),
	}, nil
}

func (a *StatusAPIV2) SetTaskStatus(ctx context.Context, req *flowstatusv2.SetTaskStatusRequest) (*flowstatusv2.SetTaskStatusResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	resulting, err := a.Service.SetTaskStatus(ctx, req.GetTaskId(), req.GetStatusId(), caller.UserID, businessID)
	if err != nil {
		return nil, statusError(err, "Failed to set task status")
	}
	return &flowstatusv2.SetTaskStatusResponse{Message: "Status set", Status: resulting}, nil
}
