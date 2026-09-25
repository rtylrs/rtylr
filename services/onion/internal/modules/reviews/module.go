package reviews

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/reviews/internal/api"
	"github.com/voxire/onion/internal/modules/reviews/internal/application"
	hrreviewv1 "github.com/voxire/voxie/gen/go/hr/review/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrreviewv1.RegisterReviewServiceServer(
		api.NewReviewAPI(m.service),
		hrreviewv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hr review service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
