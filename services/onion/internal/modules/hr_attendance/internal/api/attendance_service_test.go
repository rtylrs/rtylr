package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

type employeeResolverStub struct {
	ownerID    int64
	userID     int64
	employeeID int64
}

func (s *employeeResolverStub) GetEmployeeIDByUserID(_ context.Context, ownerID, userID int64) (int64, error) {
	s.ownerID = ownerID
	s.userID = userID
	return s.employeeID, nil
}

func TestResolveIdentityFailsClosedWithoutTenantMapping(t *testing.T) {
	resolver := &employeeResolverStub{employeeID: 42}
	api := NewAttendanceAPI(nil, resolver)
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{BusinessID: 9, UserID: 7})

	_, _, err := api.resolveIdentity(ctx)
	if oneKitStatus(err) != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", oneKitStatus(err), http.StatusForbidden)
	}
	if resolver.ownerID != 0 || resolver.userID != 0 {
		t.Fatalf("resolver should not run without tenant mapping, got owner %d, user %d", resolver.ownerID, resolver.userID)
	}
}

func TestResolveIdentityDeniesAnonymousRequest(t *testing.T) {
	api := NewAttendanceAPI(nil, &employeeResolverStub{employeeID: 42})

	_, _, err := api.resolveIdentity(context.Background())
	if err == nil {
		t.Fatal("expected anonymous request to be denied")
	}
	var httpErr *onekithttp.Error
	if !errors.As(err, &httpErr) || httpErr.HTTPStatusCode() != http.StatusForbidden {
		t.Fatalf("want a 403 permission-denied error, got %v", err)
	}
}
