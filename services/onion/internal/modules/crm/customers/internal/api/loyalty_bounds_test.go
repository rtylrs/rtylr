package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
)

func TestLoyaltyPointHandlersRejectPointsPastStorageBound(t *testing.T) {
	points := int(domain.MaxLoyaltyPoints + 1)
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "award",
			run: func() error {
				return (&AwardPointsHandler{}).Validate(context.Background(), &AwardPointsRequest{CustomerID: 7, Points: points})
			},
		},
		{
			name: "redeem",
			run: func() error {
				return (&RedeemPointsHandler{}).Validate(context.Background(), &RedeemPointsRequest{CustomerID: 7, Points: points})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if status := customerPermissionStatus(test.run()); status != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
			}
		})
	}
}

func TestLoyaltyTierListValidationPreservesUnboundedLegacyRequest(t *testing.T) {
	req := &ListLoyaltyTiersRequest{}
	if err := (&ListLoyaltyTiersHandler{}).Validate(context.Background(), req); err != nil {
		t.Fatalf("empty page request rejected: %v", err)
	}
	if req.Page != 0 || req.PageSize != 0 {
		t.Fatalf("empty page request was rewritten to %d/%d", req.Page, req.PageSize)
	}

	page := &ListLoyaltyTiersRequest{Page: 2}
	if err := (&ListLoyaltyTiersHandler{}).Validate(context.Background(), page); err != nil {
		t.Fatalf("partial page request rejected: %v", err)
	}
	if page.Page != 2 || page.PageSize != domain.DefaultLoyaltyTierPageSize {
		t.Fatalf("partial page request = %d/%d, want 2/%d", page.Page, page.PageSize, domain.DefaultLoyaltyTierPageSize)
	}
}

func TestLoyaltyPointHandlersAcceptInt32Boundary(t *testing.T) {
	points := int(domain.MaxLoyaltyPoints)
	if err := (&AwardPointsHandler{}).Validate(context.Background(), &AwardPointsRequest{CustomerID: 7, Points: points}); err != nil {
		t.Fatalf("award boundary rejected: %v", err)
	}
	if err := (&RedeemPointsHandler{}).Validate(context.Background(), &RedeemPointsRequest{CustomerID: 7, Points: points}); err != nil {
		t.Fatalf("redeem boundary rejected: %v", err)
	}
}
