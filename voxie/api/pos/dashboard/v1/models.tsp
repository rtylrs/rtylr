import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.dashboard.v1;

model DashboardRequest {}

model DashboardResponse {
  total_sales: float64;
  order_count: int32;
  avg_order: float64;
  table_turnover: float64;
}
