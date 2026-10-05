package audit

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/audit/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/audit/internal/api"
	"github.com/voxire/onion/internal/modules/pos/audit/internal/application"
	posauditv1 "github.com/voxire/voxie/gen/go/pos/audit/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posauditv1.RegisterAuditServiceServer(
		api.NewAuditAPI(m.service),
		posauditv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS audit service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
