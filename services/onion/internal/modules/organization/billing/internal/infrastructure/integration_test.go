package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/billing/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetSummary_ReturnsNonZeroValue(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.BusinessID)
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

	summary, err := repo.GetSummary(ctx, testutil.BusinessID)
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

	summary, err := repo.GetSummary(ctx, testutil.BusinessID)
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

	summary, err := repo.GetSummary(ctx, testutil.BusinessID)
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

	summary, err := repo.GetSummary(ctx, testutil.BusinessID)
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

	// An unknown business with no subscription is treated as the seeded free plan.
	summary, err := repo.GetSummary(ctx, 999999)
	if err != nil {
		t.Fatalf("GetSummary wrong business: %v", err)
	}
	if summary.Plan == "" {
		t.Error("expected non-empty Plan for the seeded free plan")
	}
}

func TestListInvoices_ReturnsUnavailable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.ListInvoices(ctx, testutil.BusinessID); !errors.Is(err, infrastructure.ErrBillingInvoicesUnavailable) {
		t.Fatalf("ListInvoices error = %v, want ErrBillingInvoicesUnavailable", err)
	}
}

func TestListInvoices_WrongOwnerReturnsUnavailable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.ListInvoices(ctx, 999999); !errors.Is(err, infrastructure.ErrBillingInvoicesUnavailable) {
		t.Fatalf("ListInvoices wrong business error = %v, want ErrBillingInvoicesUnavailable", err)
	}
}

func TestGetSummary_CurrentUsageBelowLimit(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.BusinessID)
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

	first, err := repo.GetSummary(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetSummary first: %v", err)
	}
	second, err := repo.GetSummary(ctx, testutil.BusinessID)
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

func TestListInvoices_DoesNotReturnFabricatedEmptySuccess(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.ListInvoices(ctx, testutil.BusinessID); !errors.Is(err, infrastructure.ErrBillingInvoicesUnavailable) {
		t.Fatalf("ListInvoices error = %v, want ErrBillingInvoicesUnavailable", err)
	}
}

func TestGetSummary_BusinessIDZero_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.GetSummary(ctx, 0); err == nil {
		t.Fatal("GetSummary business=0: expected an invalid-business error")
	}
}

func TestGetSummary_NegativeOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.GetSummary(ctx, -1); err == nil {
		t.Fatal("GetSummary business=-1: expected an invalid-business error")
	}
}

func TestGetSummary_StatusNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.BusinessID)
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

	summary, err := repo.GetSummary(ctx, testutil.BusinessID)
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

	s, err := repo.GetSummary(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if s.CurrentUsage > s.UsageLimit {
		t.Errorf("CurrentUsage (%f) > UsageLimit (%f): unexpected", s.CurrentUsage, s.UsageLimit)
	}
}

func TestListInvoices_MultipleCallsReturnSameUnavailableError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, firstErr := repo.ListInvoices(ctx, testutil.BusinessID)
	second, secondErr := repo.ListInvoices(ctx, testutil.BusinessID)
	if !errors.Is(firstErr, infrastructure.ErrBillingInvoicesUnavailable) || !errors.Is(secondErr, infrastructure.ErrBillingInvoicesUnavailable) {
		t.Fatalf("ListInvoices errors = %v, %v; want ErrBillingInvoicesUnavailable", firstErr, secondErr)
	}
	if first != nil || second != nil {
		t.Fatal("unavailable invoice history must not return fabricated rows")
	}
}

func TestGetSummary_TwiceReturnsSameStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.GetSummary(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("first GetSummary: %v", err)
	}
	second, err := repo.GetSummary(ctx, testutil.BusinessID)
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

	first, err := repo.GetSummary(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("first GetSummary: %v", err)
	}
	second, err := repo.GetSummary(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("second GetSummary: %v", err)
	}
	if first.Plan != second.Plan {
		t.Errorf("Plan changed: %q vs %q", first.Plan, second.Plan)
	}
}

func TestGetSummary_LargeBusinessID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, 1<<40)
	if err != nil {
		t.Fatalf("GetSummary large business: %v", err)
	}
	_ = summary
}

func TestListInvoices_ZeroOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.ListInvoices(ctx, 0); err == nil {
		t.Fatal("ListInvoices business=0: expected an invalid-business error")
	}
}

func TestGetSummary_UsageLimitExact(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetSummary(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	// Free tier limit is expected to be a round number > 0.
	if summary.UsageLimit != float64(int(summary.UsageLimit)) {
		t.Logf("UsageLimit is not a whole number: %f (may be acceptable)", summary.UsageLimit)
	}
}

func TestListInvoices_LargeOwnerReturnsUnavailable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.ListInvoices(ctx, 1<<40); !errors.Is(err, infrastructure.ErrBillingInvoicesUnavailable) {
		t.Fatalf("ListInvoices large business error = %v, want ErrBillingInvoicesUnavailable", err)
	}
}
