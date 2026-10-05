package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hremployeespayrollrunsv1 "github.com/voxire/voxie/gen/go/hr/employees/payroll/runs/v1"
)

func GetMyPayrollRunRoutes() []string { return []string{"/hr/payroll/my-runs/{id}"} }

type GetMyPayrollRunHandler struct{ Service *application.Service }

type GetMyPayrollRunRequest struct {
	ID int64 `json:"id"`
}

type GetMyPayrollRunResponse struct {
	Run *hremployeespayrollrunsv1.PayrollRunPayload `json:"run"`
}

func (h *GetMyPayrollRunHandler) Serve(ctx context.Context, _ *GetMyPayrollRunRequest) (GetMyPayrollRunResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetMyPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetMyPayrollRunResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid payroll run id")
	}

	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return GetMyPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}

	employeeID, empErr := h.Service.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
	if empErr != nil {
		logger.Ctx(ctx).Warn("failed to resolve employee for self-service payroll lookup", logger.Any("business_id", businessID), logger.Error(empErr))
		return GetMyPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}

	run, entry, err := h.Service.GetMyPayrollRun(ctx, businessID, id, employeeID)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollRunNotFound) {
			return GetMyPayrollRunResponse{}, onekiterrors.New(http.StatusNotFound, "Payroll run not found")
		}
		logger.Ctx(ctx).Error("failed to get my payroll run", logger.Any("owner_id", device.User.UserID), logger.Any("run_id", id), logger.Any("employee_id", employeeID), logger.Error(err))
		return GetMyPayrollRunResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get payroll run")
	}

	runPayload := payrollRunToProto(*run)
	runPayload.Entries = []*hremployeespayrollrunsv1.PayrollEntryPayload{payrollEntryToProto(*entry)}

	return GetMyPayrollRunResponse{Run: runPayload}, nil
}
