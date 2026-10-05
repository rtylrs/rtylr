package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hremployeespayrollrunsv1 "github.com/voxire/voxie/gen/go/hr/employees/payroll/runs/v1"
)

func TriggerPayrollRunRoutes() []string { return []string{"/hr/payroll/runs/trigger"} }

type TriggerPayrollRunHandler struct{ Service *application.Service }

type TriggerPayrollRunRequest struct {
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

type TriggerPayrollRunResponse struct {
	Run     *hremployeespayrollrunsv1.PayrollRunPayload     `json:"run"`
	Entries []*hremployeespayrollrunsv1.PayrollEntryPayload `json:"entries"`
}

func (h *TriggerPayrollRunHandler) Validate(_ context.Context, req *TriggerPayrollRunRequest) error {
	if req.PeriodStart == "" || req.PeriodEnd == "" {
		return onekiterrors.New(http.StatusBadRequest, "period_start and period_end are required")
	}
	if _, err := time.Parse("2006-01-02", req.PeriodStart); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "period_start must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", req.PeriodEnd); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "period_end must be in YYYY-MM-DD format")
	}
	return nil
}

func (h *TriggerPayrollRunHandler) Serve(ctx context.Context, req *TriggerPayrollRunRequest) (TriggerPayrollRunResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return TriggerPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requirePayrollPermission(ctx, "hr:payroll:execute"); err != nil {
		return TriggerPayrollRunResponse{}, err
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return TriggerPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}

	employeeID, empErr := h.Service.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
	if empErr != nil {
		logger.Ctx(ctx).Warn("failed to resolve employee for payroll permission check", logger.Any("business_id", businessID), logger.Error(empErr))
		return TriggerPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}
	allowed, permErr := h.Service.HasPermission(ctx, businessID, employeeID, "manage_payroll")
	if permErr != nil {
		logger.Ctx(ctx).Error("failed to check manage_payroll permission", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(permErr))
		return TriggerPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}
	if !allowed {
		return TriggerPayrollRunResponse{}, onekiterrors.PermissionDenied()
	}

	run, entries, err := h.Service.TriggerPayrollRun(ctx, businessID, device.User.UserID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollRunAlreadyExists) {
			return TriggerPayrollRunResponse{}, onekiterrors.New(http.StatusConflict, "Payroll run already exists for this period")
		}
		if errors.Is(err, domain.ErrPayrollNotConfigured) {
			return TriggerPayrollRunResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, "Payroll not configured for this company")
		}
		if errors.Is(err, domain.ErrOvertimeConfigMissing) {
			return TriggerPayrollRunResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, "Overtime is enabled but overtime rates are not configured")
		}
		if errors.Is(err, domain.ErrOvertimeRateMissing) {
			return TriggerPayrollRunResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, "Approved overtime requires an hourly rate for every employee")
		}
		if errors.Is(err, domain.ErrInvalidDateRange) {
			return TriggerPayrollRunResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid date range")
		}
		logger.Ctx(ctx).Error("failed to trigger payroll run", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return TriggerPayrollRunResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to trigger payroll run")
	}

	runPayload := payrollRunToProto(*run)
	entryPayloads := make([]*hremployeespayrollrunsv1.PayrollEntryPayload, 0, len(entries))
	for _, e := range entries {
		entryPayloads = append(entryPayloads, payrollEntryToProto(e))
	}
	runPayload.Entries = entryPayloads

	return TriggerPayrollRunResponse{Run: runPayload, Entries: entryPayloads}, nil
}
