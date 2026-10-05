package api

import (
	"context"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/daysummary/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/daysummary/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	possummaryv1 "github.com/voxire/voxie/gen/go/pos/day_summary/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "pos reports view", App: authorization.AppPOS, Permission: "pos:reports:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewDaySummaryAPI(application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))).DaySummary(ctx, &possummaryv1.DaySummaryRequest{Date: "2026-08-27"})
			return response, err
		}},
	})
}
