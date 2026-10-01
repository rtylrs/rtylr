package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestInviteDoesNotAssignPermissionProfilesByDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.EnsureBusiness(t, db, testutil.BusinessID, testutil.OwnerID)
	ctx := context.Background()
	service := application.NewService(ctx, db)
	repo := infrastructure.NewRepository(db)
	email := fmt.Sprintf("empty-profiles-invite-%d@test.local", time.Now().UnixNano())

	userID, _, err := service.InviteUser(ctx, testutil.OwnerID, testutil.BusinessID, domain.InviteUserInput{
		Email:          email,
		LicensedAppIDs: []string{"hr", "pos"},
	})
	if err != nil {
		t.Fatalf("InviteUser: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.DeleteInvitedUser(ctx, testutil.BusinessID, testutil.OwnerID, userID); err != nil {
			t.Errorf("clean up invited member: %v", err)
		}
	})

	var profiles, licenses int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM authorization_role_assignments
		WHERE business_id = ? AND user_id = ?`, testutil.BusinessID, userID).Scan(&profiles); err != nil {
		t.Fatalf("count assigned profiles: %v", err)
	}
	if profiles != 0 {
		t.Fatalf("invite created %d implicit permission profiles, want none", profiles)
	}
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM authorization_license_assignments
		WHERE business_id = ? AND user_id = ? AND status = 'pending'`, testutil.BusinessID, userID).Scan(&licenses); err != nil {
		t.Fatalf("count licensed apps: %v", err)
	}
	if licenses != 2 {
		t.Fatalf("pending app licenses = %d, want 2", licenses)
	}
}
