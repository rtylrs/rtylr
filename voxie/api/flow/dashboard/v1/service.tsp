import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.dashboard.v1;

@route("/flow/dashboard/v1")
interface DashboardService {
  @get
  @route("/dashboard")
  getFlowDashboard(...GetFlowDashboardRequest): GetFlowDashboardResponse;
}
