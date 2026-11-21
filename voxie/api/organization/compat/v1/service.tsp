import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace organization.compat.v1;

/** Legacy console administration routes, served with their existing wire format. */
@route("/console")
interface ConsoleService {
  @delete
  @route("/addon-grants/{id}")
  deleteAddonGrant(...PathIdRequest): JsonResponse;

  @delete
  @route("/grants/{id}")
  deleteGrant(...PathIdRequest): JsonResponse;

  @delete
  @route("/resources/{table}/rows/{id}")
  deleteResourceRow(...PathTableIdRequest): JsonResponse;

  @get
  @route("/addon-grants")
  listAddonGrants(...EmptyRequest): JsonResponse;

  @get
  @route("/addons")
  listAddons(...EmptyRequest): JsonResponse;

  @get
  @route("/ai-cache/analytics")
  getAiCacheAnalytics(...EmptyRequest): JsonResponse;

  @get
  @route("/grants")
  listGrants(...EmptyRequest): JsonResponse;

  @post
  @route("/addon-grants")
  createAddonGrant(...BodyRequest): JsonResponse;

  @post
  @route("/addon-grants/{id}/revoke")
  revokeAddonGrant(...PathIdBodyRequest): JsonResponse;

  @post
  @route("/grants")
  createGrant(...BodyRequest): JsonResponse;

  @post
  @route("/grants/{id}/revoke")
  revokeGrant(...PathIdBodyRequest): JsonResponse;

  @post
  @route("/resources/{table}/rows")
  createResourceRow(...PathTableBodyRequest): JsonResponse;

  @put
  @route("/resources/{table}/rows/{id}")
  updateResourceRow(...PathTableIdBodyRequest): JsonResponse;

  @get
  @route("/resources")
  @Onkt.http
  listResources(...EmptyRequest): EmptyResponse;

  @get
  @route("/resources/{table}")
  @Onkt.http
  getResources(...PathTableRequest): EmptyResponse;

  @get
  @route("/resources/{table}/rows")
  @Onkt.http
  getResourcesRows(...PathTableRequest): EmptyResponse;

  @get
  @route("/resources/{table}/rows/{id}")
  @Onkt.http
  getResourcesRows2(...PathTableIdRequest): EmptyResponse;
}

/** OAuth authorization redirect, served by its original handler. */
@route("/oauth")
interface OAuthService {
  @get
  @route("/authorize")
  @Onkt.http
  listAuthorize(...EmptyRequest): EmptyResponse;
}

/**
 * Billing provider webhooks, served by their original handlers (they verify
 * the raw body signature).
 */
@route("/subscriptions")
interface BillingWebhookService {
  @post
  @route("/paddle-webhook")
  @Onkt.http
  createPaddleWebhook(...EmptyRequest): EmptyResponse;
}
