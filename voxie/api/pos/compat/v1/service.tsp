import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.compat.v1;

/** Legacy POS order and payment routes, served with their existing wire format. */
@route("/pos")
interface CompatibilityService {
  @get
  @route("/orders")
  listOrders(...EmptyRequest): JsonResponse;

  @post
  @route("/orders/compat")
  createCompatOrder(...BodyRequest): JsonResponse;

  @post
  @route("/orders/{id}/pay")
  payOrder(...PathIdBodyRequest): JsonResponse;

  @post
  @route("/orders/{id}/refund/approve")
  approveOrderRefund(...PathIdBodyRequest): JsonResponse;

  @post
  @route("/orders/{id}/refund/override")
  overrideOrderRefund(...PathIdBodyRequest): JsonResponse;

  @post
  @route("/orders/{id}/void/approve")
  approveOrderVoid(...PathIdBodyRequest): JsonResponse;

  @post
  @route("/orders/{id}/void/override")
  overrideOrderVoid(...PathIdBodyRequest): JsonResponse;

  @put
  @route("/orders/{id}/resume")
  resumeOrder(...PathIdBodyRequest): JsonResponse;

  @get
  @route("/ws")
  @Onkt.http
  listWs(...EmptyRequest): EmptyResponse;

  @post
  @route("/payments/webhook")
  @Onkt.http
  createPaymentsWebhook(...EmptyRequest): EmptyResponse;

  @post
  @route("/ws-ticket")
  @Onkt.http
  createWsTicket(...EmptyRequest): EmptyResponse;
}
