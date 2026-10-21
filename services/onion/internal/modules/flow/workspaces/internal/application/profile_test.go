package application

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/platform/hubprofile"
)

func TestGetFlowProfileReturnsOnlyPublicProjects(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	svc.SetHubProfileResolver(fakeHubProfileResolver{profile: hubprofile.Profile{
		ID: v2UserA, FullName: "Abed Amouneh", Email: "abed@example.com",
	}})

	createV2Project(ctx, t, svc, "public-one", "public", v2UserA)
	createV2Project(ctx, t, svc, "private-one", "private", v2UserA)
	// Another member's public project must not appear on this profile.
	createV2Project(ctx, t, svc, "someone-elses", "public", v2UserB)

	profile, err := svc.GetFlowProfile(ctx, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("GetFlowProfile: %v", err)
	}

	if len(profile.PublicProjects) != 1 {
		names := make([]string, 0, len(profile.PublicProjects))
		for _, p := range profile.PublicProjects {
			names = append(names, p.Name)
		}
		t.Fatalf("public projects = %v, want exactly [public-one]", names)
	}
	if profile.PublicProjects[0].Name != "public-one" {
		t.Errorf("wrong project surfaced: %q", profile.PublicProjects[0].Name)
	}
	if profile.Name != "Abed Amouneh" || profile.Email != "abed@example.com" {
		t.Errorf("identity not carried through: %+v", profile)
	}
	if profile.RoleTitle != "" {
		t.Errorf("role title = %q, want empty until a job title is available", profile.RoleTitle)
	}
}

// An archived public project is in the recycle bin, not on a profile.
func TestGetFlowProfileExcludesArchivedProjects(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	svc.SetHubProfileResolver(fakeHubProfileResolver{profile: hubprofile.Profile{
		ID: v2UserA, FullName: "Abed Amouneh",
	}})

	p := createV2Project(ctx, t, svc, "public-then-binned", "public", v2UserA)
	if err := svc.DeleteProject(ctx, p.ID, v2Business); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	profile, err := svc.GetFlowProfile(ctx, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("GetFlowProfile: %v", err)
	}
	if len(profile.PublicProjects) != 0 {
		t.Errorf("archived project still on the profile: %+v", profile.PublicProjects)
	}
}

// Cross-business isolation is the resolver's job: a subject outside the
// caller's business is not found rather than returned with an empty project
// list, which would confirm the user id exists somewhere.
func TestGetFlowProfilePropagatesUnknownSubject(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	svc.SetHubProfileResolver(fakeHubProfileResolver{err: hubprofile.ErrUserNotFound})

	if _, err := svc.GetFlowProfile(ctx, v2UserB, v2Business); !errors.Is(err, hubprofile.ErrUserNotFound) {
		t.Fatalf("err = %v, want hubprofile.ErrUserNotFound", err)
	}
}

func TestInitialsOf(t *testing.T) {
	cases := map[string]string{
		"Abed Amouneh":           "AA",
		"abed amouneh":           "AA",
		"Abed El-Fattah Amouneh": "AA",
		"Abed":                   "AB",
		"A":                      "A",
		"":                       "",
		"   ":                    "",
	}
	for in, want := range cases {
		if got := initialsOf(in); got != want {
			t.Errorf("initialsOf(%q) = %q, want %q", in, got, want)
		}
	}
}
