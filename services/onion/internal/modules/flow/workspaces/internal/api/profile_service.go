package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/hubprofile"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowprofilev1 "github.com/voxire/voxie/gen/go/flow/profile/v1"
)

// ProfileAPI serves flow.profile.v1.
//
// The contract and its generated clients have existed since voxie #44 and the
// frontend has shipped a profile page against them, but nothing ever served the
// route, so every profile answered 404. This is that handler.
type ProfileAPI struct {
	Service *application.Service
}

func NewProfileAPI(service *application.Service) *ProfileAPI {
	return &ProfileAPI{Service: service}
}

func (a *ProfileAPI) GetProfile(
	ctx context.Context,
	req *flowprofilev1.GetProfileRequest,
) (*flowprofilev1.GetProfileResponse, error) {
	businessID, _, err := resolveBusinessIDV2(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetUserId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "user_id is required")
	}

	profile, err := a.Service.GetFlowProfile(ctx, req.GetUserId(), businessID)
	if err != nil {
		// A subject outside the caller's business is indistinguishable from one
		// who does not exist, which is the answer we want either way: it does
		// not confirm that some user id belongs to another business.
		if errors.Is(err, hubprofile.ErrUserNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Profile not found")
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load profile")
	}

	projects := make([]*flowprofilev1.ProfileProject, 0, len(profile.PublicProjects))
	for _, p := range profile.PublicProjects {
		projects = append(projects, &flowprofilev1.ProfileProject{
			Id:             p.ID,
			Name:           p.Name,
			Icon:           p.Icon,
			Color:          p.Color,
			Description:    p.Description,
			Status:         p.Status,
			TaskCount:      p.TaskCount,
			CompletedCount: p.CompletedCount,
		})
	}

	return &flowprofilev1.GetProfileResponse{
		Profile: &flowprofilev1.Profile{
			UserId:         profile.UserID,
			Name:           profile.Name,
			Email:          profile.Email,
			Initials:       profile.Initials,
			RoleTitle:      profile.RoleTitle,
			Bio:            profile.Bio,
			PublicProjects: projects,
		},
	}, nil
}
