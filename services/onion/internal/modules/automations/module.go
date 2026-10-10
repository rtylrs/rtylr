package automations

import (
	"context"
	"database/sql"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"

	"github.com/voxire/onion/internal/modules/automations/internal/api"
	"github.com/voxire/onion/internal/modules/automations/internal/application"
	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/automations/internal/jobs"
	"github.com/voxire/onion/internal/platform/emailqueue"
	"github.com/voxire/onion/internal/platform/eventbus"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type Module struct {
	service  *application.Service
	executor *application.Executor
	fanout   *application.EventFanout
}

type CustomerContact struct {
	Email string
	Name  string
}

type CustomerContactResolver func(context.Context, int64, int64) (CustomerContact, bool, error)

func NewModule(ctx context.Context, db *sql.DB, contacts CustomerContactResolver) *Module {
	repo := infrastructure.NewRepository(db)
	queue := emailQueueAdapter{db: db}
	resolveContact := func(ctx context.Context, customerID, businessID int64) (domain.CustomerContact, bool, error) {
		contact, found, err := contacts(ctx, customerID, businessID)
		return domain.CustomerContact{Email: contact.Email, Name: contact.Name}, found, err
	}
	return &Module{
		service:  application.NewService(repo),
		executor: application.NewExecutor(repo, queue, eventbus.NewSQLReader(db), resolveContact),
		fanout:   application.NewEventFanout(repo),
	}
}

func (m *Module) DispatchEvent(ctx context.Context, event eventbus.Event) error {
	return m.fanout.DispatchEvent(ctx, event)
}

type emailQueueAdapter struct{ db *sql.DB }

func (a emailQueueAdapter) Enqueue(ctx context.Context, message application.EmailMessage) error {
	return emailqueue.Enqueue(ctx, a.db, emailqueue.Job{
		To:        message.To,
		Subject:   message.Subject,
		HTML:      message.HTML,
		Text:      message.Text,
		DedupeKey: message.DedupeKey,
	})
}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

// RegisterJobs is the hook platformjobs.AutoRegister looks for. Execution is a
// worker concern: the API process serves automation CRUD and never runs one.
func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	jobs.Register(ctx, registry, m.executor)
}
