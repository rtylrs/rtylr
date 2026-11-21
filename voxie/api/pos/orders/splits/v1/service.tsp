import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.splits.v1;

@route("/pos/orders/splits/v1")
interface OrderSplitService {
  @post
  @route("/orders/{order_id}/split")
  splitOrder(...SplitOrderRequest): SplitOrderResponse;
}
