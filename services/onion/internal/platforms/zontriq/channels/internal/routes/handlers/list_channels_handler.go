package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	channelsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/channels/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/application"
)

func ListChannelsRoutes() []string {
	return []string{
		"/zontriq/channels",
	}
}

type ListChannelsHandler struct {
	ChannelService *application.Service
}

func (h *ListChannelsHandler) Serve(ctx context.Context, req *channelsv1.ListChannelsRequest) (channelsv1.ListChannelsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return channelsv1.ListChannelsResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	zontriqUserID := int32(device.User.PlatformUserID)

	platform := req.GetPlatform()

	channels, err := h.ChannelService.ListChannelsByUserID(ctx, zontriqUserID, platform)
	if err != nil {
		return channelsv1.ListChannelsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list channels",
		}
	}

	protoChannels := make([]*channelsv1.Channel, 0, len(channels))
	for i := range channels {
		ch := &channels[i]
		protoChannels = append(protoChannels, &channelsv1.Channel{
			Id:                ch.ID,
			Platform:          ch.Platform,
			PlatformChannelId: ch.PlatformChannelID,
			ChannelName:       ch.ChannelName,
			ChannelHandle:     ch.ChannelHandle,
			ChannelAvatarUrl:  ch.ChannelAvatarUrl,
			ConnectionState:   ch.ConnectionState,
			ConnectedAt:       ch.ConnectedAt,
			Status:            ch.Status,
			CreatedAt:         ch.CreatedAt,
			UpdatedAt:         ch.UpdatedAt,
		})
	}

	return channelsv1.ListChannelsResponse{
		Channels: protoChannels,
	}, nil
}
