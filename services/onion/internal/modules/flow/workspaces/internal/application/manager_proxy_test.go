package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// stubOrgchart stands in for HR. Recording the calls matters because the proxy
// must not be consulted at all for a caller who already has access: it is the
// most expensive branch and it runs last.
type stubOrgchart struct {
	reports []int64
	err     error
	calls   int
}

func (s *stubOrgchart) ManagedUserIDs(_ context.Context, _, _ int64) ([]int64, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.reports, nil
}

func TestManagedUserIDsIsInertWithoutAResolver(t *testing.T) {
	svc := &Service{}

	if got := svc.managedUserIDs(context.Background(), 7, 42); got != nil {
		t.Fatalf("no resolver wired: want nil, got %v", got)
	}
}

// A resolver failure must not deny access the caller already has, and must not
// surface as an error on a list they are entitled to see. It widens visibility,
// so the only safe failure is to widen nothing.
func TestManagedUserIDsSwallowsResolverErrors(t *testing.T) {
	stub := &stubOrgchart{err: errors.New("hr unavailable")}
	svc := &Service{orgchart: stub}

	if got := svc.managedUserIDs(context.Background(), 7, 42); got != nil {
		t.Fatalf("resolver error: want nil, got %v", got)
	}
	if stub.calls != 1 {
		t.Fatalf("resolver should have been consulted once, got %d", stub.calls)
	}
}

func TestManagedUserIDsSkipsUnidentifiedCallerOrBusiness(t *testing.T) {
	stub := &stubOrgchart{reports: []int64{9}}
	svc := &Service{orgchart: stub}

	for _, tc := range []struct{ caller, business int64 }{{0, 42}, {7, 0}, {0, 0}} {
		if got := svc.managedUserIDs(context.Background(), tc.caller, tc.business); got != nil {
			t.Fatalf("caller=%d business=%d: want nil, got %v", tc.caller, tc.business, got)
		}
	}
	if stub.calls != 0 {
		t.Fatalf("resolver must not be consulted without both ids, got %d calls", stub.calls)
	}
}

// Project.is_personal's contract comment says the My Tasks bucket is "exempt
// from manager-visibility policies unless its owner explicitly shares it", so
// the exemption is a contract requirement rather than a preference. The proxy
// has to stop before consulting HR at all.
func TestManagesSomeoneOnExemptsThePersonalBucket(t *testing.T) {
	stub := &stubOrgchart{reports: []int64{9}}
	svc := &Service{orgchart: stub}

	personal := domain.Project{ID: 1, CreatedBy: 9, OwnerID: 9, IsPersonal: true}
	if svc.managesSomeoneOn(context.Background(), personal, 7, 42) {
		t.Fatal("a personal bucket must stay invisible to the owner's manager")
	}
	if stub.calls != 0 {
		t.Fatalf("the exemption should short-circuit before HR is asked, got %d calls", stub.calls)
	}
}

func TestManagesSomeoneOnMatchesAReportsOwnedProject(t *testing.T) {
	svc := &Service{orgchart: &stubOrgchart{reports: []int64{9, 11}}}

	byCreator := domain.Project{ID: 1, CreatedBy: 11, OwnerID: 99}
	if !svc.managesSomeoneOn(context.Background(), byCreator, 7, 42) {
		t.Fatal("a project created by a report should be visible to their manager")
	}

	byOwner := domain.Project{ID: 2, CreatedBy: 99, OwnerID: 9}
	if !svc.managesSomeoneOn(context.Background(), byOwner, 7, 42) {
		t.Fatal("a project owned by a report should be visible to their manager")
	}
}

// Managing nobody is the common case, and the outcome must be indistinguishable
// from having no resolver at all.
func TestManagesSomeoneOnDeniesWhenTheCallerManagesNobody(t *testing.T) {
	svc := &Service{orgchart: &stubOrgchart{reports: nil}}

	p := domain.Project{ID: 1, CreatedBy: 99, OwnerID: 99}
	if svc.managesSomeoneOn(context.Background(), p, 7, 42) {
		t.Fatal("a caller who manages nobody gains nothing from the proxy")
	}
}

// Somebody else's project, and the caller manages people, but not this one's.
func TestManagesSomeoneOnDeniesAnUnrelatedProject(t *testing.T) {
	svc := &Service{orgchart: &stubOrgchart{reports: []int64{9, 11}}}

	p := domain.Project{ID: 1, CreatedBy: 500, OwnerID: 501}
	if svc.managesSomeoneOn(context.Background(), p, 7, 42) {
		t.Fatal("the proxy must not reach a project no report is involved in")
	}
}

func TestProjectRosterUsesProjectVisibility(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	managedProject := createV2Project(ctx, t, svc, "manager-visible", "", v2UserB)
	unrelatedProject := createV2Project(ctx, t, svc, "manager-hidden", "", v2UserB+1)
	svc.orgchart = &stubOrgchart{reports: []int64{v2UserB}}

	if _, err := svc.GetProjectV2(ctx, managedProject.ID, v2UserA, v2Business); err != nil {
		t.Fatalf("manager should be able to open a report's project: %v", err)
	}
	if !svc.IsProjectMemberOrBypass(ctx, v2UserA, v2Business, managedProject.ID) {
		t.Fatal("manager who can open the project should be able to load its member roster")
	}
	if svc.IsProjectMemberOrBypass(ctx, v2UserA, v2Business, unrelatedProject.ID) {
		t.Fatal("manager must not load an unrelated project's member roster")
	}
}

// No unit tests for markVisibleViaReport, deliberately.
//
// Service.repo is a concrete *infrastructure.Repository, so the membership
// lookup it depends on cannot be stubbed. With a nil repo every path ends in
// "not flagged", which makes any assertion here pass even with the logic
// deleted: I wrote three such tests, mutation-probed them, watched all three
// survive the logic being removed, and deleted them. A test that passes when the
// feature is gone is worse than no test, because it reads as coverage.
//
// The flagging path is verified against the database in the PR that introduced
// it. Making it unit-testable means extracting a membership seam the way
// orgchart.ManagerResolver was extracted, which is worth doing when something
// else needs it too.
