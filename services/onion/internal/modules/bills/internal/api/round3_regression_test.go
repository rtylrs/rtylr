package api

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	financebillv1 "github.com/voxire/voxie/gen/go/finance/bill/v1"
)

func TestRound3MergeBillUpdatePreservesOmittedAmount(t *testing.T) {
	current := domain.Bill{
		Amount:      12345,
		Currency:    "USD",
		DueDate:     time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
		Status:      "pending",
		Description: "original",
	}
	got, err := mergeBillUpdateInput(current, &financebillv1.UpdateBillBody{
		Status:      "overdue",
		Description: "updated",
	})
	if err != nil {
		t.Fatalf("merge update: %v", err)
	}
	if got.Amount != current.Amount || got.Currency != current.Currency || got.DueDate != "2026-10-20" {
		t.Fatalf("omitted fields changed: %+v", got)
	}
	if got.Status != "overdue" || got.Description != "updated" {
		t.Fatalf("present fields not applied: %+v", got)
	}
}
