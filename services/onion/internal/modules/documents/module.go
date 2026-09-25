package documents

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/documents/internal/api"
	"github.com/voxire/onion/internal/modules/documents/internal/application"
	hrdocumentv1 "github.com/voxire/voxie/gen/go/hr/document/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
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
