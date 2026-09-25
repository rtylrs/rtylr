package hr_onboarding

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/api"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	serverHttp.RegisterGET[api.GetEmployeeChecklistRequest, api.GetEmployeeChecklistResponse](
		ctx, router, api.GetEmployeeChecklistRoutes(), &api.GetEmployeeChecklistHandler{Service: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.CompleteChecklistItemRequest, api.CompleteChecklistItemResponse](
		ctx, router, api.CompleteChecklistItemRoutes(), &api.CompleteChecklistItemHandler{Service: m.service},
	)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	err := hronboardingv1.RegisterOnboardingServiceServer(
		api.NewOnboardingService(m.service),
		hronboardingv1.WithMux(mux),
	)
	if err != nil {
		panic(err)
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
