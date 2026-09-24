package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

func ListEmployeeChecklistsRoutes() []string { return []string{"/hr/onboarding/checklists"} }

// Mirrors the default/max clamp applied in application.Service.ListEmployeeChecklists
// so an unbounded or absent caller-provided limit never reaches the query
// layer, and so the response echoes the value actually used.
const (
	defaultListEmployeeChecklistsLimit int32 = 20
	maxListEmployeeChecklistsLimit     int32 = 100
)

type ListEmployeeChecklistsHandler struct{ Service *application.Service }

type ListEmployeeChecklistsRequest struct {
	EmployeeID int64 `json:"employee_id"`
	Limit      int32 `json:"limit"`
	Offset     int32 `json:"offset"`
}

type ListEmployeeChecklistsResponse struct {
	Checklists []*hronboardingv1.EmployeeChecklistPayload `json:"checklists"`
	Total      int64                                      `json:"total"`
	Limit      int32                                      `json:"limit"`
	Offset     int32                                      `json:"offset"`
}

func (h *ListEmployeeChecklistsHandler) Serve(ctx context.Context, req *ListEmployeeChecklistsRequest) (ListEmployeeChecklistsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListEmployeeChecklistsResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListEmployeeChecklistsResponse{}, onekiterrors.PermissionDenied()
	}

	if req.EmployeeID == 0 {
		return ListEmployeeChecklistsResponse{}, onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultListEmployeeChecklistsLimit
	}
	if limit > maxListEmployeeChecklistsLimit {
		limit = maxListEmployeeChecklistsLimit
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	checklists, total, err := h.Service.ListEmployeeChecklists(ctx, businessID, req.EmployeeID, domain.ListEmployeeChecklistsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list employee checklists", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return ListEmployeeChecklistsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list checklists")
	}

	payloads := make([]*hronboardingv1.EmployeeChecklistPayload, 0, len(checklists))
	for _, c := range checklists {
		payloads = append(payloads, employeeChecklistToProto(c))
	}

	return ListEmployeeChecklistsResponse{Checklists: payloads, Total: total, Limit: limit, Offset: offset}, nil
}
