package application

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

// ListStaffUsers returns all staff users for the given owner with their app permissions.
func (s *Service) ListStaffUsers(ctx context.Context, ownerID int64) ([]domain.StaffUser, error) {
	return s.repo.ListStaffUsers(ctx, ownerID)
}

// InviteUser creates a new user account and sets their app permissions.
func (s *Service) InviteUser(ctx context.Context, ownerID int64, input domain.InviteUserInput) (int64, error) {
	if input.Email == "" || input.Password == "" {
		return 0, errors.New("email and password are required")
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(input.Email))

	// Check if user already exists; if so, just set permissions.
	existingID, _, _, err := s.repo.GetUserByEmail(ctx, normalizedEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	userID := existingID

	if userID == 0 {
		hash, hashErr := passwordutils.HashPassword(input.Password)
		if hashErr != nil {
			return 0, hashErr
		}

		userID, err = s.repo.CreateUser(ctx, normalizedEmail, hash, strings.TrimSpace(input.FullName))
		if err != nil {
			return 0, err
		}
	}

	role := input.Role
	if role == "" {
		role = domain.RoleStaff
	}

	apps := input.Apps
	if apps == nil {
		apps = []string{}
	}

	if err := s.repo.UpsertPermission(ctx, ownerID, userID, apps, role); err != nil {
		return 0, err
	}

	return userID, nil
}

// UpdatePermissions updates a user's app list and role.
func (s *Service) UpdatePermissions(ctx context.Context, ownerID int64, input domain.UpdatePermissionsInput) error {
	if input.UserID == 0 {
		return errors.New("user id is required")
	}

	role := input.Role
	if role == "" {
		role = domain.RoleStaff
	}

	apps := input.Apps
	if apps == nil {
		apps = []string{}
	}

	return s.repo.UpsertPermission(ctx, ownerID, input.UserID, apps, role)
}

// DeactivateUser marks a user as inactive for the given owner.
func (s *Service) DeactivateUser(ctx context.Context, ownerID, userID int64) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	return s.repo.DeactivateUser(ctx, ownerID, userID)
}

// GetMe returns the current user's profile plus their apps for the given owner.
func (s *Service) GetMe(ctx context.Context, ownerID, userID int64) (string, string, []string, domain.Role, error) {
	email, fullName, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", nil, "", err
	}

	perm, err := s.repo.GetPermission(ctx, ownerID, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return email, fullName, []string{}, "", nil
	}

	if perm == nil {
		return email, fullName, []string{}, "", nil
	}

	return email, fullName, perm.Apps, perm.Role, nil
}
