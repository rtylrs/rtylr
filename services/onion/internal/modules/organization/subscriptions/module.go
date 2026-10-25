package subscriptions

import (
	"context"
	"database/sql"
	"fmt"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/organization/subscriptions/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/organization/subscriptions/internal/api"
	"github.com/voxire/onion/internal/modules/organization/subscriptions/internal/application"
	internaljobs "github.com/voxire/onion/internal/modules/organization/subscriptions/internal/jobs"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	hubsubscriptionv1 "github.com/voxire/voxie/gen/go/hub/subscription/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) GetAIQuotaEntitlement(ctx context.Context, businessID int64) (string, int, int, string, error) {
	return m.service.GetAIQuotaEntitlement(ctx, businessID)
}

func (m *Module) GetFreePlanAICredits(ctx context.Context) (int64, error) {
	return m.service.GetFreePlanAICredits(ctx)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubsubscriptionv1.RegisterSubscriptionServiceServer(
		api.NewSubscriptionAPI(m.service),
		hubsubscriptionv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub subscription service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	_ = ctx
	// Paddle signs the raw request body, so this webhook must own body reads.
	router.Post("/subscriptions/paddle-webhook", api.PaddleWebhookHTTPHandler(m.service))
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/hub/subscription/v1/plans",
		"/subscriptions/paddle-webhook",
	}
}

func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	internaljobs.Register(ctx, registry, m.service)
}
