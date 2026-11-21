import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.compat.v1;

model ListOrdersRequest {
  @query("limit")
  limit?: int32;

  @query("offset")
  offset?: int32;

  @query("status")
  status?: string;
}

model ListOrdersResponse {
  orders: unknown[];
}
