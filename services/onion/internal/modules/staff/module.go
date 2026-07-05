package staff

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hrstaffv1 "github.com/voxire/voxie/gen/go/hr/staff/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/staff/internal/api"
	"github.com/voxire/onion/internal/modules/staff/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrstaffv1.RegisterStaffServiceServer(
		api.NewStaffAPI(m.service),
		hrstaffv1.WithMux(mux),
		hrstaffv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register HR staff service: %w", err))
	}
}

// LoadDashboardStats returns headline HR dashboard metrics for cross-module OneKit wiring.
func (m *Module) LoadDashboardStats(ctx context.Context, businessID int64) (totalStaff, onShiftToday, pendingTimeOff, payrollDraft int, err error) {
	stats, err := m.service.GetDashboardStats(ctx, businessID)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return stats.TotalStaff, stats.OnShiftToday, stats.PendingTimeOff, stats.PayrollDraft, nil
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
