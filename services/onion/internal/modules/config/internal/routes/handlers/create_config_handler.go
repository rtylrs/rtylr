package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/domain"
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
	ConfigID int64  `json:"configId"`
	Message  string `json:"message"`
}

func (h *CreateConfigHandler) Validate(_ context.Context, req *CreateConfigRequest) *errorModels.ServerError {
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

func (h *CreateConfigHandler) Serve(ctx context.Context, req *CreateConfigRequest) (CreateConfigResponse, *errorModels.ServerError) {
	configInput := domain.ConfigInput{
		OwnerId:      req.Owner,
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
