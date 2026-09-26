package api

import (
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
	"github.com/voxire/onion/internal/platform/stock_core"
)

func TestStockTransferMutationErrorMapsDomainFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "missing transfer", err: domain.ErrStockTransferNotFound, want: http.StatusNotFound},
		{name: "missing reference", err: domain.ErrStockTransferReferenceNotFound, want: http.StatusNotFound},
		{name: "insufficient stock", err: stock_core.ErrInsufficientStock, want: http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusCode(t, stockTransferMutationError(tt.err)); got != tt.want {
				t.Fatalf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

func statusCode(t *testing.T, err error) int {
	t.Helper()
	var oneKitErr *onekithttp.Error
	if !errors.As(err, &oneKitErr) {
		t.Fatalf("expected oneKit HTTP error, got %T: %v", err, err)
	}
	return int(oneKitErr.GetStatusCode())
}
