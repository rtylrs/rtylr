import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.financial.compat.v1;

model CashFlowPoint {
  date: string;
  inflow_minor_units: float64;
  outflow_minor_units: float64;
}

model GetCashFlowRequest {
  @query("from")
  from_date?: string;

  @query("to")
  to_date?: string;
}

model GetCashFlowResponse {
  series: CashFlowPoint[];
}
