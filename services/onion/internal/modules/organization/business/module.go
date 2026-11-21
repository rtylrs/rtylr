package business

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	businessdomain "github.com/voxire/onion/internal/modules/organization/business/internal/domain"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/business/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/organization/business/internal/api"
	"github.com/voxire/onion/internal/modules/organization/business/internal/application"
	usermodule "github.com/voxire/onion/internal/modules/organization/user"
	hubbusinessv1 "github.com/voxire/voxie/gen/go/hub/business/v1"
)

type Module struct {
	service     *application.Service
	repository  *moduleinfra.Repository
	userService usermodule.OnboardingPort
}

type PublicLocation = businessdomain.PublicLocation

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := moduleinfra.NewRepository(db)
	return &Module{
		service:    application.NewService(ctx, repo),
		repository: repo,
	}
}

func (m *Module) WithUserService(userService usermodule.OnboardingPort) *Module {
	m.userService = userService
	return m
}

func (m *Module) WithPublicLocationReader(reader func(context.Context, int64) ([]PublicLocation, error)) *Module {
	m.service.WithPublicLocationReader(publicLocationReaderAdapter{read: reader})
	return m
}

type publicLocationReaderAdapter struct {
	read func(context.Context, int64) ([]PublicLocation, error)
}

func (a publicLocationReaderAdapter) ListPublicLocations(ctx context.Context, businessID int64) ([]businessdomain.PublicLocation, error) {
	return a.read(ctx, businessID)
}

func (m *Module) IsOwner(ctx context.Context, userID int64) (bool, error) {
	return m.service.IsOwner(ctx, userID)
}

func (m *Module) GetBusinessName(ctx context.Context, businessID int64) (string, error) {
	profile, err := m.service.GetProfile(ctx, businessID)
	if err != nil {
		return "", err
	}
	if profile.SiteName != "" {
		return profile.SiteName, nil
	}
	return profile.Name, nil
}

func (m *Module) GetBusinessOwnerID(ctx context.Context, businessID int64) (int64, error) {
	return m.service.GetBusinessOwnerID(ctx, businessID)
}

// PublicBusinessBySlug resolves the active business behind a storefront slug:
// the lookup behind /businesses/by-slug, without its locations. An unknown
// slug returns sql.ErrNoRows.
func (m *Module) PublicBusinessBySlug(ctx context.Context, slug string) (id int64, name, logoURL string, err error) {
	result, err := m.repository.LookupBySlug(ctx, strings.ToLower(strings.TrimSpace(slug)))
	if err != nil {
		return 0, "", "", err
	}
	return result.Business.ID, result.Business.Name, result.Business.LogoURL, nil
}

// PublicBusinessByID returns what an active business shows to the public. An
// unknown or inactive business returns sql.ErrNoRows.
func (m *Module) PublicBusinessByID(ctx context.Context, businessID int64) (name, logoURL string, err error) {
	business, err := m.repository.GetPublicBusiness(ctx, businessID)
	if err != nil {
		return "", "", err
	}
	return business.Name, business.LogoURL, nil
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	businessAPI := api.NewBusinessAPI(m.service)
	if m.userService != nil {
		businessAPI.WithUserService(m.userService)
	}
	if err := hubbusinessv1.RegisterBusinessServiceServer(
		businessAPI,
		hubbusinessv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub business service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/hub/business/v1/businesses/by-slug",
		"/hub/business/v1/businesses/public",
		"/hub/business/v1/businesses/name-exists",
	}
}

// ReportingPreferences is a current, uncached read. Unknown or inactive businesses return sql.ErrNoRows.
func (m *Module) ReportingPreferences(ctx context.Context, businessID int64) (currency, timezone string, err error) {
	return m.repository.ReportingPreferences(ctx, businessID)
}
func (m *Module) ReportingCurrency(ctx context.Context, businessID int64) (string, error) {
	currency, _, err := m.ReportingPreferences(ctx, businessID)
	return currency, err
}
func (m *Module) ReportingTimezone(ctx context.Context, businessID int64) (string, error) {
	_, timezone, err := m.ReportingPreferences(ctx, businessID)
	return timezone, err
}
