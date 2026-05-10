package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
)

func CreateShiftRoutes() []string { return []string{"/hr/shifts"} }

type CreateShiftHandler struct{ Service *application.Service }

type CreateShiftRequest struct {
	LocationID int64  `json:"location_id"`
	Name       string `json:"name"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Color      string `json:"color"`
}

type CreateShiftResponse struct {
	Shift *hrv1.ShiftPayload `json:"shift"`
}

func (h *CreateShiftHandler) Validate(_ context.Context, req *CreateShiftRequest) *errorModels.ServerError {
	if req.LocationID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "location_id is required"}
	}
	if req.Name == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "name is required"}
	}
	if req.StartTime == "" || req.EndTime == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "start_time and end_time are required"}
	}
	return nil
}

func (h *CreateShiftHandler) Serve(ctx context.Context, req *CreateShiftRequest) (CreateShiftResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateShiftResponse{}, errorModels.PermissionDeniedServerError()
	}

	created, err := h.Service.CreateShift(ctx, device.User.UserID, domain.ShiftRecord{
		LocationID: req.LocationID,
		Name:       req.Name,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Color:      req.Color,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create shift", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateShiftResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create shift"}
	}

	return CreateShiftResponse{Shift: shiftToProto(*created)}, nil
}
