package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListUsers_ReturnsAtLeastTwoUsers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	users, total, err := repo.ListUsers(ctx, 100, 0, 0, "", "")
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

func TestListUsers_EmailFilterNoMatch(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	users, total, err := repo.ListUsers(ctx, 100, 0, 0, "zzznomatchzz99999@example.com", "")
	if err != nil {
		t.Fatalf("ListUsers email filter: %v", err)
	}
	if total != 0 || len(users) != 0 {
		t.Errorf("expected 0 results for no-match email, got total=%d rows=%d", total, len(users))
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

	businesses, total, err := repo.ListBusinesses(ctx, 100, 0, 0, "", 0)
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

func TestListBusinesses_NameFilterIsExact(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := "Exact Business Search Xkzq"
	res, err := db.ExecContext(ctx, "INSERT INTO businesses (name, owner_id) VALUES (?, ?)", name, testutil.OwnerID)
	if err != nil {
		t.Fatalf("insert business: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("business id: %v", err)
	}
	t.Cleanup(func() { repo.DeleteBusiness(ctx, id) })

	businesses, total, err := repo.ListBusinesses(ctx, 100, 0, 0, name, 0)
	if err != nil {
		t.Fatalf("ListBusinesses exact name: %v", err)
	}
	if total != 1 || len(businesses) != 1 {
		t.Fatalf("expected one exact name match, got total=%d rows=%d", total, len(businesses))
	}
	if businesses[0].ID != id {
		t.Errorf("expected business id=%d, got %d", id, businesses[0].ID)
	}

	businesses, total, err = repo.ListBusinesses(ctx, 100, 0, 0, "Business Search", 0)
	if err != nil {
		t.Fatalf("ListBusinesses partial name: %v", err)
	}
	if total != 0 || len(businesses) != 0 {
		t.Fatalf("expected partial name to return no rows, got total=%d rows=%d", total, len(businesses))
	}
}

func TestListBusinesses_IDAndOwnerFiltersAreExact(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := "Exact Business ID Xkzq"
	res, err := db.ExecContext(ctx, "INSERT INTO businesses (name, owner_id) VALUES (?, ?)", name, testutil.OwnerID)
	if err != nil {
		t.Fatalf("insert business: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("business id: %v", err)
	}
	t.Cleanup(func() { repo.DeleteBusiness(ctx, id) })

	businesses, total, err := repo.ListBusinesses(ctx, 100, 0, id, "", 0)
	if err != nil {
		t.Fatalf("ListBusinesses id filter: %v", err)
	}
	if total != 1 || len(businesses) != 1 || businesses[0].ID != id {
		t.Fatalf("expected one exact id match, got total=%d rows=%d", total, len(businesses))
	}

	businesses, total, err = repo.ListBusinesses(ctx, 100, 0, id, "", testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBusinesses id+owner filter: %v", err)
	}
	if total != 1 || len(businesses) != 1 || businesses[0].ID != id {
		t.Fatalf("expected one exact id+owner match, got total=%d rows=%d", total, len(businesses))
	}

	businesses, total, err = repo.ListBusinesses(ctx, 100, 0, id, "", testutil.OwnerID+999999)
	if err != nil {
		t.Fatalf("ListBusinesses wrong owner filter: %v", err)
	}
	if total != 0 || len(businesses) != 0 {
		t.Fatalf("expected wrong owner to return no rows, got total=%d rows=%d", total, len(businesses))
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
