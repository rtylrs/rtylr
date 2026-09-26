package api

import (
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/stock_core"
)

func TestPurchaseOrderMutationErrorMapsDomainFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "missing purchase order", err: domain.ErrPurchaseOrderNotFound, want: http.StatusNotFound},
		{name: "missing reference", err: domain.ErrPurchaseOrderReferenceNotFound, want: http.StatusNotFound},
		{name: "insufficient stock", err: stock_core.ErrInsufficientStock, want: http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oneKitErr *onekithttp.Error
			mapped := purchaseOrderMutationError(tt.err, "fallback")
			if !errors.As(mapped, &oneKitErr) {
				t.Fatalf("expected oneKit HTTP error, got %T: %v", mapped, mapped)
			}
			if got := int(oneKitErr.GetStatusCode()); got != tt.want {
				t.Fatalf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
