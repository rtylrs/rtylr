package pos_tables

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_tables/internal/api"
	"github.com/voxire/onion/internal/modules/pos_tables/internal/application"
	postablev1 "github.com/voxire/voxie/gen/go/pos/table/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := postablev1.RegisterTableServiceServer(
		api.NewTableAPI(m.service),
		postablev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS table service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
