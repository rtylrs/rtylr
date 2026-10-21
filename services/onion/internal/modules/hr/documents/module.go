package documents

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/documents/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/documents/internal/api"
	"github.com/voxire/onion/internal/modules/hr/documents/internal/application"
	hrdocumentv1 "github.com/voxire/voxie/gen/go/hr/document/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrdocumentv1.RegisterDocumentServiceServer(
		api.NewDocumentAPI(m.service),
		hrdocumentv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register document service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
