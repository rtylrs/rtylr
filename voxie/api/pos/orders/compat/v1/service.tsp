import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.compat.v1;

/** Compatibility contract for Dash's business-wide order listing. */
@route("/pos/orders/v1")
interface CompatibilityService {
  @get
  @route("/orders")
  listOrders(...ListOrdersRequest): ListOrdersResponse;
}
