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
}

func (m *mockRepo) ResolveOwnerID(_ context.Context, _, _ int64) int64 { return 1 }
func (m *mockRepo) GetEntitlement(_ context.Context, _ int64) (domain.Entitlement, error) {
	return m.ent, m.entErr
}
func (m *mockRepo) GetMode(_ context.Context, _ int64) (domain.Mode, error) { return m.mode, nil }
func (m *mockRepo) SetMode(_ context.Context, _ int64, _ domain.Mode) error { return nil }
func (m *mockRepo) GetUsedTokens(_ context.Context, _, _ int64, _ time.Time) (int64, error) {
	return m.used, m.usedErr
}

func (m *mockRepo) AddUsage(_ context.Context, _, _ int64, _ time.Time, _ int64) error { return nil }

func (m *mockRepo) GetFreePlanCredits(_ context.Context) (int64, error) {
	return m.free, m.freeErr
}

func (m *mockRepo) Reserve(_ context.Context, _, _ int64, _ time.Time, estimate, _ int64) (bool, error) {
	if m.reserveE != nil {
		return false, m.reserveE
	}
	if m.used+estimate > allowanceOf(m.ent) {
		return false, nil
	}
	m.used += estimate
	return true, nil
}

func (m *mockRepo) Adjust(_ context.Context, _, _ int64, _ time.Time, delta int64) error {
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
	s := &Service{repo: m}
	allowed, st, err := s.Reserve(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if allowed || !st.Blocked {
		t.Errorf("allowed=%v blocked=%v, want blocked on exhausted capped plan", allowed, st.Blocked)
	}
}

func TestReserve_FailsClosedWhenEntitlementUnavailable(t *testing.T) {
	m := &mockRepo{entErr: errors.New("db down")}
	s := &Service{repo: m}
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
	s := &Service{repo: capped}
	if allowed, _, _ := s.Reserve(context.Background(), 1, 2); allowed {
		t.Error("capped plan must fail closed on usage lookup error")
	}

	overage := &mockRepo{
		ent:     domain.Entitlement{PlanSlug: "pro", AICreditsPerMonth: 1000, SeatCount: 3, Status: "active"},
		usedErr: errors.New("db down"),
	}
	s2 := &Service{repo: overage}
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
	s := &Service{repo: m}
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
	m := &mockRepo{failAdjN: 3}
	s := &Service{repo: m}
	s.Reconcile(context.Background(), 1, 2, 3000, ReservationEstimateTokens)
	if len(m.adjusts) != 1 {
		t.Fatalf("adjust calls = %d, want 1", len(m.adjusts))
	}
	if m.adjusts[0] != 3000-ReservationEstimateTokens {
		t.Errorf("delta = %d, want actual minus reservation", m.adjusts[0])
	}
}
