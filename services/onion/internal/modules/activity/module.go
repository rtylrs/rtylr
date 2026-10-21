package activity

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/eventbus"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/activity/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/activity/internal/api"
	"github.com/voxire/onion/internal/modules/activity/internal/application"
	hubactivityv1 "github.com/voxire/voxie/gen/go/hub/activity/v1"
)

type Module struct {
	service    *application.Service
	repository *moduleinfra.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := moduleinfra.NewRepository(db)
	return &Module{service: application.NewService(ctx, repo), repository: repo}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubactivityv1.RegisterActivityServiceServer(
		api.NewActivityAPI(m.service),
		hubactivityv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub activity service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) WithUserNameReader(reader func(context.Context, int64, int64) (string, error)) *Module {
	m.repository.WithUserNameReader(reader)
	return m
}

func (m *Module) DispatchEvent(ctx context.Context, fact eventbus.Event) error {
	if err := m.repository.RecordFact(ctx, fact); err != nil {
		return err
	}
	cacher.InvalidateScope(ctx, fmt.Sprintf("activity:%d:", fact.BusinessID))
	return nil
}
