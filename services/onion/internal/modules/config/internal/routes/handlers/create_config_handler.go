package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/domain"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	configv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/config/v1"
)

func CreateConfigRoutes() []string {
	return []string{
		"/config/create",
	}
}

type CreateConfigHandler struct {
	ConfigService *application.Service
}

func (h *CreateConfigHandler) Validate(_ context.Context, req *configv1.CreateConfigRequest) *errorModels.ServerError {
	if req.Owner == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Owner is required",
		}
	}

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

func (h *CreateConfigHandler) Serve(ctx context.Context, req *configv1.CreateConfigRequest) (configv1.CreateConfigResponse, *errorModels.ServerError) {
	configInput := domain.ConfigInput{
		OwnerId:      req.Owner,
		SiteName:     req.Key,
		ContactEmail: req.Value,
	}

	err := h.ConfigService.CreateConfig(ctx, configInput)
	if err != nil {
		return configv1.CreateConfigResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create config",
		}
	}

	return configv1.CreateConfigResponse{
		Message: "Config created successfully",
	}, nil
}
