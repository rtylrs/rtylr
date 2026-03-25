package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	configv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/config/v1"
	"github.com/voxire/onion/internal/platform/utils/converter"
)

func GetFeaturesRoutes() []string {
	return []string{
		"/get-features",
	}
}

type GetFeaturesHandler struct {
	ConfigService *application.Service
}

func (h *GetFeaturesHandler) Serve(ctx context.Context, _ *configv1.GetFeaturesRequest) (configv1.GetFeaturesResponse, *errorModels.ServerError) {
	features, err := h.ConfigService.EvaluateAllFeaturesForUser(ctx)
	if err != nil {
		return configv1.GetFeaturesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to evaluate features",
		}
	}

	return configv1.GetFeaturesResponse{
		Features: converter.Convert[*configv1.Features](features),
	}, nil
}
