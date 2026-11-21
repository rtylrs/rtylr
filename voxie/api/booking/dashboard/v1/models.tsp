import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.dashboard.v1;

model GetDashboardRequest {}

model GetDashboardResponse {
  today_count: int32;
  upcoming_count: int32;
  cancellation_rate: float64;
  revenue: float64;
}
