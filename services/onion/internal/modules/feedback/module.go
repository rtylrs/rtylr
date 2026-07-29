package feedback

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/feedback/internal/api"
	"github.com/voxire/onion/internal/modules/feedback/internal/application"
	crmfeedbackv1 "github.com/voxire/voxie/gen/go/crm/feedback/v1"
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
	if err := crmfeedbackv1.RegisterFeedbackServiceServer(
		api.NewFeedbackAPI(m.service),
		crmfeedbackv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register crm feedback service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
