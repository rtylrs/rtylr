package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/domain"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

func ClockInRoutes() []string { return []string{"/hr/attendance/clock-in"} }

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

func (h *ClockInHandler) Validate(ctx context.Context, req *ClockInRequest) *errorModels.ServerError {
	if req.OwnerID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "owner_id is required"}
	}
	if req.EmployeeID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "employee_id is required"}
	}
	if req.ClockInMethod == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "clock_in_method is required"}
	}
	if !domain.IsValidClockInMethod(req.ClockInMethod) {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "invalid clock_in_method"}
	}
	if req.ClockInMethod != domain.ClockInMethodPin && req.ClockInMethod != domain.ClockInMethodQR {
		return &errorModels.ServerError{HTTPCode: http.StatusForbidden, Code: http.StatusForbidden, Message: "This clock-in method requires authenticated app access"}
	}

	if req.ClockInMethod == domain.ClockInMethodQR {
		if req.QRToken == "" {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "qr_token is required for QR clock-in"}
		}
		if err := h.Service.ValidateQRToken(ctx, req.OwnerID, req.LocationID, req.QRToken); err != nil {
			return &errorModels.ServerError{HTTPCode: http.StatusUnauthorized, Code: http.StatusUnauthorized, Message: "Invalid or expired QR token"}
		}
	}

	if req.ClockInMethod == domain.ClockInMethodPin {
		if req.Pin == "" {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "pin is required for PIN clock-in"}
		}
		if err := h.Service.ValidatePIN(ctx, req.OwnerID, req.EmployeeID, req.Pin); err != nil {
			if errors.Is(err, domain.ErrPINNotSet) {
				return &errorModels.ServerError{HTTPCode: http.StatusPreconditionFailed, Code: http.StatusPreconditionFailed, Message: "Employee has not set a clock-in PIN"}
			}
			if errors.Is(err, domain.ErrInvalidPIN) {
				return &errorModels.ServerError{HTTPCode: http.StatusUnauthorized, Code: http.StatusUnauthorized, Message: "Invalid PIN"}
			}
			return &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to validate PIN"}
		}
	}

	return nil
}

func (h *ClockInHandler) Serve(ctx context.Context, req *ClockInRequest) (ClockInResponse, *errorModels.ServerError) {
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
			return ClockInResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Already clocked in"}
		}
		if errors.Is(err, domain.ErrInvalidClockInMethod) {
			return ClockInResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid clock-in method"}
		}
		if errors.Is(err, domain.ErrClockInMethodDisabled) {
			return ClockInResponse{}, &errorModels.ServerError{HTTPCode: http.StatusForbidden, Code: http.StatusForbidden, Message: "This clock-in method is not enabled"}
		}
		if errors.Is(err, domain.ErrGeofenceViolation) {
			return ClockInResponse{}, &errorModels.ServerError{HTTPCode: http.StatusForbidden, Code: http.StatusForbidden, Message: "Location is outside the allowed geofence"}
		}
		logger.Ctx(ctx).Error("failed to clock in", logger.Any("owner_id", req.OwnerID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return ClockInResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to clock in"}
	}

	return ClockInResponse{Record: attendanceToProto(*created)}, nil
}
