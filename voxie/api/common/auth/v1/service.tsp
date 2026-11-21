import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace common.auth.v1;

/** AuthService defines the public authentication endpoints. */
@route("/common/auth/v1")
interface AuthService {
  /** Register creates a new owner account and optional first business. */
  @post
  @route("/register")
  register(...RegisterRequest): RegisterResponse;

  /** Login authenticates an email and password pair. */
  @post
  @route("/login")
  login(...LoginRequest): LoginResponse;

  /** GoogleAuth completes social authentication with an OAuth code. */
  @post
  @route("/register/google")
  googleAuth(...GoogleAuthRequest): GoogleAuthResponse;

  /** Refresh renews a short-lived access token for an active session. */
  @post
  @route("/refresh")
  refresh(...RefreshRequest): RefreshResponse;

  /** Revoke ends the supplied first-party session. */
  @post
  @route("/revoke")
  revoke(...RevokeSessionRequest): RevokeSessionResponse;

  /** ResetPassword starts a password reset flow. */
  @post
  @route("/password-reset/request")
  resetPassword(...ResetPasswordRequest): ResetPasswordResponse;

  /** ResetPasswordConfirm completes a password reset flow. */
  @post
  @route("/password-reset/confirm")
  resetPasswordConfirm(
    ...ResetPasswordConfirmRequest,
  ): ResetPasswordConfirmResponse;

  /** VerifyEmail verifies an account email using a token. */
  @post
  @route("/verify-email")
  verifyEmail(...VerifyEmailRequest): VerifyEmailResponse;

  /** ResendVerification sends a fresh email verification message. */
  @post
  @route("/resend-verification")
  resendVerification(...ResendVerificationRequest): ResendVerificationResponse;

  /** InviteAccept lets an invited user activate their account. */
  @post
  @route("/invite/accept")
  inviteAccept(...InviteAcceptRequest): InviteAcceptResponse;

  /** CheckEmailExists probes whether an email is already registered. */
  @get
  @route("/email-exists")
  checkEmailExists(...CheckEmailExistsRequest): CheckEmailExistsResponse;

  /** GoogleAuthorize returns the Google consent URL for sign-in. */
  @get
  @route("/google/authorize")
  googleAuthorize(...GoogleAuthorizeRequest): GoogleAuthorizeResponse;
}
