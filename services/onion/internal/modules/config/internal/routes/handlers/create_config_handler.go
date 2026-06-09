package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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

func (h *CreateConfigHandler) Validate(_ context.Context, req *CreateConfigRequest) *errorModels.ServerError {
	// Owner is derived from the authenticated user in Serve, not trusted from
	// the request body, so it is intentionally not validated here.
	if req.Key == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Key is required",
		}
	}

	if req.Value == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Value is required",
		}
	}

	return nil
}

func (h *CreateConfigHandler) Serve(ctx context.Context, req *CreateConfigRequest) (CreateConfigResponse, *errorModels.ServerError) {
	// Derive the owner from the authenticated user — never trust an owner id
	// supplied in the request body (that was an IDOR: any user could create a
	// config for another business).
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	ownerID := strconv.FormatInt(device.User.BusinessID, 10)

	configInput := domain.ConfigInput{
		OwnerId:      ownerID,
		SiteName:     req.Key,
		ContactEmail: req.Value,
	}

	err := h.ConfigService.CreateConfig(ctx, configInput)
	if err != nil {
		return CreateConfigResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create config",
		}
	}

	return CreateConfigResponse{
		Message: "Config created successfully",
	}, nil
}
