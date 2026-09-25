package kpis

import (
	"context"
	"errors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const getBusinessKPIsToolName = "business.get_kpis"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	if name != getBusinessKPIsToolName {
		return nil, false, nil
	}
	businessID, err := kpisToolBusinessID(ctx)
	if err != nil {
		return nil, true, err
	}
	kpis, err := m.service.GetKPIs(ctx, businessID)
	if err != nil {
		return nil, true, err
	}
	return map[string]any{"kpis": kpis}, true, nil
}

func kpisToolBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, errors.New("authentication is required")
	}
	return businessID, nil
}
