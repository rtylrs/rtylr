package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	configv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/config/v1"
	"github.com/voxire/onion/internal/platform/utils/converter"
)

func ConfigByOwnerRoutes() []string {
	return []string{
		"/get-configs",
	}
}

type ConfigByOwnerHandler struct {
	ConfigService *application.Service
}

func (h *ConfigByOwnerHandler) Validate(_ context.Context, req *configv1.GetConfigByOwnerRequest) *errorModels.ServerError {
	if req.Owner == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Owner is required",
		}
	}

	return nil
}

func (h *ConfigByOwnerHandler) Serve(ctx context.Context, req *configv1.GetConfigByOwnerRequest) (configv1.GetConfigByOwnerResponse, *errorModels.ServerError) {
	config, err := h.ConfigService.GetConfig(ctx, req.Owner)
	if err != nil {
		return configv1.GetConfigByOwnerResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get config",
		}
	}

	theme, err := h.ConfigService.GetThemeByOwnerID(ctx, req.Owner)
	if err != nil {
		return configv1.GetConfigByOwnerResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get theme",
		}
	}

	return configv1.GetConfigByOwnerResponse{
		Config: converter.Convert[*configv1.Config](config),
		Theme:  converter.Convert[*configv1.Theme](theme),
	}, nil
}
