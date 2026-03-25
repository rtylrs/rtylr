package rtylr

import (
	"context"

	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/rtylr/echo"
	"github.com/voxire/onion/internal/platforms/rtylr/memory"
)

type Platform struct {
	Memory *memory.Module
	Echo   *echo.Module
}

func NewPlatform(ctx context.Context) *Platform {
	memoryModule := memory.NewModule(ctx)

	return &Platform{
		Memory: memoryModule,
		Echo:   echo.NewModule(ctx, memoryModule),
	}
}

func (p *Platform) ConfigureAccessPolicy(_ context.Context, accessPolicy *http.RouteAccessPolicy) {
	accessPolicy.RegisterAgentRoutes("rtylr/1.0", "/invalidate-cache", "/status")
}
