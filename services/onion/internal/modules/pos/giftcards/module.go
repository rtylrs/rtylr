package giftcards

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/giftcards/internal/infrastructure"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/pos/giftcards/internal/api"
	"github.com/voxire/onion/internal/modules/pos/giftcards/internal/application"
	"github.com/voxire/onion/internal/platform/eventbus"
	posgiftcardv1 "github.com/voxire/voxie/gen/go/pos/gift_card/v1"
)

type Module struct {
	service *application.Service
	links   customerLinkWriter
}

type customerLinkWriter interface {
	LinkCustomerToGiftCard(context.Context, int64, int64, int64) error
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repository := moduleinfra.NewRepository(db)
	return &Module{service: application.NewService(ctx, repository), links: repository}
}

func (m *Module) DispatchEvent(ctx context.Context, event eventbus.Event) error {
	if event.Type != crmcontract.CustomerLinkedType {
		return nil
	}
	payload, err := eventbus.PayloadJSON(event.Payload)
	if err != nil {
		return fmt.Errorf("encode crm customer link event: %w", err)
	}
	var linked crmcontract.CustomerLinked
	if err := json.Unmarshal(payload, &linked); err != nil {
		return fmt.Errorf("decode crm customer link event: %w", err)
	}
	if linked.SourceType != "pos_gift_card" {
		return nil
	}
	if event.BusinessID <= 0 || linked.SourceID <= 0 || linked.CustomerID <= 0 {
		return fmt.Errorf("crm customer link event requires positive business, source, and customer ids")
	}
	return m.links.LinkCustomerToGiftCard(ctx, event.BusinessID, linked.SourceID, linked.CustomerID)
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
