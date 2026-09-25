package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const teamStaffUserID int64 = 1002

func TestTeam_ListPermissions_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.ListPermissions(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
}

func TestTeam_BulkInvite_ReportsMissingGroup(t *testing.T) {
	db := testutil.OpenDB(t)
	email := fmt.Sprintf("missing-group-invite-%d@test.local", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM invite_tokens WHERE user_id IN (SELECT id FROM users WHERE email = ?)", email)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM user_app_permissions WHERE user_id IN (SELECT id FROM users WHERE email = ?)", email)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE email = ?", email)
	})
	service := application.NewService(context.Background(), db)
	results, err := service.BulkInvite(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.BulkInviteInput{
		Emails:  []string{email},
		Role:    domain.RoleStaff,
		GroupID: 9999999,
	})
	if err != nil {
		t.Fatalf("BulkInvite: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("BulkInvite returned %d results, want 1", len(results))
	}
	if results[0].Status != "error" {
		t.Fatalf("status = %q, want error", results[0].Status)
	}
	if !strings.Contains(results[0].Message, "group assignment failed") {
		t.Fatalf("message = %q, want group assignment failure", results[0].Message)
	}
}

func TestTeam_ListPermissions_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	perms, err := repo.ListPermissions(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	for _, p := range perms {
		if p.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, p.OwnerID)
		}
	}
}

func TestTeam_GetPermissionByUserID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	// May be nil if the user has no permissions row
	_, err := repo.GetPermissionByUserID(context.Background(), teamStaffUserID)
	if err != nil {
		t.Fatalf("GetPermissionByUserID: %v", err)
	}
}

func TestTeam_GetUserByID_Owner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	email, fullName, _, err := repo.GetUserByID(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if email == "" {
		t.Error("expected non-empty email for owner")
	}
	_ = fullName
}

func TestTeam_GetUserByID_Unknown_Error(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, _, _, err := repo.GetUserByID(context.Background(), 9999999)
	if err == nil {
		t.Fatal("expected error for unknown user ID")
	}
}

func TestTeam_CountActiveMembers_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	count, err := repo.CountActiveMembers(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountActiveMembers: %v", err)
	}
	if count < 0 {
		t.Errorf("expected count >= 0, got %d", count)
	}
}

func TestTeam_IsOnboarded_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_ = repo.IsOnboarded(context.Background(), testutil.OwnerID)
}

func TestTeam_ListGroups_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.ListGroups(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
}

func TestTeam_CreateGroup_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateGroup(context.Background(), testutil.OwnerID, domain.CreateGroupInput{
		Name:        "Integration Test Group",
		Description: "Created by integration test",
		Color:       "#FF5733",
	})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.DeleteGroup(context.Background(), testutil.OwnerID, id) })
}

func TestTeam_CreateGroup_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateGroup(context.Background(), testutil.OwnerID, domain.CreateGroupInput{Name: "ListGroup"})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteGroup(context.Background(), testutil.OwnerID, id) })

	groups, err := repo.ListGroups(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	found := false
	for _, g := range groups {
		if g.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created group id=%d not found in list", id)
	}
}

func TestTeam_CreateInviteToken_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM invite_tokens WHERE token = ?", "test-invite-token-xyz")
	})
	err := repo.CreateInviteToken(context.Background(), teamStaffUserID, testutil.OwnerID,
		"test-invite-token-xyz", time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("CreateInviteToken: %v", err)
	}
}

func TestTeam_GetInviteToken_ReturnsData(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	token := "get-invite-token-abc"
	expiry := time.Now().Add(1 * time.Hour)
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM invite_tokens WHERE token = ?", token)
	})
	if err := repo.CreateInviteToken(context.Background(), teamStaffUserID, testutil.OwnerID, token, expiry); err != nil {
		t.Fatalf("CreateInviteToken: %v", err)
	}

	userID, ownerID, _, err := repo.GetInviteToken(context.Background(), token)
	if err != nil {
		t.Fatalf("GetInviteToken: %v", err)
	}
	if userID != teamStaffUserID {
		t.Errorf("expected user_id %d, got %d", teamStaffUserID, userID)
	}
	if ownerID != testutil.OwnerID {
		t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, ownerID)
	}
}

func TestTeam_GetBusinessByID_ReturnsName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	name, ownerID, err := repo.GetBusinessByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetBusinessByID: %v", err)
	}
	if name == "" {
		t.Error("expected non-empty business name")
	}
	if ownerID != testutil.OwnerID {
		t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, ownerID)
	}
}

func TestTeam_MarkInviteTokenUsed_IsSingleUse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	token := "single-use-invite-token"
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM invite_tokens WHERE token = ?", token)
	})
	if err := repo.CreateInviteToken(ctx, teamStaffUserID, testutil.OwnerID, token, time.Now().Add(1*time.Hour)); err != nil {
		t.Fatalf("CreateInviteToken: %v", err)
	}

	if err := repo.MarkInviteTokenUsed(ctx, token); err != nil {
		t.Fatalf("first MarkInviteTokenUsed: %v", err)
	}
	if err := repo.MarkInviteTokenUsed(ctx, token); !errors.Is(err, domain.ErrTokenAlreadyUsed) {
		t.Fatalf("second MarkInviteTokenUsed error = %v, want ErrTokenAlreadyUsed", err)
	}
}

func TestTeam_MarkInviteTokenUsed_RejectsExpired(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	token := "expired-invite-token"
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM invite_tokens WHERE token = ?", token)
	})
	if err := repo.CreateInviteToken(ctx, teamStaffUserID, testutil.OwnerID, token, time.Now().Add(-1*time.Hour)); err != nil {
		t.Fatalf("CreateInviteToken: %v", err)
	}

	if err := repo.MarkInviteTokenUsed(ctx, token); !errors.Is(err, domain.ErrTokenAlreadyUsed) {
		t.Fatalf("MarkInviteTokenUsed on expired token error = %v, want ErrTokenAlreadyUsed", err)
	}
}
