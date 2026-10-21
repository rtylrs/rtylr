package hr_time_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	hrtime "github.com/voxire/onion/internal/modules/hr/hr_time"
	"github.com/voxire/onion/internal/testutil"
)

func TestFlowWorkdayEntriesOwnTheirPersistenceAndScope(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	module := hrtime.NewModule(ctx, db, nil)
	clientKey := fmt.Sprintf("flow-entry-owner-%d", time.Now().UnixNano())
	input := hrtime.FlowWorkdayEntryInput{
		OwnerID: testutil.BusinessID, BusinessID: testutil.BusinessID,
		EmployeeID: 4101, WorkDayID: 992001, TaskID: 992002, Date: "2026-10-04",
		DurationMinutes: 45, Description: "Flow entry", ClientKey: clientKey,
		LoggedBy: testutil.OwnerID,
	}
	var createdID int64
	t.Cleanup(func() {
		if createdID > 0 {
			if _, err := module.DeleteFlowWorkdayEntry(ctx, createdID, input.OwnerID, input.BusinessID); err != nil {
				t.Errorf("delete Flow entry fixture: %v", err)
			}
		}
	})

	created, err := module.CreateFlowWorkdayEntry(ctx, input)
	if err != nil {
		t.Fatalf("create Flow entry: %v", err)
	}
	createdID = created.ID
	if created.ID <= 0 || created.DurationMinutes != 45 {
		t.Fatalf("created entry = %+v", created)
	}

	memberMinutes, err := module.SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx, input.BusinessID, []int64{input.TaskID})
	if err != nil || len(memberMinutes) != 1 || memberMinutes[0].EmployeeID != input.EmployeeID || memberMinutes[0].Minutes != 45 {
		t.Fatalf("member time aggregate = %+v, err=%v", memberMinutes, err)
	}
	foreignMemberMinutes, err := module.SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx, testutil.OtherBusinessID, []int64{input.TaskID})
	if err != nil || len(foreignMemberMinutes) != 0 {
		t.Fatalf("cross-business member time aggregate = %+v, err=%v", foreignMemberMinutes, err)
	}

	retry := input
	retry.DurationMinutes = 90
	retry.Description = "retry must not replace the original"
	retried, err := module.CreateFlowWorkdayEntry(ctx, retry)
	if err != nil {
		t.Fatalf("retry Flow entry: %v", err)
	}
	if retried.ID != created.ID || retried.DurationMinutes != 45 || retried.Description != input.Description {
		t.Fatalf("retry = %+v, want original entry %+v", retried, created)
	}

	entries, total, err := module.ListFlowWorkdayEntries(ctx, input.OwnerID, input.BusinessID, input.WorkDayID)
	if err != nil || len(entries) != 1 || total != 45 {
		t.Fatalf("list entries = %+v, total=%d, err=%v", entries, total, err)
	}
	otherBusinessEntries, _, err := module.ListFlowWorkdayEntries(ctx, input.OwnerID, testutil.OtherBusinessID, input.WorkDayID)
	if err != nil || len(otherBusinessEntries) != 0 {
		t.Fatalf("cross-business list = %+v, err=%v", otherBusinessEntries, err)
	}

	foreign := input
	foreign.ID = created.ID
	foreign.BusinessID = testutil.OtherBusinessID
	foreign.DurationMinutes = 60
	if _, err := module.UpdateFlowWorkdayEntry(ctx, foreign); err == nil {
		t.Fatal("cross-business update succeeded")
	}
	entries, total, err = module.ListFlowWorkdayEntries(ctx, input.OwnerID, input.BusinessID, input.WorkDayID)
	if err != nil || len(entries) != 1 || total != 45 || entries[0].Description != input.Description {
		t.Fatalf("cross-business update changed entry: %+v, total=%d, err=%v", entries, total, err)
	}

	updatedInput := input
	updatedInput.ID = created.ID
	updatedInput.DurationMinutes = 60
	updatedInput.Description = "updated by HR Time"
	updated, err := module.UpdateFlowWorkdayEntry(ctx, updatedInput)
	if err != nil || updated.DurationMinutes != 60 || updated.Description != updatedInput.Description {
		t.Fatalf("update = %+v, err=%v", updated, err)
	}

	owner, found, err := module.FlowWorkdayEntryOwner(ctx, created.ID, input.OwnerID, input.BusinessID)
	if err != nil || !found || owner.WorkDayID != input.WorkDayID || owner.EmployeeID != input.EmployeeID {
		t.Fatalf("entry owner = %+v, found=%t, err=%v", owner, found, err)
	}
	deleted, err := module.DeleteFlowWorkdayEntry(ctx, created.ID, input.OwnerID, testutil.OtherBusinessID)
	if err != nil || deleted {
		t.Fatalf("cross-business delete = %t, err=%v", deleted, err)
	}
	deleted, err = module.DeleteFlowWorkdayEntry(ctx, created.ID, input.OwnerID, input.BusinessID)
	if err != nil || !deleted {
		t.Fatalf("owner delete = %t, err=%v", deleted, err)
	}
	createdID = 0
}
