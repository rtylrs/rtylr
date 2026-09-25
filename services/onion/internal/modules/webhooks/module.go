package webhooks

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/webhooks/internal/application"
	"github.com/voxire/onion/internal/modules/webhooks/internal/jobs"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type Module struct {
	service  *application.Service
	sender   *application.Sender
	notifier *application.Notifier
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service:  application.NewService(ctx, db),
		sender:   application.NewSender(db),
		notifier: application.NewNotifier(db),
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) GetService() *application.Service {
	return m.service
}

// RegisterJobs is the hook platformjobs.AutoRegister looks for. Delivery is a
// worker concern: the API process serves registration CRUD and never sends.
func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	jobs.Register(ctx, registry, m.sender)
	jobs.RegisterNotifier(ctx, registry, m.notifier)
}
