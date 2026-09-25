package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/aiquota/internal/domain"
	"github.com/voxire/onion/internal/modules/aiquota/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

type repository interface {
	ResolveOwnerID(ctx context.Context, businessID, fallbackUserID int64) int64
	GetEntitlement(ctx context.Context, ownerID int64) (domain.Entitlement, error)
	GetMode(ctx context.Context, ownerID int64) (domain.Mode, error)
	SetMode(ctx context.Context, ownerID int64, mode domain.Mode) error
	GetUsedTokens(ctx context.Context, ownerID, bucketUserID int64, month time.Time) (int64, error)
	AddUsage(ctx context.Context, ownerID, bucketUserID int64, month time.Time, tokens int64) error
	GetFreePlanCredits(ctx context.Context) (int64, error)
	Reserve(ctx context.Context, ownerID, bucketUserID int64, month time.Time, estimate, allowance int64) (bool, error)
	Adjust(ctx context.Context, ownerID, bucketUserID int64, month time.Time, delta int64) error
}

// ReservationEstimateTokens is optimistically debited before the provider call
// so N parallel requests cannot all pass a nearly exhausted allowance. The
// post-stream reconciliation credits back the unused portion.
// ReservationEstimateTokens is exported so the module adapter forwards the
// exact debit size to callers for reconciliation.
const ReservationEstimateTokens = int64(4000)

// activeEntitlementStatuses lists subscription states that still grant plan AI
// credits; anything else (cancelled, past_due, ...) falls back to the free tier.
var activeEntitlementStatuses = map[string]struct{}{
	"active":   {},
	"trialing": {},
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// perSeatBonusCredits mirrors the per-seat AI credit bonus described in the plan
// seed (Pro +500/seat, Scale +1000/seat). Base credits come from the DB column.
func perSeatBonusCredits(slug string) int {
	switch slug {
	case "pro":
		return 500
	case "scale":
		return 1000
	default:
		return 0
	}
}

// allowsOverage reports whether a plan keeps serving (and bills) past its
// allowance. Paid plans never hard-block; free/solo plans do.
func allowsOverage(slug string) bool {
	return slug == "pro" || slug == "scale"
}

func monthStart() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func monthEnd() time.Time {
	return monthStart().AddDate(0, 1, 0)
}

func (s *Service) status(ctx context.Context, ownerID, userID int64) (domain.QuotaStatus, error) {
	ent, err := s.repo.GetEntitlement(ctx, ownerID)
	if err != nil {
		return domain.QuotaStatus{}, err
	}
	if _, active := activeEntitlementStatuses[ent.Status]; !active && ent.Status != "none" {
		// Cancelled or past-due subscriptions no longer earn plan credits;
		// the tenant keeps only whatever the free plan grants.
		credits, ferr := s.repo.GetFreePlanCredits(ctx)
		if ferr == nil {
			ent = domain.Entitlement{PlanSlug: "free", AICreditsPerMonth: int(credits), SeatCount: ent.SeatCount, Status: "free"}
		}
	}
	mode, err := s.repo.GetMode(ctx, ownerID)
	if err != nil {
		return domain.QuotaStatus{}, err
	}

	seats := ent.SeatCount
	if seats < 1 {
		seats = 1
	}
	teamCredits := ent.AICreditsPerMonth + perSeatBonusCredits(ent.PlanSlug)*seats
	teamAllowance := int64(teamCredits) * int64(config.AITokensPerCredit)

	bucketUser := int64(0)
	allowance := teamAllowance
	if mode == domain.ModePerPerson {
		bucketUser = userID
		allowance = teamAllowance / int64(seats)
	}

	used, err := s.repo.GetUsedTokens(ctx, ownerID, bucketUser, monthStart())
	if err != nil {
		return domain.QuotaStatus{}, err
	}

	overageAllowed := allowsOverage(ent.PlanSlug)
	remaining := allowance - used
	if remaining < 0 {
		remaining = 0
	}
	overage := used - allowance
	if overage < 0 {
		overage = 0
	}

	return domain.QuotaStatus{
		PlanSlug:        ent.PlanSlug,
		Mode:            mode,
		AllowanceTokens: allowance,
		UsedTokens:      used,
		RemainingTokens: remaining,
		OverageTokens:   overage,
		AllowOverage:    overageAllowed,
		Blocked:         !overageAllowed && used >= allowance,
		PeriodEnd:       monthEnd(),
	}, nil
}

// Authorize reserves ReservationEstimateTokens against the tenant's allowance
// atomically and reports whether the caller may make an AI request. Because
// the reservation is the gate, N parallel requests can no longer all pass a
// nearly exhausted allowance.
//
// Failure policy: when the entitlement itself cannot be read the request is
// blocked (failing closed) — an outage must not make AI unmetered for
// hard-capped plans. When only the usage bucket cannot be read, plans that
// bill for overage keep serving; capped plans are blocked.
func (s *Service) Reserve(ctx context.Context, businessID, userID int64) (bool, domain.QuotaStatus, error) {
	owner := s.repo.ResolveOwnerID(ctx, businessID, userID)
	ent, entErr := s.repo.GetEntitlement(ctx, owner)
	mode, modeErr := s.repo.GetMode(ctx, owner)

	if entErr != nil {
		logger.Ctx(ctx).Error("AI QUOTA UNMETERABLE: entitlement lookup failed; failing closed",
			logger.Any("owner_id", owner), logger.Error(entErr))
		return false, domain.QuotaStatus{Blocked: true}, entErr
	}
	if _, active := activeEntitlementStatuses[ent.Status]; !active && ent.Status != "none" {
		if credits, ferr := s.repo.GetFreePlanCredits(ctx); ferr == nil {
			ent = domain.Entitlement{PlanSlug: "free", AICreditsPerMonth: int(credits), SeatCount: ent.SeatCount, Status: "free"}
		}
	}
	if modeErr != nil {
		mode = domain.ModePool
	}

	bucketUser, allowance := bucketAndAllowance(ent, mode, userID)

	used, err := s.repo.GetUsedTokens(ctx, owner, bucketUser, monthStart())
	if err != nil {
		if allowsOverage(ent.PlanSlug) {
			logger.Ctx(ctx).Error("quota usage lookup failed; failing open for overage plan",
				logger.Any("owner_id", owner), logger.Error(err))
			return true, domain.QuotaStatus{}, err
		}
		logger.Ctx(ctx).Error("AI QUOTA BLOCKING: usage lookup failed on capped plan",
			logger.Any("owner_id", owner), logger.Error(err))
		return false, domain.QuotaStatus{}, err
	}

	overageAllowed := allowsOverage(ent.PlanSlug)
	if !overageAllowed && used >= allowance {
		return false, buildStatus(ent, mode, allowance, used), nil
	}

	reserved, err := s.repo.Reserve(ctx, owner, bucketUser, monthStart(), ReservationEstimateTokens, allowance)
	if err != nil {
		if allowsOverage(ent.PlanSlug) {
			logger.Ctx(ctx).Error("quota reserve failed; failing open for overage plan",
				logger.Any("owner_id", owner), logger.Error(err))
			return true, domain.QuotaStatus{}, err
		}
		logger.Ctx(ctx).Error("AI QUOTA BLOCKING: reserve failed on capped plan",
			logger.Any("owner_id", owner), logger.Error(err))
		return false, domain.QuotaStatus{}, err
	}
	if !reserved {
		return false, buildStatus(ent, mode, allowance, used), nil
	}

	used += ReservationEstimateTokens
	return true, buildStatus(ent, mode, allowance, used), nil
}

func bucketAndAllowance(ent domain.Entitlement, mode domain.Mode, userID int64) (int64, int64) {
	seats := ent.SeatCount
	if seats < 1 {
		seats = 1
	}
	teamCredits := ent.AICreditsPerMonth + perSeatBonusCredits(ent.PlanSlug)*seats
	teamAllowance := int64(teamCredits) * int64(config.AITokensPerCredit)
	if mode == domain.ModePerPerson {
		return userID, teamAllowance / int64(seats)
	}
	return 0, teamAllowance
}

func buildStatus(ent domain.Entitlement, mode domain.Mode, allowance, used int64) domain.QuotaStatus {
	remaining := allowance - used
	if remaining < 0 {
		remaining = 0
	}
	overage := used - allowance
	if overage < 0 {
		overage = 0
	}
	return domain.QuotaStatus{
		PlanSlug:        ent.PlanSlug,
		Mode:            mode,
		AllowanceTokens: allowance,
		UsedTokens:      used,
		RemainingTokens: remaining,
		OverageTokens:   overage,
		AllowOverage:    allowsOverage(ent.PlanSlug),
		Blocked:         !allowsOverage(ent.PlanSlug) && used >= allowance,
		PeriodEnd:       monthEnd(),
	}
}

// Reconcile replaces a prior reservation with actual usage: the delta between
// what Reserve debited and what the provider reported is applied here, so a
// response that used fewer tokens credits the difference back.
//
// Writes are retried with backoff because a lost reconciliation permanently
// under-counts usage and quotas never fill.
func (s *Service) Reconcile(ctx context.Context, businessID, userID, actualTokens, reservedTokens int64) {
	delta := actualTokens - reservedTokens
	if delta == 0 {
		return
	}
	owner := s.repo.ResolveOwnerID(ctx, businessID, userID)
	mode, err := s.repo.GetMode(ctx, owner)
	if err != nil {
		mode = domain.ModePool
	}
	bucket := int64(0)
	if mode == domain.ModePerPerson {
		bucket = userID
	}

	const attempts = 3
	for i := 0; i < attempts; i++ {
		addErr := s.repo.Adjust(ctx, owner, bucket, monthStart(), delta)
		if addErr == nil {
			return
		}
		if i == attempts-1 {
			logger.Ctx(ctx).Error("AI QUOTA UNDER-COUNT: reconcile failed after retries",
				logger.Any("owner_id", owner), logger.Any("delta", delta), logger.Error(addErr))
			return
		}
		time.Sleep(time.Duration(250*(i+1)) * time.Millisecond)
	}
}

// Status returns the caller's current quota view.
func (s *Service) Status(ctx context.Context, businessID, userID int64) (domain.QuotaStatus, error) {
	owner := s.repo.ResolveOwnerID(ctx, businessID, userID)
	return s.status(ctx, owner, userID)
}

// GetMode returns the tenant's allocation mode.
func (s *Service) GetMode(ctx context.Context, businessID, userID int64) (domain.Mode, error) {
	owner := s.repo.ResolveOwnerID(ctx, businessID, userID)
	return s.repo.GetMode(ctx, owner)
}

// SetMode updates the tenant's allocation mode (admin action).
func (s *Service) SetMode(ctx context.Context, businessID, userID int64, mode domain.Mode) error {
	owner := s.repo.ResolveOwnerID(ctx, businessID, userID)
	return s.repo.SetMode(ctx, owner, mode.Normalize())
}
