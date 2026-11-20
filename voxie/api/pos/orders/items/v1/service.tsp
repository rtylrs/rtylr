import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.items.v1;

@route("/pos/orders/items/v1")
interface OrderItemService {
  @put
  @route("/orders/{order_id}/items/{item_id}/status")
  updateOrderItemStatus(
    ...UpdateOrderItemStatusRequest,
  ): UpdateOrderItemStatusResponse;
}
