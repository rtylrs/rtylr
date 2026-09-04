package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Nothing ever set users.business_id for an invited member, so their token was
// issued without a business: /users/me reported business_id "0", every Flow
// endpoint 403d, and the auth app returned them to the sign-in screen (#1074).
//
// Two behaviours matter and both are covered: an unset business is filled, and
// an existing one is left alone. The second is not incidental. A person can
// belong to more than one business and business_id is the one their session
// resolves to, so inviting somebody who already has one must not move them.
func TestAttachUserToBusinessFillsOnlyAnUnsetBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	const (
		userID          int64 = 990621
		invitedBusiness int64 = 990622
		otherBusiness   int64 = 990623
	)

	if _, err := db.ExecContext(ctx,
		"INSERT INTO users (id, email, password_hash, full_name, business_id) VALUES (?, ?, ?, ?, NULL)",
		userID, "attach-business-probe@test.local", "fixture-auth-disabled", "Probe Person",
	); err != nil {
		t.Skipf("cannot seed probe user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	})

	readBusiness := func() *int64 {
		t.Helper()
		var businessID *int64
		if err := db.QueryRowContext(ctx, "SELECT business_id FROM users WHERE id = ?", userID).Scan(&businessID); err != nil {
			t.Fatalf("read users.business_id: %v", err)
		}
		return businessID
	}

	if got := readBusiness(); got != nil {
		t.Fatalf("probe precondition failed: business_id is %d, want NULL", *got)
	}

	attached, err := repo.AttachUserToBusiness(ctx, userID, invitedBusiness)
	if err != nil {
		t.Fatalf("AttachUserToBusiness: %v", err)
	}
	if !attached {
		t.Fatal("reported no change for a user whose business was unset")
	}
	if got := readBusiness(); got == nil || *got != invitedBusiness {
		t.Fatalf("business_id = %v, want %d", got, invitedBusiness)
	}

	// A second invite, from a different business, must not move them.
	attached, err = repo.AttachUserToBusiness(ctx, userID, otherBusiness)
	if err != nil {
		t.Fatalf("AttachUserToBusiness on an attached user: %v", err)
	}
	if attached {
		t.Error("reported a change for a user who already belongs to a business")
	}
	if got := readBusiness(); got == nil || *got != invitedBusiness {
		t.Errorf("business_id = %v, want it left at %d", got, invitedBusiness)
	}

	// A zero business is treated the same as NULL, since that is what a row
	// written before this existed looks like.
	if _, err := db.ExecContext(ctx, "UPDATE users SET business_id = 0 WHERE id = ?", userID); err != nil {
		t.Fatalf("reset business_id to zero: %v", err)
	}
	attached, err = repo.AttachUserToBusiness(ctx, userID, otherBusiness)
	if err != nil {
		t.Fatalf("AttachUserToBusiness on a zero business: %v", err)
	}
	if !attached {
		t.Error("reported no change for a user whose business was zero")
	}
	if got := readBusiness(); got == nil || *got != otherBusiness {
		t.Errorf("business_id = %v, want %d", got, otherBusiness)
	}
}
