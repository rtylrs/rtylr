package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func TestGetAttendanceRequiresEmployeeViewPermission(t *testing.T) {
	h := &GetAttendanceHandler{}
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:     authorization.AppHR,
				Licensed:  true,
				HasAccess: true,
			}},
		},
	})

	_, err := h.Serve(ctx, &GetAttendanceRequest{OwnerID: 9, EmployeeID: 7})
	var oneKitErr *onekithttp.Error
	if !errors.As(err, &oneKitErr) || oneKitErr.GetStatusCode() != http.StatusForbidden {
		t.Fatalf("want a 403 permission-denied error, got %v", err)
	}
}

func TestGetAttendanceRejectsBusinessOverride(t *testing.T) {
	h := &GetAttendanceHandler{}
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHR,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"hr:employees:view"},
			}},
		},
	})

	_, err := h.Serve(ctx, &GetAttendanceRequest{OwnerID: 10, EmployeeID: 7})
	var oneKitErr *onekithttp.Error
	if !errors.As(err, &oneKitErr) || oneKitErr.GetStatusCode() != http.StatusForbidden {
		t.Fatalf("want a 403 permission-denied error, got %v", err)
	}
}
