package pos_gift_cards

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/api"
	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/application"
	posgiftcardv1 "github.com/voxire/voxie/gen/go/pos/gift_card/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posgiftcardv1.RegisterGiftCardServiceServer(
		api.NewGiftCardAPI(m.service),
		posgiftcardv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS gift card service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
