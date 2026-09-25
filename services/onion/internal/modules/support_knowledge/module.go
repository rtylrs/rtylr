package supportknowledge

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/support_knowledge/internal/api"
	"github.com/voxire/onion/internal/modules/support_knowledge/internal/application"
	supportknowledgev1 "github.com/voxire/voxie/gen/go/support/knowledge/v1"
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
