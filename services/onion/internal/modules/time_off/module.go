package time_off

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/time_off/internal/api"
	"github.com/voxire/onion/internal/modules/time_off/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	hrtimeoffv1 "github.com/voxire/voxie/gen/go/hr/time_off/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrtimeoffv1.RegisterTimeOffServiceServer(
		api.NewTimeOffAPI(m.service),
		hrtimeoffv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register HR time off service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
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
