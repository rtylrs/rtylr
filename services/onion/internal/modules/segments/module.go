package segments

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/segments/internal/api"
	"github.com/voxire/onion/internal/modules/segments/internal/application"
	crmsegmentv1 "github.com/voxire/voxie/gen/go/crm/segment/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := crmsegmentv1.RegisterSegmentServiceServer(
		api.NewSegmentAPI(m.service),
		crmsegmentv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register segment service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
