package oauth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/organization/oauth/internal/infrastructure"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/organization/oauth/internal/api"
	"github.com/voxire/onion/internal/modules/organization/oauth/internal/application"
	commonoauthv1 "github.com/voxire/voxie/gen/go/common/oauth/v1"
)

type Module struct {
	service    *application.Service
	authAppURL string
}

func NewModule(ctx context.Context, db *sql.DB, authAppURL string) *Module {
	return &Module{
		service:    application.NewService(ctx, moduleinfra.NewRepository(db)),
		authAppURL: authAppURL,
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := commonoauthv1.RegisterOauthServiceServer(
		api.NewOauthAPI(m.service),
		commonoauthv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register common oauth service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	_ = ctx
	router.Get("/oauth/authorize", (&api.AuthorizeHandler{
		Service:    m.service,
		AuthAppURL: m.authAppURL,
	}).ServeHTTP)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/oauth/authorize",
		"/oauth/token",
		"/oauth/revoke",
		"/common/oauth/v1/token",
		"/common/oauth/v1/revoke",
	}
}
