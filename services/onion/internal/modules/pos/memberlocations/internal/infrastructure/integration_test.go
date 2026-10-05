package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

// fixtureBusinessID is the business id mapped to owner_id=1001 (businesses: id=1).
const fixtureBusinessID int64 = 1

// fixtureUserID is the member user in the fixture (pos_member_locations user_id=1002).
const fixtureUserID int64 = 1002

func TestMemberLocations_ListForUser_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)

	locs, err := repo.ListForUser(context.Background(), fixtureBusinessID, fixtureUserID)
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	_ = locs
}

func TestMemberLocations_ListForUser_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)

	locs, err := repo.ListForUser(context.Background(), fixtureBusinessID, fixtureUserID)
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	if locs == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestMemberLocations_ListForUser_WrongUserReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)

	locs, err := repo.ListForUser(context.Background(), fixtureBusinessID, 999999)
	if err != nil {
		t.Fatalf("ListForUser wrong user: %v", err)
	}
	if len(locs) != 0 {
		t.Fatalf("expected 0 records for unknown user, got %d", len(locs))
	}
}

func TestMemberLocations_ListForUser_AllHaveCorrectBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)

	locs, err := repo.ListForUser(context.Background(), fixtureBusinessID, fixtureUserID)
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	for i, l := range locs {
		if l.BusinessID != fixtureBusinessID {
			t.Errorf("locs[%d].BusinessID = %d, want %d", i, l.BusinessID, fixtureBusinessID)
		}
	}
}

func TestMemberLocations_ListForUser_AllHaveCorrectUserID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)

	locs, err := repo.ListForUser(context.Background(), fixtureBusinessID, fixtureUserID)
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	for i, l := range locs {
		if l.UserID != fixtureUserID {
			t.Errorf("locs[%d].UserID = %d, want %d", i, l.UserID, fixtureUserID)
		}
	}
}

func TestMemberLocations_ListForUser_AllHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)

	locs, err := repo.ListForUser(context.Background(), fixtureBusinessID, fixtureUserID)
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	for i, l := range locs {
		if l.ID == 0 {
			t.Errorf("locs[%d].ID is zero", i)
		}
	}
}

func TestMemberLocations_Assign_ThenListContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	const testUser int64 = 7701

	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM pos_member_locations WHERE user_id = ? AND location_id = ?", testUser, int64(3001))
	})

	id, err := repo.Assign(ctx, fixtureBusinessID, testUser, 3001, false, false, true)
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	if id == 0 {
		t.Fatal("Assign returned ID=0")
	}

	locs, err := repo.ListForUser(ctx, fixtureBusinessID, testUser)
	if err != nil {
		t.Fatalf("ListForUser after assign: %v", err)
	}
	found := false
	for _, l := range locs {
		if l.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("assigned record id=%d not found in list", id)
	}
}

func TestMemberLocations_Remove_ThenListDoesNotContain(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	const testUser int64 = 7702

	id, err := repo.Assign(ctx, fixtureBusinessID, testUser, 3001, false, false, false)
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}

	err = repo.Remove(ctx, fixtureBusinessID, id)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	locs, err := repo.ListForUser(ctx, fixtureBusinessID, testUser)
	if err != nil {
		t.Fatalf("ListForUser after remove: %v", err)
	}
	for _, l := range locs {
		if l.ID == id {
			t.Errorf("removed record id=%d still found in list", id)
		}
	}
}

func TestMemberLocations_GetMyLocations_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)

	locs, err := repo.GetMyLocations(context.Background(), fixtureBusinessID, fixtureUserID)
	if err != nil {
		t.Fatalf("GetMyLocations: %v", err)
	}
	_ = locs
}

func TestMemberLocations_GetMyLocations_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)

	locs, err := repo.GetMyLocations(context.Background(), fixtureBusinessID, fixtureUserID)
	if err != nil {
		t.Fatalf("GetMyLocations: %v", err)
	}
	if locs == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestMemberLocations_GetMyLocations_WrongUserReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)

	locs, err := repo.GetMyLocations(context.Background(), fixtureBusinessID, 999999)
	if err != nil {
		t.Fatalf("GetMyLocations wrong user: %v", err)
	}
	if len(locs) != 0 {
		t.Fatalf("expected 0 for unknown user, got %d", len(locs))
	}
}

func TestMemberLocations_ListForUser_AllHaveNonZeroLocationID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)

	locs, err := repo.ListForUser(context.Background(), fixtureBusinessID, fixtureUserID)
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	for i, l := range locs {
		if l.LocationID == 0 {
			t.Errorf("locs[%d].LocationID is zero", i)
		}
	}
}
