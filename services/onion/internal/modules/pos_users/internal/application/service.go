package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_users/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	posUsersScope = "pos_users"
	posUsersTTL   = 2 * time.Minute
)

func invalidatePosUsersScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posUsersScope, businessID))
}

var ErrInvalidInput = errors.New("invalid input")

var ErrForbidden = errors.New("forbidden")

var ErrPINAlreadySet = errors.New("pin already set")

type Service struct {
	repo *infrastructure.Repository
	// pins is the narrowed repository view used by the credential path, and
	// pinAttempts the durable brute-force counter. Both are separate fields so
	// the verify path can be exercised without a database.
	pins        pinVerifier
	pinAttempts PinAttemptStore
	// now is the clock used to decide whether a key is inside its backoff
	// window. Overridable in tests so the schedule can be asserted without
	// sleeping.
	now func() time.Time
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)
	return &Service{
		repo:        repo,
		pins:        repo,
		pinAttempts: infrastructure.NewPINAttemptStore(db),
		now:         time.Now,
	}
}

func (s *Service) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func (s *Service) ListByOwner(ctx context.Context, businessID int64) ([]domain.POSUser, error) {
	return cacher.Of[[]domain.POSUser]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posUsersScope, businessID, "by_owner"),
		func(ctx context.Context) ([]domain.POSUser, error) {
			users, err := s.repo.ListByOwner(ctx, businessID)
			if err != nil {
				return nil, err
			}
			return users, nil
		},
		cacher.WithExpiry(ctx, posUsersTTL),
	)
}

func (s *Service) ListByLocation(ctx context.Context, businessID, locationID int64) ([]domain.POSUser, error) {
	key := cacheKeys.Scoped(posUsersScope, businessID, fmt.Sprintf("by_loc=%d", locationID))
	return cacher.Of[[]domain.POSUser]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.POSUser, error) {
			users, err := s.repo.ListByLocation(ctx, businessID, locationID)
			if err != nil {
				return nil, err
			}
			return users, nil
		},
		cacher.WithExpiry(ctx, posUsersTTL),
	)
}

func (s *Service) Create(ctx context.Context, businessID, createdBy int64, input domain.CreatePOSUserInput) (int64, error) {
	if strings.TrimSpace(input.Name) == "" {
		return 0, ErrInvalidInput
	}
	if input.PIN == "" {
		return 0, ErrInvalidInput
	}
	if err := domain.ValidatePIN(input.PIN); err != nil {
		return 0, err
	}
	hashed, err := passwordutils.HashPIN(input.PIN)
	if err != nil {
		return 0, err
	}
	input.PIN = ""
	id, err := s.repo.Create(ctx, businessID, createdBy, input, hashed)
	if err != nil {
		return 0, err
	}
	invalidatePosUsersScope(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.UpdatePOSUserInput) error {
	newPIN := input.PIN
	input.PIN = ""
	if newPIN != "" {
		if err := domain.ValidatePIN(newPIN); err != nil {
			return err
		}
	}
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	if newPIN != "" {
		hashed, err := passwordutils.HashPIN(newPIN)
		if err != nil {
			return err
		}
		if err := s.repo.UpdatePINHash(ctx, businessID, id, hashed); err != nil {
			return err
		}
		// A manager rotating a PIN is also the unblock path for an operator a
		// colleague locked out, so drop the counter for that row.
		s.clearPINAttemptsForPOSUser(ctx, businessID, id)
	}
	invalidatePosUsersScope(ctx, businessID)
	return nil
}

// clearPINAttemptsForPOSUser is best-effort: a failure here must not fail a PIN
// rotation that already committed, and the next successful verification
// clears the counter anyway.
func (s *Service) clearPINAttemptsForPOSUser(ctx context.Context, businessID, posUserID int64) {
	store, ok := s.pinAttempts.(*infrastructure.PINAttemptStore)
	if !ok {
		return
	}
	if err := store.ResetPINFailuresForPOSUser(ctx, businessID, posUserID); err != nil {
		logger.Ctx(ctx).Warn("failed to clear pos pin attempt counter after rotation",
			logger.Any("business_id", businessID),
			logger.Any("pos_user_id", posUserID),
			logger.Error(err),
		)
	}
}

func (s *Service) Deactivate(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Deactivate(ctx, businessID, id); err != nil {
		return err
	}
	invalidatePosUsersScope(ctx, businessID)
	return nil
}

type SetPINActor struct {
	UserID          int64
	IsOwner         bool
	TeamRole        string
	EffectiveAccess *authorization.EffectiveAccess
}

func (s *Service) SetPIN(ctx context.Context, businessID int64, actor SetPINActor, input domain.SetPINInput) error {
	if input.UserID <= 0 || input.LocationID <= 0 {
		return ErrInvalidInput
	}
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.PIN) == "" {
		return ErrInvalidInput
	}
	// Check the PIN policy before any lookup: a malformed PIN is the caller's own
	// input, so rejecting it early leaks nothing and avoids a wasted round trip.
	if err := domain.ValidatePIN(input.PIN); err != nil {
		return err
	}

	access, err := s.repo.GetMemberLocationAccess(ctx, businessID, input.UserID, input.LocationID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return ErrInvalidInput
		}
		return err
	}

	if input.UserID != actor.UserID {
		if mgrErr := s.requireLocationManager(ctx, businessID, actor, input.LocationID); mgrErr != nil {
			return mgrErr
		}
	}

	existing, err := s.repo.FindByUserAndLocation(ctx, businessID, input.UserID, input.LocationID)
	if err != nil && !errors.Is(err, infrastructure.ErrUserNotFound) {
		return err
	}
	if existing != nil {
		return ErrPINAlreadySet
	}

	hashed, err := passwordutils.HashPIN(input.PIN)
	if err != nil {
		return err
	}
	input.PIN = ""

	role := roleForAccess(access)
	caps := domain.DefaultCapabilities(role)
	_, err = s.repo.Create(ctx, businessID, input.UserID, domain.CreatePOSUserInput{
		UserID:           input.UserID,
		LocationID:       input.LocationID,
		Name:             input.Name,
		Role:             role,
		MaxDiscountPct:   caps.MaxDiscountPct,
		CanVoid:          caps.CanVoid,
		CanRefund:        caps.CanRefund,
		CanOpenDrawer:    caps.CanOpenDrawer,
		CanApplyDiscount: caps.CanApplyDiscount,
	}, hashed)
	if err != nil {
		return err
	}
	// Onboarding is the self-service unblock path for an operator a previous
	// brute-force attempt locked out.
	if s.pinAttempts != nil {
		if resetErr := s.pinAttempts.ResetPINFailures(ctx, businessID, input.LocationID, input.UserID); resetErr != nil {
			logger.Ctx(ctx).Warn("failed to clear pos pin attempt counter after set-pin",
				logger.Any("business_id", businessID),
				logger.Error(resetErr),
			)
		}
	}
	invalidatePosUsersScope(ctx, businessID)
	return nil
}

func (s *Service) pinVerifierOrDefault() pinVerifier {
	if s.pins != nil {
		return s.pins
	}
	if s.repo != nil {
		return s.repo
	}
	return missingPINVerifier{}
}

// recordPINFailure increments the brute-force counter, logging rather than
// returning a counter-store failure: the PIN itself is already being rejected,
// so there is no safer outcome to switch to and no oracle to avoid — the
// response is identical either way.
func (s *Service) recordPINFailure(ctx context.Context, store PinAttemptStore, businessID, locationID, userID int64) {
	if store == nil {
		return
	}
	if _, err := store.RecordPINFailure(ctx, businessID, locationID, userID); err != nil {
		logger.Ctx(ctx).Error("failed to record pos pin verification failure",
			logger.Any("business_id", businessID),
			logger.Any("location_id", locationID),
			logger.Any("pos_user_id", userID),
			logger.Error(err),
		)
	}
}

// missingPINVerifier stands in for a Service built without a repository, as the
// zero-value unit tests do. Every read fails as not-found, which the verify path
// maps to the same uniform denial a real unknown operator gets.
type missingPINVerifier struct{}

func (missingPINVerifier) FindForVerify(context.Context, int64, int64, int64) (*domain.POSUser, error) {
	return nil, infrastructure.ErrUserNotFound
}

func (missingPINVerifier) GetPINHash(context.Context, int64, int64) (string, error) {
	return "", infrastructure.ErrUserNotFound
}

func (missingPINVerifier) GetAuthorizationVersion(context.Context, int64, int64) (int64, error) {
	return 0, nil
}

func (s *Service) requireLocationManager(ctx context.Context, businessID int64, actor SetPINActor, locationID int64) error {
	if actor.EffectiveAccess != nil && hasPOSPermission(actor.EffectiveAccess, "pos:employees:manage") {
		return nil
	}
	access, err := s.repo.GetMemberLocationAccess(ctx, businessID, actor.UserID, locationID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return ErrForbidden
		}
		return err
	}
	if access.CanManagePOSUsers {
		return nil
	}
	return ErrForbidden
}

func hasPOSPermission(access *authorization.EffectiveAccess, permission string) bool {
	if access == nil {
		return false
	}
	for _, app := range access.Apps {
		if app.AppID != authorization.AppPOS || !app.Licensed || !app.HasAccess {
			continue
		}
		for _, granted := range app.Permissions {
			if granted == permission {
				return true
			}
		}
	}
	return false
}

func roleForAccess(access *infrastructure.LocationAccess) domain.POSUserRole {
	if access != nil && access.CanManagePOSUsers {
		return domain.RoleManager
	}
	return domain.RoleCashier
}

// VerifyPIN checks one operator PIN and, on success, returns the capability set
// that authorises voids, refunds and discounts. That makes it the endpoint worth
// brute-forcing, so the guard is deliberately blunt: a per-operator durable
// counter with exponential backoff, and a single uniform failure for every
// reason the check can fail.
//
// The uniform failure is load-bearing, not cosmetic. Unknown operator, wrong
// PIN, operator with no PIN set, counter-store error and active lockout all
// return domain.ErrInvalidPIN and all burn a real bcrypt comparison, so status,
// body and latency are indistinguishable. A distinct error or a skipped
// comparison on the unknown-operator path would let a caller enumerate which
// user ids hold a POS PIN at a location.
func (s *Service) VerifyPIN(ctx context.Context, businessID, locationID, userID int64, pin string) (*domain.VerifiedPOSUser, error) {
	pin = strings.TrimSpace(pin)
	if locationID <= 0 || userID <= 0 || pin == "" {
		burnPINCompare(pin)
		return nil, domain.ErrInvalidPIN
	}

	verifier := s.pinVerifierOrDefault()
	store := s.pinAttempts

	// Fail closed when the counter cannot be read. A counter we cannot read is
	// a counter we cannot enforce, and skipping the check would restore the
	// unbounded guess budget this guard exists to remove. Locked-out callers
	// land here too, so a correct PIN is refused until the window expires.
	if store != nil {
		retryAt, retryErr := store.PinRetryAt(ctx, businessID, locationID, userID)
		if retryErr != nil {
			logger.Ctx(ctx).Error("pos pin attempt counter unavailable; denying verification",
				logger.Any("business_id", businessID),
				logger.Any("location_id", locationID),
				logger.Any("pos_user_id", userID),
				logger.Error(retryErr),
			)
			burnPINCompare(pin)
			return nil, domain.ErrInvalidPIN
		}
		if !retryAt.IsZero() && s.clock().Before(retryAt) {
			burnPINCompare(pin)
			return nil, domain.ErrInvalidPIN
		}
	}

	user, err := verifier.FindForVerify(ctx, businessID, locationID, userID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			s.recordPINFailure(ctx, store, businessID, locationID, userID)
			burnPINCompare(pin)
			return nil, domain.ErrInvalidPIN
		}
		return nil, err
	}

	hash, err := verifier.GetPINHash(ctx, businessID, user.ID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			s.recordPINFailure(ctx, store, businessID, locationID, userID)
			burnPINCompare(pin)
			return nil, domain.ErrInvalidPIN
		}
		return nil, err
	}
	if hash == "" || !passwordutils.CheckPINHash(pin, hash) {
		if hash == "" {
			burnPINCompare(pin)
		}
		s.recordPINFailure(ctx, store, businessID, locationID, userID)
		return nil, domain.ErrInvalidPIN
	}

	authorizationVersion, err := verifier.GetAuthorizationVersion(ctx, businessID, user.UserID)
	if err != nil {
		return nil, err
	}

	// Only a real success clears the counter, and only after the capability
	// lookups above have succeeded, so a mid-verification error cannot be used
	// to reset an attacker's budget.
	if store != nil {
		if resetErr := store.ResetPINFailures(ctx, businessID, locationID, userID); resetErr != nil {
			logger.Ctx(ctx).Warn("failed to clear pos pin attempt counter after success",
				logger.Any("business_id", businessID),
				logger.Any("location_id", locationID),
				logger.Any("pos_user_id", userID),
				logger.Error(resetErr),
			)
		}
	}

	return &domain.VerifiedPOSUser{
		UserID:               user.UserID,
		POSUserID:            user.ID,
		LocationID:           user.LocationID,
		Role:                 user.Role,
		DisplayName:          user.Name,
		AuthorizationVersion: authorizationVersion,
		Capabilities: domain.Capabilities{
			CanVoid:          user.CanVoid,
			CanRefund:        user.CanRefund,
			CanOpenDrawer:    user.CanOpenDrawer,
			CanApplyDiscount: user.CanApplyDiscount,
			MaxDiscountPct:   user.MaxDiscountPct,
		},
	}, nil
}
