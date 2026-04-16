package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func GenerateQRRoutes() []string { return []string{"/hr/attendance/qr/generate"} }

type GenerateQRHandler struct{ Service *application.Service }

type GenerateQRRequest struct {
	LocationID int64 `json:"location_id"`
}

type GenerateQRResponse struct {
	Token string `json:"token"`
}

func (h *GenerateQRHandler) Validate(_ context.Context, req *GenerateQRRequest) *errorModels.ServerError {
	if req.LocationID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "location_id is required"}
	}
	return nil
}

func (h *GenerateQRHandler) Serve(ctx context.Context, req *GenerateQRRequest) (GenerateQRResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GenerateQRResponse{}, errorModels.PermissionDeniedServerError()
	}

	token, err := h.Service.GenerateQRToken(ctx, device.User.UserID, req.LocationID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to generate QR token", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GenerateQRResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to generate QR token"}
	}

	return GenerateQRResponse{Token: token}, nil
}
