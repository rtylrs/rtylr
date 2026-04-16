package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func GetEmployeeChecklistRoutes() []string { return []string{"/hr/onboarding/checklists/{id}"} }

type GetEmployeeChecklistHandler struct{ Service *application.Service }

type GetEmployeeChecklistRequest struct {
	EmployeeID int64 `json:"employee_id"`
}

type GetEmployeeChecklistResponse struct {
	Checklist *hrv1.EmployeeChecklistPayload `json:"checklist"`
}

func (h *GetEmployeeChecklistHandler) Serve(ctx context.Context, req *GetEmployeeChecklistRequest) (GetEmployeeChecklistResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetEmployeeChecklistResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	checklistID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetEmployeeChecklistResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid checklist id"}
	}

	if req.EmployeeID == 0 {
		return GetEmployeeChecklistResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "employee_id is required"}
	}

	checklist, err := h.Service.GetEmployeeChecklist(ctx, device.User.UserID, req.EmployeeID, checklistID)
	if err != nil {
		if err == domain.ErrChecklistNotFound {
			return GetEmployeeChecklistResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Checklist not found"}
		}
		logger.Ctx(ctx).Error("failed to get employee checklist", logger.Any("owner_id", device.User.UserID), logger.Any("checklist_id", checklistID), logger.Error(err))
		return GetEmployeeChecklistResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get checklist"}
	}

	return GetEmployeeChecklistResponse{Checklist: employeeChecklistToProto(*checklist)}, nil
}
