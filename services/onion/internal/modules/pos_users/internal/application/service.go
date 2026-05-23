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
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	posUsersScope = "pos_users"
	posUsersTTL   = 2 * time.Minute
)

func invalidatePosUsersScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(posUsersScope, ownerID))
	}
}

// ErrInvalidInput signals that a caller-supplied field (e.g. empty PIN on
// a create call) violates a service-layer invariant. Handlers translate
// this into a 400.
var ErrInvalidInput = errors.New("invalid input")

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) ListByOwner(ctx context.Context, ownerID int64) ([]domain.POSUser, error) {
	return cacher.Of[[]domain.POSUser]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posUsersScope, ownerID, "by_owner"),
		func(ctx context.Context) ([]domain.POSUser, error) {
			users, err := s.repo.ListByOwner(ctx, ownerID)
			if err != nil {
				return nil, err
			}
			return scrubSecrets(users), nil
		},
		cacher.WithExpiry(ctx, posUsersTTL),
	)
}

func (s *Service) ListByLocation(ctx context.Context, ownerID, locationID int64) ([]domain.POSUser, error) {
	key := cacheKeys.Scoped(posUsersScope, ownerID, fmt.Sprintf("by_loc=%d", locationID))
	return cacher.Of[[]domain.POSUser]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.POSUser, error) {
			users, err := s.repo.ListByLocation(ctx, ownerID, locationID)
			if err != nil {
				return nil, err
			}
			return scrubSecrets(users), nil
		},
		cacher.WithExpiry(ctx, posUsersTTL),
	)
}

// Create hashes the plaintext PIN with bcrypt cost 12 before persisting and
// never passes the plaintext to the repository.
func (s *Service) Create(ctx context.Context, ownerID, createdBy int64, input domain.CreatePOSUserInput) (int64, error) {
	if strings.TrimSpace(input.Name) == "" {
		return 0, ErrInvalidInput
	}
	if input.PIN == "" {
		return 0, ErrInvalidInput
	}
	hashed, err := passwordutils.HashPIN(input.PIN)
	if err != nil {
		return 0, err
	}
	// Zero the plaintext on the input struct so a slip further down the
	// call stack (logging, error-wrap) cannot surface it.
	input.PIN = ""
	id, err := s.repo.Create(ctx, ownerID, createdBy, input, hashed)
	if err != nil {
		return 0, err
	}
	invalidatePosUsersScope(ctx, ownerID)
	return id, nil
}

// Update rewrites a pos_user's profile fields. If the caller supplied a new
// PIN, it is hashed and written via the dedicated PIN-hash query; the
// plaintext never reaches the repository layer.
func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.UpdatePOSUserInput) error {
	newPIN := input.PIN
	input.PIN = ""
	if err := s.repo.Update(ctx, ownerID, id, input); err != nil {
		return err
	}
	if newPIN != "" {
		hashed, err := passwordutils.HashPIN(newPIN)
		if err != nil {
			return err
		}
		if err := s.repo.UpdatePINHash(ctx, ownerID, id, hashed); err != nil {
			return err
		}
	}
	invalidatePosUsersScope(ctx, ownerID)
	return nil
}

func (s *Service) Deactivate(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.Deactivate(ctx, ownerID, id); err != nil {
		return err
	}
	invalidatePosUsersScope(ctx, ownerID)
	return nil
}

// VerifyPIN authenticates a pos_user by (ownerID, display name, pin) and
// returns a minimal projection that intentionally omits the PIN, the hash,
// and any internal bookkeeping. A generic domain.ErrInvalidPIN is returned
// for every failure mode so callers cannot distinguish "wrong name" from
// "wrong PIN" via the API surface.
//
// This function implements the dual-path verify described in the security
// overhaul: legacy rows that still hold a plaintext PIN (pin_hash IS NULL)
// are compared against the submitted PIN in constant-ish time, then
// transparently migrated to a bcrypt hash on success. Rows with a pin_hash
// take the fast path and skip the plaintext comparison entirely.
//
// TODO(legacy-pin-cleanup): once `SELECT COUNT(*) FROM pos_users WHERE pin_hash IS NULL`
// returns 0 in production (target: ~2026-06-01 once all clients have transacted at least
// once and triggered transparent rehash), remove the legacy plaintext branch below.
func (s *Service) VerifyPIN(ctx context.Context, ownerID int64, displayName, pin string) (*domain.VerifiedPOSUser, error) {
	if displayName == "" || pin == "" {
		return nil, domain.ErrInvalidPIN
	}

	user, err := s.repo.FindByDisplayName(ctx, ownerID, displayName)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return nil, domain.ErrInvalidPIN
		}
		return nil, err
	}

	switch {
	case user.PINHash != "":
		if !passwordutils.CheckPINHash(pin, user.PINHash) {
			return nil, domain.ErrInvalidPIN
		}
	case user.PIN != "":
		// Legacy plaintext path. Only accept equality; never log either
		// side of the comparison.
		if pin != user.PIN {
			return nil, domain.ErrInvalidPIN
		}
		if err := s.migrateLegacyPIN(ctx, user, pin); err != nil {
			// Migration failure is non-fatal for the verify call itself
			// — the PIN was correct, we just couldn't upgrade storage.
			// Log without the PIN and continue.
			logger.Ctx(ctx).Warn("pos_users legacy pin migration failed",
				logger.Any("owner_id", ownerID),
				logger.Any("pos_user_id", user.ID),
				logger.Error(err),
			)
		}
	default:
		// No hash and no plaintext — nothing to verify against.
		return nil, domain.ErrInvalidPIN
	}

	return &domain.VerifiedPOSUser{
		UserID:      user.ID,
		Role:        user.Role,
		DisplayName: user.Name,
	}, nil
}

// migrateLegacyPIN hashes the just-verified plaintext PIN and persists it,
// clearing the legacy column in the same statement. It is intentionally
// separate from VerifyPIN so the verify path can log migration failures
// without polluting the hot path.
func (s *Service) migrateLegacyPIN(ctx context.Context, user *domain.POSUser, pin string) error {
	hashed, err := passwordutils.HashPIN(pin)
	if err != nil {
		return err
	}
	return s.repo.UpdatePINHash(ctx, user.OwnerID, user.ID, hashed)
}

// scrubSecrets returns the same slice with PIN and PINHash zeroed so that
// list endpoints cannot accidentally leak either value through the handler
// layer. The domain.POSUser shape still carries these fields for the
// dual-path verify flow that needs them.
func scrubSecrets(users []domain.POSUser) []domain.POSUser {
	for i := range users {
		users[i].PIN = ""
		users[i].PINHash = ""
	}
	return users
}
