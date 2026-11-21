import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.dashboard.v1;

/** TopCustomer is a customer ranked by spend for dashboard summaries. */
model TopCustomer {
  @encode(string)
  id: int64;

  first_name: string;
  last_name: string;
  email: string;
  total_spent: float64;

  @encode(string)
  loyalty_points: int64;

  loyalty_tier: string;
}

model GetDashboardRequest {}

model GetDashboardResponse {
  @encode(string)
  total_customers: int64;

  @encode(string)
  new_this_month: int64;

  @encode(string)
  active_loyalty_members: int64;

  top_customers: TopCustomer[];
}
