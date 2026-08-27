package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/hubprofile"
)

// fakeHubProfileResolver is a hand-rolled stand-in for hubprofile.Resolver.
// The repo has no mocking library in use anywhere, so every other test in
// this package that needs a fake dependency (see authorization_test.go
// equivalents in sibling modules) hand-writes one; this follows the same
// convention.
type fakeHubProfileResolver struct {
	profile hubprofile.Profile
	err     error
}

func (f fakeHubProfileResolver) GetProfile(context.Context, int64, int64) (hubprofile.Profile, error) {
	return f.profile, f.err
}

const (
	bypassUserID     int64 = 77001
	bypassBusinessID int64 = 77002
)

// Deliberately zero/no-DB entity ids below (project/task/workspace = 0). With
// no hub bypass, projectRoleFor/workspaceRoleFor short-circuit to "" for a
// zero id without ever touching the repository (see their guard clauses), and
// RequireTaskRole denies outright on a zero taskID before touching the
// repository either. That lets these tests exercise the bypass and its
// fallthrough without a live database: repo stays nil throughout and is only
// ever reached if a test's expectations are wrong.

func TestRequireProjectRole_HubBypass(t *testing.T) {
	tests := []struct {
		name       string
		resolver   hubprofile.Resolver
		wantBypass bool
	}{
		{
			name:       "owner does not bypass project scope",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{IsOwner: true}},
			wantBypass: false,
		},
		{
			name:       "admin does not bypass project scope",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{Role: "admin"}},
			wantBypass: false,
		},
		{
			name:       "operator does not bypass project scope",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{Role: "operator"}},
			wantBypass: false,
		},
		{
			name:       "staff role does not bypass, falls through",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{Role: "staff"}},
			wantBypass: false,
		},
		{
			name:       "viewer role does not bypass, falls through",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{Role: "viewer"}},
			wantBypass: false,
		},
		{
			name:       "resolver error does not bypass, falls through",
			resolver:   fakeHubProfileResolver{err: errors.New("hub lookup unavailable")},
			wantBypass: false,
		},
		{
			name:       "no resolver wired does not bypass, falls through",
			resolver:   nil,
			wantBypass: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &Service{hubProfiles: tc.resolver}

			err := svc.RequireProjectRole(context.Background(), 0, bypassUserID, bypassBusinessID, domain.FlowRoleAdmin)
			if tc.wantBypass && err != nil {
				t.Fatalf("RequireProjectRole with bypass expected to grant access, got err = %v", err)
			}
			if !tc.wantBypass && !errors.Is(err, domain.ErrUnauthorized) {
				t.Fatalf("RequireProjectRole fallthrough with no project role: err = %v, want ErrUnauthorized", err)
			}
		})
	}
}

func TestRequireWorkspaceRole_HubBypass(t *testing.T) {
	tests := []struct {
		name       string
		resolver   hubprofile.Resolver
		wantBypass bool
	}{
		{
			name:       "owner does not bypass workspace scope",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{IsOwner: true}},
			wantBypass: false,
		},
		{
			name:       "admin does not bypass workspace scope",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{Role: "admin"}},
			wantBypass: false,
		},
		{
			name:       "operator does not bypass workspace scope",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{Role: "operator"}},
			wantBypass: false,
		},
		{
			name:       "staff role does not bypass, falls through",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{Role: "staff"}},
			wantBypass: false,
		},
		{
			name:       "resolver error does not bypass, falls through",
			resolver:   fakeHubProfileResolver{err: errors.New("hub lookup unavailable")},
			wantBypass: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &Service{hubProfiles: tc.resolver}

			err := svc.RequireWorkspaceRole(context.Background(), 0, bypassUserID, bypassBusinessID, domain.FlowRoleAdmin)
			if tc.wantBypass && err != nil {
				t.Fatalf("RequireWorkspaceRole with bypass expected to grant access, got err = %v", err)
			}
			if !tc.wantBypass && !errors.Is(err, domain.ErrUnauthorized) {
				t.Fatalf("RequireWorkspaceRole fallthrough with no workspace role: err = %v, want ErrUnauthorized", err)
			}
		})
	}
}

func TestRequireTaskRole_HubBypass(t *testing.T) {
	tests := []struct {
		name       string
		resolver   hubprofile.Resolver
		wantBypass bool
	}{
		{
			name:       "owner does not bypass task scope",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{IsOwner: true}},
			wantBypass: false,
		},
		{
			name:       "admin does not bypass task scope",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{Role: "admin"}},
			wantBypass: false,
		},
		{
			name:       "operator does not bypass task scope",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{Role: "operator"}},
			wantBypass: false,
		},
		{
			name:       "viewer role does not bypass, falls through",
			resolver:   fakeHubProfileResolver{profile: hubprofile.Profile{Role: "viewer"}},
			wantBypass: false,
		},
		{
			name:       "resolver error does not bypass, falls through (and does not panic)",
			resolver:   fakeHubProfileResolver{err: errors.New("hub lookup unavailable")},
			wantBypass: false,
		},
		{
			name:       "no resolver wired does not bypass, falls through (and does not panic)",
			resolver:   nil,
			wantBypass: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &Service{hubProfiles: tc.resolver}

			// taskID 0 makes RequireTaskRole's own pre-existing guard deny
			// immediately (see its "taskID == 0" check) without ever touching
			// the repository - repo is nil on this Service, so any query
			// through it would panic, which is itself part of what these
			// fallthrough cases are checking: a missing/erroring resolver
			// must fall through to that existing logic, not panic.
			err := svc.RequireTaskRole(context.Background(), 0, bypassUserID, bypassBusinessID, domain.FlowRoleAdmin)
			if tc.wantBypass && err != nil {
				t.Fatalf("RequireTaskRole with bypass expected to grant access, got err = %v", err)
			}
			if !tc.wantBypass && !errors.Is(err, domain.ErrUnauthorized) {
				t.Fatalf("RequireTaskRole fallthrough with no task: err = %v, want ErrUnauthorized", err)
			}
		})
	}
}
