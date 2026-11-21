import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace hub.developer.v1;

model ApiKey {
  @encode(string)
  id: int64;

  name: string;

  /** First 8 chars shown */
  key_prefix: string;

  scopes: string[];
  created_at: string;
  last_used_at: string;
  expires_at: string;
  is_active: boolean;
}

model Webhook {
  @encode(string)
  id: int64;

  url: string;
  events: string[];
  is_active: boolean;
  created_at: string;
  last_triggered_at: string;
}

model CreateApiKeyRequest {
  @Onkt.required
  name: string;

  @minItems(1)
  scopes: string[];
}

model CreateApiKeyResponse {
  @encode(string)
  id: int64;

  key: string;
  message: string;
}

model CreateWebhookRequest {
  @Onkt.required
  url: string;

  @minItems(1)
  events: string[];
}

model CreateWebhookResponse {
  @encode(string)
  id: int64;

  message: string;

  /**
   * secret signs every delivery (see the X-Rtylr-Signature header). It is
   * returned only here, once, and cannot be read back afterwards.
   */
  secret: string;
}

model RotateWebhookSecretRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model RotateWebhookSecretResponse {
  /** secret signs every delivery from now on. It is returned only here, once. */
  secret: string;
}

model DeleteWebhookRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteWebhookResponse {
  message: string;
}

model ListApiKeysRequest {}

model ListApiKeysResponse {
  keys: ApiKey[];
}

/** WebhookDelivery is one attempt to deliver one event to one registration. */
model WebhookDelivery {
  @encode(string)
  id: int64;

  @encode(string)
  webhook_id: int64;

  @encode(string)
  event_id: int64;

  /** event_type is the event this delivery carries, for example "order_created". */
  event_type: string;

  /**
   * status is one of pending, delivering, delivered, failed or dead. A dead
   * delivery has exhausted its attempts and will not be retried.
   */
  status: string;

  /** attempts is how many times delivery has been tried. */
  attempts: int32;

  /** max_attempts is the cap after which a failing delivery is dead-lettered. */
  max_attempts: int32;

  /**
   * response_status is the HTTP status the endpoint answered. Zero when the
   * request never got a response, such as a timeout or a refused connection.
   */
  response_status: int32;

  /**
   * last_error is why the most recent attempt failed. It includes a short
   * excerpt of the endpoint's own response body when it sent one, which is
   * usually the fastest way for the owner to see what their endpoint said.
   */
  last_error: string;

  /**
   * next_attempt_at is when a retryable delivery will be tried again. Empty
   * for one that is delivered or dead.
   */
  next_attempt_at: string;

  /** delivered_at is when the endpoint accepted the event. Empty until then. */
  delivered_at: string;

  created_at: string;
}

model ListWebhookDeliveriesRequest {
  /**
   * webhook_id narrows to one registration. Omit for every registration in
   * the business.
   */
  @query("webhook_id")
  @encode(string)
  webhook_id: int64;

  /**
   * status narrows to one status. "dead" is the one worth watching, since
   * those events will never arrive without intervention.
   */
  @query("status")
  status: string;

  @query("page")
  page: int32;

  @query("limit")
  limit: int32;
}

model ListWebhookDeliveriesResponse {
  deliveries: WebhookDelivery[];
  pagination: PaginationResponse;
}

model ListWebhooksRequest {}

model ListWebhooksResponse {
  webhooks: Webhook[];
}

model RevokeApiKeyRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model RevokeApiKeyResponse {
  message: string;
}

model UpdateWebhookRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  url: string;

  @minItems(1)
  events: string[];

  is_active: boolean;
}

model UpdateWebhookResponse {
  webhook: Webhook;
}
