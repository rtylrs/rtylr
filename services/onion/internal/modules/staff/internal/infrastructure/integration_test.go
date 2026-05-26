package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/staff/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListStaff_ReturnsAliceAndBob(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	staff, err := repo.ListStaff(ctx, testutil.OwnerID, "", "", "")
	if err != nil {
		t.Fatalf("ListStaff: %v", err)
	}
	if len(staff) < 2 {
		t.Fatalf("expected at least 2 staff members, got %d", len(staff))
	}

	names := make(map[string]bool)
	for _, s := range staff {
		names[s.FirstName] = true
	}
	if !names["Alice"] {
		t.Error("expected Alice in staff list")
	}
	if !names["Bob"] {
		t.Error("expected Bob in staff list")
	}
}

func TestGetStaffByID_AliceID4101(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	s, err := repo.GetStaffByID(ctx, testutil.OwnerID, 4101)
	if err != nil {
		t.Fatalf("GetStaffByID: %v", err)
	}
	if s.FirstName != "Alice" {
		t.Errorf("expected FirstName=Alice, got %q", s.FirstName)
	}
	if s.EmploymentType != "full_time" {
		t.Errorf("expected EmploymentType=full_time, got %q", s.EmploymentType)
	}
}
