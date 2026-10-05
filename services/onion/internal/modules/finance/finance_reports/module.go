package financereports

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	moduleinfra "github.com/voxire/onion/internal/modules/finance/finance_reports/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/finance/finance_reports/internal/api"
	"github.com/voxire/onion/internal/modules/finance/finance_reports/internal/application"
	financestatementv1 "github.com/voxire/voxie/gen/go/finance/statement/v1"
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
	if err := financestatementv1.RegisterStatementServiceServer(
		api.NewStatementAPI(m.service),
		financestatementv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance statement service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) WithReportingCurrency(read func(context.Context, int64) (string, error)) *Module {
	m.repository.WithReportingCurrency(read)
	return m
}

func (m *Module) InsightsPaidInvoiceTotal(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return m.repository.InsightsPaidInvoiceTotal(ctx, businessID, from, to)
}

func (m *Module) InsightsExpenseTotal(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return m.repository.InsightsExpenseTotal(ctx, businessID, from, to)
}
