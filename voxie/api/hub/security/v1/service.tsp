import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.security.v1;

/** SecurityService defines hub security endpoints. */
@route("/hub/security/v1")
interface SecurityService {
  /** Returns the current security settings. */
  @get
  @route("/state")
  getSecurityState(...GetSecurityStateRequest): GetSecurityStateResponse;

  /** Returns paginated login events for the authenticated user. */
  @get
  @route("/login-history")
  loginHistory(...LoginHistoryRequest): LoginHistoryResponse;

  /** Revokes all active sessions. */
  @post
  @route("/invalidate-sessions")
  invalidateSessions(...InvalidateSessionsRequest): InvalidateSessionsResponse;

  /** Toggles two-factor authentication. */
  @patch
  @route("/2fa")
  update2FA(...Update2FARequest): Update2FAResponse;

  /** Sets whether every member in the current business must enroll in MFA. */
  @patch
  @route("/organization/mfa")
  updateOrganizationMFAEnforcement(
    ...UpdateOrganizationMFAEnforcementRequest,
  ): UpdateOrganizationMFAEnforcementResponse;

  /** Starts a TOTP enrollment and returns the secret once. */
  @post
  @route("/mfa/enroll")
  beginMFAEnrollment(...BeginMFAEnrollmentRequest): BeginMFAEnrollmentResponse;

  /** Confirms a TOTP enrollment and issues one-time recovery codes. */
  @post
  @route("/mfa/enroll/confirm")
  completeMFAEnrollment(
    ...CompleteMFAEnrollmentRequest,
  ): CompleteMFAEnrollmentResponse;

  /** Verifies a TOTP code for the current authenticated session. */
  @post
  @route("/mfa/verify")
  verifyMFA(...VerifyMFARequest): VerifyMFAResponse;

  /** Consumes one recovery code for the current authenticated session. */
  @post
  @route("/mfa/recovery")
  useMFARecoveryCode(...UseMFARecoveryCodeRequest): UseMFARecoveryCodeResponse;
}
