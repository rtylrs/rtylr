import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.holds.v1;

@route("/pos/orders/holds/v1")
interface OrderHoldService {
  @post
  @route("/orders/{order_id}/hold")
  holdOrder(...HoldOrderRequest): HoldOrderResponse;

  @get
  @route("/orders/held")
  listHeldOrders(...ListHeldOrdersRequest): ListHeldOrdersResponse;
}
