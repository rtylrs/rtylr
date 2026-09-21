package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/quotes/internal/application"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	financequotev1 "github.com/voxire/voxie/gen/go/finance/quote/v1"
)

// CreateQuote is the live proto path (module.go registers only this service),
// so malformed values must be rejected here instead of being silently coerced
// by the repository (a bad valid_until used to be stored as NULL and a
// mismatched total was persisted verbatim).
func TestQuoteAPI_CreateQuote_ValidatesInput(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.EnsureBusiness(t, db, testutil.BusinessID, testutil.OwnerID)

	service := application.NewService(context.Background(), db)
	api := NewQuoteAPI(service)
	ctx := auth.WithAuthUser(context.Background(), &auth.AuthUser{
		UserID:     testutil.OwnerID,
		BusinessID: testutil.BusinessID,
	})

	for _, tc := range []struct {
		name    string
		request *financequotev1.CreateQuoteRequest
	}{
		{
			name: "malformed valid_until",
			request: &financequotev1.CreateQuoteRequest{
				CustomerName: "Malformed Date Co",
				Subtotal:     100,
				Total:        100,
				ValidUntil:   "09/21/2026",
			},
		},
		{
			name: "inconsistent total",
			request: &financequotev1.CreateQuoteRequest{
				CustomerName: "Inconsistent Total Co",
				Subtotal:     100,
				Total:        90,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := api.CreateQuote(ctx, tc.request)
			if got := quoteValidationStatus(t, err); got != http.StatusBadRequest {
				t.Fatalf("CreateQuote(%s) = %d, want 400 (err=%v)", tc.name, got, err)
			}
		})
	}

	// A well-formed request must still be accepted and keep the requested date;
	// otherwise the rejection above would prove nothing about validity.
	resp, err := api.CreateQuote(ctx, &financequotev1.CreateQuoteRequest{
		CustomerName: "Valid Date Co",
		Subtotal:     100,
		Total:        100,
		ValidUntil:   "2026-09-21",
	})
	if err != nil {
		t.Fatalf("CreateQuote(valid input): %v", err)
	}
	if got := resp.GetQuote().GetValidUntil(); got != "2026-09-21" {
		t.Fatalf("valid_until = %q, want 2026-09-21", got)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_quotes WHERE id = ?", resp.GetQuote().GetId())
	})
}
