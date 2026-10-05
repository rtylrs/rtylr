package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

type ClockInHandler struct{ Service *application.Service }

type ClockInRequest struct {
	OwnerID       int64   `json:"owner_id"`
	EmployeeID    int64   `json:"employee_id"`
	LocationID    int64   `json:"location_id"`
	ClockInMethod string  `json:"clock_in_method"`
	ClockInLat    float64 `json:"clock_in_lat"`
	ClockInLng    float64 `json:"clock_in_lng"`
	QRToken       string  `json:"qr_token"`
	Pin           string  `json:"pin"`
}

type ClockInResponse struct {
	Record *hrattendancev1.AttendanceRecordPayload `json:"record"`
}

func (h *ClockInHandler) Validate(ctx context.Context, req *ClockInRequest) error {
	if req.OwnerID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "owner_id is required")
	}
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	if req.ClockInMethod == "" {
		return onekiterrors.New(http.StatusBadRequest, "clock_in_method is required")
	}
	if !domain.IsValidClockInMethod(req.ClockInMethod) {
		return onekiterrors.New(http.StatusBadRequest, "invalid clock_in_method")
	}
	device := authPkg.DeviceCtx(ctx)
	if (device.IsAnonymous || device.User == nil) && req.ClockInMethod != domain.ClockInMethodPin && req.ClockInMethod != domain.ClockInMethodQR {
		return onekiterrors.New(http.StatusForbidden, "This clock-in method requires authenticated app access")
	}

	if req.ClockInMethod == domain.ClockInMethodQR {
		if req.QRToken == "" {
			return onekiterrors.New(http.StatusBadRequest, "qr_token is required for QR clock-in")
		}
		if err := h.Service.ValidateQRToken(ctx, req.OwnerID, req.LocationID, req.QRToken); err != nil {
			return onekiterrors.New(http.StatusUnauthorized, "Invalid or expired QR token")
		}
	}

	if req.ClockInMethod == domain.ClockInMethodPin {
		if req.Pin == "" {
			return onekiterrors.New(http.StatusBadRequest, "pin is required for PIN clock-in")
		}
		if h.Service.IsPINLocked(ctx, req.OwnerID, req.EmployeeID) {
			return onekiterrors.New(http.StatusTooManyRequests, "Too many failed PIN attempts; try again later")
		}
		if err := h.Service.ValidatePIN(ctx, req.OwnerID, req.EmployeeID, req.Pin); err != nil {
			if errors.Is(err, domain.ErrPINNotSet) {
				return onekiterrors.New(http.StatusPreconditionFailed, "Employee has not set a clock-in PIN")
			}
			if errors.Is(err, domain.ErrInvalidPIN) {
				return onekiterrors.New(http.StatusUnauthorized, "Invalid PIN")
			}
			return onekiterrors.New(http.StatusInternalServerError, "Failed to validate PIN")
		}
	}

	return nil
}

func (h *ClockInHandler) Serve(ctx context.Context, req *ClockInRequest) (ClockInResponse, error) {
	record := domain.AttendanceRecordRecord{
		EmployeeID:    req.EmployeeID,
		LocationID:    req.LocationID,
		ClockInMethod: req.ClockInMethod,
		ClockInLat:    req.ClockInLat,
		ClockInLng:    req.ClockInLng,
	}

	created, err := h.Service.ClockIn(ctx, req.OwnerID, record)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyClockedIn) {
			return ClockInResponse{}, onekiterrors.New(http.StatusConflict, "Already clocked in")
		}
		if errors.Is(err, domain.ErrInvalidClockInMethod) {
			return ClockInResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid clock-in method")
		}
		if errors.Is(err, domain.ErrClockInMethodDisabled) {
			return ClockInResponse{}, onekiterrors.New(http.StatusForbidden, "This clock-in method is not enabled")
		}
		if errors.Is(err, domain.ErrGeofenceViolation) {
			return ClockInResponse{}, onekiterrors.New(http.StatusForbidden, "Location is outside the allowed geofence")
		}
		logger.Ctx(ctx).Error("failed to clock in", logger.Any("owner_id", req.OwnerID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return ClockInResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to clock in")
	}

	return ClockInResponse{Record: attendanceToProto(*created)}, nil
}
