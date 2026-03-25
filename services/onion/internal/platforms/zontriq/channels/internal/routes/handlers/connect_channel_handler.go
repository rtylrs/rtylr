package handlers

import (
	"context"
	"net/http"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	channelsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/channels/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/infrastructure"
)

func ConnectChannelRoutes() []string {
	return []string{
		"/zontriq/channels",
	}
}

type ConnectChannelHandler struct {
	ChannelService *application.Service
}

func (h *ConnectChannelHandler) Validate(_ context.Context, req *channelsv1.ConnectChannelRequest) *errorModels.ServerError {
	if req.Platform == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Platform is required",
		}
	}

	if req.AccessToken == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Access token is required",
		}
	}

	if req.PlatformChannelId == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Platform channel ID is required",
		}
	}

	if req.ChannelName == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Channel name is required",
		}
	}

	return nil
}

func (h *ConnectChannelHandler) Serve(ctx context.Context, req *channelsv1.ConnectChannelRequest) (channelsv1.ConnectChannelResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return channelsv1.ConnectChannelResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	zontriqUserID := int32(device.User.PlatformUserID)

	var tokenExpiresAt time.Time

	if req.TokenExpiresAt != "" {
		var err error

		tokenExpiresAt, err = time.Parse(time.RFC3339, req.TokenExpiresAt)
		if err != nil {
			return channelsv1.ConnectChannelResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadRequest,
				Message: "Invalid token_expires_at format, expected RFC3339",
			}
		}
	}

	input := infrastructure.CreateChannelInput{
		ZontriqUserID:     zontriqUserID,
		Platform:          req.Platform,
		PlatformChannelID: req.PlatformChannelId,
		ChannelName:       req.ChannelName,
		ChannelHandle:     req.ChannelHandle,
		ChannelAvatarUrl:  req.ChannelAvatarUrl,
		AccessToken:       req.AccessToken,
		RefreshToken:      req.RefreshToken,
		TokenExpiresAt:    tokenExpiresAt,
	}

	channel, err := h.ChannelService.ConnectChannel(ctx, input)
	if err != nil {
		return channelsv1.ConnectChannelResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to connect channel",
		}
	}

	return channelsv1.ConnectChannelResponse{
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
