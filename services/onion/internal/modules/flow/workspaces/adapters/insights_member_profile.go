package workspaceadapters

import (
	"context"
	"errors"

	flowworkspaces "github.com/voxire/onion/internal/modules/flow/workspaces"
	"github.com/voxire/onion/internal/modules/organization/user"
)

type flowInsightsMemberProfileAdapter struct {
	module *user.Module
}

func (a flowInsightsMemberProfileAdapter) GetInsightsMemberProfiles(ctx context.Context, userIDs []int64) (map[int64]flowworkspaces.InsightsMemberProfile, error) {
	profiles := make(map[int64]flowworkspaces.InsightsMemberProfile, len(userIDs))
	for _, userID := range userIDs {
		profile, err := a.module.GetUserProfile(ctx, userID)
		if errors.Is(err, user.ErrUserNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		profiles[userID] = flowworkspaces.InsightsMemberProfile{UserID: profile.ID, Name: profile.Name}
	}
	return profiles, nil
}

func NewInsightsMemberProfileSource(module *user.Module) flowworkspaces.InsightsMemberProfileSource {
	return flowInsightsMemberProfileAdapter{module: module}
}
