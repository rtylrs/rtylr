package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	flowworkspaces "github.com/voxire/onion/internal/modules/flow/workspaces"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
	"github.com/voxire/onion/internal/testutil/flowhrtime"
)

type insightsMemberProfiles map[int64]string

func (p insightsMemberProfiles) GetInsightsMemberProfiles(_ context.Context, userIDs []int64) (map[int64]flowworkspaces.InsightsMemberProfile, error) {
	profiles := make(map[int64]flowworkspaces.InsightsMemberProfile, len(userIDs))
	for _, userID := range userIDs {
		profiles[userID] = flowworkspaces.InsightsMemberProfile{UserID: userID, Name: p[userID]}
	}
	return profiles, nil
}

func TestGetInsightsCombinesOwnerAggregatesAndProfiles(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := testsupport.NewRepository(db)
	module := flowworkspaces.NewModule(ctx, db)
	hrTime := flowhrtime.New(ctx, db)
	module.SetWorkDayLineStore(hrTime)
	module.SetInsightsMemberProfileSource(insightsMemberProfiles{testutil.OwnerID: "Flow member"})

	workspaceID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Name: "Insights owner boundary", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, workspaceID, testutil.BusinessID) })

	projectID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: workspaceID, OwnerID: testutil.OwnerID,
		Name: "Insights owner boundary project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projectID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projectID, WorkspaceID: workspaceID, OwnerID: testutil.OwnerID,
		BusinessID: testutil.BusinessID, Title: "Insights member time", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	clientKey := fmt.Sprintf("insights-time-%d", time.Now().UnixNano())
	line, err := hrTime.CreateWorkDayLine(ctx, flowworkspaces.WorkDayLineInput{
		BusinessID: testutil.BusinessID,
		EmployeeID: testutil.OwnerID, WorkDayID: taskID, Date: "2035-01-01",
		ProjectID: projectID, TaskID: taskID, DurationMinutes: 30,
		ClientKey: clientKey, LoggedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("create HR-owned workday line: %v", err)
	}
	t.Cleanup(func() {
		if _, err := hrTime.DeleteWorkDayLine(ctx, line.ID, testutil.BusinessID); err != nil {
			t.Errorf("delete HR-owned workday line: %v", err)
		}
	})

	timeLogID, err := repo.CreateTimeLog(ctx, domain.TimeLogInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		TaskID: taskID, UserID: testutil.OwnerID, DurationMinutes: 45,
		Description: "Flow task timer",
	})
	if err != nil {
		t.Fatalf("create Flow time log: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTimeLog(ctx, timeLogID, testutil.OwnerID) })

	insights, err := module.Service().GetInsights(ctx, testutil.BusinessID, workspaceID)
	if err != nil {
		t.Fatalf("GetInsights: %v", err)
	}
	if len(insights.TimeByMember) != 1 {
		t.Fatalf("time by member = %+v, want one merged member", insights.TimeByMember)
	}
	member := insights.TimeByMember[0]
	if member.ID != fmt.Sprint(testutil.OwnerID) || member.Name != "Flow member" || member.Hours != 1.3 {
		t.Fatalf("time by member = %+v, want 1.3 hours for Flow member", member)
	}
}
