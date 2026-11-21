import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using pos.orders.v1;

namespace pos.orders.splits.v1;

// ── Split Order ───────────────────────────────────────────────────
model SplitOrderRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  @minItems(1)
  splits: SplitGroup[];
}

model SplitOrderResponse {
  @Onkt.encode("string")
  order_ids: int64[];

  message: string;
}
