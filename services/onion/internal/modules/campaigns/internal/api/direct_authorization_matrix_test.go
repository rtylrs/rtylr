package api

import (
	"context"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/campaigns/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "crm campaigns view", App: authorization.AppCRM, Permission: "crm:campaigns:view", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListCampaignsHandler{Service: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &ListCampaignsRequest{})
			return response, err
		}},
		{Name: "crm campaigns create", App: authorization.AppCRM, Permission: "crm:campaigns:create", Call: func(ctx context.Context) (any, error) {
			response, err := (&CreateCampaignHandler{Service: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &CreateCampaignRequest{Name: "authorization probe", Type: "email", Subject: "authorization probe", Body: "authorization probe"})
			return response, err
		}},
		{Name: "crm campaigns send", App: authorization.AppCRM, Permission: "crm:campaigns:send", Call: func(ctx context.Context) (any, error) {
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "7006")
			response, err := (&UpdateCampaignStatusHandler{Service: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(context.WithValue(ctx, chi.RouteCtxKey, rctx), &UpdateCampaignStatusRequest{Status: "sent"})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
