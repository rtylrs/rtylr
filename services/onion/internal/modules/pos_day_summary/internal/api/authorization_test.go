package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	possummaryv1 "github.com/voxire/voxie/gen/go/pos/day_summary/v1"
)

func daySummaryStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestDaySummaryRequiresReportsPermissionBeforeServiceAccess(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:     authorization.AppPOS,
				Licensed:  true,
				HasAccess: true,
			}},
		},
	})
	_, err := (&DaySummaryAPI{}).DaySummary(ctx, &possummaryv1.DaySummaryRequest{Date: "2026-08-27"})
	if daySummaryStatus(err) != http.StatusForbidden {
		t.Fatalf("returned %v, want 403 before service access", err)
	}
}
