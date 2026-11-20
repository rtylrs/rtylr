import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.config.v1;

/** ConfigService defines hub configuration endpoints. */
@route("/hub/config/v1")
interface ConfigService {
  /** Returns the active business configuration. */
  @get
  @route("/config")
  getConfig(...GetConfigRequest): GetConfigResponse;

  /** Creates the business configuration. */
  @post
  @route("/config")
  createConfig(...CreateConfigRequest): CreateConfigResponse;

  /** Updates the business configuration. */
  @put
  @route("/config")
  updateConfig(...UpdateConfigRequest): UpdateConfigResponse;

  @get
  @route("/config/custom-domain")
  getCustomDomain(...GetCustomDomainRequest): GetCustomDomainResponse;

  @post
  @route("/config/custom-domain")
  registerCustomDomain(
    ...RegisterCustomDomainRequest,
  ): RegisterCustomDomainResponse;

  @post
  @route("/config/custom-domain/verify")
  verifyCustomDomain(...VerifyCustomDomainRequest): VerifyCustomDomainResponse;

  @delete
  @route("/config/custom-domain")
  removeCustomDomain(...RemoveCustomDomainRequest): RemoveCustomDomainResponse;

  /** Queues a test email using the configured tenant sender settings. */
  @post
  @route("/config/test-email")
  sendTestEmail(...SendTestEmailRequest): SendTestEmailResponse;

  /** Returns the industry catalog grouped by category. */
  @get
  @route("/industries")
  getIndustries(...GetIndustriesRequest): GetIndustriesResponse;
}
