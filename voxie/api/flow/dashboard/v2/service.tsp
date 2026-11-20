import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.dashboard.v2;

/**
 * DashboardService aggregates business-wide counts for the caller. The
 * workspace parameter is gone; there is exactly one scope, the caller's
 * business, resolved from auth.
 */
@route("/flow/dashboard/v2")
interface DashboardService {
  @get
  @route("/dashboard")
  getFlowDashboard(...GetFlowDashboardRequest): GetFlowDashboardResponse;
}
