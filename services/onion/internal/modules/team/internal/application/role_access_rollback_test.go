package application

import (
	"context"
	"os"
	"testing"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestRoleAccessRollbackRestoresLegacyAuthorization(t *testing.T) {
	if os.Getenv("ROLE_ACCESS_V2_ROLLBACK_CHECK") != "1" {
		t.Skip("role access rollback check runs through scripts/test-role-access-control-v2-rollback.sh")
	}

	db := testutil.OpenDB(t)
	service := NewService(context.Background(), db)
	access, err := service.GetEffectiveAccess(context.Background(), 1001, 1002)
	if err != nil {
		t.Fatalf("GetEffectiveAccess: %v", err)
	}
	if access == nil {
		t.Fatal("GetEffectiveAccess returned nil access")
	}
	if len(access.Apps) != 1 || access.Apps[0].AppID != authorization.AppFlow || !access.Apps[0].Licensed || !access.Apps[0].HasAccess {
		t.Fatalf("legacy access after rollback = %+v, want active Flow app access", access)
	}
}
