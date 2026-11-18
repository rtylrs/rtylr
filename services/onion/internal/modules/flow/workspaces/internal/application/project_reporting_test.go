package application

import (
	"testing"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/hubprofile"
)

func reportingResolver(access authorization.EffectiveAccess) fakeHubProfileResolver {
	return fakeHubProfileResolver{profile: hubprofile.Profile{EffectiveAccess: &access}}
}

// A business admin may see reporting on a private project they never joined,
// while a plain member of the business may not. Both are refused the same way
// as a project that does not exist.
func TestCanSeeProjectReportAdmitsBypassRolesOnly(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	private := createV2Project(ctx, t, svc, "report-private", "private", v2UserB)

	cases := []struct {
		name   string
		access authorization.EffectiveAccess
		want   bool
	}{
		{"business owner", authorization.EffectiveAccess{IsOwner: true}, true},
		{"business admin", authorization.EffectiveAccess{IsAdmin: true}, true},
		{"work team manager", authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID: "flow", Licensed: true, HasAccess: true, Permissions: []string{"flow:team:manage"},
		}}}, true},
		{"plain member of the business", authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID: "flow", Licensed: true, HasAccess: true, Permissions: []string{"flow:projects:view"},
		}}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc.hubProfiles = reportingResolver(tc.access)
			got, err := svc.CanSeeProjectReport(ctx, private.ID, v2UserA, v2Business)
			if err != nil {
				t.Fatalf("CanSeeProjectReport: %v", err)
			}
			if got != tc.want {
				t.Errorf("CanSeeProjectReport = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCanSeeProjectReportFollowsProjectAccess(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	svc.hubProfiles = reportingResolver(authorization.EffectiveAccess{})
	private := createV2Project(ctx, t, svc, "report-access", "private", v2UserB)

	if ok, err := svc.CanSeeProjectReport(ctx, private.ID, v2UserB, v2Business); err != nil || !ok {
		t.Errorf("creator: got %v, %v; want true", ok, err)
	}
	if ok, err := svc.CanSeeProjectReport(ctx, private.ID, v2UserA, v2Business); err != nil || ok {
		t.Errorf("stranger: got %v, %v; want false", ok, err)
	}

	// The dept-manager proxy: v2UserA manages the project's creator.
	svc.orgchart = &stubOrgchart{reports: []int64{v2UserB}}
	if ok, err := svc.CanSeeProjectReport(ctx, private.ID, v2UserA, v2Business); err != nil || !ok {
		t.Errorf("manager of the creator: got %v, %v; want true", ok, err)
	}
}

func TestCanSeeProjectReportRefusesPersonalAndMissingProjects(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	svc.hubProfiles = reportingResolver(authorization.EffectiveAccess{IsOwner: true})

	bucketID, _, err := svc.EnsureMyTasksProject(ctx, v2UserB, v2Business, 0)
	if err != nil {
		t.Fatalf("EnsureMyTasksProject: %v", err)
	}
	if ok, err := svc.CanSeeProjectReport(ctx, bucketID, v2UserA, v2Business); err != nil || ok {
		t.Errorf("someone's My Tasks bucket: got %v, %v; want false", ok, err)
	}
	if ok, err := svc.CanSeeProjectReport(ctx, bucketID, v2UserB, v2Business); err != nil || ok {
		t.Errorf("own My Tasks bucket: got %v, %v; want false", ok, err)
	}

	other := createV2Project(ctx, t, svc, "report-missing", "private", v2UserB)
	if ok, err := svc.CanSeeProjectReport(ctx, other.ID, v2UserA, v2Business+1); err != nil || ok {
		t.Errorf("another business's project: got %v, %v; want false", ok, err)
	}
	if ok, err := svc.CanSeeProjectReport(ctx, other.ID+100000, v2UserA, v2Business); err != nil || ok {
		t.Errorf("missing project: got %v, %v; want false", ok, err)
	}
}

// Ruling P8: access is not listing. An admin may open any project's report,
// but their list is what Flow's sidebar shows them, plus reports' projects.
func TestListedProjectIDsStaysCuratedForAdmins(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	svc.hubProfiles = reportingResolver(authorization.EffectiveAccess{IsAdmin: true})

	mine := createV2Project(ctx, t, svc, "listed-mine", "private", v2UserA)
	theirs := createV2Project(ctx, t, svc, "listed-theirs", "private", v2UserB)

	ids, err := svc.ListedProjectIDs(ctx, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("ListedProjectIDs: %v", err)
	}
	if !containsID(ids, mine.ID) {
		t.Errorf("admin's own project missing from %v", ids)
	}
	if containsID(ids, theirs.ID) {
		t.Errorf("admin's list leaked someone else's private project: %v", ids)
	}

	svc.orgchart = &stubOrgchart{reports: []int64{v2UserB}}
	ids, err = svc.ListedProjectIDs(ctx, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("ListedProjectIDs as manager: %v", err)
	}
	if !containsID(ids, theirs.ID) {
		t.Errorf("manager's list should include a report's project: %v", ids)
	}
}

// Insights and Work agree: whoever may see a project's dashboard may open the
// project in Work, and nobody else. Each caller is v2UserA on a private project
// v2UserB created.
func TestCanSeeProjectReportAgreesWithOpeningTheProject(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	private := createV2Project(ctx, t, svc, "report-agrees", "private", v2UserB)

	cases := []struct {
		name     string
		resolver hubprofile.Resolver
		reports  []int64
		want     bool
	}{
		{"plain member", reportingResolver(authorization.EffectiveAccess{}), nil, false},
		{"business owner", reportingResolver(authorization.EffectiveAccess{IsOwner: true}), nil, true},
		{"business admin", adminInBusiness{businessID: v2Business}, nil, true},
		{"flow admin", flowAdminInBusiness{businessID: v2Business}, nil, true},
		{"flow admin of another business", flowAdminInBusiness{businessID: v2Business + 1}, nil, false},
		{"manager of the creator", reportingResolver(authorization.EffectiveAccess{}), []int64{v2UserB}, true},
		{"manager of someone else", reportingResolver(authorization.EffectiveAccess{}), []int64{v2UserB + 1000}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc.hubProfiles = tc.resolver
			svc.orgchart = &stubOrgchart{reports: tc.reports}

			report, err := svc.CanSeeProjectReport(ctx, private.ID, v2UserA, v2Business)
			if err != nil {
				t.Fatalf("CanSeeProjectReport: %v", err)
			}
			_, openErr := svc.GetProjectV2(ctx, private.ID, v2UserA, v2Business)
			opens := openErr == nil
			if report != tc.want || opens != tc.want {
				t.Errorf("report = %v, opens in Work = %v; want both %v", report, opens, tc.want)
			}
		})
	}
}

func containsID(ids []int64, id int64) bool {
	for _, got := range ids {
		if got == id {
			return true
		}
	}
	return false
}
