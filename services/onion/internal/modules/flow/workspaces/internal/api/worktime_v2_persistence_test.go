package api_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	flowworkspaces "github.com/voxire/onion/internal/modules/flow/workspaces"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/api"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/hubprofile"
	"github.com/voxire/onion/internal/testutil"
	"github.com/voxire/onion/internal/testutil/flowhrtime"
	timev2 "github.com/voxire/voxie/gen/go/flow/time/v2"
)

type timeOwnerProfile struct{}

func (timeOwnerProfile) GetProfile(context.Context, int64, int64) (hubprofile.Profile, error) {
	return hubprofile.Profile{EffectiveAccess: &authorization.EffectiveAccess{IsOwner: true}}, nil
}

type testWorkdayRosterSource struct {
	memberIDs   []int64
	employeeIDs map[int64]int64
}

func (s testWorkdayRosterSource) ResolveWorkdayRoster(_ context.Context, _ int64, candidates []int64, includeAll bool) (flowworkspaces.WorkdayRoster, error) {
	allowed := make(map[int64]struct{}, len(candidates))
	for _, id := range candidates {
		allowed[id] = struct{}{}
	}
	result := flowworkspaces.WorkdayRoster{EmployeeUserIDs: make(map[int64]int64)}
	for _, userID := range s.memberIDs {
		result.BusinessMemberUserIDs = append(result.BusinessMemberUserIDs, userID)
		if includeAll {
			result.ActiveMemberUserIDs = append(result.ActiveMemberUserIDs, userID)
			continue
		}
		if _, ok := allowed[userID]; ok {
			result.ActiveMemberUserIDs = append(result.ActiveMemberUserIDs, userID)
		}
	}
	for employeeID, userID := range s.employeeIDs {
		if includeAll {
			result.EmployeeUserIDs[employeeID] = userID
			continue
		}
		if _, ok := allowed[userID]; ok {
			result.EmployeeUserIDs[employeeID] = userID
		}
	}
	return result, nil
}

func TestWorkTimeV2PersistsTotalsWithoutHROnboarding(t *testing.T) {
	db := testutil.OpenDB(t)
	const ownerID, businessID, colleagueID = int64(993001), int64(993002), int64(993003)
	testutil.GrantBusinessAccess(t, db, businessID, ownerID, colleagueID)
	repo := testsupport.NewRepository(db)
	flowModule := flowworkspaces.NewModule(context.Background(), db)
	flowModule.SetWorkDayLineStore(flowhrtime.New(context.Background(), db))
	flowModule.SetWorkdayRosterSource(testWorkdayRosterSource{
		memberIDs: []int64{ownerID, colleagueID}, employeeIDs: map[int64]int64{ownerID: colleagueID},
	})
	svc := flowModule.Service()
	svc.SetHubProfileResolver(timeOwnerProfile{})
	handler := api.NewWorkTimeAPIV2(svc)
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: ownerID, BusinessID: businessID})
	const date = "2026-10-02"
	start, end := "09:00", "17:00"
	minutes, updatedMinutes := int32(60), int32(90)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM hr_time_entries WHERE business_id = ?", businessID)
		_, _ = db.Exec("DELETE FROM flow_work_days WHERE business_id = ?", businessID)
		_, _ = db.Exec("DELETE FROM hr_employees WHERE business_id = ?", businessID)
	})
	day, err := handler.UpsertWorkDay(ctx, &timev2.UpsertWorkDayRequest{Date: date, Timezone: "UTC", WorkMode: "onsite", StartTime: &start, EndTime: &end, BreakMinutes: 60})
	if err != nil {
		t.Fatal(err)
	}
	line, err := handler.CreateWorkDayLine(ctx, &timev2.CreateWorkDayLineRequest{WorkDayId: day.WorkDay.Id, ProjectId: testutil.Int64Ptr(123), DurationMinutes: &minutes, Description: "Persistence check", ClientKey: "worktime-v2-persistence"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := handler.UpdateWorkDayLine(ctx, &timev2.UpdateWorkDayLineRequest{Id: line.Line.Id, ProjectId: testutil.Int64Ptr(123), DurationMinutes: &updatedMinutes}); err != nil {
		t.Fatal(err)
	}
	own, err := handler.GetWorkDay(ctx, &timev2.GetWorkDayRequest{Date: date})
	if err != nil || own.WorkDay.GetLoggedMinutes() != 90 || len(own.Lines) != 1 {
		t.Fatalf("detail totals: %+v %v", own, err)
	}
	rangeResult, err := handler.ListWorkDays(ctx, &timev2.ListWorkDaysRequest{FromDate: date, ToDate: date, Limit: 7})
	if err != nil || len(rangeResult.WorkDays) != 1 || rangeResult.WorkDays[0].GetLoggedMinutes() != 90 || rangeResult.WorkDays[0].GetGapMinutes() != 330 {
		t.Fatalf("range totals: %+v %v", rangeResult, err)
	}
	// A legacy HR employee ID can collide with the owner's user ID without changing attribution.
	if _, err := db.Exec("INSERT INTO hr_employees (id, owner_id, business_id, user_id, first_name, last_name, email) VALUES (?, ?, ?, ?, 'Legacy', 'Test', 'worktime-v2-legacy@example.test')", ownerID, businessID, businessID, colleagueID); err != nil {
		t.Fatal(err)
	}
	const legacyDate = "2026-10-01"
	legacy, err := repo.UpsertWorkDay(ctx, domain.WorkDayInput{BusinessID: businessID, EmployeeID: ownerID, Date: legacyDate, Timezone: "UTC", WorkMode: "onsite", LoggedBy: colleagueID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateWorkDayLine(ctx, domain.WorkDayLineInput{BusinessID: businessID, EmployeeID: ownerID, WorkDayID: legacy.ID, Date: legacyDate, ProjectID: 123, DurationMinutes: 30, LoggedBy: colleagueID}); err != nil {
		t.Fatal(err)
	}
	team, err := handler.ListTeamWorkDays(ctx, &timev2.ListTeamWorkDaysRequest{FromDate: legacyDate, ToDate: date, Limit: 50})
	if err != nil || len(team.WorkDays) != 2 {
		t.Fatalf("team days: %+v %v", team, err)
	}
	for _, d := range team.WorkDays {
		if d.Id == day.WorkDay.Id && (d.GetUserId() != ownerID || d.GetLoggedMinutes() != 90) {
			t.Fatalf("user-keyed day lost: %+v", d)
		}
		if d.Id == legacy.ID && (d.GetUserId() != colleagueID || d.GetLoggedMinutes() != 30) {
			t.Fatalf("HR-keyed day lost: %+v", d)
		}
	}
	filtered, total, err := repo.ListTeamWorkDays(ctx, businessID, []int64{ownerID}, []int64{ownerID}, []int64{colleagueID}, legacyDate, date, 50, 0)
	if err != nil || total != 1 || len(filtered) != 1 || filtered[0].ID != legacy.ID {
		t.Fatalf("roster scope: %+v %d %v", filtered, total, err)
	}
	foreign, total, err := repo.ListTeamWorkDays(ctx, businessID+1, nil, nil, nil, legacyDate, date, 50, 0)
	if err != nil || total != 0 || len(foreign) != 0 {
		t.Fatalf("cross-tenant days: %+v %d %v", foreign, total, err)
	}
	if _, err := handler.DeleteWorkDayLine(ctx, &timev2.DeleteWorkDayLineRequest{Id: line.Line.Id}); err != nil {
		t.Fatal(err)
	}
	rangeResult, err = handler.ListWorkDays(ctx, &timev2.ListWorkDaysRequest{FromDate: date, ToDate: date, Limit: 7})
	if err != nil || rangeResult.WorkDays[0].GetLoggedMinutes() != 0 {
		t.Fatalf("deleted entry still counted: %+v %v", rangeResult, err)
	}
}

func TestWorkTypeV2PersistsTreeAndArchives(t *testing.T) {
	db := testutil.OpenDB(t)
	const ownerID, businessID = int64(993011), int64(993012)
	testutil.GrantBusinessAccess(t, db, businessID, ownerID)
	svc := application.NewService(context.Background(), testsupport.NewRepository(db))
	svc.SetHubProfileResolver(timeOwnerProfile{})
	handler := api.NewWorkTimeAPIV2(svc)
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: ownerID, BusinessID: businessID})
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM flow_work_types WHERE business_id = ?", businessID) })
	created, err := handler.CreateWorkType(ctx, &timev2.CreateWorkTypeRequest{Name: "Delivery", Description: "Initial description"})
	if err != nil {
		t.Fatal(err)
	}
	rootID := created.WorkType.Id
	child, err := handler.CreateWorkType(ctx, &timev2.CreateWorkTypeRequest{ParentId: testutil.Int64Ptr(rootID), Name: "Review"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := handler.CreateWorkType(ctx, &timev2.CreateWorkTypeRequest{ParentId: testutil.Int64Ptr(child.WorkType.Id), Name: "Too deep"}); err == nil {
		t.Fatal("accepted a third work-type level")
	}
	updated, err := handler.UpdateWorkType(ctx, &timev2.UpdateWorkTypeRequest{Id: rootID, Name: "Shipping", Description: "Updated description"})
	if err != nil || updated.WorkType.GetName() != "Shipping" {
		t.Fatalf("update: %+v %v", updated, err)
	}
	listed, err := handler.ListWorkTypes(ctx, &timev2.ListWorkTypesRequest{})
	if err != nil || len(listed.WorkTypes) != 2 {
		t.Fatalf("list: %+v %v", listed, err)
	}
	if _, err := handler.DeleteWorkType(ctx, &timev2.DeleteWorkTypeRequest{Id: child.WorkType.Id}); err != nil {
		t.Fatal(err)
	}
	if _, err := handler.DeleteWorkType(ctx, &timev2.DeleteWorkTypeRequest{Id: rootID}); err != nil {
		t.Fatal(err)
	}
	active, err := handler.ListWorkTypes(ctx, &timev2.ListWorkTypesRequest{})
	if err != nil || len(active.WorkTypes) != 0 {
		t.Fatalf("archive remained active: %+v %v", active, err)
	}
	archived, err := handler.ListWorkTypes(ctx, &timev2.ListWorkTypesRequest{IncludeArchived: true})
	if err != nil || len(archived.WorkTypes) != 2 {
		t.Fatalf("archive missing: %+v %v", archived, err)
	}
	for _, workType := range archived.WorkTypes {
		if !workType.GetArchived() {
			t.Fatal("type was not archived")
		}
	}
}
