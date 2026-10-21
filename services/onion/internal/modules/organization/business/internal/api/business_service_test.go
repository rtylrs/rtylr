package api

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/business/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	hubbusinessv1 "github.com/voxire/voxie/gen/go/hub/business/v1"
)

func deviceCtx(userID, businessID int64) context.Context {
	return (&authPkg.Device{
		User:        &authPkg.AuthUser{UserID: userID, BusinessID: businessID},
		Language:    "en",
		IsAnonymous: false,
	}).Ctx(context.Background())
}

func TestGetBusiness_DeniesExplicitForeignBusinessID(t *testing.T) {
	api := NewBusinessAPI(&application.Service{})
	ctx := deviceCtx(7, 1)

	_, err := api.GetBusiness(ctx, &hubbusinessv1.GetBusinessRequest{BusinessId: 999})
	if err == nil {
		t.Fatal("expected denial for a business the caller has no access to")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "permission") {
		t.Fatalf("expected permission error, got: %v", err)
	}
}

func TestGetBusiness_DeniesUnauthenticated(t *testing.T) {
	api := NewBusinessAPI(&application.Service{})

	_, err := api.GetBusiness(context.Background(), &hubbusinessv1.GetBusinessRequest{BusinessId: 1})
	if err == nil {
		t.Fatal("expected denial without authentication")
	}
}
