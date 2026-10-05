package config

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/config/internal/api"
	"github.com/voxire/onion/internal/modules/organization/config/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/config/internal/infrastructure"
	custom_domains "github.com/voxire/onion/internal/modules/organization/custom_domains"
	"github.com/voxire/onion/internal/platform/emailqueue"
	hubconfigv1 "github.com/voxire/voxie/gen/go/hub/config/v1"
)

type Module struct {
	ConfigService *application.Service
	CustomDomains *custom_domains.API
}

type emailQueueAdapter struct{ db *sql.DB }

func (a emailQueueAdapter) Enqueue(ctx context.Context, email application.TestEmail) error {
	return emailqueue.Enqueue(ctx, a.db, emailqueue.Job{
		Priority: emailqueue.PriorityHigh,
		To:       email.To,
		Subject:  email.Subject,
		HTML:     email.HTML,
		From:     email.From,
		ReplyTo:  email.ReplyTo,
	})
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	configService := application.NewService(moduleinfra.NewRepository(db), emailQueueAdapter{db: db})
	customDomainService := custom_domains.NewService(ctx, db)

	return &Module{
		ConfigService: configService,
		CustomDomains: custom_domains.NewAPI(customDomainService),
	}
}
func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubconfigv1.RegisterConfigServiceServer(
		api.NewConfigAPI(m.ConfigService, m.CustomDomains),
		hubconfigv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub config service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/hub/config/v1/industries",
	}
}
