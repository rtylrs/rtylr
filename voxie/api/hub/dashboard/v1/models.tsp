import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.dashboard.v1;

model GetKPIsRequest {}

model GetKPIsResponse {
  revenue: float64;

  @encode(string)
  order_count: int64;

  @encode(string)
  staff_count: int64;

  @encode(string)
  customer_count: int64;

  @encode(string)
  low_stock_count: int64;

  @encode(string)
  active_apps: int64;
}
