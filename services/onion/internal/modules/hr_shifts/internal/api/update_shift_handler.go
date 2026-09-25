package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrshiftv1 "github.com/voxire/voxie/gen/go/hr/shift/v1"
)

func UpdateShiftRoutes() []string { return []string{"/hr/shifts"} }

type UpdateShiftHandler struct{ Service *application.Service }

// UpdateShiftRequest is a partial edit: only the fields present are applied, so
// a caller editing one field does not blank the others. staff_id is absent on
// purpose, because it decides whether a row is a template or an attendance
// record and an update must not convert one into the other.
type UpdateShiftRequest struct {
	ID                   int64  `json:"id"`
	LocationID           int64  `json:"location_id"`
	Name                 string `json:"name"`
	StartTime            string `json:"start_time"`
	EndTime              string `json:"end_time"`
	Color                string `json:"color"`
	DepartmentID         int64  `json:"department_id"`
	Date                 string `json:"date"`
	BreakDurationMinutes int32  `json:"break_duration_minutes"`
	Status               string `json:"status"`
	Notes                string `json:"notes"`
}

type UpdateShiftResponse struct {
	Shift *hrshiftv1.ShiftPayload `json:"shift"`
}

func (h *UpdateShiftHandler) Validate(_ context.Context, req *UpdateShiftRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	return nil
}

func (h *UpdateShiftHandler) Serve(ctx context.Context, req *UpdateShiftRequest) (UpdateShiftResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateShiftResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireShiftPermission(ctx); err != nil {
		return UpdateShiftResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateShiftResponse{}, onekiterrors.PermissionDenied()
	}

	updated, err := h.Service.UpdateShift(ctx, businessID, domain.ShiftRecord{
		ID:                   req.ID,
		LocationID:           req.LocationID,
		Name:                 req.Name,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		Color:                req.Color,
		DepartmentID:         req.DepartmentID,
		Date:                 req.Date,
		BreakDurationMinutes: req.BreakDurationMinutes,
		Status:               req.Status,
		Notes:                req.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrShiftNotFound) {
			return UpdateShiftResponse{}, onekiterrors.New(http.StatusNotFound, "Shift not found")
		}
		if validationErr := shiftValidationError(err); validationErr != nil {
			return UpdateShiftResponse{}, validationErr
		}
		logger.Ctx(ctx).Error("failed to update shift",
			logger.Any("business_id", businessID), logger.Any("shift_id", req.ID), logger.Error(err))
		return UpdateShiftResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update shift")
	}

	return UpdateShiftResponse{Shift: shiftToProto(*updated)}, nil
}
