package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/domain"
)

// dashboardRepo answers the dashboard's own rule from ruled and Flow-granted
// lookups from byID.
type dashboardRepo struct {
	mockRepo
	ruled     []domain.DashboardProject
	byID      map[int64]domain.DashboardProject
	askedByID []int64
}

func (r *dashboardRepo) ListDashboardProjects(_ context.Context, _, _ int64) ([]domain.DashboardProject, error) {
	return append([]domain.DashboardProject{}, r.ruled...), nil
}

func (r *dashboardRepo) GetDashboardProject(_ context.Context, _, _, projectID int64) (domain.DashboardProject, bool, error) {
	for _, p := range r.ruled {
		if p.ID == projectID {
			return p, true, nil
		}
	}
	return domain.DashboardProject{}, false, nil
}

func (r *dashboardRepo) GetDashboardProjectByID(_ context.Context, _, projectID int64) (domain.DashboardProject, bool, error) {
	p, ok := r.byID[projectID]
	return p, ok, nil
}

func (r *dashboardRepo) ListDashboardProjectsByIDs(_ context.Context, _ int64, ids []int64) ([]domain.DashboardProject, error) {
	r.askedByID = append(r.askedByID, ids...)
	out := []domain.DashboardProject{}
	for _, id := range ids {
		if p, ok := r.byID[id]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

type fakeAccess struct {
	allowed map[int64]bool
	listed  []int64
	err     error
	asked   int
}

func (a *fakeAccess) CanSeeProjectReport(_ context.Context, _, _, projectID int64) (bool, error) {
	a.asked++
	return a.allowed[projectID], a.err
}

func (a *fakeAccess) ListedProjectIDs(_ context.Context, _, _ int64) ([]int64, error) {
	return a.listed, a.err
}

var (
	publicSite   = domain.DashboardProject{ID: 1, Name: "Website", Status: "active", Visibility: "public"}
	privateAudit = domain.DashboardProject{ID: 2, Name: "audit", Status: "active", Visibility: "private"}
	reportsWork  = domain.DashboardProject{ID: 3, Name: "Brand refresh", Status: "active", Visibility: "private"}
)

func newDashboardService() (*Service, *dashboardRepo) {
	repo := &dashboardRepo{
		ruled: []domain.DashboardProject{publicSite},
		byID:  map[int64]domain.DashboardProject{publicSite.ID: publicSite, privateAudit.ID: privateAudit, reportsWork.ID: reportsWork},
	}
	return &Service{repo: repo}, repo
}

func TestGetDashboardProjectAsksFlowWhenItsOwnRuleMisses(t *testing.T) {
	ctx := context.Background()

	svc, _ := newDashboardService()
	if _, found, err := svc.GetDashboardProject(ctx, 9, 7, privateAudit.ID); err != nil || found {
		t.Fatalf("no access wired: found=%v err=%v, want not found", found, err)
	}

	access := &fakeAccess{allowed: map[int64]bool{privateAudit.ID: true}}
	svc.SetProjectAccess(access)
	got, found, err := svc.GetDashboardProject(ctx, 9, 7, privateAudit.ID)
	if err != nil || !found || got != privateAudit {
		t.Fatalf("Flow grants: got %+v found=%v err=%v, want %+v", got, found, err, privateAudit)
	}

	if _, found, err := svc.GetDashboardProject(ctx, 9, 7, reportsWork.ID); err != nil || found {
		t.Errorf("Flow refuses: found=%v err=%v, want not found", found, err)
	}

	access.asked = 0
	if _, found, err := svc.GetDashboardProject(ctx, 9, 7, publicSite.ID); err != nil || !found {
		t.Errorf("public project: found=%v err=%v", found, err)
	}
	if access.asked != 0 {
		t.Errorf("Flow was asked %d times about a project the dashboard's own rule covers", access.asked)
	}
}

func TestGetDashboardProjectPropagatesFlowAccessErrors(t *testing.T) {
	want := errors.New("hub down")
	svc, _ := newDashboardService()
	svc.SetProjectAccess(&fakeAccess{allowed: map[int64]bool{privateAudit.ID: true}, err: want})

	if _, found, err := svc.GetDashboardProject(context.Background(), 9, 7, privateAudit.ID); !errors.Is(err, want) || found {
		t.Errorf("found=%v err=%v, want the access error and not found", found, err)
	}
}

func TestListDashboardProjectsAddsWhatFlowListsInNameOrder(t *testing.T) {
	svc, repo := newDashboardService()
	// Flow's sidebar lists the public project again and a report's project; it
	// does not list privateAudit, so the picker must not either.
	svc.SetProjectAccess(&fakeAccess{listed: []int64{publicSite.ID, reportsWork.ID, reportsWork.ID}})

	got, err := svc.ListDashboardProjects(context.Background(), 9, 7)
	if err != nil {
		t.Fatalf("ListDashboardProjects: %v", err)
	}
	want := []domain.DashboardProject{reportsWork, publicSite}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("projects = %+v, want %+v", got, want)
	}
	if len(repo.askedByID) != 1 || repo.askedByID[0] != reportsWork.ID {
		t.Errorf("read by id %v, want only the project the own rule missed", repo.askedByID)
	}
}

func TestListDashboardProjectsWithoutFlowAccessUsesItsOwnRule(t *testing.T) {
	svc, repo := newDashboardService()
	got, err := svc.ListDashboardProjects(context.Background(), 9, 7)
	if err != nil || len(got) != 1 || got[0] != publicSite {
		t.Errorf("projects = %+v, %v; want only the public project", got, err)
	}
	if len(repo.askedByID) != 0 {
		t.Errorf("read by id %v with no access wired", repo.askedByID)
	}
}
