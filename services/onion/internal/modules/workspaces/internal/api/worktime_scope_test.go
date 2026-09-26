package api_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/workspaces/internal/api"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	flowtimev1 "github.com/voxire/voxie/gen/go/flow/time/v1"
)

// These are the deny cases only, and a nil Service is enough for them: scope()
// rejects before it ever reaches the Service, so a panic here would itself be
// the failure.
//
// The allow case is deliberately not here. scope() goes on to call
// tenant.ResolveBusinessID, which fails closed on a nil pool, so a permitted
// caller and a rejected one both come back as PermissionDenied and no assertion
// through this API can tell them apart. A test that expected the nil Service to
// panic on the allow path was committed alongside scope() and never once passed.
// The decision it meant to check is callerMayActFor, covered directly in
// worktime_scope_internal_test.go. Do not re-add it here.

func withCaller(userID int64) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     userID,
		BusinessID: 1,
	})
}

func int64Ptr(v int64) *int64 { return &v }

func TestWorkTimeAPI_GetWorkDay_DeniesEmployeeIDOverride(t *testing.T) {
	workTimeAPI := api.NewWorkTimeAPI(nil)
	_, err := workTimeAPI.GetWorkDay(withCaller(1001), &flowtimev1.GetWorkDayRequest{
		Date:       "2026-08-16",
		EmployeeId: int64Ptr(9999), // someone else
	})

	var httpErr *onekithttp.Error
	if !errors.As(err, &httpErr) || httpErr.HTTPStatusCode() != http.StatusForbidden {
		t.Fatalf("want a 403 permission-denied error, got %v", err)
	}
}

func TestWorkTimeAPI_GetWorkDay_DeniesUnauthenticatedCaller(t *testing.T) {
	workTimeAPI := api.NewWorkTimeAPI(nil)
	_, err := workTimeAPI.GetWorkDay(context.Background(), &flowtimev1.GetWorkDayRequest{
		Date: "2026-08-16",
	})

	var httpErr *onekithttp.Error
	if !errors.As(err, &httpErr) || httpErr.HTTPStatusCode() != http.StatusForbidden {
		t.Fatalf("want a 403 permission-denied error, got %v", err)
	}
}
