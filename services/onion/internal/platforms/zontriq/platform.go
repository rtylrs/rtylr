package zontriq

import (
	"context"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/zontriq/audio"
	"github.com/voxire/onion/internal/platforms/zontriq/automations"
	"github.com/voxire/onion/internal/platforms/zontriq/channels"
	"github.com/voxire/onion/internal/platforms/zontriq/credits"
	"github.com/voxire/onion/internal/platforms/zontriq/series"
	"github.com/voxire/onion/internal/platforms/zontriq/users"
	"github.com/voxire/onion/internal/platforms/zontriq/videos"
	"github.com/voxire/onion/internal/platforms/zontriq/voices"
)

type Platform struct {
	Users       *users.Module
	Channels    *channels.Module
	Series      *series.Module
	Videos      *videos.Module
	Automations *automations.Module
	Audio       *audio.Module
	Voices      *voices.Module
	Credits     *credits.Module
	Provisioner authPkg.PlatformProvisioner
}

func NewPlatform(ctx context.Context) *Platform {
	usersModule := users.NewModule(ctx)

	return &Platform{
		Users:       usersModule,
		Channels:    channels.NewModule(ctx),
		Series:      series.NewModule(ctx),
		Videos:      videos.NewModule(ctx),
		Automations: automations.NewModule(ctx),
		Audio:       audio.NewModule(ctx),
		Voices:      voices.NewModule(ctx),
		Credits:     credits.NewModule(ctx),
		Provisioner: NewProvisioner(usersModule.Adapter),
	}
}

func (p *Platform) ConfigureAccessPolicy(_ context.Context, accessPolicy *http.RouteAccessPolicy) {
	accessPolicy.RegisterAgentRoutes("zontriq/1.0", "/invalidate-cache", "/status")
}
