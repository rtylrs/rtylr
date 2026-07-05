package hr_leave

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"

	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/api"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	internaljobs "github.com/voxire/onion/internal/modules/hr_leave/internal/jobs"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

type hrCoreApprovalAdapter struct {
	core *hrcore.Module
}

func (a *hrCoreApprovalAdapter) GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*domain.ApprovalChainInfo, error) {
	chain, err := a.core.GetApprovalChain(ctx, ownerID, requestType)
	if err != nil {
		return nil, err
	}
	if chain == nil {
		return nil, nil
	}
	levels := make([]domain.ApprovalLevelInfo, 0, len(chain.Levels))
	for _, lv := range chain.Levels {
		levels = append(levels, domain.ApprovalLevelInfo{
			LevelOrder:         lv.LevelOrder,
			ApproverType:       lv.ApproverType,
			ApproverRoleID:     lv.ApproverRoleID,
			ApproverEmployeeID: lv.ApproverEmployeeID,
		})
	}
	return &domain.ApprovalChainInfo{
		TotalLevels: chain.TotalLevels,
		Levels:      levels,
	}, nil
}

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB, hrCoreModule *hrcore.Module) *Module {
	adapter := &hrCoreApprovalAdapter{core: hrCoreModule}
	return &Module{
		service: application.NewService(ctx, db, adapter),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	serverHttp.RegisterGET[api.GetLeaveCalendarRequest, api.GetLeaveCalendarResponse](
		ctx, router, api.GetLeaveCalendarRoutes(), &api.GetLeaveCalendarHandler{Service: m.service},
	)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	err := hrleavev1.RegisterLeaveServiceServer(
		api.NewLeaveService(m.service),
		hrleavev1.WithMux(mux),
		hrleavev1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	)
	if err != nil {
		panic(err)
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	internaljobs.Register(ctx, registry, m.service)
}

func (m *Module) RunMonthlyAccrual(ctx context.Context) error {
	return m.service.RunMonthlyAccrual(ctx)
}
