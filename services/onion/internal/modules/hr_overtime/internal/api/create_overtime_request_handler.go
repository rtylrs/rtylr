package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *CreateOvertimeRequestHandler) Validate(_ context.Context, req *CreateOvertimeRequestRequest) error {
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	if req.Date == "" {
		return onekiterrors.New(http.StatusBadRequest, "date is required")
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "date must be in YYYY-MM-DD format")
	}
	if req.PlannedExtraMinutes <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "planned_extra_minutes must be greater than 0")
	}
	return nil
}

func (h *CreateOvertimeRequestHandler) Serve(ctx context.Context, req *CreateOvertimeRequestRequest) (CreateOvertimeRequestResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateOvertimeRequestResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateOvertimeRequestResponse{}, onekiterrors.PermissionDenied()
	}

	created, err := h.Service.CreateOvertimeRequest(ctx, businessID, device.User.UserID, domain.OvertimeRequestRecord{
		EmployeeID:          req.EmployeeID,
		Date:                req.Date,
		PlannedExtraMinutes: req.PlannedExtraMinutes,
		Reason:              req.Reason,
	})
	if err != nil {
		if errors.Is(err, domain.ErrOvertimeAccessDenied) {
			return CreateOvertimeRequestResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to create overtime request", logger.Any("business_id", businessID), logger.Error(err))
		return CreateOvertimeRequestResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create overtime request")
	}

	return CreateOvertimeRequestResponse{Request: overtimeRequestToProto(*created)}, nil
}
