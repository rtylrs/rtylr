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

// SetPIN is the self-service onboarding path: an operator picked from a
// location's roster sets or rotates their own 4-digit PIN. The operator's
// identity (userID) and the location come from the roster selection, not a
// trusted session, so the member's assignment to the location is re-verified
// here. Role and capabilities are resolved server-side from the member's
// location access — never trusted from the client.
//
// First-time setup creates the pos_users row with role-defaulted capabilities;
// a rotation only rewrites the PIN hash and preserves any admin overrides.
func (s *Service) SetPIN(ctx context.Context, ownerID int64, input domain.SetPINInput) error {
	if input.UserID <= 0 || input.LocationID <= 0 {
		return ErrInvalidInput
	}
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.PIN) == "" {
		return ErrInvalidInput
	}

	access, err := s.repo.GetMemberLocationAccess(ctx, ownerID, input.UserID, input.LocationID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			// Not assigned to this location — cannot hold a PIN here.
			return ErrInvalidInput
		}
		return err
	}

	hashed, err := passwordutils.HashPIN(input.PIN)
	if err != nil {
		return err
	}
	input.PIN = ""

	existing, err := s.repo.FindByUserAndLocation(ctx, ownerID, input.UserID, input.LocationID)
	if err != nil && !errors.Is(err, infrastructure.ErrUserNotFound) {
		return err
	}

	if existing != nil {
		// Rotation — keep role + capability overrides, swap the hash only.
		if err := s.repo.UpdatePINHash(ctx, ownerID, existing.ID, hashed); err != nil {
			return err
		}
		invalidatePosUsersScope(ctx, ownerID)
		return nil
	}

	role := roleForAccess(access)
	caps := domain.DefaultCapabilities(role)
	_, err = s.repo.Create(ctx, ownerID, input.UserID, domain.CreatePOSUserInput{
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
	invalidatePosUsersScope(ctx, ownerID)
	return nil
}

// roleForAccess maps a member's location management flag to a POS role. A
// member who can manage POS users at the location (owner/admin/store_manager
// assignment) operates as a manager; everyone else defaults to cashier.
func roleForAccess(access *infrastructure.LocationAccess) domain.POSUserRole {
	if access != nil && access.CanManagePOSUsers {
		return domain.RoleManager
	}
	return domain.RoleCashier
}

// VerifyPIN authenticates a roster operator by (ownerID, locationID, userID,
// pin) and returns a projection carrying the effective capabilities so the POS
// client can gate manager-only actions. A generic domain.ErrInvalidPIN is
// returned for every failure mode so callers cannot distinguish causes.
func (s *Service) VerifyPIN(ctx context.Context, ownerID, locationID, userID int64, pin string) (*domain.VerifiedPOSUser, error) {
	if locationID <= 0 || userID <= 0 || pin == "" {
		return nil, domain.ErrInvalidPIN
	}

	user, err := s.repo.FindForVerify(ctx, ownerID, locationID, userID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return nil, domain.ErrInvalidPIN
		}
		return nil, err
	}

	if user.PINHash == "" || !passwordutils.CheckPINHash(pin, user.PINHash) {
		return nil, domain.ErrInvalidPIN
	}

	return &domain.VerifiedPOSUser{
		UserID:      user.UserID,
		POSUserID:   user.ID,
		LocationID:  user.LocationID,
		Role:        user.Role,
		DisplayName: user.Name,
		Capabilities: domain.Capabilities{
			CanVoid:          user.CanVoid,
			CanRefund:        user.CanRefund,
			CanOpenDrawer:    user.CanOpenDrawer,
			CanApplyDiscount: user.CanApplyDiscount,
			MaxDiscountPct:   user.MaxDiscountPct,
		},
	}, nil
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
