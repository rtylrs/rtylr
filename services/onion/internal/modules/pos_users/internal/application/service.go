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
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
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
	}
	invalidatePosUsersScope(ctx, businessID)
	return nil
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
	invalidatePosUsersScope(ctx, businessID)
	return nil
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
		if app.AppID != authorization.AppPOS || !app.HasAccess {
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

func (s *Service) VerifyPIN(ctx context.Context, businessID, locationID, userID int64, pin string) (*domain.VerifiedPOSUser, error) {
	pin = strings.TrimSpace(pin)
	if locationID <= 0 || userID <= 0 || pin == "" {
		return nil, domain.ErrInvalidPIN
	}

	user, err := s.repo.FindForVerify(ctx, businessID, locationID, userID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return nil, domain.ErrInvalidPIN
		}
		return nil, err
	}

	hash, err := s.repo.GetPINHash(ctx, businessID, user.ID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return nil, domain.ErrInvalidPIN
		}
		return nil, err
	}
	if hash == "" || !passwordutils.CheckPINHash(pin, hash) {
		return nil, domain.ErrInvalidPIN
	}

	authorizationVersion, err := s.repo.GetAuthorizationVersion(ctx, businessID, user.UserID)
	if err != nil {
		return nil, err
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
