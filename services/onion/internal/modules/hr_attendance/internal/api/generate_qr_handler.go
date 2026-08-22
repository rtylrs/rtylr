package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func GenerateQRRoutes() []string { return []string{"/hr/attendance/qr/generate"} }

type GenerateQRHandler struct{ Service *application.Service }

type GenerateQRRequest struct {
	LocationID int64 `json:"location_id"`
}

type GenerateQRResponse struct {
	Token string `json:"token"`
}

func (h *GenerateQRHandler) Validate(_ context.Context, req *GenerateQRRequest) error {
	if req.LocationID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "location_id is required")
	}
	return nil
}

func (h *GenerateQRHandler) Serve(ctx context.Context, req *GenerateQRRequest) (GenerateQRResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GenerateQRResponse{}, onekiterrors.PermissionDenied()
	}

	ownerID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return GenerateQRResponse{}, onekiterrors.PermissionDenied()
	}

	token, err := h.Service.GenerateQRToken(ctx, ownerID, req.LocationID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to generate QR token", logger.Any("owner_id", ownerID), logger.Error(err))
		return GenerateQRResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to generate QR token")
	}

	return GenerateQRResponse{Token: token}, nil
}
