package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/billing/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetSummary_ReturnsNonZeroValue(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary.Plan == "" {
		t.Error("expected non-empty Plan")
	}
}

func TestGetSummary_PlanIsFreeTier(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary.Plan != "free" {
		t.Errorf("Plan = %q, want %q", summary.Plan, "free")
	}
}

func TestGetSummary_StatusIsActive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary.Status != "active" {
		t.Errorf("Status = %q, want %q", summary.Status, "active")
	}
}

func TestGetSummary_UsageLimitPositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary.UsageLimit <= 0 {
		t.Errorf("UsageLimit = %f, expected > 0", summary.UsageLimit)
	}
}

func TestGetSummary_CurrentUsageNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary.CurrentUsage < 0 {
		t.Errorf("CurrentUsage = %f, expected >= 0", summary.CurrentUsage)
	}
}

func TestGetSummary_WrongOwnerNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Stub implementation returns same result regardless of owner
	summary, err := repo.GetSummary(ctx, 999999)
	if err != nil {
		t.Fatalf("GetSummary wrong owner: %v", err)
	}
	if summary.Plan == "" {
		t.Error("expected non-empty Plan even for unknown owner (stub)")
	}
}

func TestListInvoices_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invoices, err := repo.ListInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListInvoices: %v", err)
	}
	if invoices == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestListInvoices_WrongOwnerNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invoices, err := repo.ListInvoices(ctx, 999999)
	if err != nil {
		t.Fatalf("ListInvoices wrong owner: %v", err)
	}
	if invoices == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestGetSummary_CurrentUsageBelowLimit(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary.CurrentUsage > summary.UsageLimit {
		t.Errorf("CurrentUsage (%f) exceeds UsageLimit (%f)", summary.CurrentUsage, summary.UsageLimit)
	}
}

func TestGetSummary_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary first: %v", err)
	}
	second, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary second: %v", err)
	}
	if first.Plan != second.Plan {
		t.Errorf("inconsistent Plan: %q vs %q", first.Plan, second.Plan)
	}
	if first.UsageLimit != second.UsageLimit {
		t.Errorf("inconsistent UsageLimit: %f vs %f", first.UsageLimit, second.UsageLimit)
	}
}

func TestListInvoices_IsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Stub always returns empty list
	invoices, err := repo.ListInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListInvoices: %v", err)
	}
	if len(invoices) != 0 {
		t.Errorf("expected 0 invoices from stub, got %d", len(invoices))
	}
}
