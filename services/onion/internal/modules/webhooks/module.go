package webhooks

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/webhooks/internal/application"
	"github.com/voxire/onion/internal/modules/webhooks/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/webhooks/internal/jobs"
	"github.com/voxire/onion/internal/platform/emailqueue"
	"github.com/voxire/onion/internal/platform/eventbus"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type Module struct {
	service  *application.Service
	sender   *application.Sender
	notifier *application.Notifier
	fanout   *application.EventFanout
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := infrastructure.NewRepository(db)
	return &Module{
		service:  application.NewService(repo, eventbus.NewSQLReader(db)),
		sender:   application.NewSender(repo, eventbus.NewSQLReader(db)),
		notifier: application.NewNotifier(repo, emailQueueAdapter{db: db}, nil),
		fanout:   application.NewEventFanout(repo),
	}
}

func (m *Module) SetNotificationContactResolver(resolver application.NotificationContactResolver) {
	m.notifier.SetContactResolver(resolver)
}

func (m *Module) DispatchEvent(ctx context.Context, event eventbus.Event) error {
	return m.fanout.DispatchEvent(ctx, event)
}

type emailQueueAdapter struct{ db *sql.DB }

func (a emailQueueAdapter) Enqueue(ctx context.Context, message application.NotificationEmail) error {
	return emailqueue.Enqueue(ctx, a.db, emailqueue.Job{
		To:        message.To,
		Subject:   message.Subject,
		HTML:      message.HTML,
		Text:      message.Text,
		DedupeKey: message.DedupeKey,
	})
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
