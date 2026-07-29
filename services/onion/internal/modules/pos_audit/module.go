package pos_audit

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/api"
	"github.com/voxire/onion/internal/modules/pos_audit/internal/application"
	posauditv1 "github.com/voxire/voxie/gen/go/pos/audit/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
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
