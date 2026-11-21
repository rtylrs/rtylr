import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.dashboard.v1;

/** DashboardService defines the CRM dashboard endpoint. */
@route("/crm/dashboard/v1")
interface DashboardService {
  /** Returns CRM summary metrics for the active business. */
  @get
  @route("/dashboard")
  getDashboard(...GetDashboardRequest): GetDashboardResponse;
}
