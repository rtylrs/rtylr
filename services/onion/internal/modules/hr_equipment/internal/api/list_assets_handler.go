package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
)

func ListAssetsRoutes() []string { return []string{"/hr/equipment/assets"} }

type ListAssetsHandler struct{ Service *application.Service }

type ListAssetsRequest struct {
	LocationID int64  `json:"location_id"`
	Status     string `json:"status"`
	Limit      int32  `json:"limit"`
	Offset     int32  `json:"offset"`
}

type ListAssetsResponse struct {
	Assets []*hrequipmentv1.AssetPayload `json:"assets"`
}

func (h *ListAssetsHandler) Serve(ctx context.Context, req *ListAssetsRequest) (ListAssetsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListAssetsResponse{}, onekiterrors.PermissionDenied()
	}

	params := domain.ListAssetsParams{Limit: req.Limit, Offset: req.Offset}
	if req.LocationID != 0 {
		params.LocationID = &req.LocationID
	}
	if req.Status != "" {
		params.Status = &req.Status
	}

	assets, err := h.Service.ListAssets(ctx, device.User.UserID, params)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list assets", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListAssetsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list assets")
	}

	payloads := make([]*hrequipmentv1.AssetPayload, 0, len(assets))
	for _, a := range assets {
		payloads = append(payloads, assetToProto(a))
	}

	return ListAssetsResponse{Assets: payloads}, nil
}
