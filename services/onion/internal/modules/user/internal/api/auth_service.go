package api

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"strings"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	commonauthv1 "github.com/voxire/voxie/gen/go/common/auth/v1"
)

type InviteAccepter interface {
	AcceptInvite(ctx context.Context, token, password string) (int64, error)
}

type AuthAPI struct {
	UserService         *application.Service
	ProvisionerResolver authPkg.PlatformProvisionerResolver
	InviteAccepter      InviteAccepter
	OnUserCreated       func(ctx context.Context, userID int64)
	OnUserLogin         func(ctx context.Context, userID int64)
}

func NewAuthAPI(
	userService *application.Service,
	provisionerResolver authPkg.PlatformProvisionerResolver,
	inviteAccepter InviteAccepter,
	onUserCreated func(ctx context.Context, userID int64),
	onUserLogin func(ctx context.Context, userID int64),
) *AuthAPI {
	return &AuthAPI{
		UserService:         userService,
		ProvisionerResolver: provisionerResolver,
		InviteAccepter:      inviteAccepter,
		OnUserCreated:       onUserCreated,
		OnUserLogin:         onUserLogin,
	}
}

func (a *AuthAPI) Register(ctx context.Context, req *commonauthv1.RegisterRequest) (*commonauthv1.RegisterResponse, error) {
	registerData := application.RegisterUserData{
		Email:        strings.TrimSpace(req.GetEmail()),
		Password:     req.GetPassword(),
		FullName:     strings.TrimSpace(req.GetName()),
		BusinessName: strings.TrimSpace(req.GetBusinessName()),
	}

	userID, err := a.UserService.Register(ctx, registerData)
	if err != nil {
		return nil, authErrorFromError(err, "Registration failed")
	}

	planSlug := application.DefaultSignupPlanSlug
	if req.PlanSlug != nil && strings.TrimSpace(*req.PlanSlug) != "" {
		planSlug = strings.TrimSpace(*req.PlanSlug)
	}
	if subErr := a.UserService.AssignSubscription(ctx, userID, planSlug); subErr != nil {
		logger.Ctx(ctx).Warn("failed to assign subscription on registration",
			logger.String("plan_slug", planSlug),
			logger.Error(subErr),
		)
	}

	if a.OnUserCreated != nil {
		runAsyncUserHook(ctx, userID, a.OnUserCreated)
	}

	platformClaims := a.provisionPlatformClaims(ctx, userID)
	token, err := a.UserService.GenerateTokenWithPlatform(ctx, userID, platformClaims)
	if err != nil {
		return nil, onekithttp.NewError(http.StatusInternalServerError, "Failed to generate authentication token")
	}

	return &commonauthv1.RegisterResponse{
		Token:   token,
		Message: "Registration successful",
		UserId:  userID,
		Email:   registerData.Email,
		Name:    registerData.FullName,
	}, nil
}

func (a *AuthAPI) Login(ctx context.Context, req *commonauthv1.LoginRequest) (*commonauthv1.LoginResponse, error) {
	if strings.TrimSpace(req.GetEmail()) == "" || req.GetPassword() == "" {
		return nil, onekithttp.NewError(http.StatusBadRequest, "Email and password are required")
	}

	userID, err := a.UserService.Authenticate(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return nil, onekithttp.NewError(http.StatusUnauthorized, "Invalid email or password")
		}
		logger.Ctx(ctx).Error("login: Authenticate failed", logger.Error(err))
		return nil, onekithttp.NewError(http.StatusInternalServerError, "Authentication failed")
	}

	platformClaims := a.resolvePlatformClaims(ctx, userID)
	token, err := a.UserService.GenerateTokenWithPlatform(ctx, userID, platformClaims)
	if err != nil {
		logger.Ctx(ctx).Error("login: GenerateTokenWithPlatform failed", logger.Error(err))
		return nil, onekithttp.NewError(http.StatusInternalServerError, "Failed to generate authentication token")
	}

	user, err := a.UserService.GetUserByID(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("login: GetUserByID failed", logger.Error(err))
		return nil, onekithttp.NewError(http.StatusInternalServerError, "Failed to retrieve user information")
	}

	device := authPkg.DeviceCtx(ctx)
	runAsyncLoginTasks(ctx, a.UserService, userID, req.GetEmail(), device.IP, device.UserAgent, a.OnUserLogin)

	return &commonauthv1.LoginResponse{
		Token:          token,
		Message:        "Login successful",
		UserId:         userID,
		Email:          user.Email,
		Name:           user.FullName,
		IsConsoleAdmin: user.IsConsoleAdmin,
	}, nil
}

func (a *AuthAPI) GoogleAuth(ctx context.Context, req *commonauthv1.GoogleAuthRequest) (*commonauthv1.GoogleAuthResponse, error) {
	if strings.TrimSpace(req.GetCode()) == "" {
		return nil, onekithttp.NewError(http.StatusBadRequest, "code is required")
	}

	// The register flow records the signup IP and device for abuse review;
	// Google signups are handed the same device context the login path reads,
	// instead of a zero Addr that stringifies to the literal "invalid IP".
	device := authPkg.DeviceCtx(ctx)
	signupIP, ipErr := netip.ParseAddr(device.IP)
	if ipErr != nil {
		logger.Ctx(ctx).Warn("google auth: could not parse client IP",
			logger.String("ip", device.IP),
			logger.Error(ipErr),
		)
	}

	businessName := ""
	if req.BusinessName != nil {
		businessName = strings.TrimSpace(*req.BusinessName)
	}

	result, err := a.UserService.GoogleAuth(ctx, req.GetCode(), businessName, signupIP, device.DeviceInfo)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrGoogleOAuthNotConfigured):
			return nil, onekithttp.NewError(http.StatusServiceUnavailable, "Google login is not configured on this server")
		case errors.Is(err, application.ErrGoogleOAuthUnavailable):
			return nil, onekithttp.NewError(http.StatusServiceUnavailable, "Google login is temporarily unavailable")
		case errors.Is(err, application.ErrGoogleEmailUnverified):
			return nil, onekithttp.NewError(http.StatusForbidden, "Your Google email is not verified")
		case errors.Is(err, application.ErrGoogleTokenInvalid):
			return nil, onekithttp.NewError(http.StatusUnauthorized, "Google code is invalid or expired")
		default:
			return nil, onekithttp.NewError(http.StatusInternalServerError, "Google authentication failed")
		}
	}

	if result.IsNew && a.OnUserCreated != nil {
		runAsyncUserHook(ctx, result.UserID, a.OnUserCreated)
	}
	if !result.IsNew && a.OnUserLogin != nil {
		runAsyncUserHook(ctx, result.UserID, a.OnUserLogin)
	}

	return &commonauthv1.GoogleAuthResponse{
		Token: result.Token,
		IsNew: result.IsNew,
	}, nil
}

func (a *AuthAPI) ResetPassword(ctx context.Context, req *commonauthv1.ResetPasswordRequest) (*commonauthv1.ResetPasswordResponse, error) {
	if err := a.UserService.RequestPasswordReset(ctx, req.GetEmail()); err != nil {
		logger.Ctx(ctx).Error("password reset request failed",
			logger.String("email", req.GetEmail()),
			logger.Error(err),
		)
	}
	return &commonauthv1.ResetPasswordResponse{
		Message: "If that email is registered, a reset link has been sent.",
	}, nil
}

func (a *AuthAPI) ResetPasswordConfirm(ctx context.Context, req *commonauthv1.ResetPasswordConfirmRequest) (*commonauthv1.ResetPasswordConfirmResponse, error) {
	err := a.UserService.ResetPassword(ctx, req.GetToken(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, application.ErrResetTokenNotFound):
			return nil, onekithttp.NewError(http.StatusBadRequest, "Reset link is invalid or has expired. Please request a new one.")
		case errors.Is(err, application.ErrWeakPassword):
			return nil, onekithttp.NewError(http.StatusBadRequest, "Password is too weak. Use at least 8 characters.")
		default:
			return nil, onekithttp.NewError(http.StatusInternalServerError, "Failed to reset password.")
		}
	}
	return &commonauthv1.ResetPasswordConfirmResponse{Message: "Password reset successfully."}, nil
}

func (a *AuthAPI) VerifyEmail(ctx context.Context, req *commonauthv1.VerifyEmailRequest) (*commonauthv1.VerifyEmailResponse, error) {
	if strings.TrimSpace(req.GetToken()) == "" {
		return nil, onekithttp.NewError(http.StatusBadRequest, "token is required")
	}
	if err := a.UserService.VerifyEmailCode(ctx, req.GetToken()); err != nil {
		switch {
		case errors.Is(err, application.ErrEmailCodeSuperseded):
			return nil, onekithttp.NewError(http.StatusBadRequest, "This verification code is too old. Please request a new one.")
		case errors.Is(err, application.ErrTooManyEmailCodeAttempts):
			return nil, onekithttp.NewError(http.StatusTooManyRequests, "Too many attempts. Please request a new verification code.")
		default:
			return nil, onekithttp.NewError(http.StatusBadRequest, "Invalid or expired email verification code")
		}
	}
	return &commonauthv1.VerifyEmailResponse{Verified: true, Message: "Email verified"}, nil
}

func (a *AuthAPI) ResendVerification(ctx context.Context, req *commonauthv1.ResendVerificationRequest) (*commonauthv1.ResendVerificationResponse, error) {
	if strings.TrimSpace(req.GetEmail()) == "" {
		return nil, onekithttp.NewError(http.StatusBadRequest, "email is required")
	}
	if err := a.UserService.IssueEmailVerificationCode(ctx, req.GetEmail()); err != nil {
		logger.Ctx(ctx).Warn("email verification code could not be issued", logger.Error(err))
	}
	return &commonauthv1.ResendVerificationResponse{
		Ok:      true,
		Message: "If that email is registered, a verification email has been sent.",
	}, nil
}

func (a *AuthAPI) InviteAccept(ctx context.Context, req *commonauthv1.InviteAcceptRequest) (*commonauthv1.InviteAcceptResponse, error) {
	if a.InviteAccepter == nil {
		return nil, onekithttp.NewError(http.StatusInternalServerError, "Invite accept is not configured")
	}
	if strings.TrimSpace(req.GetToken()) == "" {
		return nil, onekithttp.NewError(http.StatusBadRequest, "Token is required")
	}
	if len(req.GetPassword()) < 6 {
		return nil, onekithttp.NewError(http.StatusBadRequest, "Password must be at least 6 characters")
	}

	_, err := a.InviteAccepter.AcceptInvite(ctx, req.GetToken(), req.GetPassword())
	if err != nil {
		return nil, onekithttp.NewError(http.StatusBadRequest, "Invalid or expired invitation")
	}

	return &commonauthv1.InviteAcceptResponse{Message: "Password set successfully. You can now sign in."}, nil
}

func (a *AuthAPI) CheckEmailExists(ctx context.Context, req *commonauthv1.CheckEmailExistsRequest) (*commonauthv1.CheckEmailExistsResponse, error) {
	exists, err := a.UserService.EmailExists(ctx, req.GetEmail())
	if err != nil {
		logger.Ctx(ctx).Error("email exists check failed", logger.Error(err))
		return nil, onekithttp.NewError(http.StatusInternalServerError, "Failed to check email")
	}

	return &commonauthv1.CheckEmailExistsResponse{Exists: exists}, nil
}

func (a *AuthAPI) SocialAuthorize(_ context.Context, req *commonauthv1.SocialAuthorizeRequest) (*commonauthv1.SocialAuthorizeResponse, error) {
	url, err := a.UserService.SocialAuthorizationURL(req.GetProvider())
	if err != nil {
		if errors.Is(err, application.ErrGoogleOAuthNotConfigured) {
			return nil, onekithttp.NewError(http.StatusServiceUnavailable, "Social login is not configured on this server")
		}
		return nil, onekithttp.NewError(http.StatusBadRequest, "Failed to generate authorization URL: "+err.Error())
	}

	return &commonauthv1.SocialAuthorizeResponse{Url: url}, nil
}

func authErrorFromError(err error, fallbackMessage string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	lowerErr := strings.ToLower(errStr)

	switch {
	case errors.Is(err, application.ErrBusinessNameTaken):
		// A taken business name is a user-correctable conflict; without this it
		// fell through to the 500 default because the message does not contain
		// "duplicate" or "already exists".
		return onekithttp.NewError(http.StatusConflict, errStr)
	case errStr == "sql: no rows in result set" || strings.Contains(lowerErr, "not found"):
		return onekithttp.NewError(http.StatusNotFound, errStr)
	case strings.Contains(lowerErr, "duplicate") || strings.Contains(lowerErr, "already exists"):
		return onekithttp.NewError(http.StatusConflict, errStr)
	case strings.Contains(lowerErr, "invalid") ||
		strings.Contains(lowerErr, "weak") ||
		strings.Contains(lowerErr, "required") ||
		strings.Contains(lowerErr, "validation"):
		return onekithttp.NewError(http.StatusBadRequest, errStr)
	case strings.Contains(lowerErr, "denied") || strings.Contains(lowerErr, "unauthorized"):
		return onekithttp.NewError(http.StatusForbidden, errStr)
	default:
		if fallbackMessage == "" {
			fallbackMessage = errStr
		}
		return onekithttp.NewError(http.StatusInternalServerError, fallbackMessage)
	}
}

func (a *AuthAPI) provisionPlatformClaims(ctx context.Context, userID int64) *application.PlatformClaims {
	if a.ProvisionerResolver == nil {
		return nil
	}
	productPlatform := authPkg.ProductPlatformCtx(ctx)
	provisioner := a.ProvisionerResolver.ResolveProvisioner(productPlatform)
	if provisioner == nil {
		return nil
	}
	pc, err := provisioner.ProvisionUser(ctx, userID)
	if err != nil || pc == nil {
		return nil
	}
	return &application.PlatformClaims{
		Platform:       pc.Platform,
		PlatformUserID: pc.PlatformUserID,
		PlanID:         pc.PlanID,
	}
}

func (a *AuthAPI) resolvePlatformClaims(ctx context.Context, userID int64) *application.PlatformClaims {
	if a.ProvisionerResolver == nil {
		return nil
	}
	productPlatform := authPkg.ProductPlatformCtx(ctx)
	provisioner := a.ProvisionerResolver.ResolveProvisioner(productPlatform)
	if provisioner == nil {
		return nil
	}
	pc, err := provisioner.ResolvePlatformClaims(ctx, userID)
	if err != nil || pc == nil {
		return nil
	}
	return &application.PlatformClaims{
		Platform:       pc.Platform,
		PlatformUserID: pc.PlatformUserID,
		PlanID:         pc.PlanID,
	}
}
