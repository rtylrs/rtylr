package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestInviteUserRemovesNewUserWhenAuthorizationFails(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := newTestService(ctx, db)
	email := fmt.Sprintf("failed-invite+%d@example.com", time.Now().UnixNano())

	t.Cleanup(func() {
		var userID int64
		err := db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = ?", email).Scan(&userID)
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		if err != nil {
			t.Errorf("find failed invite user during cleanup: %v", err)
			return
		}
		if err := service.repo.DeleteInvitedUser(ctx, testutil.BusinessID, userID); err != nil {
			t.Errorf("clean up failed invite user: %v", err)
		}
	})

	_, _, err := service.InviteUser(ctx, testutil.OwnerID, testutil.BusinessID, domain.InviteUserInput{
		Email:          email,
		LicensedAppIDs: []string{"not-an-app"},
	})
	if err == nil {
		t.Fatal("InviteUser succeeded with an invalid app")
	}

	var userID int64
	err = db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("failed invite user lookup error = %v, want sql.ErrNoRows", err)
	}
}
