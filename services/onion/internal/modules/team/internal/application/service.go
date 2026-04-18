package application

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/database/dbclient"
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

func (s *Service) GetStaffUserByID(ctx context.Context, ownerID, businessID, userID int64) (*domain.StaffUser, error) {
	if userID == 0 {
		return nil, domain.ErrUserNotFound
	}

	return s.repo.GetStaffUserByID(ctx, ownerID, businessID, userID)
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
		// Hash password if provided (manual auth), otherwise empty (invite link)
		passwordHash := ""
		if input.Password != "" {
			hash, hashErr := passwordutils.HashPassword(input.Password)
			if hashErr != nil {
				return 0, "", hashErr
			}
			passwordHash = hash
		}

		userID, err = s.repo.CreateUser(ctx, normalizedEmail, passwordHash, strings.TrimSpace(input.FullName))
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

// VerifyInviteToken checks if the token is valid and returns the user info along with their app access list.
func (s *Service) VerifyInviteToken(ctx context.Context, token string) (int64, string, string, []string, error) {
	userID, ownerID, _, err := s.repo.GetInviteToken(ctx, token)
	if err != nil {
		return 0, "", "", nil, err
	}

	email, fullName, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return 0, "", "", nil, err
	}

	// Fetch the user's app permissions so the frontend can redirect to the right app after accept.
	apps := []string{}
	if perm, permErr := s.repo.GetPermission(ctx, ownerID, userID); permErr == nil && perm != nil {
		apps = perm.Apps
	}

	return userID, email, fullName, apps, nil
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

// RequestAccess stores a pending access request from a staff user.
func (s *Service) RequestAccess(ctx context.Context, businessID, userID int64, app string) error {
	return s.repo.CreateAccessRequest(ctx, businessID, userID, app)
}

// GetBusinessOwnerEmail returns the email of the business owner.
func (s *Service) GetBusinessOwnerEmail(ctx context.Context, businessID int64) string {
	_, ownerID, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return ""
	}
	email, _, err := s.repo.GetUserByID(ctx, ownerID)
	if err != nil {
		return ""
	}
	return email
}

// GetUserInfo returns email and full name for a user.
func (s *Service) GetUserInfo(ctx context.Context, userID int64) (string, string, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *Service) IsOnboarded(ctx context.Context, userID int64) bool {
	return s.repo.IsOnboarded(ctx, userID)
}

func (s *Service) MarkOnboarded(ctx context.Context, userID int64) error {
	return s.repo.MarkOnboarded(ctx, userID)
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

// GetOwnerApps returns the list of apps the owner is subscribed to.
func (s *Service) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	var appsJSON string
	// cross-schema read from subscriptions + plans — approved for owner app resolution
	err := dbclient.Pool().QueryRowContext(ctx,
		`SELECT p.apps FROM subscriptions s JOIN plans p ON p.id = s.plan_id WHERE s.owner_id = ?`,
		ownerID).Scan(&appsJSON)
	if err != nil {
		return nil, err
	}

	var apps []string
	if jsonErr := json.Unmarshal([]byte(appsJSON), &apps); jsonErr != nil {
		return []string{}, nil
	}

	return apps, nil
}

// GetOwnerSubscriptionInfo returns the plan slug and status for the owner's subscription.
func (s *Service) GetOwnerSubscriptionInfo(ctx context.Context, ownerID int64) (string, string, error) {
	var planSlug, status string
	// cross-schema read from subscriptions + plans — approved for owner subscription info
	err := dbclient.Pool().QueryRowContext(ctx,
		`SELECT p.slug, s.status FROM subscriptions s JOIN plans p ON p.id = s.plan_id WHERE s.owner_id = ?`,
		ownerID).Scan(&planSlug, &status)
	if err != nil {
		return "", "", err
	}

	return planSlug, status, nil
}

// RecordSeatChange updates the billing seat_periods table after a team member is added or removed.
// It fires asynchronously so it never blocks team operations; billing inaccuracies can be corrected
// by a future reconciliation job.
func (s *Service) RecordSeatChange(ctx context.Context, ownerID int64) {
	newCount, err := s.repo.CountActiveMembers(ctx, ownerID)
	if err != nil {
		return
	}

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	billingMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var subscriptionID int64
	var pricePerSeatCents int
	// cross-schema read from subscriptions — approved for seat period management
	if scanErr := dbclient.Pool().QueryRowContext(ctx,
		`SELECT id, price_per_seat_cents FROM subscriptions WHERE owner_id = ?`, ownerID,
	).Scan(&subscriptionID, &pricePerSeatCents); scanErr != nil {
		return
	}

	// cross-schema write to seat_periods — approved for seat billing accuracy
	_, _ = dbclient.Pool().ExecContext(ctx,
		`UPDATE seat_periods SET period_end = ? WHERE subscription_id = ? AND period_end IS NULL`,
		yesterday, subscriptionID)

	_, _ = dbclient.Pool().ExecContext(ctx,
		`INSERT INTO seat_periods (subscription_id, owner_id, seat_count, price_per_seat_cents, period_start, billing_month)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		subscriptionID, ownerID, newCount, pricePerSeatCents, now, billingMonth)

	_, _ = dbclient.Pool().ExecContext(ctx,
		`UPDATE subscriptions SET seat_count = ?, updated_at = CURRENT_TIMESTAMP WHERE owner_id = ?`,
		newCount, ownerID)
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
