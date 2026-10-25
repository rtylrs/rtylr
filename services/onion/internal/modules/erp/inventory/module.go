package inventory

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/erp/inventory/internal/infrastructure"

	poscontract "github.com/voxire/onion/internal/contracts/pos"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/api"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/application"
	posstockapi "github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/api"
	posstockapplication "github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/application"
	posstockdomain "github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/domain"
	posstockinfra "github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/infrastructure"
	"github.com/voxire/onion/internal/platform/eventbus"
	erpinventoryv1 "github.com/voxire/voxie/gen/go/erp/inventory/v1"
	posstockv1 "github.com/voxire/voxie/gen/go/pos/stock/v1"
)

type Module struct {
	service            *application.Service
	posStockService    *posstockapplication.Service
	posStockRepository *posstockinfra.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	posRepo := posstockinfra.NewRepository(db)
	return &Module{
		service:            application.NewService(ctx, moduleinfra.NewRepository(db)),
		posStockService:    posstockapplication.NewService(ctx, posRepo),
		posStockRepository: posRepo,
	}
}

func (m *Module) DispatchEvent(ctx context.Context, event eventbus.Event) error {
	var change posstockdomain.RecordedOrderInventory
	switch event.Type {
	case poscontract.OrderPaidType:
		if event.SubjectType != "pos_order" {
			return fmt.Errorf("POS order paid event has unexpected subject type %q", event.SubjectType)
		}
		var payload poscontract.PaidOrder
		if err := eventbus.DecodePayload(event.Payload, &payload); err != nil {
			return fmt.Errorf("decode POS order paid event: %w", err)
		}
		change = posstockdomain.RecordedOrderInventory{
			BusinessID:   event.BusinessID,
			SourceType:   event.SubjectType,
			SourceID:     event.SubjectID,
			LocationKind: "pos_branch",
			LocationID:   payload.BranchID,
			Channel:      "pos",
			Items:        make([]posstockdomain.OrderInventoryItem, 0, len(payload.Items)),
		}
		for _, item := range payload.Items {
			change.Items = append(change.Items, posstockdomain.OrderInventoryItem{ProductID: item.ProductID, Quantity: item.Quantity})
		}
	default:
		return nil
	}
	if event.BusinessID <= 0 || event.SubjectID <= 0 {
		return fmt.Errorf("order inventory event requires positive business and subject ids")
	}
	return m.posStockService.ApplyRecordedOrderInventory(ctx, change)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := erpinventoryv1.RegisterInventoryServiceServer(
		api.NewInventoryAPI(m.service),
		erpinventoryv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP inventory service: %w", err))
	}
	if err := posstockv1.RegisterStockServiceServer(
		posstockapi.NewStockAPI(m.posStockService),
		posstockv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS stock service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

// WithPOSBranchReader supplies the POS check that a branch belongs to a
// business.
func (m *Module) WithPOSBranchReader(exists func(context.Context, int64, int64) (bool, error)) *Module {
	m.posStockRepository.WithBranchReader(exists)
	return m
}
