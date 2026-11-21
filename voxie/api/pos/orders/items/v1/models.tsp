import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.items.v1;

// ── Update Order Item Status ──────────────────────────────────────
model UpdateOrderItemStatusBody {
  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;
}

model UpdateOrderItemStatusRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  @minValueExclusive(0)
  @encode(string)
  item_id: int64;

  /** body carries the new order item status bound from the HTTP payload. */
  @body
  body: UpdateOrderItemStatusBody;
}

model UpdateOrderItemStatusResponse {
  message: string;
}

// ── Void Order Item ───────────────────────────────────────────────
model VoidOrderItemRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  @minValueExclusive(0)
  @encode(string)
  item_id: int64;
}

model VoidOrderItemResponse {
  success: boolean;
}
