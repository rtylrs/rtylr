package flow_insights

import (
	"context"
	"errors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	getFlowInsightsToolName     = "flow.get_insights"
	getFlowTeamWorkloadToolName = "flow.get_team_workload"
	getFlowRoadmapToolName      = "flow.get_roadmap"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	ownerID, err := flowInsightsToolOwnerID(ctx)
	if err != nil {
		switch name {
		case getFlowInsightsToolName, getFlowTeamWorkloadToolName, getFlowRoadmapToolName:
			return nil, true, err
		default:
			return nil, false, nil
		}
	}

	switch name {
	case getFlowInsightsToolName:
		stats, err := m.service.GetInsights(ctx, ownerID)
		if err != nil {
			return nil, true, err
		}
		return map[string]any{"insights": stats}, true, nil
	case getFlowTeamWorkloadToolName:
		workload, err := m.service.GetTeamWorkload(ctx, ownerID)
		if err != nil {
			return nil, true, err
		}
		return map[string]any{
			"members":  workload.Members,
			"returned": len(workload.Members),
		}, true, nil
	case getFlowRoadmapToolName:
		roadmap, err := m.service.GetRoadmap(ctx, ownerID)
		if err != nil {
			return nil, true, err
		}
		return map[string]any{
			"projects": roadmap.Projects,
			"returned": len(roadmap.Projects),
		}, true, nil
	default:
		return nil, false, nil
	}
}

func flowInsightsToolOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	return authUser.UserID, nil
}
