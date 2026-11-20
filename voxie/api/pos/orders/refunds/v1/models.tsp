import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.refunds.v1;

// ── Process Refund ────────────────────────────────────────────────
model ProcessRefundRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  @Onkt.required
  @minValueExclusive(0)
  amount?: float64;

  @Onkt.required
  reason: string;
}

model ProcessRefundResponse {
  @encode(string)
  refund_id: int64;

  message: string;
}
