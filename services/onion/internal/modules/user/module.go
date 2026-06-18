package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	"github.com/voxire/onion/internal/modules/user/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/user/internal/routes"
	"github.com/voxire/onion/internal/modules/user/internal/routes/handlers"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/utils/sliceutils"
)

const tokenValidationCacheTTL = 30 * time.Second

type Module struct {
	UserService         *application.Service
	UserRepository      *infrastructure.Repository
	tokenValidator      authPkg.TokenValidator
	ProvisionerResolver authPkg.PlatformProvisionerResolver
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

func (m *Module) GetService() *application.Service {
	return m.UserService
}

func (m *Module) GetTokenValidator() authPkg.TokenValidator {
	return m.tokenValidator
}

// OnUserCreated is an optional hook called after each successful user registration.
// Set it before RegisterRoutes is called (e.g., from main to create personal workspaces).
var OnUserCreated func(ctx context.Context, userID int64)

// OnUserLogin is an optional hook called after each successful login.
// Idempotent operations (e.g. ensure personal workspace exists) should be wired here.
var OnUserLogin func(ctx context.Context, userID int64)

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...any) {
	routes.RegisterRoutes(ctx, router, m.UserService, m.ProvisionerResolver, OnUserCreated, OnUserLogin)
}

func (m *Module) GetAnonymousRoutes() []string {
	return sliceutils.Concat(
		handlers.LoginRoutes(),
		handlers.RegisterRoutes(),
		handlers.GoogleAuthRoutes(),
		handlers.GoogleAuthorizeRoutes(),
		handlers.EmailExistsRoutes(),
	)
}
