package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CreateConfigRoutes() []string {
	return []string{
		"/config/create",
	}
}

type CreateConfigHandler struct {
	ConfigService *application.Service
}

type CreateConfigRequest struct {
	Owner       string `json:"owner"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type CreateConfigResponse struct {
	ConfigID int64  `json:"config_id"`
	Message  string `json:"message"`
}

func (h *CreateConfigHandler) Validate(_ context.Context, req *CreateConfigRequest) error {
	// Owner is derived from the authenticated user in Serve, not trusted from
	// the request body, so it is intentionally not validated here.
	if req.Key == "" {
		return onekiterrors.New(http.StatusBadRequest, "Key is required")
	}

	if req.Value == "" {
		return onekiterrors.New(http.StatusBadRequest, "Value is required")
	}

	return nil
}

func (h *CreateConfigHandler) Serve(ctx context.Context, req *CreateConfigRequest) (CreateConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateConfigResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateConfigResponse{}, onekiterrors.PermissionDenied()
	}

	businessIDStr := strconv.FormatInt(businessID, 10)

	configInput := domain.ConfigInput{
		OwnerId:      businessIDStr,
		BusinessID:   businessID,
		SiteName:     req.Key,
		ContactEmail: req.Value,
	}

	err := h.ConfigService.CreateConfig(ctx, configInput)
	if err != nil {
		return CreateConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create config")
	}

	return CreateConfigResponse{
		Message: "Config created successfully",
	}, nil
}
