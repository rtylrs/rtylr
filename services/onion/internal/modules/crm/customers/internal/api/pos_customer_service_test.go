package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	poscustomerv1 "github.com/voxire/voxie/gen/go/pos/customer/v1"
)

type posCustomerScopeResolver struct {
	businessID int64
}

func (r posCustomerScopeResolver) ActiveBusinessOwner(context.Context, int64) (int64, error) {
	return r.businessID, nil
}
func (r posCustomerScopeResolver) UserBusiness(context.Context, int64, int64) (int64, error) {
	return r.businessID, nil
}

func posCustomerErrorStatus(err error) int {
	var oneKitErr *onekiterrors.Error
	if errors.As(err, &oneKitErr) {
		return oneKitErr.HTTPStatusCode()
	}
	return 0
}

func TestSalesContactToProtoIncludesLoyaltyPoints(t *testing.T) {
	proto := salesContactToProto(domain.SalesContact{
		ID:            42,
		FirstName:     "Amina",
		LastName:      "Haddad",
		Email:         "amina@example.com",
		Phone:         "+9613123456",
		LoyaltyPoints: 180,
	})

	if proto.Id != 42 || proto.FirstName != "Amina" || proto.LastName != "Haddad" {
		t.Fatalf("unexpected customer identity fields: %+v", proto)
	}
	if proto.LoyaltyPoints == nil || *proto.LoyaltyPoints != 180 {
		t.Fatalf("expected LoyaltyPoints=180, got %v", proto.LoyaltyPoints)
	}
}

func TestCreateCustomerValidatesInputAndAuth(t *testing.T) {
	t.Cleanup(tenant.SetScopeResolver(posCustomerScopeResolver{businessID: 9}))
	api := NewPOSCustomerAPI(nil)

	if _, err := api.CreateCustomer(context.Background(), &poscustomerv1.CreateCustomerRequest{FirstName: "Nadia"}); posCustomerErrorStatus(err) != http.StatusForbidden {
		t.Fatalf("expected 403 for unauthenticated request, got %v", err)
	}

	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: 7, BusinessID: 9})
	if _, err := api.CreateCustomer(ctx, &poscustomerv1.CreateCustomerRequest{FirstName: "   "}); posCustomerErrorStatus(err) != http.StatusBadRequest {
		t.Fatalf("expected 400 for blank first name, got %v", err)
	}

	badEmail := "not-an-email"
	if _, err := api.CreateCustomer(ctx, &poscustomerv1.CreateCustomerRequest{FirstName: "Nadia", Email: &badEmail}); posCustomerErrorStatus(err) != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid email, got %v", err)
	}
}
