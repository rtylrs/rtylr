package infrastructure_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure"
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

func TestResourceCRUD_HidesSensitiveUserColumns(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	columns, _, err := repo.DescribeResourceTable(ctx, "users")
	if err != nil {
		t.Fatalf("DescribeResourceTable users: %v", err)
	}
	assertNoResourceColumn(t, columns, "password_hash")

	rows, err := repo.ListResourceRows(ctx, "users", 10, 0)
	if err != nil {
		t.Fatalf("ListResourceRows users: %v", err)
	}
	assertNoResourceColumn(t, rows.Columns, "password_hash")
	for _, row := range rows.Rows {
		if _, ok := row["password_hash"]; ok {
			t.Fatal("generic resource rows exposed password_hash")
		}
	}

	row, rowColumns, _, err := repo.GetResourceRow(ctx, "users", fmt.Sprint(testutil.OwnerID))
	if err != nil {
		t.Fatalf("GetResourceRow users: %v", err)
	}
	assertNoResourceColumn(t, rowColumns, "password_hash")
	if _, ok := row["password_hash"]; ok {
		t.Fatal("generic resource row exposed password_hash")
	}
}

func assertNoResourceColumn(t *testing.T, columns []infrastructure.ResourceColumnRow, name string) {
	t.Helper()
	for _, column := range columns {
		if column.Name == name {
			t.Fatalf("generic resource metadata exposed %s", name)
		}
	}
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

func TestCreateUser_ThenGetByID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	email := "test-create-byid@example.com"
	id, err := repo.CreateUser(ctx, email, "hash1", "Create ByID", false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	user, err := repo.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if user.Email != email {
		t.Errorf("expected email=%q, got %q", email, user.Email)
	}
}

func TestCreateUser_ThenGetByEmail(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	email := "test-create-byemail@example.com"
	id, err := repo.CreateUser(ctx, email, "hash2", "Create ByEmail", false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	user, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if user.ID != id {
		t.Errorf("expected ID=%d, got %d", id, user.ID)
	}
}

func TestGetUserByEmail_UnknownReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetUserByEmail(ctx, "no-such-user-zzz99@example.com")
	if err == nil {
		t.Fatal("expected error for unknown email, got nil")
	}
}

func TestUpdateUser_ChangesFullName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	email := "test-update-name@example.com"
	id, err := repo.CreateUser(ctx, email, "hash3", "Old Name", false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	if err := repo.UpdateUser(ctx, id, "New Name", email); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}

	user, err := repo.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if !user.FullName.Valid || user.FullName.String != "New Name" {
		t.Errorf("expected FullName=New Name, got %v", user.FullName)
	}
}

func TestDeleteUser_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	email := "test-delete@example.com"
	id, err := repo.CreateUser(ctx, email, "hash4", "Delete Me", false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := repo.DeleteUser(ctx, id); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	_, err = repo.GetUserByID(ctx, id)
	if err == nil {
		t.Fatal("expected error after deletion, got nil")
	}
}

func TestBanUser_ThenUnban(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	email := "test-ban-unban@example.com"
	id, err := repo.CreateUser(ctx, email, "hash5", "Ban Test", false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	if err := repo.BanUser(ctx, id, "test ban reason"); err != nil {
		t.Fatalf("BanUser: %v", err)
	}

	user, err := repo.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID after ban: %v", err)
	}
	if !user.IsBanned {
		t.Error("expected IsBanned=true after BanUser")
	}

	if err := repo.UnbanUser(ctx, id); err != nil {
		t.Fatalf("UnbanUser: %v", err)
	}

	user, err = repo.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID after unban: %v", err)
	}
	if user.IsBanned {
		t.Error("expected IsBanned=false after UnbanUser")
	}
}

func TestListUsers_EmailFilterReturnsExactMatch(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	email := "searchable-unique-xkzq@example.com"
	id, err := repo.CreateUser(ctx, email, "hash6", "Searchable User", false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	users, total, err := repo.ListUsers(ctx, 100, 0, 0, email, "")
	if err != nil {
		t.Fatalf("ListUsers with email filter: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
	if users[0].Email != email {
		t.Errorf("expected email=%q, got %q", email, users[0].Email)
	}
}

func TestListUsers_IDFilterReturnsExactMatch(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	email := "search-by-id-xkzq@example.com"
	id, err := repo.CreateUser(ctx, email, "hash-id", "Search By ID", false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	users, total, err := repo.ListUsers(ctx, 100, 0, id, "", "")
	if err != nil {
		t.Fatalf("ListUsers with id filter: %v", err)
	}
	if total != 1 || len(users) != 1 {
		t.Fatalf("expected one exact id match, got total=%d rows=%d", total, len(users))
	}
	if users[0].ID != id {
		t.Errorf("expected id=%d, got %d", id, users[0].ID)
	}
}

func TestListUsers_NameFilterReturnsExactMatch(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := "Exact Search Name Xkzq"
	email := "search-by-name-xkzq@example.com"
	id, err := repo.CreateUser(ctx, email, "hash-name", name, false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	users, total, err := repo.ListUsers(ctx, 100, 0, 0, "", name)
	if err != nil {
		t.Fatalf("ListUsers with name filter: %v", err)
	}
	if total != 1 || len(users) != 1 {
		t.Fatalf("expected one exact name match, got total=%d rows=%d", total, len(users))
	}
	if users[0].Email != email {
		t.Errorf("expected email=%q, got %q", email, users[0].Email)
	}
}

func TestListUsers_OffsetSkipsUsers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, total0, err := repo.ListUsers(ctx, 100, 0, 0, "", "")
	if err != nil {
		t.Fatalf("ListUsers offset=0: %v", err)
	}

	_, total1, err := repo.ListUsers(ctx, 100, 1, 0, "", "")
	if err != nil {
		t.Fatalf("ListUsers offset=1: %v", err)
	}

	if total0 != total1 {
		t.Errorf("total count should be same regardless of offset: %d vs %d", total0, total1)
	}
}

func TestCreateUser_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateUser(ctx, "positive-id@example.com", "hash7", "Positive ID", false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestBanUser_StoresBanReason(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateUser(ctx, "ban-reason@example.com", "hash8", "Ban Reason Test", false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	reason := "policy violation"
	if err := repo.BanUser(ctx, id, reason); err != nil {
		t.Fatalf("BanUser: %v", err)
	}

	user, err := repo.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if !user.BanReason.Valid || user.BanReason.String != reason {
		t.Errorf("expected BanReason=%q, got %v", reason, user.BanReason)
	}
}

func TestCreateUser_ConsoleAdminFlag(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateUser(ctx, "console-admin@example.com", "hash9", "Admin User", true)
	if err != nil {
		t.Fatalf("CreateUser admin: %v", err)
	}
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	user, err := repo.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if !user.IsConsoleAdmin {
		t.Error("expected IsConsoleAdmin=true")
	}
}

func TestUpdateUser_ChangesEmail(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	origEmail := "update-email-orig@example.com"
	id, err := repo.CreateUser(ctx, origEmail, "hash10", "Email Changer", false)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	newEmail := "update-email-new@example.com"
	t.Cleanup(func() { repo.DeleteUser(ctx, id) })

	if err := repo.UpdateUser(ctx, id, "Email Changer", newEmail); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}

	user, err := repo.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if user.Email != newEmail {
		t.Errorf("expected email=%q, got %q", newEmail, user.Email)
	}
}
