import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.discounts.v1;

// ── Apply Discount ────────────────────────────────────────────────
model DiscountOrderRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  @Onkt.required
  discount_type: string;

  @Onkt.required
  @minValue(0)
  discount_value?: float64;

  @Onkt.required
  reason: string;
}

model DiscountOrderResponse {
  message: string;
}
