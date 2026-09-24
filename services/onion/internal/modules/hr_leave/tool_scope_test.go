package hr_leave

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/database/dbclient"
)

func TestHRLeaveToolOwnerIDDoesNotTreatUserIDAsBusiness(t *testing.T) {
	restore := dbclient.SetPoolForTesting(nil)
	defer restore()

	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, &authPkg.AuthUser{UserID: 77, BusinessID: 5})
	businessID, err := hrLeaveToolOwnerID(ctx)
	if err == nil {
		t.Fatalf("unverifiable tool scope returned business %d without an error", businessID)
	}
	if businessID == 77 {
		t.Fatal("tool owner scope returned the platform user id as a business id")
	}
}
