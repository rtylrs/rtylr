import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.day_summary.v1;

model DaySummaryRequest {
  @Onkt.required
  @query("date")
  date: string;
}

model DaySummaryResponse {
  total_sales: float64;
  total_tax: float64;
  total_tips: float64;
  total_refunds: float64;
  total_orders: int32;
  payment_breakdown: Record<float64>;
}
