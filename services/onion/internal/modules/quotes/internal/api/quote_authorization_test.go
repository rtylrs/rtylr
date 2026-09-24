package api

import (
	"context"
	"net/http"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	financequotev1 "github.com/voxire/voxie/gen/go/finance/quote/v1"
)

func quoteUser(permissions ...string) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     1,
		BusinessID: 2,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID:       authorization.AppFinance,
			Licensed:    true,
			HasAccess:   true,
			Permissions: permissions,
		}}},
	})
}

func TestQuoteAPI_RequiresExactFeaturePermission(t *testing.T) {
	api := &QuoteAPI{}
	viewer := quoteUser("finance:quotes:view")

	if _, err := api.CreateQuote(viewer, &financequotev1.CreateQuoteRequest{}); quoteValidationStatus(t, err) != http.StatusForbidden {
		t.Fatalf("viewer create status = %d, want 403", quoteValidationStatus(t, err))
	}
	if _, err := api.ConvertQuote(viewer, &financequotev1.ConvertQuoteRequest{Id: 1}); quoteValidationStatus(t, err) != http.StatusForbidden {
		t.Fatalf("viewer convert status = %d, want 403", quoteValidationStatus(t, err))
	}

	manager := quoteUser("finance:quotes:manage")
	if _, err := api.ListQuotes(manager, &financequotev1.ListQuotesRequest{}); quoteValidationStatus(t, err) != http.StatusForbidden {
		t.Fatalf("manager-without-view list status = %d, want 403", quoteValidationStatus(t, err))
	}
}
