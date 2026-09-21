package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hremployeespayrollrunsv1 "github.com/voxire/voxie/gen/go/hr/employees/payroll/runs/v1"
)

func ListMyPayrollRunsRoutes() []string { return []string{"/hr/payroll/my-runs"} }

type ListMyPayrollRunsHandler struct{ Service *application.Service }

type ListMyPayrollRunsRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListMyPayrollRunsResponse struct {
	Runs   []*hremployeespayrollrunsv1.PayrollRunPayload `json:"runs"`
	Total  int64                                         `json:"total"`
	Limit  int32                                         `json:"limit"`
	Offset int32                                         `json:"offset"`
}

func (h *ListMyPayrollRunsHandler) Serve(ctx context.Context, req *ListMyPayrollRunsRequest) (ListMyPayrollRunsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListMyPayrollRunsResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return ListMyPayrollRunsResponse{}, onekiterrors.PermissionDenied()
	}

	employeeID, empErr := h.Service.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
	if empErr != nil {
		logger.Ctx(ctx).Warn("failed to resolve employee for self-service payroll lookup", logger.Any("business_id", businessID), logger.Error(empErr))
		return ListMyPayrollRunsResponse{}, onekiterrors.PermissionDenied()
	}

	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	runs, total, err := h.Service.ListMyPayrollRuns(ctx, businessID, employeeID, domain.ListPayrollRunsParams{
		Limit: req.Limit, Offset: req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list my payroll runs", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", employeeID), logger.Error(err))
		return ListMyPayrollRunsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list payroll runs")
	}

	payloads := make([]*hremployeespayrollrunsv1.PayrollRunPayload, 0, len(runs))
	for _, r := range runs {
		payloads = append(payloads, payrollRunToProto(r))
	}

	return ListMyPayrollRunsResponse{Runs: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
