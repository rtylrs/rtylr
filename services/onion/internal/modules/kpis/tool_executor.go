package kpis

import (
	"context"
	"errors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const getBusinessKPIsToolName = "business.get_kpis"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	if name != getBusinessKPIsToolName {
		return nil, false, nil
	}
	ownerID, err := kpisToolOwnerID(ctx)
	if err != nil {
		return nil, true, err
	}
	kpis, err := m.service.GetKPIs(ctx, ownerID)
	if err != nil {
		return nil, true, err
	}
	return map[string]any{"kpis": kpis}, true, nil
}

func kpisToolOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	return authUser.UserID, nil
}
