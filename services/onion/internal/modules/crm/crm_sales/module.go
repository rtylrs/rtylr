package crmsales

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/api"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	moduleinfra "github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/eventbus"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

// Module is CRM sales: leads, deals in pipelines, activities, tasks, lead
// imports and contracts. docs/crm-model.md describes the flow.
type Module struct {
	service    *application.Service
	repository *moduleinfra.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repository := moduleinfra.NewRepository(db)
	return &Module{
		service:    application.NewService(ctx, repository),
		repository: repository,
	}
}

// DispatchEvent records the customer a converted lead's person became. Links
// for pre-redesign lead contacts still in flight are ignored: nothing reads
// those contacts any more.
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
	if linked.SourceType != moduleinfra.LeadSubjectType {
		return nil
	}
	if event.BusinessID <= 0 || linked.SourceID <= 0 || linked.CustomerID <= 0 {
		return fmt.Errorf("crm customer link event requires positive business, source, and customer ids")
	}
	return m.service.LinkLeadCustomer(ctx, event.BusinessID, linked.SourceID, linked.CustomerID)
}

// RegisterAPI mounts the generated crm.sales.v1 services on the shared oneKit
// mux.
func (m *Module) RegisterAPI(mux *http.ServeMux) {
	salesAPI := api.New(m.service)
	registrations := []struct {
		name     string
		register func() error
	}{
		{"lead", func() error { return crmsalesv1.RegisterLeadServiceServer(mux, salesAPI) }},
		{"activity", func() error { return crmsalesv1.RegisterActivityServiceServer(mux, salesAPI) }},
		{"deal", func() error { return crmsalesv1.RegisterDealServiceServer(mux, salesAPI) }},
		{"lead import", func() error { return crmsalesv1.RegisterLeadImportServiceServer(mux, salesAPI) }},
		{"contract", func() error { return crmsalesv1.RegisterContractServiceServer(mux, salesAPI) }},
	}
	for _, registration := range registrations {
		if err := registration.register(); err != nil {
			panic(fmt.Errorf("register crm sales %s service: %w", registration.name, err))
		}
	}
}

func (m *Module) WithMembershipReader(reader func(context.Context, int64, int64) (bool, error)) *Module {
	m.repository.WithMembershipReader(reader)
	return m
}

// Rep is a teammate work can be assigned to.
type Rep struct {
	ID   int64
	Name string
}

// WithRepReader supplies the business's assignable teammates (owner and
// active members) for the rep pickers.
func (m *Module) WithRepReader(reader func(context.Context, int64) ([]Rep, error)) *Module {
	m.repository.WithRepReader(func(ctx context.Context, businessID int64) ([]domain.Rep, error) {
		reps, err := reader(ctx, businessID)
		if err != nil {
			return nil, err
		}
		result := make([]domain.Rep, 0, len(reps))
		for _, rep := range reps {
			result = append(result, domain.Rep(rep))
		}
		return result, nil
	})
	return m
}
