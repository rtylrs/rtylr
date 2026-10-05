package receipts

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/receipts/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/receipts/internal/api"
	"github.com/voxire/onion/internal/modules/pos/receipts/internal/application"
	posreceiptv1 "github.com/voxire/voxie/gen/go/pos/receipt/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posreceiptv1.RegisterReceiptServiceServer(
		api.NewReceiptAPI(m.service),
		posreceiptv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS receipt service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
