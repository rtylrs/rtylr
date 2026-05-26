package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListUsers_ReturnsAtLeastTwoUsers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	users, total, err := repo.ListUsers(ctx, 100, 0, "")
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if total < 2 {
		t.Errorf("expected at least 2 users, got total=%d", total)
	}
	if len(users) < 2 {
		t.Errorf("expected at least 2 users in page, got %d", len(users))
	}
}

func TestListUsers_SearchFilters(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Search for a term that should match no users.
	users, total, err := repo.ListUsers(ctx, 100, 0, "zzznomatchzz99999")
	if err != nil {
		t.Fatalf("ListUsers search: %v", err)
	}
	if total != 0 || len(users) != 0 {
		t.Errorf("expected 0 results for no-match search, got total=%d rows=%d", total, len(users))
	}
}

func TestGetUserByID_OwnerUser(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	user, err := repo.GetUserByID(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if user.ID != testutil.OwnerID {
		t.Errorf("expected ID=%d, got %d", testutil.OwnerID, user.ID)
	}
}

func TestGetUserByID_StaffUser(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	user, err := repo.GetUserByID(ctx, 1002)
	if err != nil {
		t.Fatalf("GetUserByID staff: %v", err)
	}
	if user.ID != 1002 {
		t.Errorf("expected ID=1002, got %d", user.ID)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetUserByID(ctx, 999999999)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestListBusinesses_ReturnsTestBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	businesses, total, err := repo.ListBusinesses(ctx, 100, 0, "")
	if err != nil {
		t.Fatalf("ListBusinesses: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1 business, got total=%d", total)
	}

	found := false
	for _, b := range businesses {
		if b.Name == "Test Business" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Test Business' in list, got %v", businesses)
	}
}

func TestGetBusinessByID_TestBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	biz, err := repo.GetBusinessByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetBusinessByID: %v", err)
	}
	if biz.Name != "Test Business" {
		t.Errorf("expected Name='Test Business', got %q", biz.Name)
	}
	if biz.OwnerID != testutil.OwnerID {
		t.Errorf("expected OwnerID=%d, got %d", testutil.OwnerID, biz.OwnerID)
	}
}

func TestGetBusinessByID_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetBusinessByID(ctx, 999999)
	if err == nil {
		t.Fatal("expected error for non-existent business, got nil")
	}
}

func TestGetStats_ReturnsNonNegativeCounts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalUsers < 0 {
		t.Errorf("expected non-negative TotalUsers, got %d", stats.TotalUsers)
	}
	if stats.TotalBusinesses < 0 {
		t.Errorf("expected non-negative TotalBusinesses, got %d", stats.TotalBusinesses)
	}
	if stats.TotalUsers < 2 {
		t.Errorf("expected at least 2 users in stats, got %d", stats.TotalUsers)
	}
	if stats.TotalBusinesses < 1 {
		t.Errorf("expected at least 1 business in stats, got %d", stats.TotalBusinesses)
	}
}

func TestListPlans_ReturnsPlans(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	plans, err := repo.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	// Plans may be empty in test DB but call must succeed.
	_ = plans
}

func TestGetUserByEmail_OwnerEmailWorks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Resolve the owner's email from the DB.
	var email string
	if err := db.QueryRowContext(ctx, `SELECT email FROM users WHERE id = ?`, testutil.OwnerID).Scan(&email); err != nil {
		t.Fatalf("resolve email: %v", err)
	}

	user, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if user.ID != testutil.OwnerID {
		t.Errorf("expected ID=%d, got %d", testutil.OwnerID, user.ID)
	}
}
