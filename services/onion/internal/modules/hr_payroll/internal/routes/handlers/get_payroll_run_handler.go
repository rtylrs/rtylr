package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
)

func GetPayrollRunRoutes() []string { return []string{"/hr/payroll/runs/{id}"} }

type GetPayrollRunHandler struct{ Service *application.Service }

type GetPayrollRunRequest struct {
	ID int64 `json:"id"`
}

type GetPayrollRunResponse struct {
	Run     *hrv1.PayrollRunPayload     `json:"run"`
	Entries []*hrv1.PayrollEntryPayload `json:"entries"`
}

func (h *GetPayrollRunHandler) Serve(ctx context.Context, req *GetPayrollRunRequest) (GetPayrollRunResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetPayrollRunResponse{}, errorModels.PermissionDeniedServerError()
	}

	run, entries, err := h.Service.GetPayrollRun(ctx, device.User.UserID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollRunNotFound) {
			return GetPayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Payroll run not found"}
		}
		logger.Ctx(ctx).Error("failed to get payroll run", logger.Any("owner_id", device.User.UserID), logger.Any("run_id", req.ID), logger.Error(err))
		return GetPayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get payroll run"}
	}

	runPayload := payrollRunToProto(*run)
	entryPayloads := make([]*hrv1.PayrollEntryPayload, 0, len(entries))
	for _, e := range entries {
		entryPayloads = append(entryPayloads, payrollEntryToProto(e))
	}
	runPayload.Entries = entryPayloads

	return GetPayrollRunResponse{Run: runPayload, Entries: entryPayloads}, nil
}
