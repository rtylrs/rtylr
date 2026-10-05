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

func GetPayrollRunRoutes() []string { return []string{"/hr/payroll/runs/{id}"} }

type GetPayrollRunHandler struct{ Service *application.Service }

type GetPayrollRunRequest struct {
	ID int64 `json:"id"`
}

type GetPayrollRunResponse struct {
	Run     *hremployeespayrollrunsv1.PayrollRunPayload     `json:"run"`
	Entries []*hremployeespayrollrunsv1.PayrollEntryPayload `json:"entries"`
}

func (h *GetPayrollRunHandler) Serve(ctx context.Context, _ *GetPayrollRunRequest) (GetPayrollRunResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requirePayrollPermission(ctx, "hr:payroll:view"); err != nil {
		return GetPayrollRunResponse{}, err
	}

	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return GetPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}

	employeeID, empErr := h.Service.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
	if empErr != nil {
		logger.Ctx(ctx).Warn("failed to resolve employee for payroll permission check", logger.Any("business_id", businessID), logger.Error(empErr))
		return GetPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}
	allowed, permErr := h.Service.HasPermission(ctx, businessID, employeeID, "manage_payroll")
	if permErr != nil {
		logger.Ctx(ctx).Error("failed to check manage_payroll permission", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(permErr))
		return GetPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}
	if !allowed {
		return GetPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetPayrollRunResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid payroll run id")
	}

	run, entries, err := h.Service.GetPayrollRun(ctx, businessID, id)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollRunNotFound) {
			return GetPayrollRunResponse{}, onekiterrors.New(http.StatusNotFound, "Payroll run not found")
		}
		logger.Ctx(ctx).Error("failed to get payroll run", logger.Any("business_id", businessID), logger.Any("run_id", id), logger.Error(err))
		return GetPayrollRunResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get payroll run")
	}

	runPayload := payrollRunToProto(*run)
	entryPayloads := make([]*hremployeespayrollrunsv1.PayrollEntryPayload, 0, len(entries))
	for _, e := range entries {
		entryPayloads = append(entryPayloads, payrollEntryToProto(e))
	}
	runPayload.Entries = entryPayloads

	return GetPayrollRunResponse{Run: runPayload, Entries: entryPayloads}, nil
}
