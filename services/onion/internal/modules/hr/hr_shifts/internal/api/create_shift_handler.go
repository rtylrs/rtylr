package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (r *CreateShiftRequest) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decodeString := func(keys ...string) (string, error) {
		for _, k := range keys {
			raw, ok := fields[k]
			if !ok || len(raw) == 0 || string(raw) == "null" {
				continue
			}
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return "", fmt.Errorf("%s: %w", k, err)
			}
			return s, nil
		}
		return "", nil
	}
	decodeInt64 := func(keys ...string) (int64, error) {
		for _, k := range keys {
			raw, ok := fields[k]
			if !ok || len(raw) == 0 || string(raw) == "null" {
				continue
			}
			var s string
			if err := json.Unmarshal(raw, &s); err == nil {
				s = strings.TrimSpace(s)
				if s == "" {
					return 0, nil
				}
				v, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return 0, fmt.Errorf("%s: %w", k, err)
				}
				return v, nil
			}
			var v int64
			if err := json.Unmarshal(raw, &v); err != nil {
				return 0, fmt.Errorf("%s: %w", k, err)
			}
			return v, nil
		}
		return 0, nil
	}
	decodeInt32 := func(keys ...string) (int32, error) {
		v, err := decodeInt64(keys...)
		return int32(v), err // #nosec G115 -- Bounded shift break duration minutes.
	}

	var err error
	if r.LocationID, err = decodeInt64("location_id", "locationId"); err != nil {
		return err
	}
	if r.StaffID, err = decodeInt64("staff_id", "staffId"); err != nil {
		return err
	}
	if r.DepartmentID, err = decodeInt64("department_id", "departmentId"); err != nil {
		return err
	}
	if r.BreakDurationMinutes, err = decodeInt32("break_duration_minutes", "breakDurationMinutes"); err != nil {
		return err
	}
	if r.Name, err = decodeString("name"); err != nil {
		return err
	}
	if r.StartTime, err = decodeString("start_time", "startTime"); err != nil {
		return err
	}
	if r.EndTime, err = decodeString("end_time", "endTime"); err != nil {
		return err
	}
	if r.Color, err = decodeString("color"); err != nil {
		return err
	}
	if r.Date, err = decodeString("date"); err != nil {
		return err
	}
	if r.Status, err = decodeString("status"); err != nil {
		return err
	}
	if r.Notes, err = decodeString("notes"); err != nil {
		return err
	}
	return nil
}

type CreateShiftResponse struct {
	Shift *hrshiftv1.ShiftPayload `json:"shift"`
}

func isValidShiftTime(value string) bool {
	trimmed := strings.TrimSpace(value)
	if _, err := time.Parse("15:04", trimmed); err == nil {
		return true
	}
	_, err := time.Parse("15:04:05", trimmed)
	return err == nil
}

func isValidShiftStatus(value string) bool {
	switch value {
	case "", "scheduled", "in_progress", "completed", "absent", "cancelled":
		return true
	}
	return false
}

func (h *CreateShiftHandler) Validate(_ context.Context, req *CreateShiftRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "request body is required")
	}
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)
	req.Date = strings.TrimSpace(req.Date)
	req.Status = strings.TrimSpace(req.Status)
	req.Color = strings.TrimSpace(req.Color)

	if req.StartTime == "" || req.EndTime == "" {
		return onekiterrors.New(http.StatusBadRequest, "start_time and end_time are required")
	}
	if !isValidShiftTime(req.StartTime) || !isValidShiftTime(req.EndTime) {
		return onekiterrors.New(http.StatusBadRequest, "start_time and end_time must use HH:MM or HH:MM:SS")
	}
	if !isValidShiftStatus(req.Status) {
		return onekiterrors.New(http.StatusBadRequest, "invalid shift status")
	}
	if len(req.Color) > 10 {
		return onekiterrors.New(http.StatusBadRequest, "color must be at most 10 characters")
	}
	if req.Date != "" {
		if _, err := time.Parse(time.DateOnly, req.Date); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "date must use YYYY-MM-DD")
		}
	}

	if req.StaffID != 0 {
		if req.Date == "" {
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
