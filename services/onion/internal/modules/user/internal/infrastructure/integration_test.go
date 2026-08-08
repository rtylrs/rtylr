package infrastructure_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/modules/user/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUser_GetUserByID_Owner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	u, err := repo.GetUserByID(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if u.ID != testutil.OwnerID {
		t.Errorf("expected ID %d, got %d", testutil.OwnerID, u.ID)
	}
	if u.Email == "" {
		t.Error("expected non-empty email")
	}
}

func TestUser_GetUserByID_Staff(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	u, err := repo.GetUserByID(context.Background(), 1002)
	if err != nil {
		t.Fatalf("GetUserByID staff: %v", err)
	}
	if u.ID != 1002 {
		t.Errorf("expected ID 1002, got %d", u.ID)
	}
}

func TestUser_GetUserByEmail_Owner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	// First get the email
	u, err := repo.GetUserByID(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	// Then look up by email
	u2, err := repo.GetUserByEmail(context.Background(), u.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if u2.ID != testutil.OwnerID {
		t.Errorf("expected ID %d, got %d", testutil.OwnerID, u2.ID)
	}
}

func TestUser_CreateUser_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	email := fmt.Sprintf("testuser+%d@example.com", time.Now().UnixNano())
	u := &domain.User{
		Email:        email,
		PasswordHash: "$2a$10$fakehashedpassword",
		FullName:     "Test User",
	}
	id, err := repo.CreateUser(context.Background(), u)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.DeleteUser(context.Background(), id) })
}

func TestUser_UpdateUser_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	email := fmt.Sprintf("updatename+%d@example.com", time.Now().UnixNano())
	id, err := repo.CreateUser(context.Background(), &domain.User{
		Email: email, PasswordHash: "hash", FullName: "OldName",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteUser(context.Background(), id) })

	if err := repo.UpdateUser(context.Background(), &domain.User{ID: id, FullName: "NewName"}); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}

	u, err := repo.GetUserByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetUserByID after update: %v", err)
	}
	if u.FullName != "NewName" {
		t.Errorf("expected FullName NewName, got %s", u.FullName)
	}
}

func TestUser_UpdateUserPassword_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	email := fmt.Sprintf("updatepw+%d@example.com", time.Now().UnixNano())
	id, err := repo.CreateUser(context.Background(), &domain.User{
		Email: email, PasswordHash: "oldhash",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteUser(context.Background(), id) })

	if err := repo.UpdateUserPassword(context.Background(), id, "newhash"); err != nil {
		t.Fatalf("UpdateUserPassword: %v", err)
	}
}

func TestUser_DeleteUser_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	email := fmt.Sprintf("delete+%d@example.com", time.Now().UnixNano())
	id, err := repo.CreateUser(context.Background(), &domain.User{
		Email: email, PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := repo.DeleteUser(context.Background(), id); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	_, err = repo.GetUserByID(context.Background(), id)
	if err == nil {
		t.Error("expected error after deleting user")
	}
}

func TestUser_GetUserBusinessID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetUserBusinessID(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetUserBusinessID: %v", err)
	}
}

func TestUser_GetUserRole_Owner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	role, err := repo.GetUserRole(context.Background(), 1, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetUserRole: %v", err)
	}
	if role != "owner" {
		t.Errorf("expected role owner, got %s", role)
	}
}

func TestUser_RecordLogin_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.RecordLogin(context.Background(), testutil.OwnerID, "owner@example.com", "password", "127.0.0.1", "IntegrationTest/1.0", true)
	if err != nil {
		t.Fatalf("RecordLogin: %v", err)
	}
}

func TestUser_AssignSubscription_FreeCreatesActiveSubscription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	email := fmt.Sprintf("free-subscription+%d@example.com", time.Now().UnixNano())
	id, err := repo.CreateUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: "hash",
		FullName:     "Free Owner",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM subscriptions WHERE owner_id = ?", id)
		_ = repo.DeleteUser(context.Background(), id)
	})

	if err := repo.AssignSubscription(ctx, id, "free"); err != nil {
		t.Fatalf("AssignSubscription: %v", err)
	}

	var slug, status string
	var trialEndsAt sql.NullTime
	err = db.QueryRowContext(ctx, `
SELECT p.slug, s.status, s.trial_ends_at
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.owner_id = ?`, id).Scan(&slug, &status, &trialEndsAt)
	if err != nil {
		t.Fatalf("read subscription: %v", err)
	}
	if slug != "free" {
		t.Fatalf("expected free plan slug, got %q", slug)
	}
	if status != "active" {
		t.Fatalf("expected active status for free signup, got %q", status)
	}
	if trialEndsAt.Valid {
		t.Fatalf("expected no trial end for free signup, got %v", trialEndsAt.Time)
	}
}

func TestUser_GetUserByID_Unknown_Error(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetUserByID(context.Background(), 9999999)
	if err == nil {
		t.Fatal("expected error for unknown user ID")
	}
}
