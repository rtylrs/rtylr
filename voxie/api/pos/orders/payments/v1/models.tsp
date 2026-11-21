import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.payments.v1;

// ── Process Payment ───────────────────────────────────────────────
model ProcessPaymentRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  @Onkt.required
  payment_method: string;

  @Onkt.required
  @minValueExclusive(0)
  amount_paid?: float64;
}

model ProcessPaymentResponse {
  message: string;
}
