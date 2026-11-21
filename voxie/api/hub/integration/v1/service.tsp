import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.integration.v1;

/** IntegrationService defines hub integration endpoints. */
@route("/hub/integration/v1")
interface IntegrationService {
  /** Returns connected integrations. */
  @get
  @route("/integrations")
  listIntegrations(...ListIntegrationsRequest): ListIntegrationsResponse;

  /** Connects a third-party integration. */
  @post
  @route("/integrations/connect")
  connectIntegration(...ConnectIntegrationRequest): ConnectIntegrationResponse;

  /** Disconnects a third-party integration. */
  @post
  @route("/integrations/disconnect")
  disconnectIntegration(
    ...DisconnectIntegrationRequest,
  ): DisconnectIntegrationResponse;
}
