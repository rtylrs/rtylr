import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.payments.v1;

@route("/pos/orders/payments/v1")
interface OrderPaymentService {
  @post
  @route("/orders/{order_id}/payment")
  processPayment(...ProcessPaymentRequest): ProcessPaymentResponse;
}
