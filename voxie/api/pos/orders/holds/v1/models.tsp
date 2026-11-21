import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using pos.orders.v1;

namespace pos.orders.holds.v1;

// ── Hold Order ────────────────────────────────────────────────────
model HoldOrderRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  @Onkt.required
  reason: string;
}

model HoldOrderResponse {
  message: string;
}

// ── List Held Orders ──────────────────────────────────────────────
model ListHeldOrdersRequest {}

model ListHeldOrdersResponse {
  orders: Order[];
}
