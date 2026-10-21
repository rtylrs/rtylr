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

func (m *Module) RegisterAPI(_ *http.ServeMux) {}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	serverHttp.RegisterGET[api.ListTimeOffRequest, []api.TimeOffEntryDTO](
		ctx,
		router,
		api.ListTimeOffRoutes(),
		&api.ListTimeOffHandler{Service: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.SubmitTimeOffRequest, api.TimeOffEntryDTO](
		ctx,
		router,
		api.SubmitTimeOffRoutes(),
		&api.SubmitTimeOffHandler{Service: m.service},
	)
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
