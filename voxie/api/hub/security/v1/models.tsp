import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.security.v1;

model SecurityState {
  two_factor_enabled: boolean;
  mfa_required: boolean;
  mfa_enrolled: boolean;
  organization_mfa_enforced: boolean;
  can_manage_organization_mfa: boolean;
  active_sessions: ActiveSession[];
}

model ActiveSession {
  id: string;
  ip: string;
  user_agent: string;

  /** When the record was created. */
  created_at: string;

  is_current: boolean;
}

model LoginEntry {
  /** Database identifier for this login event. */
  @encode(string)
  id: int64;

  /** Email address used for lookup or contact. */
  email: string;

  method: string;
  ip: string;
  user_agent: string;
  success: boolean;

  /** When the record was created. */
  created_at: string;
}

model GetSecurityStateRequest {}

model GetSecurityStateResponse {
  state: SecurityState;
}

model LoginHistoryRequest {
  /** 1-based page number to return. */
  @query("page")
  page: int32;

  /** Maximum number of login events to return in one page. */
  @query("page_size")
  page_size: int32;
}

model LoginHistoryResponse {
  history: LoginEntry[];

  /** Total number of login events for the authenticated user. */
  total: int32;

  /** The 1-based page number returned. */
  page: int32;

  /** The number of login events requested per page after server limits. */
  page_size: int32;

  /** Whether another page of login events is available. */
  has_more: boolean;
}

model InvalidateSessionsRequest {}

model InvalidateSessionsResponse {
  message: string;
}

model Update2FARequest {
  enabled: boolean;
}

model Update2FAResponse {
  message: string;
}

model UpdateOrganizationMFAEnforcementRequest {
  enabled: boolean;
}

model UpdateOrganizationMFAEnforcementResponse {
  enforced: boolean;
}

model BeginMFAEnrollmentRequest {}

model BeginMFAEnrollmentResponse {
  secret: string;
  otpauth_uri: string;
}

model CompleteMFAEnrollmentRequest {
  @Onkt.required
  code: string;
}

model CompleteMFAEnrollmentResponse {
  verified: boolean;
  recovery_codes: string[];
}

model VerifyMFARequest {
  @Onkt.required
  code: string;
}

model VerifyMFAResponse {
  verified: boolean;
}

model UseMFARecoveryCodeRequest {
  @Onkt.required
  code: string;
}

model UseMFARecoveryCodeResponse {
  verified: boolean;
}
