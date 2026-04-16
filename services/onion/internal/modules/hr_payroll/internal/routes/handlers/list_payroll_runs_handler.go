package handlers

import (
	"context"
	"net/http"

	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func ListPayrollRunsRoutes() []string { return []string{"/hr/payroll/runs"} }

type ListPayrollRunsHandler struct{ Service *application.Service }

type ListPayrollRunsRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListPayrollRunsResponse struct {
	Runs   []*hrv1.PayrollRunPayload `json:"runs"`
	Total  int64                     `json:"total"`
	Limit  int32                     `json:"limit"`
	Offset int32                     `json:"offset"`
}

func (h *ListPayrollRunsHandler) Serve(ctx context.Context, req *ListPayrollRunsRequest) (ListPayrollRunsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListPayrollRunsResponse{}, errorModels.PermissionDeniedServerError()
	}

	runs, total, err := h.Service.ListPayrollRuns(ctx, device.User.UserID, domain.ListPayrollRunsParams{
		Limit: req.Limit, Offset: req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list payroll runs", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListPayrollRunsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list payroll runs"}
	}

	payloads := make([]*hrv1.PayrollRunPayload, 0, len(runs))
	for _, r := range runs {
		payloads = append(payloads, payrollRunToProto(r))
	}

	return ListPayrollRunsResponse{Runs: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
