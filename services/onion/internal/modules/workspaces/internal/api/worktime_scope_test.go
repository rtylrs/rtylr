package api_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/api"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	flowtimev1 "github.com/voxire/voxie/gen/go/flow/time/v1"
)

// scope() denies before touching the Service, so a nil Service is enough to
// prove the permission check runs first - if it did not, this would panic
// instead of returning a permission error, which fails the test either way.

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

func TestWorkTimeAPI_GetWorkDay_AllowsOwnEmployeeID(t *testing.T) {
	workTimeAPI := api.NewWorkTimeAPI(nil)

	// Passes scope() and reaches the (nil) Service, which is expected to panic -
	// recovered here so the test only asserts scope() let it through rather than
	// denying a caller who is explicit about acting on their own behalf.
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("want a panic from the nil Service, meaning scope() let the self id through")
		}
	}()

	_, err := workTimeAPI.GetWorkDay(withCaller(1001), &flowtimev1.GetWorkDayRequest{
		Date:       "2026-08-16",
		EmployeeId: int64Ptr(1001), // self, explicit
	})
	if err != nil {
		t.Fatalf("want scope() to pass and panic in Service, got a returned error instead: %v", err)
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
