package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
)

func ListAssetsRoutes() []string { return []string{"/hr/equipment/assets"} }

type ListAssetsHandler struct{ Service *application.Service }

type ListAssetsRequest struct {
	EmployeeID int64 `json:"employee_id"`
	LocationID int64 `json:"location_id"`
	// Status is the legacy lifecycle filter used by the JSON route.
	Status string `json:"status"`
	// Condition is the generated API's physical-condition filter.
	Condition string `json:"condition"`
	Limit     int32  `json:"limit"`
	Offset    int32  `json:"offset"`
}

type ListAssetsResponse struct {
	Assets []*hrequipmentv1.AssetPayload `json:"assets"`
	Total  int64                         `json:"total"`
	Limit  int32                         `json:"limit"`
	Offset int32                         `json:"offset"`
}

func (h *ListAssetsHandler) Serve(ctx context.Context, req *ListAssetsRequest) (ListAssetsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListAssetsResponse{}, onekiterrors.PermissionDenied()
	}
	if req == nil {
		req = &ListAssetsRequest{}
	}
	req.Condition = strings.TrimSpace(req.Condition)
	req.Status = strings.TrimSpace(req.Status)
	if req.Condition != "" && !domain.IsValidAssetCondition(req.Condition) {
		return ListAssetsResponse{}, onekiterrors.New(http.StatusBadRequest, "condition must be one of new, good, fair, damaged")
	}
	if req.Status != "" && !domain.IsValidAssetStatus(req.Status) {
		return ListAssetsResponse{}, onekiterrors.New(http.StatusBadRequest, "status must be one of available, assigned, retired, repair")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListAssetsResponse{}, onekiterrors.PermissionDenied()
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	params := domain.ListAssetsParams{Limit: req.Limit, Offset: req.Offset}
	if req.EmployeeID != 0 {
		params.AssignedTo = &req.EmployeeID
	}
	if req.LocationID != 0 {
		params.LocationID = &req.LocationID
	}
	if req.Status != "" {
		params.Status = &req.Status
	}
	if req.Condition != "" {
		params.Condition = &req.Condition
	}

	assets, total, err := h.Service.ListAssets(ctx, businessID, params)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAssetCondition) {
			return ListAssetsResponse{}, onekiterrors.New(http.StatusBadRequest, "condition must be one of new, good, fair, damaged")
		}
		logger.Ctx(ctx).Error("failed to list assets", logger.Any("business_id", businessID), logger.Error(err))
		return ListAssetsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list assets")
	}

	payloads := make([]*hrequipmentv1.AssetPayload, 0, len(assets))
	for _, a := range assets {
		payloads = append(payloads, assetToProto(a))
	}

	return ListAssetsResponse{Assets: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
