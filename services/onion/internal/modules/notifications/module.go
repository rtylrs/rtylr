package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/notifications/internal/api"
	"github.com/voxire/onion/internal/modules/notifications/internal/application"
	hubnotificationv1 "github.com/voxire/voxie/gen/go/hub/notification/v1"
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
	if err := hubnotificationv1.RegisterNotificationServiceServer(
		api.NewNotificationAPI(m.service),
		hubnotificationv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub notification service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/dash/ws"}
}
