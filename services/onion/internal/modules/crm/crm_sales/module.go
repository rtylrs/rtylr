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
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

type Module struct {
	service    *application.Service
	repository *moduleinfra.Repository
	links      customerLinkWriter
}

type customerLinkWriter interface {
	LinkCustomerToContact(context.Context, int64, int64, int64) error
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repository := moduleinfra.NewRepository(db)
	return &Module{
		service:    application.NewService(ctx, repository),
		links:      repository,
		repository: repository,
	}
}

func (m *Module) WithBusinessOwnerResolver(resolver application.BusinessOwnerResolver) *Module {
	m.service.WithBusinessOwnerResolver(resolver)
	return m
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
	if linked.SourceType != "crm_contact" {
		return nil
	}
	if event.BusinessID <= 0 || linked.SourceID <= 0 || linked.CustomerID <= 0 {
		return fmt.Errorf("crm customer link event requires positive business, source, and customer ids")
	}
	if err := m.links.LinkCustomerToContact(ctx, event.BusinessID, linked.SourceID, linked.CustomerID); err != nil {
		return err
	}
	m.service.CustomerLinkUpdated(ctx, event.BusinessID)
	return nil
}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

// RegisterAPI mounts the generated crm.sales.v1 contract service on the
// shared oneKit mux (#809). The legacy chi routes stay until every caller is
// switched to the generated client.
func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := crmsalesv1.RegisterContractServiceServer(
		api.NewContractAPI(m.service),
		crmsalesv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register crm sales service: %w", err))
	}
}

func (m *Module) WithMembershipReader(reader func(context.Context, int64, int64) (bool, error)) *Module {
	m.repository.WithMembershipReader(reader)
	return m
}

// Rep is a teammate a lead can be assigned to.
type Rep struct {
	ID   int64
	Name string
}

// WithRepReader supplies the business's assignable teammates (owner and
// active members) for the lead rep picker.
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
