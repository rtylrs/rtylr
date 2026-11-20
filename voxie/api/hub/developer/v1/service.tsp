import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.developer.v1;

/** DeveloperService defines hub developer API endpoints. */
@route("/hub/developer/v1")
interface DeveloperService {
  /** Returns API keys for the active business. */
  @get
  @route("/api-keys")
  listApiKeys(...ListApiKeysRequest): ListApiKeysResponse;

  /** Creates a new API key. */
  @post
  @route("/api-keys")
  createApiKey(...CreateApiKeyRequest): CreateApiKeyResponse;

  /** Revokes an existing API key. */
  @post
  @route("/api-keys/revoke")
  revokeApiKey(...RevokeApiKeyRequest): RevokeApiKeyResponse;

  /** Returns configured webhooks. */
  @get
  @route("/webhooks")
  listWebhooks(...ListWebhooksRequest): ListWebhooksResponse;

  /** Registers a new webhook endpoint. */
  @post
  @route("/webhooks")
  createWebhook(...CreateWebhookRequest): CreateWebhookResponse;

  /** Updates an existing webhook. */
  @put
  @route("/webhooks/{id}")
  updateWebhook(...UpdateWebhookRequest): UpdateWebhookResponse;

  /** Removes a webhook by ID. */
  @delete
  @route("/webhooks/{id}")
  deleteWebhook(...DeleteWebhookRequest): DeleteWebhookResponse;

  /**
   * Replaces a webhook's signing secret. The old secret stops signing
   * immediately and the new one is returned once.
   */
  @post
  @route("/webhooks/{id}/rotate-secret")
  rotateWebhookSecret(
    ...RotateWebhookSecretRequest,
  ): RotateWebhookSecretResponse;

  /**
   * Returns delivery attempts for the business's webhooks, most recent first.
   *
   * A registration whose endpoint has started failing is otherwise invisible
   * to the tenant who owns it: deliveries retry, exhaust their attempts and
   * are dead-lettered with nothing surfacing. Filter on status to find those.
   *
   * Literal path segment rather than a parameter, so it does not collide with
   * the /webhooks/{id} routes above. Adding a GET /webhooks/{id} later would
   * make this ambiguous.
   */
  @get
  @route("/webhooks/deliveries")
  listWebhookDeliveries(
    ...ListWebhookDeliveriesRequest,
  ): ListWebhookDeliveriesResponse;
}
