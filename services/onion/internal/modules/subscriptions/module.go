package subscriptions

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hubsubscriptionv1 "github.com/voxire/voxie/gen/go/hub/subscription/v1"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/api"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubsubscriptionv1.RegisterSubscriptionServiceServer(
		api.NewSubscriptionAPI(m.service),
		hubsubscriptionv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub subscription service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	_ = ctx
	router.Post("/subscriptions/paddle-webhook", api.PaddleWebhookHTTPHandler(m.service))
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/api/hub/subscription/v1/plans",
		"/subscriptions/paddle-webhook",
	}
}
