package user

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/api"
	"github.com/voxire/onion/internal/modules/user/internal/application"
	"github.com/voxire/onion/internal/modules/user/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/hubprofile"
	"github.com/voxire/onion/internal/platform/utils/sliceutils"
	commonauthv1 "github.com/voxire/voxie/gen/go/common/auth/v1"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

const tokenValidationCacheTTL = 30 * time.Second

type Module struct {
	UserService         *application.Service
	UserRepository      *infrastructure.Repository
	tokenValidator      authPkg.TokenValidator
	ProvisionerResolver authPkg.PlatformProvisionerResolver
	ProfileResolver     hubprofile.Resolver
	ProvisioningStore   authPkg.ProvisioningStateStore
	inviteAccepter      api.InviteAccepter
}

type userTokenValidator struct {
	userService *application.Service
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
		Role:           claims.Role,
		Platform:       claims.Platform,
		PlatformUserID: claims.PlatformUserID,
		PlanID:         claims.PlanID,
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
	userService := application.NewService(ctx, db, googleClientID)

	module := &Module{
		UserService:    userService,
		UserRepository: userRepo,
		tokenValidator: authPkg.NewCachingTokenValidator(
			&userTokenValidator{userService: userService},
			tokenValidationCacheTTL,
		),
	}

	userService.Initialize(ctx)

	return module
}

// WithWorkOS configures the module to use WorkOS for Google OAuth.
// apiKey, clientID, and redirectURI come from the WorkOS dashboard.
func (m *Module) WithWorkOS(apiKey, clientID, redirectURI string) *Module {
	p := application.NewWorkOSProvider(apiKey, clientID, redirectURI)
	m.UserService.WithWorkOSProvider(p)
	return m
}

func (m *Module) WithBusinessProvisioner(provisioner application.BusinessProvisioner) *Module {
	m.UserService.WithBusinessProvisioner(provisioner)
	return m
}

func (m *Module) GetService() *application.Service {
	return m.UserService
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
		api.NewUserAPI(m.UserService, m.ProfileResolver, m.ProvisioningStore),
		hubuserv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub user service: %w", err))
	}

	if err := commonauthv1.RegisterAuthServiceServer(
		api.NewAuthAPI(m.UserService, m.ProvisionerResolver, m.inviteAccepter, OnUserCreated, OnUserLogin),
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
			"/common/auth/v1/password-reset/request",
			"/common/auth/v1/password-reset/confirm",
			"/common/auth/v1/verify-email",
			"/common/auth/v1/resend-verification",
			"/common/auth/v1/invite/accept",
			"/common/auth/v1/email-exists",
			"/common/auth/v1/social/authorize",
		},
	)
}
