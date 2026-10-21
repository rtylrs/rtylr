package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/aiquota/internal/domain"
	"github.com/voxire/onion/internal/platform/config"
)

type mockRepo struct {
	ent      domain.Entitlement
	entErr   error
	mode     domain.Mode
	used     int64
	usedErr  error
	free     int64
	freeErr  error
	reserveE error
	adjusts  []int64
	failAdjN int

	// resolveErr reproduces an unverified tenant scope: owner resolution fails
	// the way tenant.ResolveOwnerID does when the business cannot be confirmed.
	resolveErr error
	// scoped records every owner id the service used for a tenant-scoped call,
	// so a test can prove a failed resolution reached no read and no write.
	scoped []int64
}

func (m *mockRepo) ResolveOwnerID(_ context.Context, _, _ int64) (int64, error) {
	if m.resolveErr != nil {
		return 0, m.resolveErr
	}
	return 1, nil
}

func (m *mockRepo) scope(ownerID int64) { m.scoped = append(m.scoped, ownerID) }

func (m *mockRepo) GetAIQuotaEntitlement(_ context.Context, ownerID int64) (string, int, int, string, error) {
	m.scope(ownerID)
	return m.ent.PlanSlug, m.ent.AICreditsPerMonth, m.ent.SeatCount, m.ent.Status, m.entErr
}
func (m *mockRepo) GetMode(_ context.Context, ownerID int64) (domain.Mode, error) {
	m.scope(ownerID)
	return m.mode, nil
}
func (m *mockRepo) SetMode(_ context.Context, ownerID int64, _ domain.Mode) error {
	m.scope(ownerID)
	return nil
}
func (m *mockRepo) GetUsedTokens(_ context.Context, ownerID, _ int64, _ time.Time) (int64, error) {
	m.scope(ownerID)
	return m.used, m.usedErr
}

func (m *mockRepo) AddUsage(_ context.Context, _, _ int64, _ time.Time, _ int64) error { return nil }

func (m *mockRepo) GetFreePlanAICredits(_ context.Context) (int64, error) {
	return m.free, m.freeErr
}

func (m *mockRepo) Reserve(_ context.Context, ownerID, _ int64, _ time.Time, estimate, _ int64, allowOverage bool) (bool, error) {
	m.scope(ownerID)
	if m.reserveE != nil {
		return false, m.reserveE
	}
	if !allowOverage && m.used+estimate > allowanceOf(m.ent) {
		return false, nil
	}
	m.used += estimate
	return true, nil
}

func (m *mockRepo) Adjust(_ context.Context, ownerID, _ int64, _ time.Time, delta int64) error {
	m.scope(ownerID)
	if m.failAdjN > 0 {
		m.failAdjN--
		if m.failAdjN == 0 {
			return errors.New("adjust failed once")
		}
	}
	m.adjusts = append(m.adjusts, delta)
	return nil
}

func allowanceOf(ent domain.Entitlement) int64 {
	_, allowance := bucketAndAllowance(ent, domain.ModePool, 7)
	return allowance
}

func TestReserve_BlocksCappedPlanWhenExhausted(t *testing.T) {
	m := &mockRepo{
		ent:  domain.Entitlement{PlanSlug: "free", AICreditsPerMonth: 10, SeatCount: 1, Status: "active"},
		used: 10_000_000,
	}
	s := &Service{repo: m, entitlements: m}
	allowed, st, err := s.Reserve(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if allowed || !st.Blocked {
		t.Errorf("allowed=%v blocked=%v, want blocked on exhausted capped plan", allowed, st.Blocked)
	}
}

func TestReserve_AllowsOveragePlanPastAllowance(t *testing.T) {
	allowance := allowanceOf(domain.Entitlement{PlanSlug: "pro", AICreditsPerMonth: 1000, SeatCount: 2, Status: "active"})
	m := &mockRepo{
		ent:  domain.Entitlement{PlanSlug: "pro", AICreditsPerMonth: 1000, SeatCount: 2, Status: "active"},
		used: allowance + 50_000, // already in overage
	}
	s := &Service{repo: m, entitlements: m}
	allowed, st, err := s.Reserve(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if !allowed || st.Blocked {
		t.Errorf("allowed=%v blocked=%v, want allowed for active pro plan in overage", allowed, st.Blocked)
	}
	if st.OverageTokens <= 0 {
		t.Errorf("expected positive overage tokens, got %d", st.OverageTokens)
	}
}

func TestReserve_FailsClosedWhenEntitlementUnavailable(t *testing.T) {
	m := &mockRepo{entErr: errors.New("db down")}
	s := &Service{repo: m, entitlements: m}
	allowed, _, err := s.Reserve(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("expected error")
	}
	if allowed {
		t.Error("entitlement lookup failure must fail closed, got allowed")
	}
}

func TestReserve_FailPolicyOnUsageLookupError(t *testing.T) {
	capped := &mockRepo{
		ent:     domain.Entitlement{PlanSlug: "free", AICreditsPerMonth: 10, SeatCount: 1, Status: "active"},
		usedErr: errors.New("db down"),
	}
	s := &Service{repo: capped, entitlements: capped}
	if allowed, _, _ := s.Reserve(context.Background(), 1, 2); allowed {
		t.Error("capped plan must fail closed on usage lookup error")
	}

	overage := &mockRepo{
		ent:     domain.Entitlement{PlanSlug: "pro", AICreditsPerMonth: 1000, SeatCount: 3, Status: "active"},
		usedErr: errors.New("db down"),
	}
	s2 := &Service{repo: overage, entitlements: overage}
	if allowed, _, _ := s2.Reserve(context.Background(), 1, 2); !allowed {
		t.Error("overage plan should keep serving through usage lookup errors")
	}
}

func TestReserve_InactiveSubscriptionFallsBackToFreeCredits(t *testing.T) {
	freeAllowance := int64(100) * int64(config.AITokensPerCredit)
	m := &mockRepo{
		ent:  domain.Entitlement{PlanSlug: "pro", AICreditsPerMonth: 5000, SeatCount: 2, Status: "cancelled"},
		free: 100,
		used: freeAllowance,
	}
	s := &Service{repo: m, entitlements: m}
	allowed, st, err := s.Reserve(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if st.PlanSlug != "free" {
		t.Errorf("plan = %q, want free after cancellation", st.PlanSlug)
	}
	if allowed {
		t.Error("cancelled subscription must be capped at free allowance")
	}
}

func TestReconcile_AppliesDeltaAndRetriesOnFailure(t *testing.T) {
	m := &mockRepo{failAdjN: 2} // fails twice, succeeds on attempt 3
	s := &Service{repo: m, entitlements: m}
	s.Reconcile(context.Background(), 1, 2, 3500, 4000)
	if len(m.adjusts) != 1 || m.adjusts[0] != -500 {
		t.Errorf("adjusts = %v, want [-500]", m.adjusts)
	}
}

func TestUnverifiedOwnerReachesNoWrite(t *testing.T) {
	m := &mockRepo{resolveErr: errors.New("no business scope for request")}
	s := &Service{repo: m, entitlements: m}

	t.Run("SetMode does not write another tenant's mode", func(t *testing.T) {
		if err := s.SetMode(context.Background(), 1, 2, domain.ModePerPerson); err == nil {
			t.Fatal("expected SetMode to fail")
		}
		if len(m.scoped) != 0 {
			t.Errorf("SetMode touched owner IDs %v on unverified business", m.scoped)
		}
	})

	t.Run("Reserve neither serves nor debits", func(t *testing.T) {
		allowed, _, err := s.Reserve(context.Background(), 1, 2)
		if err == nil {
			t.Fatal("expected Reserve to fail")
		}
		if allowed {
			t.Error("Reserve must not report allowed for an unverified business")
		}
		if len(m.scoped) != 0 {
			t.Errorf("Reserve touched owner IDs %v on unverified business", m.scoped)
		}
	})

	t.Run("Reconcile does not debit another tenant's usage", func(t *testing.T) {
		s.Reconcile(context.Background(), 1, 2, 3000, 4000)
		if len(m.scoped) != 0 || len(m.adjusts) != 0 {
			t.Errorf("Reconcile touched owner IDs %v (adjusts=%v) on unverified business", m.scoped, m.adjusts)
		}
	})
}

func TestUnverifiedOwnerFailsClosedOnReads(t *testing.T) {
	m := &mockRepo{resolveErr: errors.New("no business scope for request")}
	s := &Service{repo: m, entitlements: m}

	if _, err := s.Status(context.Background(), 1, 2); err == nil {
		t.Error("expected Status to fail on unverified business")
	}
	if _, err := s.GetMode(context.Background(), 1, 2); err == nil {
		t.Error("expected GetMode to fail on unverified business")
	}
	if len(m.scoped) != 0 {
		t.Errorf("read methods touched owner IDs %v on unverified business", m.scoped)
	}
}

func TestResolvedOwnerScopesEveryScopedCall(t *testing.T) {
	m := &mockRepo{
		ent: domain.Entitlement{PlanSlug: "free", AICreditsPerMonth: 10, SeatCount: 1, Status: "active"},
	}
	s := &Service{repo: m, entitlements: m}

	ctx := context.Background()
	_, _ = s.Status(ctx, 1, 2)
	_, _ = s.GetMode(ctx, 1, 2)
	_ = s.SetMode(ctx, 1, 2, domain.ModePerPerson)
	_, _, _ = s.Reserve(ctx, 1, 2)
	s.Reconcile(ctx, 1, 2, 3000, 4000)

	for i, id := range m.scoped {
		if id != 1 {
			t.Errorf("scoped[%d] = %d, want resolved owner 1", i, id)
		}
	}
}
