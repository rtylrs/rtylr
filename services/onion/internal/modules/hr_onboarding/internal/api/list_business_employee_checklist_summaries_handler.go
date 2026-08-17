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

func ListBusinessEmployeeChecklistSummariesRoutes() []string {
	return []string{"/hr/onboarding/checklists/summary"}
}

// Mirrors the default/max clamp applied in application.Service.ListBusinessEmployeeChecklistSummaries
// so an unbounded or absent caller-provided limit never reaches the query
// layer, and so the response echoes the value actually used.
const (
	defaultListBusinessEmployeeChecklistSummariesLimit int32 = 20
	maxListBusinessEmployeeChecklistSummariesLimit     int32 = 100
)

type ListBusinessEmployeeChecklistSummariesHandler struct{ Service *application.Service }

type ListBusinessEmployeeChecklistSummariesRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListBusinessEmployeeChecklistSummariesResponse struct {
	Summaries []*hronboardingv1.EmployeeChecklistSummaryPayload `json:"summaries"`
	Total     int64                                             `json:"total"`
	Limit     int32                                             `json:"limit"`
	Offset    int32                                             `json:"offset"`
}

func (h *ListBusinessEmployeeChecklistSummariesHandler) Serve(ctx context.Context, req *ListBusinessEmployeeChecklistSummariesRequest) (ListBusinessEmployeeChecklistSummariesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListBusinessEmployeeChecklistSummariesResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListBusinessEmployeeChecklistSummariesResponse{}, onekiterrors.PermissionDenied()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultListBusinessEmployeeChecklistSummariesLimit
	}
	if limit > maxListBusinessEmployeeChecklistSummariesLimit {
		limit = maxListBusinessEmployeeChecklistSummariesLimit
	}

	summaries, total, err := h.Service.ListBusinessEmployeeChecklistSummaries(ctx, businessID, domain.ListBusinessEmployeeChecklistSummariesParams{
		Limit:  limit,
		Offset: req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list business employee checklist summaries", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListBusinessEmployeeChecklistSummariesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list checklist summaries")
	}

	payloads := make([]*hronboardingv1.EmployeeChecklistSummaryPayload, 0, len(summaries))
	for _, s := range summaries {
		payloads = append(payloads, employeeChecklistSummaryToProto(s))
	}

	return ListBusinessEmployeeChecklistSummariesResponse{Summaries: payloads, Total: total, Limit: limit, Offset: req.Offset}, nil
}
