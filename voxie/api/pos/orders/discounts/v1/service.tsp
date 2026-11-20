import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.discounts.v1;

@route("/pos/orders/discounts/v1")
interface OrderDiscountService {
  @post
  @route("/orders/{order_id}/discount")
  discountOrder(...DiscountOrderRequest): DiscountOrderResponse;
}
