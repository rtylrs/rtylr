package hr_equipment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/api"
	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/infrastructure"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
)

type hrCoreApprovalAdapter struct {
	core *hrcore.Module
}

func (a *hrCoreApprovalAdapter) EmployeeBelongsToBusiness(ctx context.Context, businessID, employeeID int64) (bool, error) {
	return a.core.EmployeeBelongsToBusiness(ctx, businessID, employeeID)
}

func (a *hrCoreApprovalAdapter) EmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error) {
	employeeID, err := a.core.GetEmployeeIDByUserID(ctx, businessID, userID)
	if errors.Is(err, hrcore.ErrEmployeeNotFound) {
		return 0, application.ErrEquipmentEmployeeNotFound
	}
	return employeeID, err
}

func (a *hrCoreApprovalAdapter) EmployeeHasRole(ctx context.Context, businessID, employeeID, roleID int64) (bool, error) {
	return a.core.HasRole(ctx, businessID, employeeID, roleID)
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
		service: application.NewService(infrastructure.NewRepository(db), adapter, adapter),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrequipmentv1.RegisterEquipmentServiceServer(
		api.NewEquipmentAPI(m.service),
		hrequipmentv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register HR equipment service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
