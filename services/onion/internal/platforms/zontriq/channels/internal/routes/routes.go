package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	channelsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/channels/v1"
	http "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, channelService *application.Service) {
	http.RegisterGET[channelsv1.ListChannelsRequest, channelsv1.ListChannelsResponse](
		ctx,
		router,
		handlers.ListChannelsRoutes(),
		&handlers.ListChannelsHandler{ChannelService: channelService},
	)

	http.RegisterGET[channelsv1.GetChannelRequest, channelsv1.GetChannelResponse](
		ctx,
		router,
		handlers.GetChannelRoutes(),
		&handlers.GetChannelHandler{ChannelService: channelService},
	)

	http.RegisterPOSTWithValidation[channelsv1.ConnectChannelRequest, channelsv1.ConnectChannelResponse](
		ctx,
		router,
		handlers.ConnectChannelRoutes(),
		&handlers.ConnectChannelHandler{ChannelService: channelService},
	)

	http.RegisterDELETE[channelsv1.DisconnectChannelRequest, channelsv1.DisconnectChannelResponse](
		ctx,
		router,
		handlers.DisconnectChannelRoutes(),
		&handlers.DisconnectChannelHandler{ChannelService: channelService},
	)

	http.RegisterPOSTWithValidation[channelsv1.RefreshChannelTokenRequest, channelsv1.RefreshChannelTokenResponse](
		ctx,
		router,
		handlers.RefreshChannelTokenRoutes(),
		&handlers.RefreshChannelTokenHandler{ChannelService: channelService},
	)
}
