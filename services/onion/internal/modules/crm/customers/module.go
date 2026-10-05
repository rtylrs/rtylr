package customers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	moduleinfra "github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure"

	"github.com/go-chi/chi/v5"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/api"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/platform/eventbus"
	crmcustomerv1 "github.com/voxire/voxie/gen/go/crm/customer/v1"
	crmdashboardv1 "github.com/voxire/voxie/gen/go/crm/dashboard/v1"
	crmloyaltyv1 "github.com/voxire/voxie/gen/go/crm/loyalty/v1"
	poscustomerv1 "github.com/voxire/voxie/gen/go/pos/customer/v1"
)

type Module struct {
	service    *application.Service
	repository *moduleinfra.Repository
}

type EmailRecipient struct {
	Email string
	Name  string
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := moduleinfra.NewRepository(db)
	return &Module{
		service:    application.NewService(ctx, repo),
		repository: repo,
	}
}

func (m *Module) WithBusinessOwnerResolver(resolve func(context.Context, int64) (int64, error)) *Module {
	m.service.WithBusinessOwnerResolver(businessOwnerResolverAdapter{resolve: resolve})
	return m
}

type businessOwnerResolverAdapter struct {
	resolve func(context.Context, int64) (int64, error)
}

func (a businessOwnerResolverAdapter) GetBusinessOwnerID(ctx context.Context, businessID int64) (int64, error) {
	return a.resolve(ctx, businessID)
}

func (m *Module) DispatchEvent(ctx context.Context, event eventbus.Event) error {
	if event.Type == crmcontract.CustomerLinkedType {
		return nil
	}
	if event.Type != crmcontract.CustomerContactCapturedType {
		return nil
	}
	payload, err := eventbus.PayloadJSON(event.Payload)
	if err != nil {
		return fmt.Errorf("encode customer contact event payload: %w", err)
	}
	var contact crmcontract.CustomerContactCaptured
	if err := json.Unmarshal(payload, &contact); err != nil {
		return fmt.Errorf("decode customer contact event payload: %w", err)
	}
	return m.service.CaptureContactFromEvent(ctx, event, contact)
}

func (m *Module) ResolveEmailRecipient(ctx context.Context, businessID, customerID int64) (EmailRecipient, bool, error) {
	customer, err := m.service.GetCustomer(ctx, businessID, customerID)
	if errors.Is(err, domain.ErrCustomerNotFound) {
		return EmailRecipient{}, false, nil
	}
	if err != nil {
		return EmailRecipient{}, false, err
	}
	name := customer.FirstName
	if customer.LastName != "" {
		if name != "" {
			name += " "
		}
		name += customer.LastName
	}
	return EmailRecipient{Email: customer.Email, Name: name}, true, nil
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := poscustomerv1.RegisterCustomerServiceServer(api.NewPOSCustomerAPI(m.service), poscustomerv1.WithMux(mux)); err != nil {
		panic(fmt.Errorf("register POS customer service: %w", err))
	}
	if err := crmcustomerv1.RegisterCustomerServiceServer(api.NewCustomerAPI(m.service), crmcustomerv1.WithMux(mux)); err != nil {
		panic(fmt.Errorf("register crm customer service: %w", err))
	}
	if err := crmdashboardv1.RegisterDashboardServiceServer(api.NewDashboardAPI(m.service), crmdashboardv1.WithMux(mux)); err != nil {
		panic(fmt.Errorf("register crm dashboard service: %w", err))
	}

	if err := crmloyaltyv1.RegisterLoyaltyServiceServer(
		api.NewLoyaltyAPI(m.service),
		crmloyaltyv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register crm loyalty service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) WithReportingTimezone(read func(context.Context, int64) (string, error)) *Module {
	m.repository.WithReportingTimezone(read)
	return m
}

func (m *Module) CountCustomersCreatedBetween(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return m.repository.CountCustomersCreatedBetween(ctx, businessID, from, to)
}
