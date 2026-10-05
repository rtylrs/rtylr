package supportknowledge

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/support/knowledge/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/support/knowledge/internal/api"
	"github.com/voxire/onion/internal/modules/support/knowledge/internal/application"
	supportknowledgev1 "github.com/voxire/voxie/gen/go/support/knowledge/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := supportknowledgev1.RegisterKnowledgeServiceServer(
		api.NewKnowledgeAPI(m.service),
		supportknowledgev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register support knowledge service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
