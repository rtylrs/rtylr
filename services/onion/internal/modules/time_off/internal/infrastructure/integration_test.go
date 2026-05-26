package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/time_off/internal/domain"
	"github.com/voxire/onion/internal/modules/time_off/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const timeOffStaffID int64 = 4101

func TestTimeOff_ListTimeOffRequests_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.ListTimeOffRequests(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListTimeOffRequests: %v", err)
	}
}

func TestTimeOff_ListTimeOffRequests_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	requests, err := repo.ListTimeOffRequests(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListTimeOffRequests: %v", err)
	}
	for _, r := range requests {
		if r.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, r.OwnerID)
		}
	}
}

func TestTimeOff_CreateTimeOffRequest_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   timeOffStaffID,
		Type:      "vacation",
		StartDate: "2026-07-01",
		EndDate:   "2026-07-05",
		DaysCount: 5.0,
		Reason:    "Summer vacation",
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestTimeOff_CreateTimeOffRequest_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   timeOffStaffID,
		Type:      "sick",
		StartDate: "2026-08-01",
		EndDate:   "2026-08-02",
		DaysCount: 2.0,
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest: %v", err)
	}

	requests, err := repo.ListTimeOffRequests(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListTimeOffRequests: %v", err)
	}
	found := false
	for _, r := range requests {
		if r.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created request id=%d not found in list", id)
	}
}

func TestTimeOff_ApproveTimeOffRequest_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   timeOffStaffID,
		Type:      "personal",
		StartDate: "2026-09-01",
		EndDate:   "2026-09-01",
		DaysCount: 1.0,
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest: %v", err)
	}

	if err := repo.ApproveTimeOffRequest(context.Background(), testutil.OwnerID, domain.TimeOffApprovalInput{
		ID:         id,
		Status:     "approved",
		ApprovedBy: testutil.OwnerID,
		Notes:      "Looks good",
	}); err != nil {
		t.Fatalf("ApproveTimeOffRequest: %v", err)
	}
}

func TestTimeOff_ApproveTimeOffRequest_UpdatesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   timeOffStaffID,
		Type:      "vacation",
		StartDate: "2026-10-01",
		EndDate:   "2026-10-03",
		DaysCount: 3.0,
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest: %v", err)
	}

	if err := repo.ApproveTimeOffRequest(context.Background(), testutil.OwnerID, domain.TimeOffApprovalInput{
		ID: id, Status: "rejected",
	}); err != nil {
		t.Fatalf("ApproveTimeOffRequest reject: %v", err)
	}

	requests, _ := repo.ListTimeOffRequests(context.Background(), testutil.OwnerID)
	for _, r := range requests {
		if r.ID == id && r.Status != "rejected" {
			t.Errorf("expected status rejected, got %s", r.Status)
		}
	}
}

func TestTimeOff_ListTimeOffRequests_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	requests, err := repo.ListTimeOffRequests(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("ListTimeOffRequests: %v", err)
	}
	if len(requests) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(requests))
	}
}

func TestTimeOff_CreateTimeOffRequest_SickType(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   timeOffStaffID,
		Type:      "sick",
		StartDate: "2026-11-01",
		EndDate:   "2026-11-01",
		DaysCount: 1.0,
		Notes:     "Doctor appointment",
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest sick: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTimeOff_CreateTimeOffRequest_HalfDay(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   timeOffStaffID,
		Type:      "personal",
		StartDate: "2026-12-01",
		EndDate:   "2026-12-01",
		DaysCount: 0.5,
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest half day: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}
