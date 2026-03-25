package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	channelsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/channels/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/application"
)

func GetChannelRoutes() []string {
	return []string{
		"/zontriq/channels/{id}",
	}
}

type GetChannelHandler struct {
	ChannelService *application.Service
}

func (h *GetChannelHandler) Serve(ctx context.Context, _ *channelsv1.GetChannelRequest) (channelsv1.GetChannelResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return channelsv1.GetChannelResponse{}, &errorModels.ServerError{
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
		return channelsv1.GetChannelResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid channel ID",
		}
	}

	channel, err := h.ChannelService.GetChannelByID(ctx, int32(channelID))
	if err != nil {
		return channelsv1.GetChannelResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get channel",
		}
	}

	return channelsv1.GetChannelResponse{
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
