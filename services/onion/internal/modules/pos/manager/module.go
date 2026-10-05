package manager

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos/manager/internal/api"
	"github.com/voxire/onion/internal/modules/pos/manager/internal/application"
	"github.com/voxire/onion/internal/modules/pos/manager/internal/infrastructure"
	posmanagerv1 "github.com/voxire/voxie/gen/go/pos/manager/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := infrastructure.NewRepository(db)
	return &Module{service: application.NewService(repo, repo, nil, infrastructure.NewPINAttemptStore(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posmanagerv1.RegisterManagerServiceServer(
		api.NewManagerAPI(m.service),
		posmanagerv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS manager service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) WithPasswordVerifier(read func(context.Context, int64, string) (bool, error)) *Module {
	m.service.WithPasswordVerifier(read)
	return m
}
