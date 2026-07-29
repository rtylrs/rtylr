package billing

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/billing/internal/api"
	"github.com/voxire/onion/internal/modules/billing/internal/application"
	hubbillingv1 "github.com/voxire/voxie/gen/go/hub/billing/v1"
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
	if err := hubbillingv1.RegisterBillingServiceServer(
		api.NewBillingAPI(m.service),
		hubbillingv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub billing service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
