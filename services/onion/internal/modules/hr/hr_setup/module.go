package hr_setup

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_setup/internal/infrastructure"

	"github.com/go-chi/chi/v5"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_setup/internal/api"
	"github.com/voxire/onion/internal/modules/hr/hr_setup/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	hrsetupv1 "github.com/voxire/voxie/gen/go/hr/setup/v1"
)

type Module struct {
	service    *application.Service
	coreModule *hrcore.Module
}

func NewModule(ctx context.Context, db *sql.DB, coreModule *hrcore.Module) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db)), coreModule: coreModule}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrsetupv1.RegisterSetupServiceServer(
		api.NewSetupAPI(m.service),
		hrsetupv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register HR setup service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	serverHttp.RegisterGET[api.GetSetupStatusRequest, api.GetSetupStatusResponse](
		ctx, router, api.GetSetupStatusRoutes(), &api.GetSetupStatusHandler{Service: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.CompleteSetupStepRequest, api.CompleteSetupStepResponse](
		ctx, router, api.CompleteSetupStepRoutes(), &api.CompleteSetupStepHandler{Service: m.service},
	)
	// /hr/config is the frontend compatibility endpoint. It composes setup
	// state with the business-scoped HR-core configuration.
	serverHttp.RegisterGET[api.GetHRConfigRequest, api.GetHRConfigResponse](
		ctx, router, api.GetHRConfigRoutes(), &api.GetHRConfigHandler{Service: m.service, CoreService: m.coreModule},
	)
	serverHttp.RegisterPUTWithValidation[api.UpsertHRConfigRequest, api.UpsertHRConfigResponse](
		ctx, router, api.UpsertHRConfigRoutes(), &api.UpsertHRConfigHandler{Service: m.service, CoreService: m.coreModule},
	)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
