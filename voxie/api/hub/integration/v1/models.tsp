import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.integration.v1;

/** Integration is a connected third-party integration. */
model Integration {
  @encode(string)
  id: int64;

  provider: string;

  /** The current lifecycle state. */
  status: string;

  config: bytes;
  connected_at: string;

  /** When the record was created. */
  created_at: string;
}

model ConnectIntegrationRequest {
  @Onkt.required
  provider: string;

  config: bytes;
}

model ConnectIntegrationResponse {
  @encode(string)
  id: int64;

  message: string;
}

model DisconnectIntegrationRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  id?: int64;
}

model DisconnectIntegrationResponse {
  message: string;
}

model ListIntegrationsRequest {}

model ListIntegrationsResponse {
  integrations: Integration[];
}
