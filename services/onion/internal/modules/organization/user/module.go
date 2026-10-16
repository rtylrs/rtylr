package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/organization/user/internal/api"
	"github.com/voxire/onion/internal/modules/organization/user/internal/application"
	userdomain "github.com/voxire/onion/internal/modules/organization/user/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/user/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/hubprofile"
	"github.com/voxire/onion/internal/platform/utils/sliceutils"
	commonauthv1 "github.com/voxire/voxie/gen/go/common/auth/v1"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

const tokenValidationCacheTTL = 30 * time.Second

type Module struct {
	userService         *application.Service
	tokenValidator      authPkg.TokenValidator
	ProvisionerResolver authPkg.PlatformProvisionerResolver
	ProfileResolver     hubprofile.Resolver
	ProvisioningStore   authPkg.ProvisioningStateStore
	inviteAccepter      api.InviteAccepter
}

type UserProfile struct {
	ID   int64
	Name string
}

type EmployeeProvisioner interface {
	EnsureEmployeeForUser(ctx context.Context, ownerID, businessID, userID int64, firstName, lastName, email string) error
}

type userTokenValidator struct {
	userService *application.Service
}

type emailNotificationAdapter struct{ db *sql.DB }

func (a emailNotificationAdapter) SendVerification(ctx context.Context, to, subject, htmlBody, textBody string) error {
	_, err := email.Send(ctx, email.SendParams{To: []string{to}, Subject: subject, HTML: htmlBody, Text: textBody})
	return err
}

func (a emailNotificationAdapter) QueueWelcome(ctx context.Context, to, name string) error {
	return email.EnqueueTemplate(ctx, a.db, email.PriorityNormal, to, "Welcome to rtylr", email.TemplateWelcome,
		email.WelcomeData{Subject: "Welcome to rtylr", Name: name},
		email.WithFrom("Mohamad at rtylr <mohamad@voxire.com>"),
		email.WithReplyTo("mohamad@voxire.com"),
	)
}

func (a emailNotificationAdapter) QueuePasswordReset(ctx context.Context, to, resetLink string) error {
	return email.EnqueueTemplate(ctx, a.db, email.PriorityCritical, to, "Reset your rtylr password", email.TemplatePasswordReset,
		email.PasswordResetData{Subject: "Reset your rtylr password", ResetLink: resetLink})
}

func (v *userTokenValidator) ValidateToken(ctx context.Context, token string) (int64, error) {
	claims, err := v.userService.ValidateToken(ctx, token)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

func (v *userTokenValidator) ValidateTokenFull(ctx context.Context, token string) (*authPkg.TokenClaims, error) {
	claims, err := v.userService.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &authPkg.TokenClaims{
		UserID:         claims.UserID,
		BusinessID:     claims.BusinessID,
		IsConsoleAdmin: claims.IsConsoleAdmin,
		Platform:       claims.Platform,
		PlatformUserID: claims.PlatformUserID,
		PlanID:         claims.PlanID,
		OAuthClientID:  claims.OAuthClientID,
		Scope:          claims.Scope,
		ExpiresAt: func() time.Time {
			if claims.ExpiresAt == nil {
				return time.Time{}
			}
			return claims.ExpiresAt.Time
		}(),
	}, nil
}

func NewModule(ctx context.Context, db *sql.DB, googleClientID string) *Module {
	userRepo := infrastructure.NewRepository(db)
	userService := application.NewService(userRepo, googleClientID).WithEmailNotifications(emailNotificationAdapter{db: db})

	module := &Module{
		userService: userService,
		tokenValidator: authPkg.NewCachingTokenValidator(
			&userTokenValidator{userService: userService},
			tokenValidationCacheTTL,
		),
	}

	userService.Initialize(ctx)

	return module
}

// WithGoogleOAuth configures direct Google OAuth for the auth application.
func (m *Module) WithGoogleOAuth(clientID, clientSecret, redirectURI string) *Module {
	p := application.NewGoogleOAuthProvider(clientID, clientSecret, redirectURI)
	m.userService.WithGoogleOAuthProvider(p)
	return m
}

func (m *Module) WithBusinessProvisioner(provisioner application.BusinessProvisioner) *Module {
	m.userService.WithBusinessProvisioner(provisioner)
	return m
}

func (m *Module) SetEmployeeProvisioner(provisioner EmployeeProvisioner) {
	if provisioner != nil {
		m.userService.WithEmployeeProvisioner(provisioner)
	}
}

func (m *Module) Onboarding() OnboardingPort {
	return m.userService
}

func (m *Module) GetUserProfile(ctx context.Context, userID int64) (UserProfile, error) {
	user, err := m.userService.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return UserProfile{}, ErrUserNotFound
		}
		return UserProfile{}, err
	}
	name := strings.TrimSpace(user.FullName)
	if name == "" {
		name = user.Email
	}
	return UserProfile{ID: user.ID, Name: name}, nil
}

func (m *Module) GetTokenValidator() authPkg.TokenValidator {
	return m.tokenValidator
}

func (m *Module) WithInviteAccepter(accepter api.InviteAccepter) *Module {
	m.inviteAccepter = accepter
	return m
}

func (m *Module) WithProvisioningStore(store authPkg.ProvisioningStateStore) *Module {
	m.ProvisioningStore = store
	return m
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubuserv1.RegisterUserServiceServer(
		api.NewUserAPI(m.userService, m.ProfileResolver, m.ProvisioningStore),
		hubuserv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub user service: %w", err))
	}

	if err := commonauthv1.RegisterAuthServiceServer(
		api.NewAuthAPI(m.userService, m.ProvisionerResolver, m.inviteAccepter, OnUserCreated, OnUserLogin),
		commonauthv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register common auth service: %w", err))
	}
}

// OnUserCreated is an optional hook called after each successful user registration.
// Set it before RegisterRoutes is called (e.g., from main to create personal workspaces).
var OnUserCreated func(ctx context.Context, userID int64)

// OnUserLogin is an optional hook called after each successful login.
// Idempotent operations (e.g. ensure personal workspace exists) should be wired here.
var OnUserLogin func(ctx context.Context, userID int64)

func (m *Module) GetAnonymousRoutes() []string {
	return sliceutils.Concat(
		[]string{
			"/common/auth/v1/register",
			"/common/auth/v1/login",
			"/common/auth/v1/register/google",
			"/common/auth/v1/refresh",
			"/common/auth/v1/revoke",
			"/common/auth/v1/password-reset/request",
			"/common/auth/v1/password-reset/confirm",
			"/common/auth/v1/verify-email",
			"/common/auth/v1/resend-verification",
			"/common/auth/v1/invite/accept",
			"/common/auth/v1/email-exists",
			"/common/auth/v1/google/authorize",
		},
	)
}

var ErrUserNotFound = userdomain.ErrUserNotFound
