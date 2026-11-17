package infrastructure_test

import (
	"testing"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/infrastructure"
)

// The by-id readers carry no viewer rule, because Flow's access check has
// already decided. They must still never return a personal My Tasks project or
// another business's project.
func TestDashboardProjectByIDReaders(t *testing.T) {
	f := newProjectDashboardFixture(t)
	repo := infrastructure.NewRepository(f.db)

	private := f.project(dashBiz, "B private audit", "private", "active", false)
	public := f.project(dashBiz, "A public site", "public", "active", false)
	archived := f.project(dashBiz, "C archived private", "private", "archived", false)
	personal := f.project(dashBiz, "D My Tasks", "private", "active", true)
	foreign := f.project(dashOther, "E other business", "private", "active", false)

	lookups := []struct {
		name      string
		projectID int64
		found     bool
	}{
		{"a private project opens with no member row", private, true},
		{"an archived project still opens by id", archived, true},
		{"a personal My Tasks project never opens", personal, false},
		{"another business's project never opens", foreign, false},
		{"a missing project never opens", foreign + 1000, false},
	}
	for _, tc := range lookups {
		t.Run(tc.name, func(t *testing.T) {
			got, found, err := repo.GetDashboardProjectByID(f.ctx, dashBiz, tc.projectID)
			if err != nil {
				t.Fatalf("GetDashboardProjectByID: %v", err)
			}
			if found != tc.found {
				t.Fatalf("found = %v, want %v (project %+v)", found, tc.found, got)
			}
			if found && got.ID != tc.projectID {
				t.Errorf("ID = %d, want %d", got.ID, tc.projectID)
			}
		})
	}

	listed, err := repo.ListDashboardProjectsByIDs(f.ctx, dashBiz, []int64{private, public, archived, personal, foreign})
	if err != nil {
		t.Fatalf("ListDashboardProjectsByIDs: %v", err)
	}
	if want := []int64{public, private}; !sameIDs(projectIDs(listed), want) {
		t.Errorf("listed = %+v, want ids %v in name order", listed, want)
	}

	empty, err := repo.ListDashboardProjectsByIDs(f.ctx, dashBiz, nil)
	if err != nil || empty == nil || len(empty) != 0 {
		t.Errorf("no ids = %+v, %v; want an empty list", empty, err)
	}
}
