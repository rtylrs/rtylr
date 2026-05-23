package console

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/console/internal/application"
	"github.com/voxire/onion/internal/modules/console/internal/routes"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	svc := application.NewService(ctx, db)

	jwtSecret, _ := ctx.Value(jwtutils.JWTSecretKey).(string)
	jwtDuration, _ := ctx.Value(jwtutils.JWTDurationKey).(time.Duration)
	if jwtSecret != "" {
		svc.Initialize(ctx, jwtSecret, jwtDuration)
	}

	return &Module{service: svc}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/console/auth/login"}
}
