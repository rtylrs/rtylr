package daysummary

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/daysummary/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/daysummary/internal/api"
	"github.com/voxire/onion/internal/modules/pos/daysummary/internal/application"
	possummaryv1 "github.com/voxire/voxie/gen/go/pos/day_summary/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := possummaryv1.RegisterDaySummaryServiceServer(
		api.NewDaySummaryAPI(m.service),
		possummaryv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS day summary service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
