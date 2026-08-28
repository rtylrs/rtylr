package handlers

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

func ListPayrollRunsRoutes() []string { return []string{"/hr/payroll/runs"} }

type ListPayrollRunsHandler struct{ Service *application.Service }

type ListPayrollRunsRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListPayrollRunsResponse struct {
	Runs   []*hremployeespayrollrunsv1.PayrollRunPayload `json:"runs"`
	Total  int64                                         `json:"total"`
	Limit  int32                                         `json:"limit"`
	Offset int32                                         `json:"offset"`
}

func (h *ListPayrollRunsHandler) Serve(ctx context.Context, req *ListPayrollRunsRequest) (ListPayrollRunsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListPayrollRunsResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requirePayrollPermission(ctx, "hr:payroll:view"); err != nil {
		return ListPayrollRunsResponse{}, err
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return ListPayrollRunsResponse{}, onekiterrors.PermissionDenied()
	}

	employeeID, empErr := h.Service.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
	if empErr != nil {
		logger.Ctx(ctx).Warn("failed to resolve employee for payroll permission check", logger.Any("business_id", businessID), logger.Error(empErr))
		return ListPayrollRunsResponse{}, onekiterrors.PermissionDenied()
	}
	allowed, permErr := h.Service.HasPermission(ctx, businessID, employeeID, "manage_payroll")
	if permErr != nil {
		logger.Ctx(ctx).Error("failed to check manage_payroll permission", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(permErr))
		return ListPayrollRunsResponse{}, onekiterrors.PermissionDenied()
	}
	if !allowed {
		return ListPayrollRunsResponse{}, onekiterrors.PermissionDenied()
	}

	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	runs, total, err := h.Service.ListPayrollRuns(ctx, businessID, domain.ListPayrollRunsParams{
		Limit: req.Limit, Offset: req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list payroll runs", logger.Any("business_id", businessID), logger.Error(err))
		return ListPayrollRunsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list payroll runs")
	}

	payloads := make([]*hremployeespayrollrunsv1.PayrollRunPayload, 0, len(runs))
	for _, r := range runs {
		payloads = append(payloads, payrollRunToProto(r))
	}

	return ListPayrollRunsResponse{Runs: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
