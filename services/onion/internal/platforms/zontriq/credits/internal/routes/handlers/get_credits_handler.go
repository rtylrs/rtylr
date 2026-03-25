package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	creditsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/credits/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/credits/internal/application"
)

func GetCreditsRoutes() []string {
	return []string{
		"/zontriq/credits",
	}
}

type GetCreditsHandler struct {
	CreditsService *application.Service
}

func (h *GetCreditsHandler) Serve(ctx context.Context, _ *creditsv1.GetCreditsRequest) (creditsv1.GetCreditsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return creditsv1.GetCreditsResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	zontriqUserID := int32(device.User.PlatformUserID)

	credits, err := h.CreditsService.GetCredits(ctx, zontriqUserID)
	if err != nil {
		return creditsv1.GetCreditsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get credits",
		}
	}

	return creditsv1.GetCreditsResponse{
		Credits: &creditsv1.Credits{
			Id:             int64(credits.ID),
			Balance:        credits.Balance,
			TotalPurchased: credits.TotalPurchased,
			TotalSpent:     credits.TotalSpent,
			Status:         credits.Status,
			CreatedAt:      credits.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      credits.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}
