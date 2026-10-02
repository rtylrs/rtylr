package api_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/api"
	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/hubprofile"
	"github.com/voxire/onion/internal/testutil"
	timev2 "github.com/voxire/voxie/gen/go/flow/time/v2"
)

type timeOwnerProfile struct{}

func (timeOwnerProfile) GetProfile(context.Context, int64, int64) (hubprofile.Profile, error) {
	return hubprofile.Profile{EffectiveAccess: &authorization.EffectiveAccess{IsOwner: true}}, nil
}

func TestWorkTimeV2PersistsTotalsWithoutHROnboarding(t *testing.T) {
	db := testutil.OpenDB(t)
	const ownerID, businessID, colleagueID = int64(993001), int64(993002), int64(993003)
	testutil.GrantBusinessAccess(t, db, businessID, ownerID, colleagueID)
	repo := infrastructure.NewRepository(db)
	svc := application.NewService(context.Background(), db)
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
	legacy, err := repo.UpsertWorkDay(ctx, domain.WorkDayInput{OwnerID: ownerID, BusinessID: businessID, EmployeeID: ownerID, Date: legacyDate, Timezone: "UTC", WorkMode: "onsite", LoggedBy: colleagueID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateWorkDayLine(ctx, domain.WorkDayLineInput{OwnerID: businessID, BusinessID: businessID, EmployeeID: ownerID, WorkDayID: legacy.ID, Date: legacyDate, DurationMinutes: 30, LoggedBy: colleagueID}); err != nil {
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
	filtered, total, err := repo.ListTeamWorkDays(ctx, businessID, []int64{colleagueID}, legacyDate, date, 50, 0)
	if err != nil || total != 1 || len(filtered) != 1 || filtered[0].ID != legacy.ID {
		t.Fatalf("roster scope: %+v %d %v", filtered, total, err)
	}
	foreign, total, err := repo.ListTeamWorkDays(ctx, businessID+1, nil, legacyDate, date, 50, 0)
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
