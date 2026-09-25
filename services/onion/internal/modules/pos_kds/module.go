package pos_kds

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_kds/internal/api"
	"github.com/voxire/onion/internal/modules/pos_kds/internal/application"
	poskdsv1 "github.com/voxire/voxie/gen/go/pos/kds/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := poskdsv1.RegisterKDSServiceServer(
		api.NewKDSAPI(m.service),
		poskdsv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS KDS service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
