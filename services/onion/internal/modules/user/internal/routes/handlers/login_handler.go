package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	"github.com/voxire/onion/internal/modules/user/internal/infrastructure/persistence/sqlc"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func LoginRoutes() []string {
	return []string{
		"/login",
		"/auth/login",
	}
}

type LoginHandler struct {
	UserService         *application.Service
	ProvisionerResolver authPkg.PlatformProvisionerResolver
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
	UserID  int64  `json:"userId"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}

func (h *LoginHandler) Validate(_ context.Context, req *LoginRequest) *errorModels.ServerError {
	if req.Email == "" || req.Password == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Email and password are required",
		}
	}

	return nil
}

func (h *LoginHandler) Serve(ctx context.Context, req *LoginRequest) (LoginResponse, *errorModels.ServerError) {
	userId, err := h.UserService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return LoginResponse{}, &errorModels.ServerError{
				Code:    http.StatusUnauthorized,
				Message: "Invalid email or password",
			}
		}

		return LoginResponse{}, &errorModels.ServerError{
			Code:    http.StatusTeapot,
			Message: "Authentication failed",
		}
	}

	var platformClaims *application.PlatformClaims

	productPlatform := authPkg.ProductPlatformCtx(ctx)
	if h.ProvisionerResolver != nil {
		provisioner := h.ProvisionerResolver.ResolveProvisioner(productPlatform)
		if provisioner != nil {
			pc, resolveErr := provisioner.ResolvePlatformClaims(ctx, userId)
			if resolveErr == nil && pc != nil {
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
		return LoginResponse{}, &errorModels.ServerError{
			Code:    http.StatusTeapot,
			Message: "Failed to generate authentication token",
		}
	}

	user, err := h.UserService.GetUserByID(ctx, userId)
	if err != nil {
		return LoginResponse{}, &errorModels.ServerError{
			Code:    http.StatusTeapot,
			Message: "Failed to retrieve user information",
		}
	}

	// Record login in history (non-blocking, via sqlc)
	go func() {
		ip := ""
		if device := authPkg.DeviceCtx(ctx); device.User != nil {
			ip = device.User.IP
		}
		q := sqlc.New(dbclient.Pool())
		_ = q.RecordLoginHistory(context.Background(), sqlc.RecordLoginHistoryParams{
			UserID:    userId,
			Email:     req.Email,
			Method:    "email",
			IpAddress: sql.NullString{String: ip, Valid: ip != ""},
			Success:   sql.NullBool{Bool: true, Valid: true},
		})
	}()

	return LoginResponse{
		Token:   token,
		Message: "Login successful",
		UserID:  userId,
		Email:   user.Email,
		Name:    user.FullName,
	}, nil
}
