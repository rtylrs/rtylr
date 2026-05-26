package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
)

func TriggerPayrollRunRoutes() []string { return []string{"/hr/payroll/runs/trigger"} }

type TriggerPayrollRunHandler struct{ Service *application.Service }

type TriggerPayrollRunRequest struct {
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

type TriggerPayrollRunResponse struct {
	Run     *hrv1.PayrollRunPayload     `json:"run"`
	Entries []*hrv1.PayrollEntryPayload `json:"entries"`
}

func (h *TriggerPayrollRunHandler) Validate(_ context.Context, req *TriggerPayrollRunRequest) *errorModels.ServerError {
	if req.PeriodStart == "" || req.PeriodEnd == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "period_start and period_end are required"}
	}
	if _, err := time.Parse("2006-01-02", req.PeriodStart); err != nil {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "period_start must be in YYYY-MM-DD format"}
	}
	if _, err := time.Parse("2006-01-02", req.PeriodEnd); err != nil {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "period_end must be in YYYY-MM-DD format"}
	}
	return nil
}

func (h *TriggerPayrollRunHandler) Serve(ctx context.Context, req *TriggerPayrollRunRequest) (TriggerPayrollRunResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return TriggerPayrollRunResponse{}, errorModels.PermissionDeniedServerError()
	}

	run, entries, err := h.Service.TriggerPayrollRun(ctx, device.User.UserID, device.User.UserID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollNotConfigured) {
			return TriggerPayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusUnprocessableEntity, Code: http.StatusUnprocessableEntity, Message: "Payroll not configured for this company"}
		}
		if errors.Is(err, domain.ErrOvertimeConfigMissing) {
			return TriggerPayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusUnprocessableEntity, Code: http.StatusUnprocessableEntity, Message: "Overtime is enabled but overtime rates are not configured"}
		}
		if errors.Is(err, domain.ErrInvalidDateRange) {
			return TriggerPayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid date range"}
		}
		logger.Ctx(ctx).Error("failed to trigger payroll run", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return TriggerPayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to trigger payroll run"}
	}

	runPayload := payrollRunToProto(*run)
	entryPayloads := make([]*hrv1.PayrollEntryPayload, 0, len(entries))
	for _, e := range entries {
		entryPayloads = append(entryPayloads, payrollEntryToProto(e))
	}
	runPayload.Entries = entryPayloads

	return TriggerPayrollRunResponse{Run: runPayload, Entries: entryPayloads}, nil
}
