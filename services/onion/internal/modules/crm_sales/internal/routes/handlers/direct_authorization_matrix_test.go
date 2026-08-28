package handlers

import (
	"context"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "crm customers view", App: authorization.AppCRM, Permission: "crm:customers:view", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListContactsHandler{Service: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &ListContactsRequest{})
			return response, err
		}},
		{Name: "crm customers create", App: authorization.AppCRM, Permission: "crm:customers:create", Call: func(ctx context.Context) (any, error) {
			response, err := (&CreateContactHandler{Service: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &CreateContactRequest{LeadID: 7101, Name: "Authorization Probe"})
			return response, err
		}},
		{Name: "crm customers edit", App: authorization.AppCRM, Permission: "crm:customers:edit", Call: func(ctx context.Context) (any, error) {
			response, err := (&UpdateContactHandler{Service: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(withContactID(ctx), &UpdateContactRequest{Name: "Authorization Probe"})
			return response, err
		}},
		{Name: "crm customers delete", App: authorization.AppCRM, Permission: "crm:customers:delete", Call: func(ctx context.Context) (any, error) {
			response, err := (&DeleteContactHandler{Service: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(withContactID(ctx), &DeleteContactRequest{})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}

func withContactID(ctx context.Context) context.Context {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "7102")
	return context.WithValue(ctx, chi.RouteCtxKey, routeCtx)
}
