import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace common.auth.v1;

model RegisterRequest {
  /** email is the email address used for lookup or contact. */
  @Onkt.required
  @format("email")
  email: string;

  /** password is the user-provided secret for authentication. */
  @Onkt.required
  password: string;

  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** business_name is the optional business name to create for the user. */
  business_name?: string;

  /** plan_slug optionally selects the subscription plan to assign after registration. */
  plan_slug?: string;
}

model RegisterResponse {
  /** token is the opaque token used to complete the operation. */
  token: string;

  /** message is a human-readable result message. */
  message: string;

  /** user_id identifies the user associated with this request. */
  @encode(string)
  user_id: int64;

  /** email is the email address used for lookup or contact. */
  email: string;

  /** name is the name value for this contract. */
  name: string;

  /** refresh_token is a long-lived opaque session token used to renew the access token. */
  refresh_token: string;
}

model LoginRequest {
  /** email is the email address used for lookup or contact. */
  @Onkt.required
  @format("email")
  email: string;

  /** password is the user-provided secret for authentication. */
  @Onkt.required
  password: string;
}

model LoginResponse {
  /** token is the opaque token used to complete the operation. */
  token: string;

  /** message is a human-readable result message. */
  message: string;

  /** user_id identifies the user associated with this request. */
  @encode(string)
  user_id: int64;

  /** email is the email address used for lookup or contact. */
  email: string;

  /** name is the name value for this contract. */
  name: string;

  /** is_console_admin indicates the user can access the internal console. */
  is_console_admin: boolean;

  /** refresh_token is a long-lived opaque session token used to renew the access token. */
  refresh_token: string;
}

model GoogleAuthRequest {
  /** business_name is the optional business name to create for the user. */
  business_name?: string;

  /** code is the OAuth authorization code returned by the provider callback. */
  @Onkt.required
  code: string;
}

model GoogleAuthResponse {
  /** token is the opaque token used to complete the operation. */
  token: string;

  /** is_new is the is new value for this contract. */
  is_new: boolean;

  /** refresh_token is a long-lived opaque session token used to renew the access token. */
  refresh_token: string;
}

model RefreshRequest {
  /** refresh_token is the opaque session credential issued during sign-in. */
  @Onkt.required
  refresh_token: string;
}

model RefreshResponse {
  /** token is a newly issued short-lived access token. */
  token: string;
}

model RevokeSessionRequest {
  /** refresh_token is the opaque session credential to revoke. */
  @Onkt.required
  refresh_token: string;
}

model RevokeSessionResponse {
  /** revoked indicates whether the session was revoked. */
  revoked: boolean;
}

model ResetPasswordRequest {
  /** email is the email address used for lookup or contact. */
  @Onkt.required
  @format("email")
  email: string;
}

model ResetPasswordResponse {
  /** message is a human-readable result message. */
  message: string;
}

model ResetPasswordConfirmRequest {
  /** token is the opaque reset token from the password reset link. */
  @Onkt.required
  token: string;

  /** password is the new password selected by the user. */
  @Onkt.required
  password: string;
}

model ResetPasswordConfirmResponse {
  /** message is a human-readable result message. */
  message: string;
}

model VerifyEmailRequest {
  /** token is the opaque email verification token. */
  @Onkt.required
  token: string;
}

model VerifyEmailResponse {
  /** verified indicates the email was verified successfully. */
  verified: boolean;

  /** message is a human-readable result message. */
  message: string;
}

model ResendVerificationRequest {
  /** email is the address that should receive a verification email. */
  @Onkt.required
  @format("email")
  email: string;
}

model ResendVerificationResponse {
  /** ok indicates the request was accepted. */
  ok: boolean;

  /** message is a human-readable result message. */
  message: string;
}

model InviteAcceptRequest {
  /** token is the opaque token used to complete the operation. */
  @Onkt.required
  token: string;

  /** password is the user-provided secret for authentication. */
  @Onkt.required
  password: string;
}

model InviteAcceptResponse {
  /** message is a human-readable result message. */
  message: string;
}

model CheckEmailExistsRequest {
  /** email is the address to probe during inline registration/login validation. */
  @Onkt.required
  @format("email")
  @query("email")
  email: string;
}

model CheckEmailExistsResponse {
  /** exists is true when an account with this email is already registered. */
  exists: boolean;
}

model GoogleAuthorizeRequest {}

model GoogleAuthorizeResponse {
  /** url is the Google consent URL to redirect the browser to. */
  url: string;
}
