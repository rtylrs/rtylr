import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace common.oauth.v1;

/** OauthService defines OAuth authorization endpoints. */
@route("/common/oauth/v1")
interface OauthService {
  /** Issue starts the OAuth authorization redirect flow. */
  @post
  @route("/issue")
  issue(...IssueRequest): IssueResponse;

  @post
  @route("/token")
  token(...TokenRequest): TokenResponse;

  @post
  @route("/revoke")
  revoke(...RevokeRequest): RevokeResponse;

  @get
  @route("/me")
  me(...MeRequest): MeResponse;
}
