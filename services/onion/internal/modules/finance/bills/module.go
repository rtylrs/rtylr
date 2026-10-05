package bills

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/finance/bills/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/finance/bills/internal/api"
	"github.com/voxire/onion/internal/modules/finance/bills/internal/application"
	financebillv1 "github.com/voxire/voxie/gen/go/finance/bill/v1"
)

type Module struct {
	service *application.Service
	repo    *moduleinfra.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := moduleinfra.NewRepository(db)
	return &Module{repo: repo, service: application.NewService(ctx, repo)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financebillv1.RegisterBillServiceServer(
		api.NewBillAPI(m.service),
		financebillv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance bill service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) WithPurchaseOrderReader(read func(context.Context, int64, int64) (bool, error)) *Module {
	m.repo.WithPurchaseOrderReader(read)
	return m
}
