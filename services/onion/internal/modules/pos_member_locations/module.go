package pos_member_locations

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/api"
	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/application"
	posaccessv1 "github.com/voxire/voxie/gen/go/pos/access/v1"
)

type Module struct{ service *application.Service }

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posaccessv1.RegisterAccessServiceServer(
		api.NewAccessAPI(m.service),
		posaccessv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS access service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string { return []string{} }
