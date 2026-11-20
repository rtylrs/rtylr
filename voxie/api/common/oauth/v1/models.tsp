import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using hub.user.v1;

namespace common.oauth.v1;

/** IssueRequest creates an OAuth authorization redirect for the signed-in user. */
model IssueRequest {
  /** client_id identifies the OAuth client requesting authorization. */
  @Onkt.required
  client_id: string;

  /** redirect_uri is the callback URI that receives the authorization code. */
  @Onkt.required
  @format("uri")
  redirect_uri: string;

  /** scope contains the requested OAuth scopes. */
  @Onkt.required
  scope: string;

  /** state is echoed back to the redirect URI. */
  @Onkt.required
  state: string;

  /** code_challenge is the PKCE challenge for the authorization code. */
  @Onkt.required
  code_challenge: string;

  /** code_challenge_method is the PKCE challenge method. */
  @Onkt.required
  code_challenge_method: string;
}

/** IssueResponse returns the authorization redirect URL. */
model IssueResponse {
  /** redirect_url is the URL the browser should navigate to. */
  redirect_url: string;
}

model TokenRequest {
  @Onkt.required
  grant_type: string;

  code: string;
  code_verifier: string;
  redirect_uri: string;

  @Onkt.required
  client_id: string;

  refresh_token: string;
}

model TokenResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: int32;

  /** scope is the space-delimited set of scopes the access token carries. */
  scope: string;
}

model RevokeRequest {
  @Onkt.required
  token: string;

  client_id: string;
}

model RevokeResponse {
  ok: boolean;
}

model MeRequest {}

model MeResponse {
  @encode(string)
  user_id: int64;

  @encode(string)
  business_id: int64;

  /** Server-resolved authorization for the active business. */
  effective_access?: EffectiveAccess;
}
