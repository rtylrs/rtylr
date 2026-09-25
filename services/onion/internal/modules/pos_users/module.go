package pos_users

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_users/internal/api"
	"github.com/voxire/onion/internal/modules/pos_users/internal/application"
	posuserv1 "github.com/voxire/voxie/gen/go/pos/user/v1"
)

type Module struct{ service *application.Service }

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posuserv1.RegisterUserServiceServer(
		api.NewUserAPI(m.service),
		posuserv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS user service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string { return []string{} }
