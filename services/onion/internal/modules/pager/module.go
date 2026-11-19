// Package pager is the order pager: a customer scans a venue's poster, enters
// the number on their receipt and their phone rings when the staff mark that
// number ready. A venue is an ordinary business.
package pager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/api"
	"github.com/voxire/onion/internal/modules/pager/internal/application"
	"github.com/voxire/onion/internal/modules/pager/internal/domain"
	moduleinfra "github.com/voxire/onion/internal/modules/pager/internal/infrastructure"
	serverhttp "github.com/voxire/onion/internal/platform/server/http"
	pagerv1 "github.com/voxire/voxie/gen/go/pager/v1"
)

// BusinessDirectory is what the pager reads from the business module: the
// slug lookup the public menu uses, and the public profile by id. Both return
// sql.ErrNoRows for an unknown or inactive business.
type BusinessDirectory interface {
	PublicBusinessBySlug(ctx context.Context, slug string) (id int64, name, logoURL string, err error)
	PublicBusinessByID(ctx context.Context, businessID int64) (name, logoURL string, err error)
}

type Module struct {
	service *application.Service
}

func NewModule(_ context.Context, db *sql.DB, businesses BusinessDirectory) *Module {
	return &Module{
		service: application.NewService(moduleinfra.NewRepository(db), venues{businesses}, time.Now),
	}
}

// RegisterAPI serves pager.v1. The public routes carry an @meta("rate_limit")
// budget in the contract, enforced per client and concrete path.
func (m *Module) RegisterAPI(mux *http.ServeMux) {
	rateLimit := serverhttp.RouteRateLimitMiddleware(func(r *http.Request) (string, string) {
		metadata, _ := pagerv1.RequestMetadataFromContext(r.Context())
		return metadata.Route, metadata.MetaValue(serverhttp.RouteRateLimitMetaKey)
	})
	if err := pagerv1.RegisterPagerServiceServer(
		&api.PagerAPI{Service: m.service},
		pagerv1.WithMux(mux),
		pagerv1.WithMiddleware(rateLimit),
		pagerv1.WithErrorWriter(api.WriteError),
	); err != nil {
		panic(fmt.Errorf("register pager service: %w", err))
	}
}

// GetAnonymousRoutes lists the customer routes. Everything else in pager.v1
// needs a signed-in member of the business.
func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/pager/v1/public/venues/{slug}",
		"/pager/v1/public/venues/{slug}/join",
		"/pager/v1/public/tickets/{public_token}",
		"/pager/v1/public/tickets/{public_token}/acknowledge",
	}
}

// venues adapts the business directory to the pager's own venue type.
type venues struct {
	businesses BusinessDirectory
}

func (v venues) VenueBySlug(ctx context.Context, slug string) (domain.Venue, error) {
	id, name, logoURL, err := v.businesses.PublicBusinessBySlug(ctx, slug)
	if err != nil {
		return domain.Venue{}, venueError(err)
	}
	return domain.Venue{BusinessID: id, Name: name, LogoURL: logoURL}, nil
}

func (v venues) VenueByID(ctx context.Context, businessID int64) (domain.Venue, error) {
	name, logoURL, err := v.businesses.PublicBusinessByID(ctx, businessID)
	if err != nil {
		return domain.Venue{}, venueError(err)
	}
	return domain.Venue{BusinessID: businessID, Name: name, LogoURL: logoURL}, nil
}

func venueError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrVenueNotFound
	}
	return err
}
