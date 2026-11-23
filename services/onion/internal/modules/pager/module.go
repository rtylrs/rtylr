// Package pager is the order pager: a customer scans a venue's QR and gets a
// short code, the cashier labels that phone's ticket with the order number or
// the customer's name, and the phone rings when the staff mark it ready. A
// venue is an ordinary business.
package pager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/api"
	"github.com/voxire/onion/internal/modules/pager/internal/application"
	"github.com/voxire/onion/internal/modules/pager/internal/domain"
	moduleinfra "github.com/voxire/onion/internal/modules/pager/internal/infrastructure"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	"github.com/voxire/onion/internal/platform/logger"
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

func NewModule(ctx context.Context, db *sql.DB, businesses BusinessDirectory) *Module {
	service := application.NewService(moduleinfra.NewRepository(db), venues{businesses}, time.Now)
	configurePush(ctx, service)
	return &Module{service: service}
}

// defaultPushSubject is who the push services can contact about our pushes
// when WEB_PUSH_SUBJECT is unset.
const defaultPushSubject = "mailto:info@voxire.com"

// configurePush turns on lock-screen notifications when the VAPID key pair is
// set. Both unset leaves them off; a partial or unreadable pair is logged and
// also leaves them off rather than failing startup.
func configurePush(ctx context.Context, service *application.Service) {
	publicKey := strings.TrimSpace(os.Getenv("WEB_PUSH_VAPID_PUBLIC_KEY"))
	privateKey := strings.TrimSpace(os.Getenv("WEB_PUSH_VAPID_PRIVATE_KEY"))
	if publicKey == "" && privateKey == "" {
		return
	}
	subject := strings.TrimSpace(os.Getenv("WEB_PUSH_SUBJECT"))
	if subject == "" {
		subject = defaultPushSubject
	}
	sender, err := moduleinfra.NewWebPush(publicKey, privateKey, subject)
	if err != nil {
		logger.Ctx(ctx).Error("pager lock-screen notifications stay off: WEB_PUSH_VAPID_* is not a valid key pair", logger.Error(err))
		return
	}
	service.EnablePush(sender, nil)
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
		pagerv1.WithMiddleware(rateLimit, scanBudget()),
		pagerv1.WithErrorWriter(api.WriteError),
	); err != nil {
		panic(fmt.Errorf("register pager service: %w", err))
	}
}

// RegisterJobs schedules the sweep that closes tickets past their lifetime.
func (m *Module) RegisterJobs(_ context.Context, registry *platformjobs.Registry) {
	if registry == nil {
		return
	}
	_ = registry.Register(platformjobs.Job{
		Name:        "pager.sweep",
		Description: "Close pager tickets past their lifetime and drop their push subscriptions.",
		Schedule:    platformjobs.ScheduleFromEnv("PAGER_SWEEP_SCHEDULE", "*/15 * * * *"),
		Handler:     m.service.Sweep,
	})
}

// scanBudget limits each client's scans at a venue over an unlabelled ticket's
// lifetime, on top of the contract's per-minute budget.
func scanBudget() pagerv1.Middleware {
	limit := serverhttp.PerRouteRateLimit(domain.MaxScansPerClient, domain.UnlabelledLifetime)
	return func(next http.Handler) http.Handler {
		limited := limit(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if metadata, _ := pagerv1.RequestMetadataFromContext(r.Context()); metadata.Method == "scanTicket" {
				limited.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetAnonymousRoutes lists the customer routes. Everything else in pager.v1
// needs a signed-in member of the business.
func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/pager/v1/public/venues/{slug}",
		"/pager/v1/public/venues/{slug}/scan",
		"/pager/v1/public/tickets/{public_token}",
		"/pager/v1/public/tickets/{public_token}/acknowledge",
		"/pager/v1/public/push-key",
		"/pager/v1/public/tickets/{public_token}/push",
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
