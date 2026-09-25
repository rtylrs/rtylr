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
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "vacation",
		StartDate:  "2026-07-01",
		EndDate:    "2026-07-05",
		DaysCount:  5.0,
		Reason:     "Summer vacation",
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
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "sick",
		StartDate:  "2026-08-01",
		EndDate:    "2026-08-02",
		DaysCount:  2.0,
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
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "personal",
		StartDate:  "2026-09-01",
		EndDate:    "2026-09-01",
		DaysCount:  1.0,
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
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "vacation",
		StartDate:  "2026-10-01",
		EndDate:    "2026-10-03",
		DaysCount:  3.0,
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
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "sick",
		StartDate:  "2026-11-01",
		EndDate:    "2026-11-01",
		DaysCount:  1.0,
		Notes:      "Doctor appointment",
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
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "personal",
		StartDate:  "2026-12-01",
		EndDate:    "2026-12-01",
		DaysCount:  0.5,
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest half day: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTimeOff_CreateTimeOffRequest_VacationType(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "vacation",
		StartDate:  "2027-01-10",
		EndDate:    "2027-01-14",
		DaysCount:  5.0,
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest vacation: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTimeOff_CreateTimeOffRequest_MultipleRequests(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	types := []string{"vacation", "sick", "personal"}
	for i, typ := range types {
		id, err := repo.CreateTimeOffRequest(ctx, domain.TimeOffInput{
			OwnerID:    testutil.OwnerID,
			BusinessID: testutil.BusinessID,
			StaffID:    timeOffStaffID,
			Type:       typ,
			StartDate:  "2027-02-01",
			EndDate:    "2027-02-02",
			DaysCount:  2.0 + float64(i),
		})
		if err != nil {
			t.Fatalf("type=%s: %v", typ, err)
		}
		if id <= 0 {
			t.Errorf("type=%s: expected positive ID", typ)
		}
	}
}

func TestTimeOff_ApproveTimeOffRequest_WithNotes(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTimeOffRequest(ctx, domain.TimeOffInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "sick",
		StartDate:  "2027-03-01",
		EndDate:    "2027-03-03",
		DaysCount:  3.0,
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest: %v", err)
	}

	if err := repo.ApproveTimeOffRequest(ctx, testutil.OwnerID, domain.TimeOffApprovalInput{
		ID:         id,
		Status:     "approved",
		ApprovedBy: testutil.OwnerID,
		Notes:      "Approved via integration test",
	}); err != nil {
		t.Fatalf("ApproveTimeOffRequest: %v", err)
	}
}

func TestTimeOff_ApproveTimeOffRequest_Rejected(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTimeOffRequest(ctx, domain.TimeOffInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "personal",
		StartDate:  "2027-04-01",
		EndDate:    "2027-04-01",
		DaysCount:  1.0,
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest: %v", err)
	}

	if err := repo.ApproveTimeOffRequest(ctx, testutil.OwnerID, domain.TimeOffApprovalInput{
		ID: id, Status: "rejected",
	}); err != nil {
		t.Fatalf("ApproveTimeOffRequest reject: %v", err)
	}

	requests, _ := repo.ListTimeOffRequests(ctx, testutil.OwnerID)
	for _, r := range requests {
		if r.ID == id && r.Status != "rejected" {
			t.Errorf("expected status=rejected, got %s", r.Status)
		}
	}
}

func TestTimeOff_ListTimeOffRequests_PositiveIDs(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Create one so the list is not empty.
	_, _ = repo.CreateTimeOffRequest(ctx, domain.TimeOffInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "vacation",
		StartDate:  "2027-05-01",
		EndDate:    "2027-05-02",
		DaysCount:  2.0,
	})

	requests, err := repo.ListTimeOffRequests(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListTimeOffRequests: %v", err)
	}
	for _, r := range requests {
		if r.ID <= 0 {
			t.Errorf("expected positive ID, got %d", r.ID)
		}
	}
}

func TestTimeOff_CreateTimeOffRequest_WithReason(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "personal",
		StartDate:  "2027-06-01",
		EndDate:    "2027-06-01",
		DaysCount:  1.0,
		Reason:     "Family event",
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest with reason: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTimeOff_CreateTimeOffRequest_LargeDaysCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "vacation",
		StartDate:  "2027-07-01",
		EndDate:    "2027-07-30",
		DaysCount:  22.0,
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest large days: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTimeOff_ListTimeOffRequests_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	a, err := repo.ListTimeOffRequests(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	b, err := repo.ListTimeOffRequests(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if len(a) != len(b) {
		t.Errorf("idempotency: got %d then %d", len(a), len(b))
	}
}

func TestTimeOff_ApproveTimeOffRequest_AppearsPending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTimeOffRequest(ctx, domain.TimeOffInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		StaffID:    timeOffStaffID,
		Type:       "vacation",
		StartDate:  "2027-09-01",
		EndDate:    "2027-09-05",
		DaysCount:  5.0,
	})
	if err != nil {
		t.Fatalf("CreateTimeOffRequest: %v", err)
	}

	// Before approval the request should appear in the list.
	requests, err := repo.ListTimeOffRequests(ctx, testutil.OwnerID)
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
		t.Errorf("pending request id=%d not found in list", id)
	}
}

func TestTimeOff_CreateTimeOffRequest_UniqueIDs(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.CreateTimeOffRequest(ctx, domain.TimeOffInput{
		OwnerID: testutil.OwnerID, StaffID: timeOffStaffID, Type: "sick",
		BusinessID: testutil.BusinessID,
		StartDate:  "2027-08-01", EndDate: "2027-08-01", DaysCount: 1.0,
	})
	if err != nil {
		t.Fatalf("req 1: %v", err)
	}
	id2, err := repo.CreateTimeOffRequest(ctx, domain.TimeOffInput{
		OwnerID: testutil.OwnerID, StaffID: timeOffStaffID, Type: "personal",
		BusinessID: testutil.BusinessID,
		StartDate:  "2027-08-02", EndDate: "2027-08-02", DaysCount: 1.0,
	})
	if err != nil {
		t.Fatalf("req 2: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected distinct IDs, both got %d", id1)
	}
}
