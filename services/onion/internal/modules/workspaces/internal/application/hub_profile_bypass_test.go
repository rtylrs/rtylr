package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/hubprofile"
)

type fakeHubProfileResolver struct {
	profile hubprofile.Profile
	err     error
}

func (f fakeHubProfileResolver) GetProfile(context.Context, int64, int64) (hubprofile.Profile, error) {
	return f.profile, f.err
}

const (
	accessUserID     int64 = 77001
	accessBusinessID int64 = 77002
)

func TestBusinessAdministratorsBypassFlowResourceAssignments(t *testing.T) {
	for _, test := range []struct {
		name   string
		access authorization.EffectiveAccess
	}{
		{name: "owner", access: authorization.EffectiveAccess{IsOwner: true}},
		{name: "admin", access: authorization.EffectiveAccess{IsAdmin: true}},
	} {
		t.Run(test.name, func(t *testing.T) {
			svc := &Service{hubProfiles: fakeHubProfileResolver{profile: hubprofile.Profile{EffectiveAccess: &test.access}}}
			ctx := context.Background()

			if err := svc.RequireProjectRole(ctx, 41, accessUserID, accessBusinessID, domain.FlowRoleAdmin); err != nil {
				t.Errorf("RequireProjectRole() = %v, want owner/admin access", err)
			}
			if err := svc.RequireWorkspaceRole(ctx, 42, accessUserID, accessBusinessID, domain.FlowRoleAdmin); err != nil {
				t.Errorf("RequireWorkspaceRole() = %v, want owner/admin access", err)
			}
			if err := svc.RequireTaskRole(ctx, 43, accessUserID, accessBusinessID, domain.FlowRoleAdmin); err != nil {
				t.Errorf("RequireTaskRole() = %v, want owner/admin access", err)
			}
		})
	}
}

func TestFlowResourceAccessStillRequiresMemberAssignment(t *testing.T) {
	memberAccess := &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
		AppID: "flow", Licensed: true, HasAccess: true, Permissions: []string{"flow:team:manage"},
	}}}
	svc := &Service{hubProfiles: fakeHubProfileResolver{profile: hubprofile.Profile{EffectiveAccess: memberAccess}}}

	if err := svc.RequireProjectRole(context.Background(), 0, accessUserID, accessBusinessID, domain.FlowRoleViewer); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("RequireProjectRole() = %v, want resource-membership denial", err)
	}
	if err := svc.RequireWorkspaceRole(context.Background(), 0, accessUserID, accessBusinessID, domain.FlowRoleViewer); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("RequireWorkspaceRole() = %v, want resource-membership denial", err)
	}
	if err := svc.RequireTaskRole(context.Background(), 0, accessUserID, accessBusinessID, domain.FlowRoleViewer); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("RequireTaskRole() = %v, want invalid-resource denial", err)
	}
}

func TestBusinessAdministratorResolutionFailsClosed(t *testing.T) {
	for _, resolver := range []hubprofile.Resolver{
		fakeHubProfileResolver{err: errors.New("hub lookup unavailable")},
		fakeHubProfileResolver{profile: hubprofile.Profile{}},
		nil,
	} {
		svc := &Service{hubProfiles: resolver}
		if err := svc.RequireProjectRole(context.Background(), 0, accessUserID, accessBusinessID, domain.FlowRoleViewer); !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("RequireProjectRole() with unresolved administrator = %v, want denial", err)
		}
	}
}
