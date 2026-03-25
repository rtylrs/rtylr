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

func DisconnectChannelRoutes() []string {
	return []string{
		"/zontriq/channels/{id}",
	}
}

type DisconnectChannelHandler struct {
	ChannelService *application.Service
}

func (h *DisconnectChannelHandler) Serve(ctx context.Context, _ *channelsv1.DisconnectChannelRequest) (channelsv1.DisconnectChannelResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return channelsv1.DisconnectChannelResponse{}, &errorModels.ServerError{
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
		return channelsv1.DisconnectChannelResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid channel ID",
		}
	}

	err = h.ChannelService.DisconnectChannel(ctx, int32(channelID))
	if err != nil {
		return channelsv1.DisconnectChannelResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to disconnect channel",
		}
	}

	return channelsv1.DisconnectChannelResponse{}, nil
}
