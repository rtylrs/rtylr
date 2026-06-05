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

// Authorize reports whether the caller may make an AI request right now and
// returns the current quota status. On a lookup error it fails open (allows)
// rather than blocking paying users on an infrastructure hiccup.
func (s *Service) Authorize(ctx context.Context, businessID, userID int64) (bool, domain.QuotaStatus, error) {
	owner := s.repo.ResolveOwnerID(ctx, businessID, userID)
	st, err := s.status(ctx, owner, userID)
	if err != nil {
		logger.Ctx(ctx).Error("quota authorize failed; failing open", logger.Error(err))
		return true, domain.QuotaStatus{}, err
	}
	return !st.Blocked, st, nil
}

// Record adds consumed tokens to the correct bucket for the tenant's mode.
func (s *Service) Record(ctx context.Context, businessID, userID int64, tokens int) {
	if tokens <= 0 {
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
	if addErr := s.repo.AddUsage(ctx, owner, bucket, monthStart(), int64(tokens)); addErr != nil {
		logger.Ctx(ctx).Error("quota record failed",
			logger.Any("owner_id", owner), logger.Error(addErr))
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
