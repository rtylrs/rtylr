package users

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos/users/internal/api"
	"github.com/voxire/onion/internal/modules/pos/users/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/pos/users/internal/infrastructure"
	posuserv1 "github.com/voxire/voxie/gen/go/pos/user/v1"
)

type Module struct {
	service    *application.Service
	repository *moduleinfra.Repository
}

func NewModule(_ context.Context, db *sql.DB) *Module {
	repo := moduleinfra.NewRepository(db)
	return &Module{service: application.NewService(repo, moduleinfra.NewPINAttemptStore(db)), repository: repo}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posuserv1.RegisterUserServiceServer(
		api.NewUserAPI(m.service),
		posuserv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS user service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string { return []string{} }

func (m *Module) WithAuthorizationVersionReader(reader func(context.Context, int64, int64) (int64, error)) *Module {
	m.repository.WithAuthorizationVersionReader(reader)
	return m
}
