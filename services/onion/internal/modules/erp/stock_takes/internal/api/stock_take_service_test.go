package api

import (
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/domain"
)

func TestStockTakeMutationErrorMapsDomainFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "missing stock take", err: domain.ErrStockTakeNotFound, want: http.StatusNotFound},
		{name: "missing reference", err: domain.ErrStockTakeReferenceNotFound, want: http.StatusNotFound},
		{name: "incomplete count", err: domain.ErrStockTakeIncomplete, want: http.StatusConflict},
		{name: "already completed", err: domain.ErrStockTakeAlreadyCompleted, want: http.StatusConflict},
		{name: "insufficient stock", err: stock.ErrInsufficientStock, want: http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stockTakeMutationError(tt.err); got == nil {
				t.Fatal("expected mapped error")
			}
			var oneKitErr *onekithttp.Error
			if !errors.As(stockTakeMutationError(tt.err), &oneKitErr) {
				t.Fatalf("expected oneKit HTTP error, got %T", stockTakeMutationError(tt.err))
			}
			if got := int(oneKitErr.GetStatusCode()); got != tt.want {
				t.Fatalf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
