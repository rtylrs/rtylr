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

func (m *mockRepo) GetEntitlement(_ context.Context, ownerID int64) (domain.Entitlement, error) {
	m.scope(ownerID)
	return m.ent, m.entErr
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

func (m *mockRepo) GetFreePlanCredits(_ context.Context) (int64, error) {
	return m.free, m.freeErr
}

func (m *mockRepo) Reserve(_ context.Context, ownerID, _ int64, _ time.Time, estimate, _ int64) (bool, error) {
	m.scope(ownerID)
	if m.reserveE != nil {
		return false, m.reserveE
	}
	if m.used+estimate > allowanceOf(m.ent) {
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

// unverifiedOwner is the resolution error the service must treat as a denial.
// Every write path in this module is scoped by the owner id that resolution
// returns, so a substituted value is a cross-tenant write.
var unverifiedOwner = errors.New("no business scope for request")

// TestUnverifiedOwnerReachesNoWrite is the regression test for the fail-open
// that ResolveOwnerID used to reintroduce: on an unresolved tenant the service
// fell through with the caller's user id and then read and wrote under it.
func TestUnverifiedOwnerReachesNoWrite(t *testing.T) {
	t.Run("SetMode does not write another tenant's mode", func(t *testing.T) {
		m := &mockRepo{resolveErr: unverifiedOwner}
		s := &Service{repo: m}

		err := s.SetMode(context.Background(), 1, 2, domain.ModePerPerson)

		if !errors.Is(err, unverifiedOwner) {
			t.Fatalf("SetMode error = %v, want %v", err, unverifiedOwner)
		}
		if len(m.scoped) != 0 {
			t.Errorf("repo calls under owner ids %v, want none", m.scoped)
		}
	})

	t.Run("Reserve neither serves nor debits", func(t *testing.T) {
		m := &mockRepo{
			ent:        domain.Entitlement{PlanSlug: "pro", AICreditsPerMonth: 1000, SeatCount: 3, Status: "active"},
			resolveErr: unverifiedOwner,
		}
		s := &Service{repo: m}

		allowed, st, err := s.Reserve(context.Background(), 1, 2)

		if !errors.Is(err, unverifiedOwner) {
			t.Fatalf("Reserve error = %v, want %v", err, unverifiedOwner)
		}
		if allowed {
			t.Error("an unverified tenant scope must not be allowed through the AI quota gate")
		}
		if !st.Blocked {
			t.Error("status must be blocked when the tenant scope is unverified")
		}
		if len(m.scoped) != 0 {
			t.Errorf("repo calls under owner ids %v, want none", m.scoped)
		}
	})

	t.Run("Reconcile does not debit another tenant's usage", func(t *testing.T) {
		m := &mockRepo{resolveErr: unverifiedOwner}
		s := &Service{repo: m}

		s.Reconcile(context.Background(), 1, 2, 3000, ReservationEstimateTokens)

		if len(m.adjusts) != 0 {
			t.Errorf("adjusts = %v, want none: reconciliation must not write to an unverified tenant", m.adjusts)
		}
		if len(m.scoped) != 0 {
			t.Errorf("repo calls under owner ids %v, want none", m.scoped)
		}
	})
}

func TestUnverifiedOwnerFailsClosedOnReads(t *testing.T) {
	s := &Service{repo: &mockRepo{resolveErr: unverifiedOwner}}

	if _, err := s.Status(context.Background(), 1, 2); !errors.Is(err, unverifiedOwner) {
		t.Errorf("Status error = %v, want %v", err, unverifiedOwner)
	}
	if _, err := s.GetMode(context.Background(), 1, 2); !errors.Is(err, unverifiedOwner) {
		t.Errorf("GetMode error = %v, want %v", err, unverifiedOwner)
	}
}

// TestResolvedOwnerScopesEveryScopedCall pins that a successful resolution is
// still the owner id the reads and writes run under, so the fail-closed path
// above is not achieved by simply ignoring the resolver.
func TestResolvedOwnerScopesEveryScopedCall(t *testing.T) {
	m := &mockRepo{ent: domain.Entitlement{PlanSlug: "pro", AICreditsPerMonth: 1000, SeatCount: 3, Status: "active"}}
	s := &Service{repo: m}

	if _, _, err := s.Reserve(context.Background(), 1, 2); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	s.Reconcile(context.Background(), 1, 2, 3000, ReservationEstimateTokens)
	if _, err := s.Status(context.Background(), 1, 2); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if _, err := s.GetMode(context.Background(), 1, 2); err != nil {
		t.Fatalf("GetMode: %v", err)
	}
	if err := s.SetMode(context.Background(), 1, 2, domain.ModePerPerson); err != nil {
		t.Fatalf("SetMode: %v", err)
	}

	if len(m.scoped) == 0 {
		t.Fatal("no scoped repo calls recorded")
	}
	for i, ownerID := range m.scoped {
		if ownerID != 1 {
			t.Fatalf("scoped call %d used owner id %d, want the resolved owner 1", i, ownerID)
		}
	}
}
