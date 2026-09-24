package api

import (
	"errors"
	"net/http"
	"testing"

	onkithttp "github.com/corezio/onekit/http"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
)

func TestSplitOrderRejectsNonPartitioningGroups(t *testing.T) {
	order := &domain.Order{Items: []domain.OrderItem{
		{ID: 1, Status: domain.ItemStatusPending},
		{ID: 2, Status: domain.ItemStatusPreparing},
		{ID: 3, Status: domain.ItemStatusPending},
		{ID: 4, Status: domain.ItemStatusCancelled},
	}}
	cases := []struct {
		name   string
		splits [][]int64
		valid  bool
	}{
		{name: "exact partition", splits: [][]int64{{1}, {2, 3}}, valid: true},
		{name: "omitted item", splits: [][]int64{{1}, {2}}},
		{name: "duplicate item", splits: [][]int64{{1, 2}, {2, 3}}},
		{name: "foreign item", splits: [][]int64{{1, 2}, {3, 99}}},
		{name: "empty group", splits: [][]int64{{1}, {}}},
		{name: "cancelled item", splits: [][]int64{{1, 2, 3}, {4}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isValidOrderItemPartition(order, tc.splits); got != tc.valid {
				t.Fatalf("isValidOrderItemPartition() = %v, want %v", got, tc.valid)
			}
		})
	}
}

func TestPaddlePaymentErrorMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int32
	}{
		{name: "not found", err: domain.ErrOrderNotFound, want: http.StatusNotFound},
		{name: "already paid legacy error", err: errors.New("order already paid"), want: http.StatusConflict},
		{name: "zero total", err: errors.New("order total must be positive"), want: http.StatusBadRequest},
		{name: "internal", err: errors.New("provider unavailable"), want: http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got *onkithttp.Error
			if !errors.As(paddlePaymentError(tc.err), &got) {
				t.Fatal("paddlePaymentError() did not return a onekit HTTP error")
			}
			if got.GetStatusCode() != tc.want {
				t.Fatalf("status = %d, want %d", got.GetStatusCode(), tc.want)
			}
		})
	}
}
