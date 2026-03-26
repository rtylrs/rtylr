package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func GetFeaturesRoutes() []string {
	return []string{
		"/get-features",
	}
}

type GetFeaturesHandler struct {
	ConfigService *application.Service
}

type GetFeaturesRequest struct{}

type FeaturesPayload struct {
	Flags       map[string]bool `json:"flags"`
	IsAnonymous bool            `json:"isAnonymous"`
	SessionID   string          `json:"sessionId"`
}

type GetFeaturesResponse struct {
	Features *FeaturesPayload `json:"features"`
}

func (h *GetFeaturesHandler) Serve(ctx context.Context, _ *GetFeaturesRequest) (GetFeaturesResponse, *errorModels.ServerError) {
	features, err := h.ConfigService.EvaluateAllFeaturesForUser(ctx)
	if err != nil {
		return GetFeaturesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to evaluate features",
		}
	}

	return GetFeaturesResponse{
		Features: &FeaturesPayload{
			Flags:       features.Flags,
			IsAnonymous: features.IsAnonymous,
			SessionID:   features.SessionId,
		},
	}, nil
}
