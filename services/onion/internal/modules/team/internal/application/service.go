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
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	meCacheTTL   = 5 * time.Minute
	teamCacheTTL = 5 * time.Minute
)

func invalidateUserScope(ctx context.Context, userID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.UserPrefix(userID))
	}
}

func invalidateBusinessTeamScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.BusinessPrefix(businessID))
	}
}

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func (s *Service) ListStaffUsers(ctx context.Context, ownerID int64) ([]domain.StaffUser, error) {
	return s.repo.ListStaffUsers(ctx, ownerID)
}

// ListStaffUsersByBusiness resolves the business owner and lists all staff
// under that owner. Lets non-owner members see the full team, not just the
// rows they personally own (which would be empty). Result is cached per
// businessID and invalidated by any team mutation under that business.
func (s *Service) ListStaffUsersByBusiness(ctx context.Context, businessID int64) ([]domain.StaffUser, error) {
	return cacher.Of[[]domain.StaffUser]().GetOrComputeSingle(
		ctx,
		cacheKeys.BusinessTeam(businessID),
		func(ctx context.Context) ([]domain.StaffUser, error) {
			_, ownerID, err := s.repo.GetBusinessByID(ctx, businessID)
			if err != nil {
				return nil, err
			}
			return s.repo.ListStaffUsers(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, teamCacheTTL),
	)
}

// GetMyProfile builds the composite /user/me payload (identity + business +
// subscription + effective resources). Cached per-user; any team or
// subscription mutation under the user invalidates `cache:user:{id}:*`.
func (s *Service) GetMyProfile(ctx context.Context, userID, businessID int64) (domain.MyProfile, error) {
	return cacher.Of[domain.MyProfile]().GetOrComputeSingle(
		ctx,
		cacheKeys.UserMe(userID),
		func(ctx context.Context) (domain.MyProfile, error) {
			return s.computeMyProfile(ctx, userID, businessID)
		},
		cacher.WithExpiry(ctx, meCacheTTL),
	)
}

func (s *Service) computeMyProfile(ctx context.Context, userID, businessID int64) (domain.MyProfile, error) {
	email, fullName, apps, role, err := s.GetMeByBusiness(ctx, businessID, userID)
	if err != nil {
		return domain.MyProfile{}, err
	}
	if apps == nil {
		apps = []string{}
	}

	businessName, isOwner := s.GetBusinessInfo(ctx, businessID, userID)

	var planSlug, subscriptionStatus string
	if isOwner {
		role = "owner"
		ownerApps, appsErr := s.GetOwnerApps(ctx, userID)
		if appsErr != nil || len(ownerApps) == 0 {
			apps = []string{"pos", "erp", "hr", "crm", "finance", "flow"}
		} else {
			apps = ownerApps
		}
		if slug, status, subErr := s.GetOwnerSubscriptionInfo(ctx, userID); subErr == nil {
			planSlug = slug
			subscriptionStatus = status
		}
	}

	effectiveResources := map[string]string{}
	if !isOwner {
		if resPerms, resErr := s.GetUserEffectiveResources(ctx, businessID, userID); resErr == nil {
			for k, v := range resPerms {
				effectiveResources[k] = string(v)
			}
		}
	}

	return domain.MyProfile{
		ID:                 userID,
		Email:              email,
		FullName:           fullName,
		Apps:               apps,
		Role:               string(role),
		BusinessID:         businessID,
		BusinessName:       businessName,
		IsOwner:            isOwner,
		Onboarded:          s.IsOnboarded(ctx, userID),
		PlanSlug:           planSlug,
		SubscriptionStatus: subscriptionStatus,
		EffectiveResources: effectiveResources,
	}, nil
}

func (s *Service) GetStaffUserByID(ctx context.Context, ownerID, businessID, userID int64) (*domain.StaffUser, error) {
	if userID == 0 {
		return nil, domain.ErrUserNotFound
	}

	return s.repo.GetStaffUserByID(ctx, ownerID, businessID, userID)
}

// InviteUser creates a user account (no password) and generates an invite token.
// Returns userID and invite token string.
//
// businessID is passed in (rather than looked up) so the service can invalidate
// the business-scoped team list cache without an extra query. Tests that don't
// care about the cache key may pass 0.
func (s *Service) InviteUser(ctx context.Context, ownerID, businessID int64, input domain.InviteUserInput) (int64, string, error) {
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

	// New team member changes the business's team list cache + the new user's
	// /me payload (apps/permissions just got attached).
	invalidateUserScope(ctx, userID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
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
	if err := s.repo.MarkOnboarded(ctx, userID); err != nil {
		return err
	}
	invalidateUserScope(ctx, userID)
	return nil
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

func (s *Service) UpdatePermissions(ctx context.Context, ownerID, businessID int64, input domain.UpdatePermissionsInput) error {
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

	if err := s.repo.UpsertPermission(ctx, ownerID, input.UserID, apps, role); err != nil {
		return err
	}

	invalidateUserScope(ctx, input.UserID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}
	return nil
}

func (s *Service) DeactivateUser(ctx context.Context, ownerID, businessID, userID int64) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	if err := s.repo.DeactivateUser(ctx, ownerID, userID); err != nil {
		return err
	}

	invalidateUserScope(ctx, userID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}
	return nil
}

// GetOwnerApps returns the list of apps the owner is subscribed to.
func (s *Service) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	return s.repo.GetOwnerApps(ctx, ownerID)
}

// GetOwnerSubscriptionInfo returns the plan slug and status for the owner's subscription.
func (s *Service) GetOwnerSubscriptionInfo(ctx context.Context, ownerID int64) (string, string, error) {
	return s.repo.GetOwnerSubscriptionInfo(ctx, ownerID)
}

// RecordSeatChange updates the billing seat_periods table after a team member is added or removed.
// It fires asynchronously so it never blocks team operations; billing inaccuracies can be corrected
// by a future reconciliation job.
func (s *Service) RecordSeatChange(ctx context.Context, ownerID int64) {
	newCount, err := s.repo.CountActiveMembers(ctx, ownerID)
	if err != nil {
		return
	}

	_ = s.repo.RecordSeatChange(ctx, ownerID, newCount)
}

// ─── Groups ──────────────────────────────────────────────────────────────────

func (s *Service) CreateGroup(ctx context.Context, ownerID int64, input domain.CreateGroupInput) (int64, error) {
	if input.Name == "" {
		return 0, errors.New("name is required")
	}
	if input.Color == "" {
		input.Color = "#6366f1"
	}
	return s.repo.CreateGroup(ctx, ownerID, input)
}

func (s *Service) ListGroups(ctx context.Context, ownerID int64) ([]domain.Group, error) {
	return s.repo.ListGroups(ctx, ownerID)
}

func (s *Service) GetGroup(ctx context.Context, ownerID, groupID int64) (*domain.Group, error) {
	return s.repo.GetGroup(ctx, ownerID, groupID)
}

func (s *Service) GetGroupDetail(ctx context.Context, ownerID, groupID int64) (*domain.GroupDetail, error) {
	return s.repo.GetGroupDetail(ctx, ownerID, groupID)
}

func (s *Service) UpdateGroup(ctx context.Context, ownerID, groupID int64, input domain.UpdateGroupInput) error {
	if input.Name == "" {
		return errors.New("name is required")
	}
	return s.repo.UpdateGroup(ctx, ownerID, groupID, input)
}

func (s *Service) DeleteGroup(ctx context.Context, ownerID, groupID int64) error {
	return s.repo.DeleteGroup(ctx, ownerID, groupID)
}

func (s *Service) AddGroupMembers(ctx context.Context, ownerID, groupID, callerID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return errors.New("no user IDs provided")
	}
	return s.repo.AddGroupMembers(ctx, ownerID, groupID, callerID, userIDs)
}

func (s *Service) RemoveGroupMember(ctx context.Context, ownerID, groupID, userID int64) error {
	if _, err := s.repo.GetGroup(ctx, ownerID, groupID); err != nil {
		return domain.ErrPermissionNotFound
	}
	return s.repo.RemoveGroupMember(ctx, groupID, userID)
}

func (s *Service) GetUserGroups(ctx context.Context, ownerID, userID int64) ([]domain.GroupSummary, error) {
	return s.repo.GetUserGroups(ctx, ownerID, userID)
}

func (s *Service) GetUserEffectiveResources(ctx context.Context, ownerID, userID int64) (map[string]domain.PermissionLevel, error) {
	return s.repo.GetUserEffectiveResources(ctx, ownerID, userID)
}

// BulkInvite invites multiple users with the same role/apps/group in one call.
func (s *Service) BulkInvite(ctx context.Context, ownerID, businessID int64, input domain.BulkInviteInput) ([]domain.BulkInviteResult, error) {
	var results []domain.BulkInviteResult
	for _, email := range input.Emails {
		email = strings.ToLower(strings.TrimSpace(email))
		if email == "" {
			continue
		}
		userID, token, err := s.InviteUser(ctx, ownerID, businessID, domain.InviteUserInput{
			Email: email,
			Apps:  input.Apps,
			Role:  input.Role,
		})
		if err != nil {
			results = append(results, domain.BulkInviteResult{
				Email:   email,
				Status:  "error",
				Message: err.Error(),
			})
			continue
		}
		if input.GroupID > 0 {
			_ = s.repo.AddGroupMembers(ctx, ownerID, input.GroupID, ownerID, []int64{userID})
		}
		status := "invited"
		if token == "" {
			status = "existing"
		}
		results = append(results, domain.BulkInviteResult{
			Email:   email,
			UserID:  userID,
			Token:   token,
			Status:  status,
			Message: "ok",
		})
	}
	return results, nil
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
