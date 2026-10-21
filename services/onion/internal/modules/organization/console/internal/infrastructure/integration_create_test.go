package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
