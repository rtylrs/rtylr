package memberlocations

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos/memberlocations/internal/api"
	"github.com/voxire/onion/internal/modules/pos/memberlocations/internal/application"
	"github.com/voxire/onion/internal/modules/pos/memberlocations/internal/infrastructure"
	posaccessv1 "github.com/voxire/voxie/gen/go/pos/access/v1"
)

type Module struct {
	service    *application.Service
	repository *infrastructure.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := infrastructure.NewRepository(db)
	return &Module{service: application.NewService(repo), repository: repo}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posaccessv1.RegisterAccessServiceServer(
		api.NewAccessAPI(m.service),
		posaccessv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS access service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string { return []string{} }

func (m *Module) WithIdentityReaders(name func(context.Context, int64, int64) (string, error), member func(context.Context, int64, int64) (bool, error)) *Module {
	m.repository.WithIdentityReaders(name, member)
	return m
}
