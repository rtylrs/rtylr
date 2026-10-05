package locations

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/locations/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/locations/internal/api"
	"github.com/voxire/onion/internal/modules/pos/locations/internal/application"
	poslocationv1 "github.com/voxire/voxie/gen/go/pos/location/v1"
)

type Module struct {
	service    *application.Service
	repository *moduleinfra.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := moduleinfra.NewRepository(db)
	return &Module{service: application.NewService(ctx, repo), repository: repo}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := poslocationv1.RegisterLocationServiceServer(
		api.NewLocationAPI(m.service),
		poslocationv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS location service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

// StockBranchOwner returns the current branch owner within a business, including
// inactive branches referenced by paid orders. Missing or foreign branches return sql.ErrNoRows.
func (m *Module) StockBranchOwner(ctx context.Context, businessID, branchID int64) (int64, error) {
	return m.repository.StockBranchOwner(ctx, businessID, branchID)
}

func (m *Module) WithLocationLimitReader(reader func(context.Context, int64) (int64, bool, error)) *Module {
	m.repository.WithLocationLimitReader(reader)
	return m
}
