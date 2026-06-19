package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	hrovertimev1 "github.com/voxire/voxie/gen/go/hr/overtime/v1"
)

func CreateOvertimeRequestRoutes() []string { return []string{"/hr/overtime/requests"} }

type CreateOvertimeRequestHandler struct{ Service *application.Service }

type CreateOvertimeRequestRequest struct {
	EmployeeID          int64  `json:"employee_id"`
	Date                string `json:"date"`
	PlannedExtraMinutes int32  `json:"planned_extra_minutes"`
	Reason              string `json:"reason"`
}

type CreateOvertimeRequestResponse struct {
	Request *hrovertimev1.OvertimeRequestPayload `json:"request"`
}

func (h *CreateOvertimeRequestHandler) Validate(_ context.Context, req *CreateOvertimeRequestRequest) *errorModels.ServerError {
	if req.EmployeeID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "employee_id is required"}
	}
	if req.Date == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "date is required"}
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "date must be in YYYY-MM-DD format"}
	}
	if req.PlannedExtraMinutes <= 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "planned_extra_minutes must be greater than 0"}
	}
	return nil
}

func (h *CreateOvertimeRequestHandler) Serve(ctx context.Context, req *CreateOvertimeRequestRequest) (CreateOvertimeRequestResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateOvertimeRequestResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateOvertimeRequestResponse{}, errorModels.PermissionDeniedServerError()
	}

	created, err := h.Service.CreateOvertimeRequest(ctx, businessID, domain.OvertimeRequestRecord{
		EmployeeID:          req.EmployeeID,
		Date:                req.Date,
		PlannedExtraMinutes: req.PlannedExtraMinutes,
		Reason:              req.Reason,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create overtime request", logger.Any("business_id", businessID), logger.Error(err))
		return CreateOvertimeRequestResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create overtime request"}
	}

	return CreateOvertimeRequestResponse{Request: overtimeRequestToProto(*created)}, nil
}
