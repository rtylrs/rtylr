package team

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/hubprofile"
)

type hubProfileResolver struct {
	service *application.Service
}

func (r *hubProfileResolver) GetProfile(ctx context.Context, userID, businessID int64) (hubprofile.Profile, error) {
	profile, err := r.service.GetMyProfile(ctx, userID, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return hubprofile.Profile{}, hubprofile.ErrUserNotFound
		}
		return hubprofile.Profile{}, err
	}

	return hubprofile.Profile{
		ID:                 profile.ID,
		Email:              profile.Email,
		FullName:           profile.FullName,
		Apps:               profile.Apps,
		Role:               profile.Role,
		BusinessID:         profile.BusinessID,
		BusinessName:       profile.BusinessName,
		IsOwner:            profile.IsOwner,
		IsConsoleAdmin:     profile.IsConsoleAdmin,
		Onboarded:          profile.Onboarded,
		EffectiveAccess:    profile.EffectiveAccess,
		ProvisioningStatus: profile.ProvisioningStatus,
	}, nil
}

func NewHubProfileResolver(service *application.Service) hubprofile.Resolver {
	return &hubProfileResolver{service: service}
}
