package hr_core

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr_core/internal/routes"
)

type ApprovalChainInfo struct {
	ID          int64
	RequestType string
	Name        string
	TotalLevels int32
	Levels      []ApprovalLevelInfo
}

type ApprovalLevelInfo struct {
	LevelOrder         int32
	ApproverType       string
	ApproverRoleID     int64
	ApproverEmployeeID int64
}

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/hr/employees/invite/verify",
		"/hr/employees/invite/accept",
	}
}

func (m *Module) GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error) {
	return m.service.GetEmployeeIDByUserID(ctx, ownerID, userID)
}

func (m *Module) HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error) {
	return m.service.HasPermission(ctx, ownerID, employeeID, permissionKey)
}

func (m *Module) GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainInfo, error) {
	chain, err := m.service.GetApprovalChain(ctx, ownerID, requestType)
	if err != nil {
		return nil, err
	}
	if chain == nil {
		return nil, nil
	}
	levels := make([]ApprovalLevelInfo, 0, len(chain.Levels))
	for _, lv := range chain.Levels {
		levels = append(levels, ApprovalLevelInfo{
			LevelOrder:         lv.LevelOrder,
			ApproverType:       lv.ApproverType,
			ApproverRoleID:     lv.ApproverRoleID,
			ApproverEmployeeID: lv.ApproverEmployeeID,
		})
	}
	return &ApprovalChainInfo{
		ID:          chain.ID,
		RequestType: chain.RequestType,
		Name:        chain.Name,
		TotalLevels: int32(len(chain.Levels)),
		Levels:      levels,
	}, nil
}
