package api

import (
	"context"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/crm/campaigns/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/crm/campaigns/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "crm campaigns view", App: authorization.AppCRM, Permission: "crm:campaigns:view", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListCampaignsHandler{Service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))}).Serve(ctx, &ListCampaignsRequest{})
			return response, err
		}},
		{Name: "crm campaigns create", App: authorization.AppCRM, Permission: "crm:campaigns:create", Call: func(ctx context.Context) (any, error) {
			response, err := (&CreateCampaignHandler{Service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))}).Serve(ctx, &CreateCampaignRequest{Name: "authorization probe", Type: "email", Subject: "authorization probe", Body: "authorization probe"})
			return response, err
		}},
		{Name: "crm campaigns send", App: authorization.AppCRM, Permission: "crm:campaigns:send", Call: func(ctx context.Context) (any, error) {
			rctx := map[string]string{}
			rctx["id"] = ("7006")
			response, err := (&UpdateCampaignStatusHandler{Service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))}).Serve(serverHttp.WithPathParams(ctx, rctx), &UpdateCampaignStatusRequest{Status: "sent"})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
