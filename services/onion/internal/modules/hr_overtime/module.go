package hr_overtime

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"

	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/api"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	internaljobs "github.com/voxire/onion/internal/modules/hr_overtime/internal/jobs"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	hrovertimev1 "github.com/voxire/voxie/gen/go/hr/overtime/v1"
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
	serverHttp.RegisterGET[api.GetOvertimeConfigRequest, api.GetOvertimeConfigResponse](
		ctx, router, api.GetOvertimeConfigRoutes(), &api.GetOvertimeConfigHandler{Service: m.service},
	)
	serverHttp.RegisterPUTWithValidation[api.UpsertOvertimeConfigRequest, api.UpsertOvertimeConfigResponse](
		ctx, router, api.UpsertOvertimeConfigRoutes(), &api.UpsertOvertimeConfigHandler{Service: m.service},
	)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	err := hrovertimev1.RegisterOvertimeServiceServer(
		api.NewOvertimeService(m.service),
		hrovertimev1.WithMux(mux),
		hrovertimev1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
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

func (m *Module) RunDailyDetection(ctx context.Context) error {
	return m.service.RunDailyDetection(ctx)
}
