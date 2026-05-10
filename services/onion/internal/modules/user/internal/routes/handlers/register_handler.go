package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	authv1 "github.com/voxire/voxie/gen/go/auth/v1"
)

func RegisterRoutes() []string {
	return []string{
		"/register",
		"/auth/register",
	}
}

type RegisterHandler struct {
	UserService         *application.Service
	ProvisionerResolver authPkg.PlatformProvisionerResolver
	// OnUserCreated is called non-fatally after a user is successfully registered.
	// Used to create a personal workspace, provision initial data, etc.
	OnUserCreated func(ctx context.Context, userID int64)
}

type RegisterRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	BusinessName string `json:"business_name"`
	PlanSlug     string `json:"plan_slug"`
}

type RegisterResponse = authv1.RegisterResponse

func (h *RegisterHandler) Validate(_ context.Context, req *RegisterRequest) *errorModels.ServerError {
	if req.Email == "" || req.Password == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Email and password are required",
		}
	}

	if err := h.UserService.ValidateEmail(req.Email); err != nil {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid email format",
		}
	}

	if err := h.UserService.ValidatePassword(req.Password); err != nil {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Password must be at least 8 characters long and contain uppercase, lowercase, and numeric characters",
		}
	}

	return nil
}

func (h *RegisterHandler) Serve(ctx context.Context, req *RegisterRequest) (RegisterResponse, *errorModels.ServerError) {
	var clientIP netip.Addr

	var deviceInfo string

	registerData := application.RegisterUserData{
		Email:           strings.TrimSpace(req.Email),
		Password:        req.Password,
		FullName:        strings.TrimSpace(req.Name),
		SignupIPAddress: clientIP,
		DeviceInfo:      deviceInfo,
	}

	userId, err := h.UserService.Register(ctx, registerData)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrUserAlreadyExists):
			return RegisterResponse{}, &errorModels.ServerError{
				Code:    http.StatusConflict,
				Message: "User with this email already exists",
			}
		case errors.Is(err, application.ErrInvalidEmail):
			return RegisterResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadRequest,
				Message: "Invalid email format",
			}
		case errors.Is(err, application.ErrWeakPassword):
			return RegisterResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadRequest,
				Message: "Password does not meet security requirements",
			}
		default:
			return RegisterResponse{}, &errorModels.ServerError{
				Code:    http.StatusTeapot,
				Message: "Registration failed",
			}
		}
	}

	// Create business for this user (they are the owner)
	businessName := strings.TrimSpace(req.BusinessName)
	if businessName == "" {
		businessName = strings.TrimSpace(req.Name) + "'s Business"
	}

	if bizErr := h.UserService.CreateBusinessForUser(ctx, userId, businessName); bizErr != nil {
		// Non-fatal: user is created but business creation failed
		// Log it but don't block registration
		_ = bizErr
	}

	// Assign subscription plan after registration
	planSlug := req.PlanSlug
	if planSlug == "" {
		planSlug = "starter"
	}
	if subErr := assignSubscription(ctx, userId, planSlug); subErr != nil {
		logger.Ctx(ctx).Warn("failed to assign subscription on registration",
			logger.String("plan_slug", planSlug),
			logger.Error(subErr),
		)
	}

	// Create personal workspace for the new user (non-fatal)
	if h.OnUserCreated != nil {
		h.OnUserCreated(ctx, userId)
	}

	var platformClaims *application.PlatformClaims

	productPlatform := authPkg.ProductPlatformCtx(ctx)
	if h.ProvisionerResolver != nil {
		provisioner := h.ProvisionerResolver.ResolveProvisioner(productPlatform)
		if provisioner != nil {
			pc, provisionErr := provisioner.ProvisionUser(ctx, userId)
			if provisionErr == nil && pc != nil {
				platformClaims = &application.PlatformClaims{
					Platform:       pc.Platform,
					PlatformUserID: pc.PlatformUserID,
					PlanID:         pc.PlanID,
				}
			}
		}
	}

	token, err := h.UserService.GenerateTokenWithPlatform(ctx, userId, platformClaims)
	if err != nil {
		return RegisterResponse{}, &errorModels.ServerError{
			Code:    http.StatusTeapot,
			Message: "Failed to generate authentication token",
		}
	}

	return RegisterResponse{
		Token:   token,
		Message: "Registration successful",
		UserId:  userId,
		Email:   registerData.Email,
		Name:    registerData.FullName,
	}, nil
}

// assignSubscription looks up the plan by slug and creates a trial subscription for the owner.
func assignSubscription(ctx context.Context, ownerID int64, planSlug string) error {
	var planID int64
	// cross-schema read from plans — approved for subscription assignment on registration
	err := dbclient.Pool().QueryRowContext(ctx,
		`SELECT id FROM plans WHERE slug = ? AND is_active = TRUE`,
		planSlug).Scan(&planID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("plan not found: " + planSlug)
		}

		return err
	}

	trialEnd := time.Now().AddDate(0, 0, 14)
	_, execErr := dbclient.Pool().ExecContext(ctx,
		`INSERT INTO subscriptions (owner_id, plan_id, status, trial_ends_at)
		 VALUES (?, ?, 'trial', ?)
		 ON DUPLICATE KEY UPDATE plan_id = VALUES(plan_id), updated_at = CURRENT_TIMESTAMP`,
		ownerID, planID, trialEnd)

	return execErr
}
