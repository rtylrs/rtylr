package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrshiftv1 "github.com/voxire/voxie/gen/go/hr/shift/v1"
)

func CreateShiftRoutes() []string { return []string{"/hr/shifts"} }

type CreateShiftHandler struct{ Service *application.Service }

// CreateShiftRequest creates either kind of row. Supplying staff_id makes it an
// attendance row, which needs a date; otherwise it is a template row, which
// needs a location and a name.
type CreateShiftRequest struct {
	LocationID           int64  `json:"location_id"`
	Name                 string `json:"name"`
	StartTime            string `json:"start_time"`
	EndTime              string `json:"end_time"`
	Color                string `json:"color"`
	StaffID              int64  `json:"staff_id"`
	DepartmentID         int64  `json:"department_id"`
	Date                 string `json:"date"`
	BreakDurationMinutes int32  `json:"break_duration_minutes"`
	Status               string `json:"status"`
	Notes                string `json:"notes"`
}

type CreateShiftResponse struct {
	Shift *hrshiftv1.ShiftPayload `json:"shift"`
}

func (h *CreateShiftHandler) Validate(_ context.Context, req *CreateShiftRequest) error {
	if req.StartTime == "" || req.EndTime == "" {
		return onekiterrors.New(http.StatusBadRequest, "start_time and end_time are required")
	}

	if req.StaffID != 0 {
		if strings.TrimSpace(req.Date) == "" {
			return onekiterrors.New(http.StatusBadRequest, "date is required when staff_id is set")
		}
		return nil
	}

	if req.LocationID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "location_id is required for a shift template")
	}
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required for a shift template")
	}
	return nil
}

func (h *CreateShiftHandler) Serve(ctx context.Context, req *CreateShiftRequest) (CreateShiftResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateShiftResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireShiftPermission(ctx); err != nil {
		return CreateShiftResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateShiftResponse{}, onekiterrors.PermissionDenied()
	}

	created, err := h.Service.CreateShift(ctx, businessID, domain.ShiftRecord{
		LocationID:           req.LocationID,
		Name:                 req.Name,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		Color:                req.Color,
		StaffID:              req.StaffID,
		DepartmentID:         req.DepartmentID,
		Date:                 req.Date,
		BreakDurationMinutes: req.BreakDurationMinutes,
		Status:               req.Status,
		Notes:                req.Notes,
	})
	if err != nil {
		if validationErr := shiftValidationError(err); validationErr != nil {
			return CreateShiftResponse{}, validationErr
		}
		logger.Ctx(ctx).Error("failed to create shift", logger.Any("business_id", businessID), logger.Error(err))
		return CreateShiftResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create shift")
	}

	return CreateShiftResponse{Shift: shiftToProto(*created)}, nil
}

// shiftValidationError maps the domain's validation sentinels onto a 400 so a
// bad request is not reported as a server fault. It returns nil when err is not
// a validation failure.
func shiftValidationError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEmptyShiftName),
		errors.Is(err, domain.ErrEmptyShiftDate),
		errors.Is(err, domain.ErrEmptyShiftTime):
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return nil
}
