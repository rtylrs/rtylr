package websitelead

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/websitelead/internal/api"
	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/websitelead/internal/infrastructure"
	serverhttp "github.com/voxire/onion/internal/platform/server/http"
	websiteleadv1 "github.com/voxire/voxie/gen/go/website/lead/v1"
)

// Notify re-exports the recipient configuration so cmd/api does not import the
// module's internal packages.
type Notify = application.Notify

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB, notify Notify) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db), notify),
	}
}

// RegisterAPI serves website.lead.v1. Each route's @meta("rate_limit") budget
// is enforced per client.
func (m *Module) RegisterAPI(mux *http.ServeMux) {
	rateLimit := serverhttp.RouteRateLimitMiddleware(func(r *http.Request) (string, string) {
		metadata, _ := websiteleadv1.RequestMetadataFromContext(r.Context())
		return metadata.Route, metadata.MetaValue(serverhttp.RouteRateLimitMetaKey)
	})
	if err := websiteleadv1.RegisterWebsiteLeadServiceServer(
		&api.WebsiteLeadAPI{Service: m.service},
		websiteleadv1.WithMux(mux),
		websiteleadv1.WithMiddleware(rateLimit),
	); err != nil {
		panic(fmt.Errorf("register website lead service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/quote", "/contact", "/newsletter", "/analytics/event"}
}
