package hr_setup

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	hrsetupv1 "github.com/voxire/voxie/gen/go/hr/setup/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/api"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrsetupv1.RegisterSetupServiceServer(
		api.NewSetupAPI(m.service),
		hrsetupv1.WithMux(mux),
		hrsetupv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
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
	serverHttp.RegisterGET[api.GetHRConfigRequest, api.GetHRConfigResponse](
		ctx, router, api.GetHRConfigRoutes(), &api.GetHRConfigHandler{Service: m.service},
	)
	serverHttp.RegisterPUTWithValidation[api.UpsertHRConfigRequest, api.UpsertHRConfigResponse](
		ctx, router, api.UpsertHRConfigRoutes(), &api.UpsertHRConfigHandler{Service: m.service},
	)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
