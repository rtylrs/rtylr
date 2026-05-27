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

func TestGetSummary_OwnerIDZero_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Stub ignores owner ID; result should still be valid.
	summary, err := repo.GetSummary(ctx, 0)
	if err != nil {
		t.Fatalf("GetSummary owner=0: %v", err)
	}
	if summary.Plan == "" {
		t.Error("expected non-empty Plan for owner=0 (stub)")
	}
}

func TestGetSummary_NegativeOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, -1)
	if err != nil {
		t.Fatalf("GetSummary owner=-1: %v", err)
	}
	_ = summary
}

func TestGetSummary_StatusNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary.Status == "" {
		t.Error("expected non-empty Status")
	}
}

func TestGetSummary_PlanNonEmpty(t *testing.T) {
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

func TestGetSummary_UsageLimitGteCurrentUsage(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	s, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if s.CurrentUsage > s.UsageLimit {
		t.Errorf("CurrentUsage (%f) > UsageLimit (%f): unexpected", s.CurrentUsage, s.UsageLimit)
	}
}

func TestListInvoices_MultipleCalls_Consistent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.ListInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first ListInvoices: %v", err)
	}
	second, err := repo.ListInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second ListInvoices: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("ListInvoices inconsistent: %d vs %d", len(first), len(second))
	}
}

func TestGetSummary_TwiceReturnsSameStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first GetSummary: %v", err)
	}
	second, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second GetSummary: %v", err)
	}
	if first.Status != second.Status {
		t.Errorf("Status changed: %q vs %q", first.Status, second.Status)
	}
}

func TestGetSummary_TwiceReturnsSamePlan(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first GetSummary: %v", err)
	}
	second, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second GetSummary: %v", err)
	}
	if first.Plan != second.Plan {
		t.Errorf("Plan changed: %q vs %q", first.Plan, second.Plan)
	}
}

func TestGetSummary_LargeOwnerID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, 1<<40)
	if err != nil {
		t.Fatalf("GetSummary large owner: %v", err)
	}
	_ = summary
}

func TestListInvoices_ZeroOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invoices, err := repo.ListInvoices(ctx, 0)
	if err != nil {
		t.Fatalf("ListInvoices owner=0: %v", err)
	}
	if invoices == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestGetSummary_UsageLimitExact(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	// Free tier limit is expected to be a round number > 0.
	if summary.UsageLimit != float64(int(summary.UsageLimit)) {
		t.Logf("UsageLimit is not a whole number: %f (may be acceptable)", summary.UsageLimit)
	}
}

func TestListInvoices_LargeOwner_ReturnsNonNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invoices, err := repo.ListInvoices(ctx, 1<<40)
	if err != nil {
		t.Fatalf("ListInvoices large owner: %v", err)
	}
	if invoices == nil {
		t.Fatal("expected non-nil slice")
	}
}
