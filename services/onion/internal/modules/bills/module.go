package bills

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/bills/internal/api"
	"github.com/voxire/onion/internal/modules/bills/internal/application"
	financebillv1 "github.com/voxire/voxie/gen/go/finance/bill/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
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
