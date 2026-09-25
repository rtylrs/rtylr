package training

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/training/internal/api"
	"github.com/voxire/onion/internal/modules/training/internal/application"
	hrtrainingv1 "github.com/voxire/voxie/gen/go/hr/training/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrtrainingv1.RegisterTrainingServiceServer(
		api.NewTrainingAPI(m.service),
		hrtrainingv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register training service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
