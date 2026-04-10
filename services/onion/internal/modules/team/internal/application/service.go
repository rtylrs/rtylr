package application

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

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

func (s *Service) ListStaffUsers(ctx context.Context, ownerID int64) ([]domain.StaffUser, error) {
	return s.repo.ListStaffUsers(ctx, ownerID)
}

// InviteUser creates a user account (no password) and generates an invite token.
// Returns userID and invite token string.
func (s *Service) InviteUser(ctx context.Context, ownerID int64, input domain.InviteUserInput) (int64, string, error) {
	if input.Email == "" {
		return 0, "", errors.New("email is required")
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(input.Email))

	existingID, _, _, err := s.repo.GetUserByEmail(ctx, normalizedEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, "", err
	}

	userID := existingID

	if userID == 0 {
		// Create user without a password (they'll set it via invite link)
		userID, err = s.repo.CreateUser(ctx, normalizedEmail, "", strings.TrimSpace(input.FullName))
		if err != nil {
			return 0, "", err
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
		return 0, "", err
	}

	// Generate invite token (32 bytes = 64 hex chars)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return 0, "", err
	}
	token := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	if err := s.repo.CreateInviteToken(ctx, userID, ownerID, token, expiresAt); err != nil {
		return 0, "", err
	}

	return userID, token, nil
}

// VerifyInviteToken checks if the token is valid and returns the user info.
func (s *Service) VerifyInviteToken(ctx context.Context, token string) (int64, string, string, error) {
	userID, _, _, err := s.repo.GetInviteToken(ctx, token)
	if err != nil {
		return 0, "", "", err
	}

	email, fullName, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return 0, "", "", err
	}

	return userID, email, fullName, nil
}

// AcceptInvite sets the user's password and marks the token as used.
func (s *Service) AcceptInvite(ctx context.Context, token, password string) (int64, error) {
	userID, _, _, err := s.repo.GetInviteToken(ctx, token)
	if err != nil {
		return 0, err
	}

	hash, err := passwordutils.HashPassword(password)
	if err != nil {
		return 0, err
	}

	if err := s.repo.SetPassword(ctx, userID, hash); err != nil {
		return 0, err
	}

	if err := s.repo.MarkInviteTokenUsed(ctx, token); err != nil {
		return 0, err
	}

	return userID, nil
}

// GetMeByBusiness returns user info with permissions looked up by business_id.
func (s *Service) GetMeByBusiness(ctx context.Context, businessID, userID int64) (string, string, []string, domain.Role, error) {
	email, fullName, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", nil, "", err
	}

	perm, err := s.repo.GetPermissionByBusiness(ctx, businessID, userID)
	if err != nil {
		return email, fullName, []string{}, "", nil
	}

	if perm == nil {
		return email, fullName, []string{}, "", nil
	}

	return email, fullName, perm.Apps, perm.Role, nil
}

// GetBusinessInfo returns the business name and whether the user is the owner.
func (s *Service) GetBusinessInfo(ctx context.Context, businessID, userID int64) (string, bool) {
	if businessID == 0 {
		return "", false
	}

	name, ownerID, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return "", false
	}

	return name, ownerID == userID
}

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

func (s *Service) DeactivateUser(ctx context.Context, ownerID, userID int64) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	return s.repo.DeactivateUser(ctx, ownerID, userID)
}

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
