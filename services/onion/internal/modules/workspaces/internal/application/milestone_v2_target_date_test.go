package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCreateMilestoneV2RejectsInvalidTargetDate(t *testing.T) {
	service := NewService(context.Background(), nil)

	_, err := service.CreateMilestoneV2(context.Background(), 1, 2, "Ship it", "", "not-a-date")
	if !errors.Is(err, ErrInvalidMilestoneDate) {
		t.Fatalf("CreateMilestoneV2 error = %v, want ErrInvalidMilestoneDate", err)
	}
}

func TestCreateMilestoneV2PersistsAndListsTargetDate(t *testing.T) {
	service, ctx, db := v2Svc(t)
	project := createV2Project(ctx, t, service, "milestone target date", "", v2UserA)

	const targetDate = "2026-10-15"
	milestone, err := service.CreateMilestoneV2(ctx, v2Business, project.ID, "Ship it", "Release the next version", targetDate)
	if err != nil {
		t.Fatalf("CreateMilestoneV2: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := db.ExecContext(ctx, "DELETE FROM flow_milestones WHERE id = ?", milestone.ID); cleanupErr != nil {
			t.Errorf("cleanup milestone: %v", cleanupErr)
		}
	})

	wantDate, parseErr := time.Parse("2006-01-02", targetDate)
	if parseErr != nil {
		t.Fatalf("parse test target date: %v", parseErr)
	}
	if !milestone.DueDate.Equal(wantDate) {
		t.Errorf("created milestone due date = %s, want %s", milestone.DueDate.Format("2006-01-02"), targetDate)
	}
	if milestone.Description != "Release the next version" {
		t.Errorf("created milestone description = %q, want %q", milestone.Description, "Release the next version")
	}

	milestones, total, err := service.ListMilestonesV2(ctx, v2Business, &project.ID, v2UserA, 50, 0)
	if err != nil {
		t.Fatalf("ListMilestonesV2: %v", err)
	}
	if total != 1 || len(milestones) != 1 {
		t.Fatalf("ListMilestonesV2 returned %d milestones and total %d, want 1 and 1", len(milestones), total)
	}
	if !milestones[0].DueDate.Equal(wantDate) {
		t.Errorf("listed milestone due date = %s, want %s", milestones[0].DueDate.Format("2006-01-02"), targetDate)
	}
}
