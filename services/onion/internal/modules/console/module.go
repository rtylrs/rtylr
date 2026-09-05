package console

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/console/internal/api"
	"github.com/voxire/onion/internal/modules/console/internal/application"
	consoleadminv1 "github.com/voxire/voxie/gen/go/console/admin/v1"
)

type Module struct {
	service *application.Service
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := consoleadminv1.RegisterConsoleAdminServiceServer(
		api.NewConsoleAdminAPI(m.service),
		consoleadminv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register console admin service: %w", err))
	}
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	svc := application.NewService(ctx, db)
	return &Module{service: svc}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return nil
}
