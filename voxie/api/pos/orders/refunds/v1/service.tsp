import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.refunds.v1;

@route("/pos/orders/refunds/v1")
interface OrderRefundService {
  @post
  @route("/orders/{order_id}/refund")
  processRefund(...ProcessRefundRequest): ProcessRefundResponse;
}
