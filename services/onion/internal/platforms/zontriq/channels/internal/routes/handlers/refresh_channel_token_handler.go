package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	channelsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/channels/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/infrastructure"
)

func RefreshChannelTokenRoutes() []string {
	return []string{
		"/zontriq/channels/{id}/token",
	}
}

type RefreshChannelTokenHandler struct {
	ChannelService *application.Service
}

func (h *RefreshChannelTokenHandler) Validate(_ context.Context, req *channelsv1.RefreshChannelTokenRequest) *errorModels.ServerError {
	if req.AccessToken == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Access token is required",
		}
	}

	return nil
}

func (h *RefreshChannelTokenHandler) Serve(ctx context.Context, req *channelsv1.RefreshChannelTokenRequest) (channelsv1.RefreshChannelTokenResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return channelsv1.RefreshChannelTokenResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	idStr := ""
	if rctx := chi.RouteContext(ctx); rctx != nil {
		idStr = rctx.URLParam("id")
	}

	channelID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return channelsv1.RefreshChannelTokenResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid channel ID",
		}
	}

	var tokenExpiresAt time.Time
	if req.TokenExpiresAt != "" {
		tokenExpiresAt, err = time.Parse(time.RFC3339, req.TokenExpiresAt)
		if err != nil {
			return channelsv1.RefreshChannelTokenResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadRequest,
				Message: "Invalid token_expires_at format, expected RFC3339",
			}
		}
	}

	input := infrastructure.UpdateTokensInput{
		ChannelID:       int32(channelID),
		AccessToken:     req.AccessToken,
		RefreshToken:    req.RefreshToken,
		TokenExpiresAt:  tokenExpiresAt,
		ConnectionState: "connected",
	}

	channel, err := h.ChannelService.RefreshChannelToken(ctx, input)
	if err != nil {
		return channelsv1.RefreshChannelTokenResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to refresh channel token",
		}
	}

	return channelsv1.RefreshChannelTokenResponse{
		Channel: &channelsv1.Channel{
			Id:                channel.ID,
			Platform:          channel.Platform,
			PlatformChannelId: channel.PlatformChannelID,
			ChannelName:       channel.ChannelName,
			ChannelHandle:     channel.ChannelHandle,
			ChannelAvatarUrl:  channel.ChannelAvatarUrl,
			ConnectionState:   channel.ConnectionState,
			ConnectedAt:       channel.ConnectedAt,
			Status:            channel.Status,
			CreatedAt:         channel.CreatedAt,
			UpdatedAt:         channel.UpdatedAt,
		},
	}, nil
}
