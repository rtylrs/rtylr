package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListAddonGrantsClampsNegativeOffset(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	note := fmt.Sprintf("negative-offset-regression-%d", time.Now().UnixNano())

	result, err := db.ExecContext(ctx, `
		INSERT INTO manual_addon_grants
			(user_id, business_id, addon_id, granted_by, note, enabled_by_grant)
		VALUES (?, ?, 'menu', ?, ?, FALSE)`,
		testutil.OwnerID, testutil.BusinessID, testutil.OwnerID, note,
	)
	if err != nil {
		t.Fatalf("seed addon grant: %v", err)
	}
	grantID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("seeded grant id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM manual_addon_grants WHERE id = ?`, grantID)
	})

	grants, _, err := repo.ListAddonGrants(ctx, 200, -1)
	if err != nil {
		t.Fatalf("ListAddonGrants with negative offset: %v", err)
	}
	for _, grant := range grants {
		if grant.ID == grantID {
			return
		}
	}
	t.Fatalf("seeded grant %d was not returned when offset was clamped to zero", grantID)
}
