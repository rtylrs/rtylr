package time_off

import (
	"context"
	"database/sql"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/time_off/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/time_off/internal/api"
	"github.com/voxire/onion/internal/modules/hr/time_off/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

// RegisterAPI is intentionally empty: the hr.time_off.v1 OneKit RPC surface
// (listTimeOff/submitTimeOff) was removed after its only consumer — the dash
// home PendingWidget — was deleted. HR surfaces read through hr/compat/v1 and
// the approve REST route below.
func (m *Module) RegisterAPI(_ *http.ServeMux) {}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	serverHttp.RegisterPOSTWithValidation[api.ApproveTimeOffRequest, api.ApproveTimeOffResponse](
		ctx,
		router,
		api.ApproveTimeOffRoutes(),
		&api.ApproveTimeOffHandler{Service: m.service},
	)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
