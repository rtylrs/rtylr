package hubmenu

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/api"
	"github.com/voxire/onion/internal/modules/hub_menu/internal/application"
	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	moduleinfra "github.com/voxire/onion/internal/modules/hub_menu/internal/infrastructure"
	hubmenuv1 "github.com/voxire/voxie/gen/go/hub/menu/v1"
)

type (
	HostResolver           = application.HostResolver
	EventDispatcher        = application.EventDispatcher
	AITranslator           = application.AITranslator
	TenantValidator        = application.TenantValidator
	MenuProfile            = domain.MenuProfile
	SaveMenuProfileDraftIn = domain.SaveMenuProfileDraftInput
	TranslationSourceEntry = domain.TranslationSourceEntry
	MenuTranslation        = domain.MenuTranslation
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) WithHostResolver(fn HostResolver) *Module {
	m.service.WithHostResolver(fn)
	return m
}

func (m *Module) WithEventDispatcher(fn EventDispatcher) *Module {
	m.service.WithEventDispatcher(fn)
	return m
}

func (m *Module) WithTranslator(fn AITranslator) *Module {
	m.service.WithTranslator(fn)
	return m
}

func (m *Module) WithTenantValidator(v TenantValidator) *Module {
	m.service.WithTenantValidator(v)
	return m
}

// GetAnonymousRoutes exposes the anonymous customer-facing branded menu endpoints
// so the HTTP auth middleware allows unauthenticated guest access.
func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/hub/menu/v1/public/resolve-host",
		"/hub/menu/v1/public/bundle",
		"/hub/menu/v1/public/table-actions",
		"/hub/menu/v1/public/events",
		"/hub/menu/v1/public/stamps",
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubmenuv1.RegisterMenuServiceServer(
		api.NewMenuAPI(m.service),
		hubmenuv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub menu service: %w", err))
	}
}

// Exported cross-module helpers for POS Orders and Caret AI tools.

func (m *Module) ResolvePaymentRedirectURL(ctx context.Context, businessID, locationID int64) string {
	return m.service.ResolvePaymentRedirectURL(ctx, businessID, locationID)
}

func (m *Module) ResolveBranchItemOverride(ctx context.Context, businessID, locationID, menuItemID int64) (visible bool, hasPriceOverride bool, priceOverride float64, err error) {
	return m.service.ResolveBranchItemOverride(ctx, businessID, locationID, menuItemID)
}

func (m *Module) AwardStampsForOrder(ctx context.Context, businessID, locationID, orderID int64, phone, name string, orderTotal float64) error {
	return m.service.AwardStampsForOrder(ctx, businessID, locationID, orderID, phone, name, orderTotal)
}

func (m *Module) GetMenuProfile(ctx context.Context, businessID, locationID int64) (MenuProfile, error) {
	return m.service.GetMenuProfile(ctx, businessID, locationID)
}

func (m *Module) SaveMenuProfileDraft(ctx context.Context, businessID int64, input SaveMenuProfileDraftIn) (MenuProfile, error) {
	return m.service.SaveMenuProfileDraft(ctx, businessID, input)
}

func (m *Module) PublishMenuProfile(ctx context.Context, businessID, locationID int64) (MenuProfile, error) {
	return m.service.PublishMenuProfile(ctx, businessID, locationID)
}

func (m *Module) CloneMenuProfile(ctx context.Context, businessID, sourceLocationID, targetLocationID int64) (MenuProfile, error) {
	return m.service.CloneMenuProfile(ctx, businessID, sourceLocationID, targetLocationID)
}
