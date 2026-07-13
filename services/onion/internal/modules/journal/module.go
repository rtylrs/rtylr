package journal

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	financejournalv1 "github.com/voxire/voxie/gen/go/finance/journal/v1"

	"github.com/voxire/onion/internal/modules/journal/internal/api"
	"github.com/voxire/onion/internal/modules/journal/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financejournalv1.RegisterJournalServiceServer(
		api.NewJournalAPI(m.service),
		financejournalv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance journal service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
