package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/giftcards/internal/application"
	"github.com/voxire/onion/internal/modules/pos/giftcards/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posgiftcardv1 "github.com/voxire/voxie/gen/go/pos/gift_card/v1"
)

type testScopeResolver struct {
	businessID int64
}

func (r testScopeResolver) ActiveBusinessOwner(context.Context, int64) (int64, error) {
	return r.businessID, nil
}
func (r testScopeResolver) UserBusiness(context.Context, int64, int64) (int64, error) {
	return r.businessID, nil
}

type fakeGiftCardRepo struct {
	lastInput domain.GiftCardInput
	createErr error
}

func (f *fakeGiftCardRepo) List(_ context.Context, _ int64) ([]domain.GiftCard, error) {
	return nil, nil
}
func (f *fakeGiftCardRepo) CreateForOwner(_ context.Context, _, _ int64, input domain.GiftCardInput) (int64, error) {
	f.lastInput = input
	if f.createErr != nil {
		return 0, f.createErr
	}
	return 77, nil
}
func (f *fakeGiftCardRepo) GetByCode(_ context.Context, _ int64, _ string) (*domain.GiftCard, error) {
	return nil, domain.ErrGiftCardNotFound
}
func (f *fakeGiftCardRepo) RedeemForOrder(_ context.Context, _, _ int64, _ domain.RedeemInput) (*domain.GiftCard, error) {
	return nil, domain.ErrGiftCardNotFound
}

func giftCardManagerCtx() context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     10,
		BusinessID: 5,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppPOS,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"pos:gift_cards:manage"},
			}},
		},
	})
}

func giftCardErrorStatus(err error) int {
	var oneKitErr *onekiterrors.Error
	if errors.As(err, &oneKitErr) {
		return oneKitErr.HTTPStatusCode()
	}
	return 0
}

func TestCreateGiftCard_NormalizesRFC3339ExpiryAndMapsDuplicateConflict(t *testing.T) {
	t.Cleanup(tenant.SetScopeResolver(testScopeResolver{businessID: 5}))
	repo := &fakeGiftCardRepo{}
	api := NewGiftCardAPI(application.NewService(context.Background(), repo))

	paidOrderID := int64(3301)
	amount := 25.0
	resp, err := api.CreateGiftCard(giftCardManagerCtx(), &posgiftcardv1.CreateGiftCardRequest{
		Code:          " GC-TEST-1 ",
		PaidOrderId:   &paidOrderID,
		InitialAmount: &amount,
		CustomerName:  " Nadia ",
		CustomerEmail: " nadia@example.com ",
		ExpiresAt:     "2027-11-27T11:30:00Z",
	})
	if err != nil {
		t.Fatalf("CreateGiftCard failed: %v", err)
	}
	if resp.GiftCardId != 77 {
		t.Fatalf("expected GiftCardId 77, got %d", resp.GiftCardId)
	}
	if repo.lastInput.ExpiresAt != "2027-11-27" {
		t.Fatalf("expected normalized YYYY-MM-DD expiry, got %q", repo.lastInput.ExpiresAt)
	}
	if repo.lastInput.Code != "GC-TEST-1" {
		t.Fatalf("expected trimmed code, got %q", repo.lastInput.Code)
	}

	repo.createErr = domain.ErrGiftCardCodeExists
	_, dupErr := api.CreateGiftCard(giftCardManagerCtx(), &posgiftcardv1.CreateGiftCardRequest{
		Code:          "GC-TEST-1",
		PaidOrderId:   &paidOrderID,
		InitialAmount: &amount,
		CustomerName:  "Nadia",
		CustomerEmail: "nadia@example.com",
		ExpiresAt:     "2027-11-27",
	})
	if giftCardErrorStatus(dupErr) != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate code, got %v (status %d)", dupErr, giftCardErrorStatus(dupErr))
	}
}
